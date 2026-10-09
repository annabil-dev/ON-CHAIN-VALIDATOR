# ZYRA ↔ Mythchain Task Adapter — Integration Contract

Status: local ZYRA/Mythchain source supports weighted criteria registration,
canonical claim/query, result submission, and per-criterion judge votes. The Go
keeper computes canonical results and stores scores/evidence. The current working
tree replaces the Miner/Client/Judge CLI subprocess boundary with native Cosmos
direct signing through CosmPy `0.12.2`; each role uses a local Cosmos key file and
the chain's gRPC/REST endpoint. This adapter change is **not yet in the published
PyPI `zyra-network 2.1.65` package**. Native mode has unit coverage; a shared-chain
smoke is still required before publishing it.
Native reward settlement remains unimplemented.

## Responsibility boundary

- **P2P** discovers and relays tasks/results/votes for responsiveness.
- **Mythchain** is authoritative for task registration, acceptance hash, active
  attempt, result commitment, and any future settlement.
- **ZYRA role process** signs native Cosmos messages and coordinates P2P with
  chain transactions/queries; it must treat chain query results as authoritative.
- A missing/unreachable chain is an unavailable state, never permission to
  submit a chain-dependent result as finalized or award a canonical reward.

## Current chain contract

The prototype module exposes these protobuf messages and gRPC query service in
package `mythprotocol.mythprotocol.v1`:

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
   applies 3-of-4 majority per criterion, and derives canonical score/status.
   Legacy tasks without criteria retain the old three-matching-verdict behavior.

See `TASK_LEASE_PROTOCOL.md` for state-transition rules and known limitations.

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

## Native signing configuration

Native signing is the default in the current source tree. CosmPy builds standard
Cosmos `TxBody`, `AuthInfo`, `SignDoc`, and `TxRaw` structures and signs with
`SIGN_MODE_DIRECT`; custom Mythchain messages are packed in protobuf `Any` with
their canonical type URLs. No `mythprotocold` executable or WSL installation is
needed on Miner/Client/Judge hosts. This code is not yet in the published
`zyra-network==2.1.65` package.

The Cosmos gRPC service must be reachable at the configured endpoint, for example
`grpc+http://<GRPC_HOST>:9090`. Cosmos REST (`rest+http://<REST_HOST>:1317`) is also
supported when the API is enabled. `tcp://<HOST>:26657` is the CometBFT RPC, not a
gRPC endpoint and should not be put in `MYTHCHAIN_GRPC_ENDPOINT`.

Native adapter config:

```text
ZYRA_MYTHCHAIN_MODE=required
MYTHCHAIN_CHAIN_ID=myth-testnet-1
MYTHCHAIN_GRPC_ENDPOINT=grpc+http://<GRPC_HOST>:9090
MYTHCHAIN_CLIENT_ADDRESS=<myth... client address>
MYTHCHAIN_CLIENT_MNEMONIC_FILE=<protected mnemonic file>
MYTHCHAIN_MINER_ADDRESS=<myth... miner address>
MYTHCHAIN_MINER_PRIVATE_KEY_FILE=<protected raw-hex key file>
MYTHCHAIN_JUDGE_ADDRESS=<myth... judge address>
MYTHCHAIN_JUDGE_PRIVATE_KEY_FILE=<protected raw-hex key file>
MYTHCHAIN_LEASE_BLOCKS=500
MYTHCHAIN_TX_FEES=<optional fee amount such as 1000umtc>
MYTHCHAIN_GAS_LIMIT=1000000
MYTHCHAIN_MINER_GRPC_ENDPOINT=<optional miner-specific endpoint>
MYTHCHAIN_JUDGE_GRPC_ENDPOINT=<optional judge-specific endpoint>
MYTHCHAIN_SECOND_JUDGE_ADDRESS=<second judge's myth... address>
MYTHCHAIN_SECOND_JUDGE_PRIVATE_KEY_FILE=<second protected key file>
```

Set exactly one role secret file: `MYTHCHAIN_<ROLE>_MNEMONIC_FILE` or
`MYTHCHAIN_<ROLE>_PRIVATE_KEY_FILE`. Keep it outside the repository with access
restricted to that role's OS account. The adapter derives the `myth1...` address
and rejects a mismatch. Do not place mnemonic/private-key material in CLI arguments,
environment literals, P2P payloads, or logs. Weighted tasks still require independent
Judge accounts and secret files.

`/submit` registers the task before P2P broadcast when mode is `required`. A
task published in this mode carries `lease_mode=mythchain`. `/mine` will not
start that task unless the chain query confirms an active canonical lease owned
by the configured miner. If another miner owns it, this miner skips the task
and continues polling. After delivery passes, the miner commits the artifact
CID/proof hash before broadcasting its P2P trajectory. A `/judge` node submits
its Cosmos vote before forwarding the P2P vote. Chain gRPC/REST failure is
fail-closed. Use distinct client/miner/judge Cosmos keys as appropriate; the
existing ZYRA/EVM wallet is not assumed to be a Cosmos account.

The active lease check reads committed height from the Cosmos query service and
treats an expiry height at or below that height as inactive. The adapter waits for
the signed transaction result and then confirms task ownership from canonical chain
state. Transaction broadcast success alone never counts as ownership.

Before enabling this against a shared testnet, verify the chain ID, gRPC/REST endpoint,
role-specific key-derived addresses, fee policy, and query/transaction behavior. The
existing EVM/ZYRA wallet must not be assumed compatible with a Cosmos account.

## First integration milestone and tests

The opt-in mode is deliberately not the default until a chain endpoint and
role-specific native key files are configured. Without it, P2P task leases remain
advisory and may allow duplicate work. The local multi-validator test confirmed one winner
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
- gRPC/REST timeout or non-zero transaction result leaves task non-final and retryable;
- malformed or stale query response is rejected;
- no secret appears in logs or P2P payload; native transaction path starts no subprocess;
- local three-validator result/vote commit and query consistency (passed).
- native signing integration against the `myth-testnet-1` gRPC endpoint (pending).
- independent live judge processes over the local relay (not yet run).
- weighted per-criterion votes have unit/integration tests and the daemon builds,
  but a weighted task has not yet been submitted to a freshly upgraded live
  three-validator network.

Reward settlement, judge eligibility/Sybil resistance, escrow, and chain/P2P
cross-system atomicity are separate milestones and are not implied by this
adapter contract.
