# ZYRA ↔ Mythchain Task Adapter — Integration Contract

Status: local ZYRA/Mythchain source supports weighted criteria registration,
canonical claim/query, result submission, and per-criterion judge votes. The Go
keeper computes canonical 2-of-3 results per criterion and stores score weights
and evidence. Python/Go unit tests and the updated daemon build pass. The existing
three-validator WSL testnet has been upgraded locally; an adapter smoke committed a
weighted task and three Cosmos votes (one FAIL, two PASS), and all RPCs returned
the same `APPROVED` 80/100 score. The ZYRA adapter/runtime and Miner polling
hotfix are included in PyPI `zyra-network 2.1.62`; the weighted Mythchain daemon
source/binary remains a separate upgrade and has not been deployed to a public
network.
Native reward settlement remains unimplemented.

## Responsibility boundary

- **P2P** discovers and relays tasks/results/votes for responsiveness.
- **Mythchain** is authoritative for task registration, acceptance hash, active
  attempt, result commitment, and any future settlement.
- **ZYRA CLI** coordinates P2P with chain transactions/queries and must treat
  chain query results as authoritative before acting on a task.
- A missing/unreachable chain is an unavailable state, never permission to
  submit a chain-dependent result as finalized or award a canonical reward.

## Current chain contract

The prototype module exposes the following logical operations through
`mythprotocold` AutoCLI (exact flags/output must be verified against the target
binary and chain configuration):

1. Register `(task_id, acceptance_sha256)` and optional immutable
   `criteria_json` from the Client account. Weighted criteria include executable
   check definitions.
2. Claim a registered task from a miner account for a bounded number of blocks
   with a client-generated nonce.
3. Query canonical lease/task state and obtain `attempt_id`/expiry height.
4. Submit the active attempt's artifact CID and proof hash. The adapter waits for
   committed state showing `SUBMITTED` before it reports success.
5. Vote against the submitted attempt from Judge accounts. Weighted votes include
   per-criterion PASS/FAIL evidence; the keeper stores each authenticated result,
   applies 2-of-3 majority per criterion, and derives canonical score/status.
   Legacy tasks without criteria retain the old two-matching-verdict behavior.

See `TASK_LEASE_PROTOCOL.md` for current command examples and known limitations.

## Required adapter behavior

### Client submission

- Validate and canonicalize the acceptance contract locally.
- If the Client has not supplied criteria in `--spec`, ZYRA drafts executable
  criteria from the prompt using its configured local model, with a conservative
  runtime-check fallback. Only supported check types are accepted; the rubric is
  shown and hash-locked before registration/claim.
- Compute SHA-256 over the existing canonical contract serialization used by
  ZYRA (`contract_hash`); do not recompute using a different JSON encoding.
- Register task ID + hash on chain before advertising a chain-backed task.
- Persist the chain ID, task ID, acceptance hash, and transaction hash in local
  task state. Retries must query before attempting duplicate registration.
- Never include API secrets, local absolute paths, or unsigned fee assumptions
  in public P2P payloads.

### Miner claim and result

- Query task state and verify the acceptance hash before claiming.
- Submit claim using the configured Cosmos key/account; do not treat P2P lease
  arbitration as a chain claim.
- Wait for chain inclusion/commit, then query state and verify this miner and
  attempt ID own the active lease before starting or resuming work.
- Before result submission, query again and confirm the same attempt is active
  and unexpired. Submit CID/proof hash once; retries query state first.
- The `/mine` path commits the ZIP's `sha256:` CID and proof hash before
  broadcasting its trajectory. It does not broadcast a chain-required result if
  the result transaction cannot be confirmed.
- P2P may carry the result for judges, but must include chain ID, task ID,
  attempt ID, acceptance hash, CID, and transaction reference consistently.

### Judge and finality

- A P2P judge vote is advisory until an authenticated chain vote transaction is
  committed and appears in the canonical task query.
- For `lease_mode=mythchain`, `/judge` submits its Cosmos vote before forwarding
  the P2P vote. Set `MYTHCHAIN_JUDGE_KEY` and `MYTHCHAIN_JUDGE_ADDRESS`; optional
  `MYTHCHAIN_JUDGE_NODE` and `MYTHCHAIN_JUDGE_HOME` select the judge RPC/keyring.
- Verify judge account, attempt ID, CID, acceptance hash, weighted rubric, result
  JSON, and individual verdict match the chain record. A stale attempt must not count.
- Until chain settlement exists, do not present local SQLite credit as native
  Mythchain reward. Preserve distinct states such as `pending`, `approved`,
  `rejected`, and `settled`.

## Adapter transport decision for first increment

Use the installed/configured `mythprotocold` CLI as a subprocess boundary rather
than embedding Cosmos signing logic in ZYRA. The initial adapter reads:

