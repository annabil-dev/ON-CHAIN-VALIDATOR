# macOS Founder Package — Installation

**Production status: BLOCKED until the production manifest, candidate test, release
pin, and macOS signing/notarization are approved.**

1. Download the `darwin-amd64` or `darwin-arm64` archive and verify it against the
   official `SHA256SUMS` file using `shasum -a 256`.
2. Extract and install the binary:

   ```sh
   tar -xzf mythprotocold-<version>-darwin-<arch>.tar.gz
   sudo install -m 0755 mythprotocold /usr/local/bin/mythprotocold
   mythprotocold version
   ```

After signing/notarization and all ceremony approvals, run the founder candidate
workflow with a new output path:

```sh
mythprotocold founder-init --chain-id "$MYTH_CHAIN_ID" \
  --confirm-chain-id "$MYTH_CHAIN_ID" \
  --output-dir "$HOME/mythchain-founder" \
  --external-address "$FOUNDER_P2P_HOST:26656"
```

This creates a candidate; rebuild/use the binary pinned to its exact genesis hash
before start. Do not bypass Gatekeeper or keep keys in the package.
