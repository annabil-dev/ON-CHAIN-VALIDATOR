# Mythchain Task Lease Prototype

This prototype adds task registration, block-height-based ownership, and
judge-vote finalization to the existing Cosmos module `x/mythprotocol`. It does
not yet settle ZYRA rewards.

## State machine

```text
no lease / released / expired
            │ MsgClaimTask (first valid transaction in committed block order)
            ▼
         LEASED ── MsgReleaseTask ──> RELEASED
            │
            ├── MsgSubmitTaskResult ──> SUBMITTED
            │
            └── block height reaches expires_at_height ──> claimable by next miner

SUBMITTED ── legacy: three matching PASS votes ──> APPROVED (terminal)
SUBMITTED ── legacy: three matching FAIL votes ──> REJECTED (immediately retryable)
SUBMITTED ── weighted: 3-of-4 majority per criterion ──> canonical score/status
```

Only one lease is stored per task. At height `H`, a claim for `N` blocks expires
at `H + N`. The lease is active while `current_height < expires_at_height`; at
the expiry height a new valid claim can replace it. CometBFT's committed block
and transaction execution order determine which claim is first. Local mempool
arrival order and miner wall clocks are not authoritative.

`attempt_id` is a deterministic SHA-256 commitment to chain ID, task ID, miner
address, acceptance hash, client-supplied 32-byte nonce, and claim block height.
It prevents a result from an older lease from being attached to its replacement.
Lease length is bounded at 10,000 blocks.

Each task ID must first be registered by its client with an immutable acceptance
hash. Weighted tasks also register immutable executable `criteria_json` as part
of the task declaration. A submitted attempt accepts votes only until its lease
expiry. Cosmos ante authentication verifies each vote transaction signer; the
keeper counts at most one vote per judge address and excludes the client and
miner. Legacy votes use three matching overall verdicts. Weighted votes store each
Judge's per-criterion outcomes; the keeper waits for 3 of 4 per criterion, then
recomputes the canonical weighted score and hard-gate result. A 2-2 split on any
criterion waits for the fourth Judge or lease expiry.

## Transactions and query

Claim:

```sh
mythprotocold tx mythprotocol claim-task \
  --task-id <task-id> --acceptance-hash <acceptance-sha256> \
  --lease-blocks <lease-blocks> --nonce <64-hex-nonce> \
  --from <miner-key> --chain-id mythprotocol --yes
```

Register task first:

```sh
mythprotocold tx mythprotocol register-task \
  --task-id <task-id> --acceptance-hash <acceptance-sha256> \
  --criteria-json '<criteria-json>' \
  --from <client-key> --chain-id mythprotocol --yes
```

Omit `--criteria-json` for legacy binary-vote tasks. Weighted criteria JSON is
generated from the Client acceptance contract and contains each criterion's id,
description, weight, hard-gate flag, and executable check definition.

Query the canonical lease:

```sh
mythprotocold query mythprotocol task-lease --task-id <task-id>
```

Release an unused lease:

```sh
mythprotocold tx mythprotocol release-task \
  --task-id <task-id> --attempt-id <attempt-id> \
  --from <miner-key> --chain-id mythprotocol --yes
```

Submit the result CID and proof commitment:

```sh
mythprotocold tx mythprotocol submit-task-result \
  --task-id <task-id> --attempt-id <attempt-id> \
  --result-cid sha256:<64-hex-artifact-digest> --proof-hash <64-hex-proof-hash> \
  --from <miner-key> --chain-id mythprotocol --yes
```

The message signer must match the miner recorded in the active lease. Result
submission is accepted only before expiry, once per attempt. A `SUBMITTED` task
can be reclaimed after that lease expires; this lets stale work eventually be
replaced if judge processing does not finish.

Submit a judge vote:

```sh
mythprotocold tx mythprotocol vote-task \
  --task-id <task-id> --attempt-id <attempt-id> --verdict PASS \
  --criteria-results-json '<per-criterion-results-json>' \
  --reason "acceptance checks passed" \
  --from <judge-key> --chain-id mythprotocol --yes
```

For weighted tasks, `criteria-results-json` maps each criterion ID to a `passed`
boolean and evidence string. The keeper rejects incomplete results or an overall
verdict that does not match the Judge's score. Legacy tasks omit this flag.

## Explicit prototype limits

- The task registry, lease, and vote prototype is in the existing scaffold
  module `x/mythprotocol`; there is not yet a separate production `x/pouw` module.
- Local ZYRA CLI source submits registration, claim, result, and weighted vote
  transactions when Mythchain mode is required. P2P remains the task/artifact
  relay; the keeper's committed state is canonical. The updated CLI/daemon have
  not been released or deployed to the previously running testnet validators.