```text
ZYRA_MYTHCHAIN_MODE=required
MYTHCHAIN_BINARY=mythprotocold
MYTHCHAIN_NODE=tcp://127.0.0.1:26657
MYTHCHAIN_CHAIN_ID=mythprotocol
MYTHCHAIN_CLIENT_KEY=<local keyring key name>
MYTHCHAIN_CLIENT_ADDRESS=<myth... client address>
MYTHCHAIN_MINER_KEY=<local keyring key name>
MYTHCHAIN_MINER_ADDRESS=<myth... miner address>
MYTHCHAIN_JUDGE_KEY=<local Cosmos judge key name>
MYTHCHAIN_JUDGE_ADDRESS=<myth... judge address>
MYTHCHAIN_LEASE_BLOCKS=500
MYTHCHAIN_TX_FEES=<optional chain fee string>
MYTHCHAIN_KEYRING_BACKEND=os
MYTHCHAIN_HOME=<optional chain CLI home directory>
MYTHCHAIN_WSL_DISTRO=<optional; e.g. Ubuntu when ZYRA runs on Windows and daemon is in WSL>
MYTHCHAIN_MINER_NODE=<optional miner-specific RPC; overrides MYTHCHAIN_NODE>
MYTHCHAIN_MINER_HOME=<optional miner-specific keyring home; overrides MYTHCHAIN_HOME>
MYTHCHAIN_JUDGE_NODE=<optional judge-specific RPC>
MYTHCHAIN_JUDGE_HOME=<optional judge-specific keyring home>
MYTHCHAIN_SECOND_JUDGE_KEY=<second independent Cosmos judge key>
MYTHCHAIN_SECOND_JUDGE_ADDRESS=<second judge's myth... address>
```

Weighted-task deployment also requires a third Judge process with its own
`MYTHCHAIN_JUDGE_KEY` and `MYTHCHAIN_JUDGE_ADDRESS`; configure each process with
its own environment/keyring rather than reusing one account.

`/submit` registers the task before P2P broadcast when mode is `required`. A
task published in this mode carries `lease_mode=mythchain`. `/mine` will not
start that task unless the chain query confirms an active canonical lease owned
by the configured miner. If another miner owns it, this miner skips the task
and continues polling. After delivery passes, the miner commits the artifact
CID/proof hash before broadcasting its P2P trajectory. A `/judge` node submits
its Cosmos vote before forwarding the P2P vote. Chain CLI/RPC failure is
fail-closed. Use distinct client/miner/judge Cosmos keys as appropriate; the
existing ZYRA/EVM wallet is not assumed to be a Cosmos account.

The active lease check reads committed block height using the CLI `status`
command and treats an expiry height at or below that height as inactive. The
current Mythchain CLI supports `--broadcast-mode sync|async`, so transactions
are broadcast with `sync` and the adapter waits until a subsequent committed
state query confirms the owner/attempt. Local command success alone never
counts as ownership. No mnemonic/private key is put into arguments by the
adapter; keyring setup/unlock remains the operator's responsibility.

Before enabling this against a shared testnet, repeat the smoke test against
that network's exact binary/version and verify its CLI output. The current Go
bootstrap and daemon are isolated under the WSL user's `.cache`; they do not
replace a system-wide Go install. The existing EVM/ZYRA wallet must not be
assumed compatible with a Cosmos account.

## First integration milestone and tests

The opt-in mode is deliberately not the default until a chain endpoint and
Cosmos keyring are configured. Without it, P2P task leases remain advisory and
may allow duplicate work. The local multi-validator test confirmed one winner
for two concurrent claims through the adapter. `scripts/local_mythchain_miner_smoke.py`
then exercised the winning worker through scripted Planner/Coder responses,
the real Docker unit/startup checks, and two independent Cosmos judge votes; it
did not contact the public P2P relay.
Two real interactive ZYRA `/mine` processes were run against a private loopback
WebSocket relay and the same single-validator RPC. The winner logged canonical
lease confirmation and entered the Swarm; the other logged that another miner
owned the task and kept polling. With the validator-feedback-to-Coder change,
the actual Qwen 7B Planner/Coder run passed local delivery checks in 11 cycles.
The real-model process was stopped while waiting for judges. Result submission
and two judge votes were then exercised through the adapter on the
three-validator testnet; all three RPCs returned the same `APPROVED` lease and
vote records. A separate two-FAIL test returned `REJECTED`, then allowed a new
canonical attempt to be claimed. No reward transfer occurred, and there was no
live end-to-end P2P judge process in the local smoke harness.

Required tests before enabling claim/result/vote:

- acceptance hash serialization parity with ZYRA;
- duplicate registration retry discovers existing matching task;
- mismatching task hash is a hard failure;
- command timeout/nonzero exit leaves task non-final and retryable;
- malformed or stale query response is rejected;
- no secret appears in subprocess arguments, logs, or P2P payload;
- local three-validator result/vote commit and query consistency (passed).
- independent live judge processes over the local relay (not yet run).
- weighted per-criterion votes have unit/integration tests and the daemon builds,
  but a weighted task has not yet been submitted to a freshly upgraded live
  three-validator network.

Reward settlement, judge eligibility/Sybil resistance, escrow, and chain/P2P
cross-system atomicity are separate milestones and are not implied by this
adapter contract.
