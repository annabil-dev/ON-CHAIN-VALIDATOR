# Mythchain Founder Guide

## Genesis and release controls

The founder is responsible for preparing the canonical genesis, preserving the
network's initial chain ID, validating validator gentxs and tokenomics, and publishing
the exact genesis SHA-256, bootstrap peer IDs, RPC endpoints, and fee policy. Generate
a new chain ID for a new network; never replace genesis on a running chain.

After the production manifest is approved, create a candidate in a new directory.
The chain ID must be supplied twice, and the founder P2P endpoint must be public and
reachable:

```sh
mythprotocold founder-init \
  --chain-id "$MYTH_CHAIN_ID" \
  --confirm-chain-id "$MYTH_CHAIN_ID" \
  --output-dir "$HOME/mythchain-founder" \
  --external-address "$FOUNDER_P2P_HOST:26656"
```

The command creates a one-validator candidate, a local encrypted file keyring,
validator/P2P identities, a public genesis and SHA-256, and a manifest under the
output directory. It refuses an existing output path. Do not start this candidate
with a binary pinned to another genesis; rebuild a release with the candidate hash,
review/approve the published hash independently, then install that binary on the
founder host while preserving the generated node-home keys.

Before packaging a candidate genesis, run:

```sh
mythprotocold treasury-audit --genesis release/genesis.json
sha256sum release/genesis.json
```

The audit must show 21,000,000 MTC total, 1,000,000 MTC in the founder account, and
20,000,000 MTC in the distribution Community Pool. Keep the resulting checksum with
the release assets and announce the same hash through the trusted release channel.

## Treasury control

The current genesis treasury is the deterministic `distribution` module account. Its
20,000,000 MTC is also accounted for in `x/distribution`'s Community Pool. It has no
founder-held private key. `MsgCommunityPoolSpend` is controlled by the governance
module authority; spend proposals must be reviewed and approved under the chain's
governance parameters before execution. A future governance-controlled treasury
migration should be proposed and audited on-chain; it must not be performed by
changing genesis on a live chain.

## Build and publish

Stable releases use annotated tags `vMAJOR.MINOR.PATCH`. The release workflow runs
dependency verification, tests, static analysis, cross-platform builds, and SHA-256
manifest generation before creating a stable GitHub Release. Do not publish a tag
until the release genesis, chain configuration, treasury audit, and security review
are approved.

The supported release helper is:

```powershell
python release_mythchain.py v1.2.3
```

It requires a clean worktree, an authenticated GitHub CLI, and confirmation before
pushing the branch/tag. Review and verify every resulting asset before announcing it.
