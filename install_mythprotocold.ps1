param(
    [Parameter(Mandatory = $true, Position = 0)]
    [ValidatePattern('^(\d+\.\d+\.\d+|latest)$')]
    [string]$Version,
    [string]$InstallDir = (Join-Path $env:LOCALAPPDATA 'Programs\MythChain'),
    [string]$NodeHome = (Join-Path $HOME '.mythprotocol')
)

$ErrorActionPreference = 'Stop'
$Repository = if ($env:MYTHCHAIN_REPO) { $env:MYTHCHAIN_REPO } else { 'annabil-dev/ON-CHAIN-VALIDATOR' }
$Architecture = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq 'Arm64') { 'arm64' } else { 'amd64' }
$Archive = "mythprotocold-$Version-windows-$Architecture.zip"
$BaseUrl = "https://github.com/$Repository/releases/download/v$Version"
if ($Version -eq 'latest') {
    $Latest = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repository/releases/latest"
    $Version = ($Latest.tag_name -replace '^v', '')
    if ($Version -notmatch '^\d+\.\d+\.\d+$') { throw "Could not resolve a stable release version from GitHub" }
    Write-Output "Resolved latest stable release: $Version"
}
$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ([guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Force -Path $TempDir | Out-Null
try {
    $ArchivePath = Join-Path $TempDir $Archive
    $ChecksumsPath = Join-Path $TempDir 'SHA256SUMS'
    Invoke-WebRequest -Uri "$BaseUrl/$Archive" -OutFile $ArchivePath
    Invoke-WebRequest -Uri "$BaseUrl/SHA256SUMS" -OutFile $ChecksumsPath
    $Line = Get-Content $ChecksumsPath | Where-Object { $_ -match "^([0-9a-fA-F]{64})\s+\*?$([regex]::Escape($Archive))$" } | Select-Object -First 1
    if (-not $Line) { throw "Release checksum missing for $Archive" }
    $Expected = [regex]::Match($Line, '^([0-9a-fA-F]{64})').Groups[1].Value.ToLowerInvariant()
    $Actual = (Get-FileHash -Path $ArchivePath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($Expected -ne $Actual) { throw "Checksum mismatch for $Archive" }

    $ExtractDir = Join-Path $TempDir 'extract'
    Expand-Archive -Path $ArchivePath -DestinationPath $ExtractDir
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    New-Item -ItemType Directory -Force -Path (Join-Path $InstallDir 'release') | Out-Null
    New-Item -ItemType Directory -Force -Path (Join-Path $InstallDir 'docs') | Out-Null
    Copy-Item (Join-Path $ExtractDir 'mythprotocold.exe') (Join-Path $InstallDir 'mythprotocold.exe') -Force
    Copy-Item (Join-Path $ExtractDir 'release\genesis.json') (Join-Path $InstallDir 'release\genesis.json') -Force
    Copy-Item (Join-Path $ExtractDir 'release\genesis.sha256') (Join-Path $InstallDir 'release\genesis.sha256') -Force
    Copy-Item (Join-Path $ExtractDir 'PRODUCTION_USER_GUIDE.md') (Join-Path $InstallDir 'docs\') -Force
    Copy-Item (Join-Path $ExtractDir 'PRODUCTION_VALIDATOR_GUIDE.md') (Join-Path $InstallDir 'docs\') -Force
    Copy-Item (Join-Path $ExtractDir 'PRODUCTION_FOUNDER_GUIDE.md') (Join-Path $InstallDir 'docs\') -Force
    Copy-Item (Join-Path $ExtractDir 'SECURITY_POLICY.md') (Join-Path $InstallDir 'docs\') -Force
    $Binary = Join-Path $InstallDir 'mythprotocold.exe'
    & $Binary version

    $Initialize = Read-Host "Initialize node at '$NodeHome' now? [y/N]"
    if ($Initialize -match '^(y|yes)$') {
        & $Binary init-node --home $NodeHome --genesis (Join-Path $InstallDir 'release\genesis.json')
        $Join = Read-Host 'Join a network now? [y/N]'
        if ($Join -match '^(y|yes)$') {
            $ChainId = Read-Host 'Chain ID'
            $Peers = Read-Host 'Persistent peer(s), comma separated (ID@host:26656)'
            $Seeds = Read-Host 'Seed(s), optional (ID@host:26656)'
            $JoinArgs = @('join', '--home', $NodeHome, '--chain-id', $ChainId)
            if ($Peers) { $JoinArgs += @('--persistent-peers', $Peers) }
            if ($Seeds) { $JoinArgs += @('--seeds', $Seeds) }
            & $Binary @JoinArgs
            $Start = Read-Host 'Start the node in this terminal now? [y/N]'
            if ($Start -match '^(y|yes)$') { & $Binary start --home $NodeHome }
        }
    }
    $GenesisPath = Join-Path $InstallDir 'release\genesis.json'
    Write-Output ''
    Write-Output 'Next steps (replace CHAIN_ID and PEERS with the official network manifest values):'
    Write-Output "  & '$Binary' init-node --home '$NodeHome' --genesis '$GenesisPath'"
    Write-Output "  & '$Binary' join --home '$NodeHome' --chain-id CHAIN_ID --persistent-peers PEERS"
    Write-Output "  & '$Binary' start --home '$NodeHome'"
} finally {
    Remove-Item -LiteralPath $TempDir -Recurse -Force -ErrorAction SilentlyContinue
}
