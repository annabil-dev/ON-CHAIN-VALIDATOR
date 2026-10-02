# MTC / ZYRA Tokenomics and Launch Plan

**Status:** the prior `umyth` chain is being replaced by an MTC-denominated genesis.
The `v0.1.3-myth-phase1` candidate uses `umtc`; a fresh genesis and new chain ID are
required. The old genesis/checksum is not compatible with this binary.

## Token roles

| Token | Denom | Role | Supply |
|---|---|---|---:|
| MTC | `umtc` | PoS bond/security token; Phase 1 gas; accepted gas token after ZYRA activation | 21,000,000 MTC, fully minted at genesis; no MTC halving or inflation |
| ZYRA | `uzyra` | PoUW task-reward token; accepted for gas after activation | 21,000,000 ZYRA cap; zero genesis supply; emitted through PoUW |

MTC genesis allocation: **1,000,000 MTC to the founder's mini-PC genesis-validator
account** and **20,000,000 MTC to the distribution Community Pool/Treasury**. The
Treasury is locked for on-chain governance spending. Validators earn transaction-fee
distributions; MTC is not inflated after genesis. The founder plans to self-bond
800,000 MTC from the 1,000,000 MTC allocation, leaving 200,000 MTC liquid;
Treasury grant/reward policy remains to be finalized.

## Launch phases

1. **MTC-only base chain:** bond and pay gas in MTC. ZYRA PoUW emission is disabled
   by default with `enable_pouw_emissions=false`. Stabilize validator operations,
   genesis supply, Community Pool accounting, and upgrades.
2. **ZYRA PoUW activation:** after the stability gate, governance enables the PoUW
   parameter. Fee ante then accepts MTC or ZYRA; validator minimum gas prices must
   include both denoms. Gasless access is available at most once per minute while an
   account has zero balance in all currently accepted fee denoms. Any positive
   accepted-fee-token balance revokes gasless permanently, even after being spent.

The public genesis validator is the founder's mini PC at `114.10.44.157`, using chain
ID `<NEW_MTC_CHAIN_ID>` (do not reuse the old `myth-mainnet-1` genesis). The canonical genesis/hash, peer seeds, RPC, commission,
unbonding, and slashing parameters must be published before external operators join.
New operators acquire MTC from the founder and submit a normal staking
`create-validator` transaction. The genesis builder defaults the founder gentx to a
5% commission, 6% maximum, and 1-percentage-point maximum change; these rates are
configurable through explicit CLI flags.

The local four-validator Phase 1 fixture evenly splits the test MTC validator
allocation across four test accounts. That fixture distribution is not the public
genesis allocation.

## PoUW issuance and task payout

- Emit **0.2 ZYRA per committed block** into the PoUW reward pool. Emission accrues
  even when there are no tasks. The separate 10,000 ZYRA/day cap is removed.
- After each **3,000,000 ZYRA** of cumulative PoUW issuance, reduce the per-block rate
  by **25%**, beginning on the next block. This step-down applies only to ZYRA; MTC
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

- `go test ./...` passes.
- The previous four-validator `umyth` genesis test produced 21M legacy MYTH units,
  including the 20M Community Pool. That genesis is incompatible with the MTC reset.
- The prior one-validator `umyth` file-keyring flow validated locally, kept wallet
  recovery material out of plaintext files, and ran without test-token balances. The
  MTC-denom genesis still needs its own validation before launch.
- Separate four-validator ZYRA smoke tests verified 3-of-4 payout, the 2–2
  no-quorum expiry/reclaim path, gasless rate limiting/permanent balance revocation,
  and adapter fee fallback.
- Local tests passed; public genesis and validator rollout remain pending.
- The fresh MTC genesis must validate `umtc` supply, bond denom and metadata before it
  replaces the previously running chain. A disposable one-validator `umtc` genesis
  fixture has now validated with 21M `umtc`, 20M Community Pool, MTC denom metadata,
  and the requested 800K self-bond.

## Validator installation

See [MYTH Public Validator and Delegator Onboarding](MYTH_PUBLIC_VALIDATOR_ONBOARDING.md)
for the Ubuntu `.deb`, checksum verification, founder mini-PC preparation, validator
and delegator commands, and public rollout placeholders. The
[`build_validator_release.sh`](../build_validator_release.sh) script prepares Linux
amd64/arm64 tarballs and `.deb` packages, and [`release_mythchain.py`](../release_mythchain.py)
publishes them through GitHub Actions without requiring a manual WSL shell. The
`v0.1.3-myth-phase1` package is planned as the `umtc`-base-denom development
prerelease, not a public-chain configuration.
