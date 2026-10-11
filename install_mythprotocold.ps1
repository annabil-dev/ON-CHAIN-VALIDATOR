param(
    [Parameter(Mandatory = $true, Position = 0)]
    [ValidatePattern('^(\d+\.\d+\.\d+|latest)$')]
    [string]$Version,
    [string]$InstallDir = (Join-Path $env:LOCALAPPDATA 'Programs\MythChain'),
    [string]$NodeHome = (Join-Path $HOME '.mythprotocol'),
    [string]$VerifiedArchive = '',
    [string]$VerifiedChecksums = ''
)

$ErrorActionPreference = 'Stop'
$Repository = if ($env:MYTHCHAIN_REPO) { $env:MYTHCHAIN_REPO } else { 'annabil-dev/ON-CHAIN-VALIDATOR' }
$Architecture = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq 'Arm64') { 'arm64' } else { 'amd64' }
if ($Version -eq 'latest') {
    $Latest = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repository/releases/latest"
    $Version = ($Latest.tag_name -replace '^v', '')
    if ($Version -notmatch '^\d+\.\d+\.\d+$') { throw "Could not resolve a stable release version from GitHub" }
    Write-Output "Resolved latest stable release: $Version"
}
$Archive = "mythprotocold-$Version-windows-$Architecture.zip"
$BaseUrl = "https://github.com/$Repository/releases/download/v$Version"
# Never try to overwrite an executable currently serving blocks. In particular,
# do not kill a running node or disturb its validator/node home automatically.
$CurrentBinary = Join-Path $InstallDir 'mythprotocold.exe'
if (Test-Path -LiteralPath $CurrentBinary) {
    $ActiveNodes = @(Get-Process -Name 'mythprotocold' -ErrorAction SilentlyContinue | Where-Object {
        try { $_.Path -ieq $CurrentBinary } catch { $false }
    })
    if ($ActiveNodes.Count -gt 0) {
        $Pids = ($ActiveNodes | ForEach-Object { $_.Id }) -join ', '
        throw "Mythchain node is still running (PID $Pids). Stop it normally with Ctrl+C in its node terminal, then re-run the installer. The existing node home '$NodeHome' must be preserved; do not initialize or join again."
    }
}
$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ([guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Force -Path $TempDir | Out-Null
try {
    $ArchivePath = Join-Path $TempDir $Archive
    $ChecksumsPath = Join-Path $TempDir 'SHA256SUMS'
    if ($VerifiedArchive -or $VerifiedChecksums) {
        if (-not $VerifiedArchive -or -not $VerifiedChecksums) { throw 'Both verified package paths are required' }
        if ([System.IO.Path]::GetFileName($VerifiedArchive) -ne $Archive) { throw 'Verified archive does not match requested release' }
        Copy-Item -LiteralPath $VerifiedArchive -Destination $ArchivePath
        Copy-Item -LiteralPath $VerifiedChecksums -Destination $ChecksumsPath
    } else {
        Invoke-WebRequest -Uri "$BaseUrl/$Archive" -OutFile $ArchivePath
        Invoke-WebRequest -Uri "$BaseUrl/SHA256SUMS" -OutFile $ChecksumsPath
    }
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

    # Short command for subsequent sessions: mythprotocol start / mythprotocol status.
    $Launcher = Join-Path $InstallDir 'mythprotocol.cmd'
    $LauncherText = '@echo off' + "`r`n" + '"%~dp0mythprotocold.exe" %*' + "`r`n"
    [System.IO.File]::WriteAllText($Launcher, $LauncherText, [System.Text.Encoding]::ASCII)
    $UserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    $UserPathItems = @($UserPath -split ';' | Where-Object { $_ })
    $AlreadyInPath = @($UserPathItems | Where-Object { $_.TrimEnd('\') -ieq $InstallDir.TrimEnd('\') }).Count -gt 0
    if (-not $AlreadyInPath) {
        [Environment]::SetEnvironmentVariable('Path', (($UserPathItems + $InstallDir) -join ';'), 'User')
    }
    if ($env:Path -notlike "*$InstallDir*") { $env:Path += ";$InstallDir" }
    Write-Output "CLI shortcut installed. Open a new PowerShell and run: mythprotocol start"

    $ExistingGenesis = Join-Path $NodeHome 'config\genesis.json'
    if (Test-Path -LiteralPath $ExistingGenesis) {
        Write-Output ''
        Write-Output "Existing node home found: $NodeHome"
        Write-Output 'Preserving existing chain database, configuration and node identity.'
        Write-Output 'Upgrade complete; initialization and peer setup were SKIPPED.'
        Write-Output 'To resume your node in a NEW PowerShell terminal: mythprotocol start'
        Write-Output 'To check sync: mythprotocol status --node tcp://127.0.0.1:26657'
        return
    }

    $Initialize = Read-Host "Initialize node at '$NodeHome' now? [y/N]"
    if ($Initialize -match '^(y|yes)$') {
        & $Binary init-node --home $NodeHome --genesis (Join-Path $InstallDir 'release\genesis.json')
        if ($LASTEXITCODE -ne 0) { throw 'Node initialization failed; join aborted' }
        $Join = Read-Host 'Join a network now? [y/N]'
        if ($Join -match '^(y|yes)$') {
            $ManifestChain = ''
            $ManifestPeers = ''
            try {
                $Manifest = Invoke-RestMethod -Uri 'https://mythchain.pages.dev/network.json' -TimeoutSec 15
                $ManifestChain = $Manifest.chain_id
                $ManifestPeers = ($Manifest.persistent_peers -join ',')
                if ($ManifestChain) { Write-Output "Official testnet: $ManifestChain (press Enter to accept)" }
            } catch { Write-Output 'Could not fetch the network manifest; enter values manually.' }
            $ChainId = Read-Host "Chain ID [$ManifestChain]"
            if (-not $ChainId) { $ChainId = $ManifestChain }
            $Peers = Read-Host "Persistent peer(s) [$ManifestPeers]"
            if (-not $Peers) { $Peers = $ManifestPeers }
            $Seeds = Read-Host 'Seed(s), optional (ID@host:26656)'
            $JoinArgs = @('join', '--home', $NodeHome, '--chain-id', $ChainId)
            if ($Peers) { $JoinArgs += @('--persistent-peers', $Peers) }
            if ($Seeds) { $JoinArgs += @('--seeds', $Seeds) }
            & $Binary @JoinArgs
            if ($LASTEXITCODE -ne 0) { throw 'Node join failed; start aborted' }
            $Start = Read-Host 'Start the node in this terminal now? [y/N]'
            if ($Start -match '^(y|yes)$') { & $Binary start --home $NodeHome; if ($LASTEXITCODE -ne 0) { throw 'Node exited with an error' } }
        }
    }
    $GenesisPath = Join-Path $InstallDir 'release\genesis.json'
    Write-Output ''
    Write-Output 'Next steps (replace CHAIN_ID and PEERS with the official network manifest values):'
    Write-Output "  & '$Binary' init-node --home '$NodeHome' --genesis '$GenesisPath'"
    Write-Output "  & '$Binary' join --home '$NodeHome' --chain-id CHAIN_ID --persistent-peers PEERS"
    Write-Output "  mythprotocol start --home '$NodeHome'"
    Write-Output "  mythprotocol status --node tcp://127.0.0.1:26657"
} finally {
    Remove-Item -LiteralPath $TempDir -Recurse -Force -ErrorAction SilentlyContinue
}
