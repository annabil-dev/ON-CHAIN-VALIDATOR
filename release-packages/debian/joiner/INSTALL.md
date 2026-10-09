# Debian Joiner Package — Installation

Quick path (recommended):

```sh
sha256sum -c mythprotocold_<version>_<arch>.deb.sha256
sudo apt install ./mythprotocold_<version>_<arch>.deb
mythprotocold init-node --home "$HOME/.mythprotocol" \
  --genesis /usr/share/mythprotocold/release/genesis.json
```

Then join with the official manifest values and start:

```sh
mythprotocold join --home "$HOME/.mythprotocol" --chain-id "$CHAIN_ID" \
  --persistent-peers "$PERSISTENT_PEERS" --seeds "$SEEDS"
mythprotocold start --home "$HOME/.mythprotocol"
```

Confirm the genesis SHA-256, chain ID, and peer IDs against the same manifest
before starting. (Debian here means the `.deb` path for Debian/Ubuntu; generic
Linux distros use the `linux/` tarball instead.)
