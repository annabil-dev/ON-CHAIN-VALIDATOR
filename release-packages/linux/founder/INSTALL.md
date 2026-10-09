# Linux Founder Package — Installation

**Production status: BLOCKED until the production manifest, candidate test, and
release pin are approved.** `founder-init` creates a candidate; it does not publish or
approve a production chain.

1. Select the release `linux` archive matching the host (`amd64` or `arm64`) and
   verify it using the release `SHA256SUMS` file:

   ```sh
   sha256sum -c SHA256SUMS
   ```

2. Extract and install the binary:

   ```sh
   tar -xzf mythprotocold-<version>-linux-<arch>.tar.gz
   sudo install -m 0755 mythprotocold /usr/local/bin/mythprotocold
   mythprotocold version
   ```

3. Provision a dedicated, empty output path. After all ceremony gates are approved,
   run the founder candidate command:

   ```sh
   mythprotocold founder-init --chain-id "$MYTH_CHAIN_ID" \
     --confirm-chain-id "$MYTH_CHAIN_ID" \
     --output-dir "$HOME/mythchain-founder" \
     --external-address "$FOUNDER_P2P_HOST:26656"
   ```

   Keep validator keys out of the package and repository; generate them on the
   founder host and back them up encrypted.

Do not start the candidate until its public genesis has been approved and a release
binary has been rebuilt with the candidate genesis SHA-256 linked in.
