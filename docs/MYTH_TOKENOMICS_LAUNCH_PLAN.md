# MYTH / ZYRA Tokenomics and Launch Plan

**Status:** MYTH Phase 1 is locally tested; the public chain has not launched.
The current validator package is a development prerelease and does not include a
public genesis, chain ID, seed list, or RPC endpoint.

## Token roles

| Token | Denom | Role | Supply |
|---|---|---|---:|
| MYTH | `umyth` | PoS bond/security token; Phase 1 gas; accepted gas token after ZYRA activation | 21,000,000 MYTH, fully minted at genesis; no MYTH halving or inflation |
| ZYRA | `uzyra` | PoUW task-reward token; accepted for gas after activation | 21,000,000 ZYRA cap; zero genesis supply; emitted through PoUW |

MYTH genesis allocation: **1,000,000 MYTH to the founder's mini-PC genesis-validator
account** and **20,000,000 MYTH to the distribution Community Pool/Treasury**. The
Treasury is locked for on-chain governance spending. Validators earn transaction-fee
distributions; MYTH is not inflated after genesis. The exact founder self-bond and
Treasury grant/reward policy are still to be finalized.

## Launch phases

1. **MYTH-only base chain:** bond and pay gas in MYTH. ZYRA PoUW emission is disabled
   by default with `enable_pouw_emissions=false`. Stabilize validator operations,
   genesis supply, Community Pool accounting, and upgrades.
2. **ZYRA PoUW activation:** after the stability gate, governance enables the PoUW
   parameter. Fee ante then accepts MYTH or ZYRA; validator minimum gas prices must
   include both denoms. Gasless access is available at most once per minute while an
   account has zero balance in all currently accepted fee denoms. Any positive
   accepted-fee-token balance revokes gasless permanently, even after being spent.

The public genesis validator is the founder's mini PC. Public chain ID, canonical
genesis/hash, peer seeds, RPC, self-bond amount, commission, unbonding and slashing
parameters must be published before external operators join. New operators acquire
MYTH from the founder and submit a normal staking `create-validator` transaction.

The local four-validator Phase 1 fixture evenly splits the test MYTH validator
allocation across four test accounts. That fixture distribution is not the public
genesis allocation.

## PoUW issuance and task payout

- Emit **0.2 ZYRA per committed block** into the PoUW reward pool. Emission accrues
  even when there are no tasks. The separate 10,000 ZYRA/day cap is removed.
- After each **3,000,000 ZYRA** of cumulative PoUW issuance, reduce the per-block rate
  by **25%**, beginning on the next block. This step-down applies only to ZYRA; MYTH
  has no step-down. Stop at the **21,000,000 ZYRA** cap.
- Task base reward uses its automatically assigned category:

  | Category | Base reward |
  |---|---:|
  | Light | 0.10 ZYRA |
  | Medium | 0.25 ZYRA |
  | Heavy | 0.50 ZYRA |
  | Very Heavy | 1.00 ZYRA |

- Apply a direct score multiplier to the base reward only when the task is quorum-
  approved and every hard gate passes:

  | Canonical score | Multiplier |
  |---|---:|
  | Exactly 75% | 1.00x |
  | >75–<80% | 1.50x |
  | 80–<90% | 1.75x |
  | 90–100% | 2.00x |
  | <75% or a hard gate fails | 0 payout |

- Split an approved payout **60% Miner / 40% Judge pool**. Allow up to four Judges,
  require a 3-of-4 matching quorum for every criterion, and share the Judge pool
  equally only among Judges matching every canonical criterion.
- Each Task ID pays once, on its first approval. Failed attempts pay zero and may be
  retried. If the reward pool is temporarily short, an approved payout waits for
  later emission.

## Category automation

The ZYRA client maps explicit task difficulty to reward category; when difficulty is
not present it uses the runtime profile (`python` → Light, `flask-web` → Medium).
The category is sent in `MsgRegisterTask` before a Miner claims the task.

## Local verification

- `go test ./...` passes; the complete Python suite reports **176 passed, 14 skipped**.
- Four-validator MYTH Phase 1 genesis test produced exactly 21M `umyth`, including
  the 20M Community Pool, kept ZYRA supply at zero, and accepted a MYTH gas-fee tx.
- Separate four-validator ZYRA smoke tests verified 3-of-4 payout, the 2–2
  no-quorum expiry/reclaim path, gasless rate limiting/permanent balance revocation,
  and adapter fee fallback.
- Local tests passed; public genesis and validator rollout remain pending.

## Validator installation

See [MYTH Public Validator and Delegator Onboarding](MYTH_PUBLIC_VALIDATOR_ONBOARDING.md)
for the Ubuntu `.deb`, checksum verification, founder mini-PC preparation, validator
and delegator commands, and public rollout placeholders. The
[`build_validator_release.sh`](../build_validator_release.sh) script prepares Linux
  amd64/arm64 tarballs and `.deb` packages. The development `v0.1.1-myth-phase1` assets
are development packages, not a public-chain configuration.
