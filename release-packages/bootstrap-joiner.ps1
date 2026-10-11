# One-command Mythchain joiner bootstrap (Windows).
# Usage: irm https://mythchain.pages.dev/install.ps1 | iex
# This bootstrap checks the downloaded archive SHA-256 before executing its installer.
param([string]$Version = 'latest')

$ErrorActionPreference = 'Stop'
$Repo = if ($env:MYTHCHAIN_REPO) { $env:MYTHCHAIN_REPO } else { 'annabil-dev/ON-CHAIN-VALIDATOR' }
if ($Version -eq 'latest') {
    $Version = ((Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest").tag_name -replace '^v', '')
}
if ($Version -notmatch '^\d+\.\d+\.\d+$') { throw 'Could not resolve the latest stable release' }
Write-Output "Installing Mythchain release $Version"

$Arch = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq 'Arm64') { 'arm64' } else { 'amd64' }
$ZipName = "mythprotocold-$Version-windows-$Arch.zip"
$BaseUrl = "https://github.com/$Repo/releases/download/v$Version"
$WorkingDir = Join-Path ([System.IO.Path]::GetTempPath()) ("mythchain-bootstrap-" + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $WorkingDir | Out-Null

try {
    $ZipPath = Join-Path $WorkingDir $ZipName
    $ManifestPath = Join-Path $WorkingDir 'SHA256SUMS'
    Invoke-WebRequest -Uri "$BaseUrl/$ZipName" -OutFile $ZipPath
    Invoke-WebRequest -Uri "$BaseUrl/SHA256SUMS" -OutFile $ManifestPath

    $Pattern = '^([0-9a-fA-F]{64})\s+\*?' + [regex]::Escape($ZipName) + '$'
    $Line = @(Get-Content -LiteralPath $ManifestPath | Where-Object { $_ -match $Pattern })
    if ($Line.Count -ne 1) { throw "Expected exactly one checksum for $ZipName" }
    $Expected = [regex]::Match($Line[0], '^([0-9a-fA-F]{64})').Groups[1].Value.ToLowerInvariant()
    $Actual = (Get-FileHash -LiteralPath $ZipPath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($Actual -ne $Expected) { throw "SHA-256 mismatch for $ZipName; refusing to execute installer." }
    Write-Output "Verified SHA-256: $ZipName"

    $ExtractDir = Join-Path $WorkingDir 'package'
    Expand-Archive -LiteralPath $ZipPath -DestinationPath $ExtractDir
    $Installer = Join-Path $ExtractDir 'install_mythprotocold.ps1'
    if (-not (Test-Path -LiteralPath $Installer -PathType Leaf)) { throw 'Verified package is missing the Windows installer' }

    # A separate Windows PowerShell process permits the verified unsigned script.
    # This does not change CurrentUser/LocalMachine policies. Organization policies
    # (MachinePolicy / UserPolicy) still take precedence and cannot be overridden.
    Write-Output 'Running verified installer with process-only execution policy; no persistent policy changes.'
    & powershell.exe -NoProfile -ExecutionPolicy Bypass -File $Installer -Version $Version -VerifiedArchive $ZipPath -VerifiedChecksums $ManifestPath
    if ($LASTEXITCODE -ne 0) {
        throw "Installer failed (exit $LASTEXITCODE). The node data directory was not removed; see the error above."
    }
} finally {
    Remove-Item -LiteralPath $WorkingDir -Recurse -Force -ErrorAction SilentlyContinue
}
