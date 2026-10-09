# Windows Joiner Package — Installation

Quick path (recommended):

```powershell
irm https://mythchain.pages.dev/install.ps1 | iex
```

(Linux/macOS: `curl -fsSL https://mythchain.pages.dev/install.sh | bash`.)

The wizard verifies the checksum, installs the binary and genesis, then walks
through `init-node` → `join` → `start`. It prefills the official values from the
live manifest — just press Enter to accept:

- Manifest: https://mythchain.pages.dev/network.json
- Current testnet: `mythchain-testnet-v2`
- Current peer: `39e6c1180ab191c363b085b47505754506af1f86@0.tcp.ap.ngrok.io:24981`

Manual fallback (only if the wizard cannot run):

```powershell
Expand-Archive .\mythprotocold-<version>-windows-<arch>.zip -DestinationPath .\mythchain-release
.\mythchain-release\mythprotocold.exe init-node --home "$HOME\.mythprotocol" --genesis .\genesis.json
.\mythchain-release\mythprotocold.exe join --home "$HOME\.mythprotocol" --chain-id $ChainId --persistent-peers $Peers --seeds $Seeds
.\mythchain-release\mythprotocold.exe start --home "$HOME\.mythprotocol"
```

Never substitute the development chain ID or a test peer for the release manifest.

## Uninstall (Windows)

Stop the node first (`Ctrl+C` in its terminal), then delete the install folder,
node home, and download leftovers. **This deletes keys and chain data** — back up
anything valuable first (testnet throwaway homes are safe to delete):

```powershell
Remove-Item "$env:LOCALAPPDATA\Programs\MythChain" -Recurse -Force
Remove-Item "$HOME\.mythprotocol" -Recurse -Force
Remove-Item .\mythchain-release, .\mythchain-*.zip, .\SHA256SUMS -Recurse -Force -ErrorAction SilentlyContinue
```

If you added the install folder to `PATH` manually, remove that entry too.
