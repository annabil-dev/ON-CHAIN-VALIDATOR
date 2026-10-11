# Linux Joiner Package — Installation

Quick path (recommended):

```sh
curl -fsSL https://mythchain.pages.dev/install.sh | bash
```

The bootstrap verifies SHA-256, installs the binary and genesis, then automatically
runs `init-node` → `join` → `start` for a fresh node. On upgrades it preserves
the existing node home and skips re-initialization. Peer data comes from:

- Manifest: https://mythchain.pages.dev/network.json
- Current testnet: `mythchain-testnet-v2`
- Current peer: use `persistent_peers` from the live manifest; ngrok port numbers can change.

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
