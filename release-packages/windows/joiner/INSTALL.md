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
- Current peer: get the live `persistent_peers` from the manifest above; tunnel ports may change.

Manual fallback (only if the wizard cannot run):

```powershell
Expand-Archive .\mythprotocold-<version>-windows-<arch>.zip -DestinationPath .\mythchain-release
.\mythchain-release\mythprotocold.exe init-node --home "$HOME\.mythprotocol" --genesis .\mythchain-release\release\genesis.json
.\mythchain-release\mythprotocold.exe join --home "$HOME\.mythprotocol" --chain-id $ChainId --persistent-peers $Peers --seeds $Seeds
.\mythchain-release\mythprotocold.exe start --home "$HOME\.mythprotocol"
```

Never substitute the development chain ID or a test peer for the release manifest.

## Restart and safe upgrade

When the node is already installed, start it in a NEW PowerShell:

```powershell
mythprotocol start
mythprotocol status --node tcp://127.0.0.1:26657
```

Before upgrading, stop any currently running instance with `Ctrl+C` in its terminal.
Windows cannot replace `mythprotocold.exe` while it is running. Re-run the one-command
installer after the process exits. The installer preserves an existing node home
and skips `init-node` / `join` automatically; do **not** delete `~\.mythprotocol`.

The bootstrap verifies the ZIP SHA-256 and runs the contained installer in a
separate Windows PowerShell process with process-scoped execution policy. It
**does not** persistently change execution policies. Organization-managed
MachinePolicy/UserPolicy restrictions may still prevent execution.

## Uninstall (Windows)

Stop the node first (`Ctrl+C` in its terminal), then delete the install folder,
node home, and download leftovers. **This deletes keys and chain data** — back up
anything valuable first. Do not delete a node home as part of an upgrade:

```powershell
Remove-Item "$env:LOCALAPPDATA\Programs\MythChain" -Recurse -Force
Remove-Item "$HOME\.mythprotocol" -Recurse -Force
Remove-Item .\mythchain-release, .\mythchain-*.zip, .\SHA256SUMS -Recurse -Force -ErrorAction SilentlyContinue
```

If you added the install folder to `PATH` manually, remove that entry too.
