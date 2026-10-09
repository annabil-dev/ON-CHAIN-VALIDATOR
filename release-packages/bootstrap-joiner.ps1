# One-command MythChain joiner bootstrap (Windows).
# Usage: irm <release-url>/bootstrap-joiner.ps1 | iex
param([string]$Version = 'latest')
$ErrorActionPreference = 'Stop'
$Repo = if ($env:MYTHCHAIN_REPO) { $env:MYTHCHAIN_REPO } else { 'annabil-dev/ON-CHAIN-VALIDATOR' }
if ($Version -eq 'latest') {
    $Version = ((Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest").tag_name -replace '^v', '')
    if ($Version -notmatch '^\d+\.\d+\.\d+$') { throw 'Could not resolve the latest stable release' }
    Write-Output "Joining release $Version"
}
$Arch = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq 'Arm64') { 'arm64' } else { 'amd64' }
$Zip = "mythprotocold-$Version-windows-$Arch.zip"
Invoke-WebRequest -Uri "https://github.com/$Repo/releases/download/v$Version/$Zip" -OutFile $Zip
Expand-Archive -Path $Zip -DestinationPath .\mythchain-release -Force
.\mythchain-release\install_mythprotocold.ps1 -Version $Version
