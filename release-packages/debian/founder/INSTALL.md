# Debian Founder Package — Installation

**Production status: BLOCKED until the production manifest, candidate test, and
release pin are approved.** `founder-init` creates a candidate but does not approve
or publish a production network.

1. Download the `amd64` or `arm64` `.deb` and its `.sha256` sidecar.
2. Verify before installing:

   ```sh
   sha256sum -c mythprotocold_<version>_<arch>.deb.sha256
   ```

3. Install and verify the binary:

   ```sh
   sudo apt install ./mythprotocold_<version>_<arch>.deb
   mythprotocold version
   ```

The package installs the daemon and release genesis under
`/usr/share/mythprotocold/release/`. After all ceremony gates are approved, create
the candidate into a fresh founder directory:

```sh
mythprotocold founder-init --chain-id "$MYTH_CHAIN_ID" \
  --confirm-chain-id "$MYTH_CHAIN_ID" \
  --output-dir "$HOME/mythchain-founder" \
  --external-address "$FOUNDER_P2P_HOST:26656"
```

This creates a candidate; rebuild/use the binary pinned to its exact genesis hash
before start. Preserve the keyring and validator keys only on the protected host and
encrypted backups.
