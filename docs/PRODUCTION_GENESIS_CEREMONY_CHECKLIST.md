# Mythchain Production Genesis Ceremony Checklist

**Status: BLOCKED for production — founder-init now creates a candidate, but final
network parameters, release pinning, and ceremony approval are still required.**

This checklist records the ceremony and the current blockers so a development
genesis or test key is not mistaken for a production network.

## Verified repository state

- The current `release/genesis.json` is chain ID `mythchain-testnet-v2`, has
  `initial_height: 1`, founder balance 1,000,000 MTC, and 20,000,000 MTC in the
  `distribution` module/Community Pool. Its SHA-256 is
  `0e9c91fbfdef9e42f27a3f64326ad5b4ef0afb45e086b4c6c945ba8a85d6e3cf`.
  (The old dev genesis `chain-rlkh6n` / `55f420c9…fb41f` is retired.)
- The genesis founder address currently recorded is
  `myth12zfy420wyx7qc2lpl2yllhuwjngdfrkdkc9tqd`. Confirm that this is the intended
  wallet before using any genesis derived from this file.
- The treasury address is the deterministic `distribution` module address
  `myth1jv65s3grqf6v6jl3dp4t6c9t9rk99cd86qepld`; its funds are governed through
  `x/gov`, not a founder-held private key.
- The existing `release-genesis/validator0` directory is a test fixture. Do not
  copy its validator or node keys to the production mini PC.
- Current public-testnet candidate created 2026-10-08: chain `mythchain-testnet-v2`,
  home `/home/annabilardec/mythchain-testnet-v2-20261008-180035`, founder address
  `myth12zfy420wyx7qc2lpl2yllhuwjngdfrkdkc9tqd`, node ID
  `39e6c1180ab191c363b085b47505754506af1f86`, external P2P
  `0.tcp.ap.ngrok.io:29687`, genesis SHA-256
  `0e9c91fbfdef9e42f27a3f64326ad5b4ef0afb45e086b4c6c945ba8a85d6e3cf`. It has
  been validated, hash-pinned, started on the mini PC, and sync-verified with a
  WSL joiner; public onboarding is still closed.

## Implemented founder workflow

- `founder-init` now creates a one-validator candidate genesis, founder wallet in a
  local encrypted file keyring, validator key, P2P node identity, node configuration,
  public genesis/checksum, and a founder manifest.
- The command requires a new output directory, a chain ID entered twice, and a P2P
  external address. It validates the fixed token allocation and preserves partial
  output on failure rather than deleting founder data.
- The generic Cosmos SDK `init` command is disabled. A restricted `genesis` command
  permits validation only; `multi-node` remains an explicit local-testnet generator.
- No founder secret/hash is embedded in the binary. Founder controls whether/where
  the command is run; publication of the approved genesis hash remains the network
  trust decision.

## Remaining production blockers

- [ ] Select and approve the production chain ID and final network parameters. The
  current `chain-rlkh6n` is the existing release genesis value, not an approved
  production ID. Confirm genesis time, initial height, validator set/gentxs,
  commission, unbonding/slashing, gas prices, and P2P/RPC policy.
- [x] Founder workflow candidate smoke test passed in disposable WSL homes: candidate
  genesis validated, treasury audit passed, startup with a hash-pinned binary
  produced blocks, and a separate `init-node`/`join` node synchronized. At height 471,
  both nodes matched block hash `8A11D6306E2BE2C51B5651387104BD1CCE0343C1F4AE56DBEEDE47E7F274154A`,
  app hash `29CDCB18843EB144B731B7B2FF4892A4D2078A0CFE2A7849EB1BE7A8AFD3BF32`, and
  validator set (one validator, voting power 800000). This was a disposable test chain,
  not production chain approval or a backup/restore rehearsal.
- [ ] Confirm the generated production founder address and encrypted keyring backup
  offline. The workflow creates a new founder wallet; it does not import an existing
  founder wallet. Do not reuse a development or repository test key.
- [ ] Generate the canonical production genesis, then update its checksum and build
  binaries with that exact genesis SHA-256 linked in. The current `init-node` binary
  is pinned to the existing release genesis and will reject a different production
  genesis until rebuilt.
- [ ] Review the founder ceremony and release authorization operational controls.
  Do not put a production mnemonic, private key, or secret in source, environment
  logs, release assets, or this checklist.

## Pre-ceremony approval

- [ ] A fresh, dedicated production mini PC is provisioned and patched; operator
  access, firewall, time synchronization, storage, and monitoring are checked.
- [ ] A dedicated, empty production node home is selected. Never reuse a dev/test
  home or copy an old validator state into the new chain.
- [ ] Founder wallet control and recovery are verified offline; the public address
  matches the approved 1,000,000 MTC allocation.
- [ ] Treasury audit passes: 20,000,000 MTC, distribution Community Pool control,
  no founder private-key control, and 21,000,000 MTC total supply.
- [ ] Validator key creation, node P2P identity creation, and encrypted offline
  backup/restore procedures are rehearsed using a disposable testnet. Production
  validator private keys never enter the source repository or release archive.
- [ ] Candidate genesis passes application genesis validation and
  `mythprotocold treasury-audit`; two independent operators compare the exact
  genesis SHA-256 and chain ID through a trusted channel.
- [ ] A fresh candidate release passes tests, static analysis, installer checks,
  and a disposable multi-node test using separate keys and a separate test chain ID.
- [ ] Founder, validator, treasury governance, bootstrap peer, RPC, fee, incident,
  and upgrade policies are approved and published in the production network manifest.

## Ceremony execution — only after all blockers and approvals are complete

1. Download the approved stable release and verify the published SHA-256 manifest.
2. Verify the packaged genesis hash, chain ID, and founder allocation against the
   independently approved manifest.
3. On the mini PC, run the tested `founder-init` command with the approved chain ID
   entered twice, a fresh output directory, and the approved public P2P address.
   Confirm that the generated founder address matches the approved allocation record.
4. Transfer only the public genesis and checksum through the approved release
   process. Rebuild and install a binary whose linked genesis checksum matches the
   candidate; retain the generated validator/keyring material only on the founder
   host and its encrypted backups.
5. Confirm fresh node and validator identities with the supported `comet show-node-id`
   and `comet show-validator` commands. Make encrypted backups and test restoration
   before starting consensus.
6. Start the founder node. The current genesis format has `initial_height: 1`, so
   expect the first committed height to be 1, not “block 0”. Confirm consecutive
   heights, validator signing, app hash, peer/RPC health, and logs.
7. Publish the chain ID, canonical genesis SHA-256, founder node ID/P2P endpoint,
   RPC endpoint, and fee policy. Do not publish private keys or validator state.
8. Only then invite users to run `init-node`, `join`, and `start` using the same
   verified production genesis and official bootstrap peers.

## Abort conditions

Stop before publication/start if any production value is unapproved, the genesis/hash
differ between operators, the founder address is not verified, any key came from a
test fixture, the release binary hash pin does not match, or a node reports a
genesis/checksum/chain-ID mismatch. Never “fix” a mismatch by editing genesis on an
already-running network; that requires a deliberate new chain and chain ID.
