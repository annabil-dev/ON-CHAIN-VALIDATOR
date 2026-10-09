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

Confirm the genesis SHA-256, chain ID, and peer IDs against the live manifest
(https://mythchain.pages.dev/network.json) before starting. Current testnet:
`mythchain-testnet-v2`, peer
`39e6c1180ab191c363b085b47505754506af1f86@0.tcp.ap.ngrok.io:24981`.
(Debian here means the `.deb` path for Debian/Ubuntu; generic Linux distros use
the `linux/` tarball instead.)
