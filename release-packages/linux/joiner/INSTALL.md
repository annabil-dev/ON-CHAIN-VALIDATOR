# Linux Joiner Package — Installation

Quick path (recommended):

```sh
curl -fsSL https://mythchain.pages.dev/install.sh | bash
```

The wizard verifies `SHA256SUMS`, installs the binary and genesis, then walks
through `init-node` → `join` → `start`. It prefills the official values from the
live manifest — just press Enter to accept:

- Manifest: https://mythchain.pages.dev/network.json
- Current testnet: `mythchain-testnet-v2`
- Current peer: `39e6c1180ab191c363b085b47505754506af1f86@0.tcp.ap.ngrok.io:24981`

Manual fallback (only if the wizard cannot run):

```sh
tar -xzf mythprotocold-<version>-linux-<arch>.tar.gz
sudo install -m 0755 mythprotocold /usr/local/bin/mythprotocold
mythprotocold init-node --home "$HOME/.mythprotocol" --genesis /path/to/genesis.json
mythprotocold join --home "$HOME/.mythprotocol" --chain-id "$CHAIN_ID" \
  --persistent-peers "$PERSISTENT_PEERS" --seeds "$SEEDS"
mythprotocold start --home "$HOME/.mythprotocol"
```

Use only the chain ID and peer IDs from the matching official manifest.

## Uninstall (Linux)

```sh
rm -rf ~/.local/bin/mythprotocold ~/.mythprotocol
# or, for an isolated folder install:
rm -rf ~/myth-test
```
