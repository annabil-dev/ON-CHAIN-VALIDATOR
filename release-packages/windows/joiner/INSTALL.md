# Windows Joiner Package — Installation

Quick path (recommended):

```powershell
$V='0.2.0'; $A="mythprotocold-$V-windows-amd64.zip"; Invoke-WebRequest "https://github.com/annabil-dev/ON-CHAIN-VALIDATOR/releases/download/v$V/$A" -OutFile $A; Expand-Archive .\$A -DestinationPath .\mythchain-release; .\mythchain-release\install_mythprotocold.ps1 -Version $V
```

The wizard verifies the checksum, installs the binary and genesis, then walks
through `init-node` → `join` → `start`. Use the chain ID and peer addresses from
the official network manifest for this release.

Manual fallback (only if the wizard cannot run):

```powershell
Expand-Archive .\mythprotocold-<version>-windows-<arch>.zip -DestinationPath .\mythchain-release
.\mythchain-release\mythprotocold.exe init-node --home "$HOME\.mythprotocol" --genesis .\genesis.json
.\mythchain-release\mythprotocold.exe join --home "$HOME\.mythprotocol" --chain-id $ChainId --persistent-peers $Peers --seeds $Seeds
.\mythchain-release\mythprotocold.exe start --home "$HOME\.mythprotocol"
```

Never substitute the development chain ID or a test peer for the release manifest.
