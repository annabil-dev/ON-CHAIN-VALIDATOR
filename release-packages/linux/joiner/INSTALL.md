# Linux Joiner Package — Installation

Quick path (recommended):

```sh
bash install_mythprotocold.sh latest
```

Ganti `latest` dengan nomor versi (misal `1.2.3`) kalau mau pin ke rilis tertentu.

The wizard verifies `SHA256SUMS`, installs the binary and genesis, then walks
through `init-node` → `join` → `start`. Use the chain ID and peer addresses from
the official network manifest for this release.

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