- Judge quorum is prototype-only: judge staking/eligibility,
  Sybil resistance, staking-weighted quorum, slashing, reward distribution, and
  client fee/escrow are not implemented.
- `proof_hash` and content CID are format-checked commitments. This keeper does
  not download artefacts or prove that their contents satisfy the task.
- The submitting transaction is authenticated by Cosmos account signing. Task
  registration anchors the acceptance hash, but not task prompt, payment budget,
  or proof bytes. Treat this as a state-machine prototype, not mainnet economics.

## Tests

The keeper tests cover task registration, conflicting claims, block-height
expiry/reclaim, release ownership, active-attempt result submission, stale
attempts, duplicate results, judge quorum and conflicts, query behavior, and
genesis export/import.

```sh
go test ./...
```

## Local concurrent-claim smoke test (single validator)

This exercises canonical transaction ordering with two independent miner
accounts. Use a disposable home and keyring; these commands are for a local
development chain only. They do not create real tokens or connect to a public
network.

Build and initialize:

```sh
go test -mod=readonly ./...
go build -mod=readonly -o "$HOME/.local/bin/mythprotocold" ./cmd/mythprotocold
export DAEMON="$HOME/.local/bin/mythprotocold"
export TEST_HOME="$HOME/.cache/mythchain-lease-test"
export CHAIN_ID="mythprotocol-lease-test"

"$DAEMON" init lease-test-validator --chain-id "$CHAIN_ID" \
  --default-denom stake --home "$TEST_HOME"
mkdir -p "$TEST_HOME/config/gentx"
"$DAEMON" keys add lease-validator --keyring-backend test --no-backup --home "$TEST_HOME"
"$DAEMON" keys add lease-client --keyring-backend test --no-backup --home "$TEST_HOME"
"$DAEMON" keys add lease-miner-a --keyring-backend test --no-backup --home "$TEST_HOME"
"$DAEMON" keys add lease-miner-b --keyring-backend test --no-backup --home "$TEST_HOME"
"$DAEMON" genesis add-genesis-account lease-validator 1000000000stake \
  --keyring-backend test --home "$TEST_HOME"
"$DAEMON" genesis add-genesis-account lease-client 100000000stake \
  --keyring-backend test --home "$TEST_HOME"
"$DAEMON" genesis add-genesis-account lease-miner-a 100000000stake \
  --keyring-backend test --home "$TEST_HOME"
"$DAEMON" genesis add-genesis-account lease-miner-b 100000000stake \
  --keyring-backend test --home "$TEST_HOME"
"$DAEMON" genesis gentx lease-validator 500000000stake --keyring-backend test \
  --home "$TEST_HOME" --chain-id "$CHAIN_ID" --moniker lease-test-validator
"$DAEMON" genesis collect-gentxs --home "$TEST_HOME"
"$DAEMON" genesis validate --home "$TEST_HOME"
```

Start the validator in one terminal:

```sh
"$DAEMON" start --home "$TEST_HOME" --minimum-gas-prices 0stake
```

Register one task, then submit the following two claim commands from separate
terminals at nearly the same time. Replace the key aliases/home if needed.
`--broadcast-mode sync` only acknowledges mempool admission; query committed
state and each transaction to determine the canonical winner.

```sh
"$DAEMON" tx mythprotocol register-task --task-id race-001 \
  --acceptance-hash aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa \
  --from lease-client --chain-id "$CHAIN_ID" --broadcast-mode sync --yes \
  --keyring-backend test --home "$TEST_HOME"
```

Miner A:

```sh
"$DAEMON" tx mythprotocol claim-task --task-id race-001 \
  --acceptance-hash aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa \
  --lease-blocks 100 --nonce 1111111111111111111111111111111111111111111111111111111111111111 \
  --from lease-miner-a --chain-id "$CHAIN_ID" --broadcast-mode sync --yes \
  --keyring-backend test --home "$TEST_HOME"
```

Miner B uses the same fields but `--nonce 2222...2222` and `--from lease-miner-b`.
After the block commits, query:

```sh
"$DAEMON" query mythprotocol task-lease --task-id race-001 --output json \
  --home "$TEST_HOME"
"$DAEMON" query tx <tx-hash-A> --output json --home "$TEST_HOME"
"$DAEMON" query tx <tx-hash-B> --output json --home "$TEST_HOME"
```

Expected: query returns one `miner_address` and `attempt_id`; that miner's claim
has code `0`, and the losing claim is included with `ErrTaskAlreadyLeased`.
Repeat with reversed submission order. This proves single-validator execution
ordering only; multi-validator behavior and the ZYRA adapter must still be
tested separately.
