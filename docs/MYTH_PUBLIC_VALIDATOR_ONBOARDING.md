# MYTH Public Validator and Delegator Onboarding

**Status: MTC base-denom migration in progress.** `v0.1.3-myth-phase1` changes the
L1 bond and base-fee denom from `umyth` to `umtc` (display symbol MTC). The running
`umyth` state/genesis is incompatible with this binary. A fresh MTC genesis and new
chain ID are required; seed/RPC details for the reset network are not yet published.

## MTC and ZYRA roles

- **MTC (`umtc`)** is the proof-of-stake bond/security token. MTC has a 21M hard
  cap, fully allocated at genesis: 1M to the founder's mini-PC genesis-validator account and
  20M locked in the distribution Community Pool/Treasury. MTC is not inflated after
  genesis. At launch, validator/delegator rewards come from transaction fees; any
  later Treasury distribution requires on-chain governance.
- **ZYRA (`uzyra`)** is the separate PoUW task-reward token. It starts at zero and
  remains inactive during the MTC-only stabilization phase. A coordinated upgrade
  enables its emissions after the base chain passes its stability gate.
- Phase 1 network fees are paid in MTC. After ZYRA activation, the chain accepts
  either MTC or ZYRA for fees. Gasless access is limited to one transaction per
  minute while the account has zero balance in all currently accepted fee denoms;
  a positive balance in either accepted fee denom revokes gasless permanently.

MTC is sold/distributed by the founder outside the validator software. Validator
software does not custody keys or sell tokens. A prospective operator first obtains
MTC, funds an account, and then self-bonds through the standard staking transaction.
The initial public genesis allocation is planned for the founder's mini-PC
genesis-validator account (1M MTC); future operators acquire MTC from the founder
and join through staking. The 20M Community Pool cannot be spent without on-chain
governance approval.

## Release artifacts required before public launch

The official release page must publish and sign-off all of the following:

| Artifact/configuration | Public launch value |
|---|---|
| Chain ID for fresh MTC genesis | `<NEW_MTC_CHAIN_ID>` (do not reset under `myth-mainnet-1`) |
| Founder mini-PC public IP | `114.10.44.157` |
| Founder genesis allocation / self-bond | 1,000,000 MTC allocation; 800,000 MTC bonded; 200,000 MTC liquid |
| Genesis validator commission | 5% rate / 6% max / 1 percentage point max-change |
| Validator package | `v0.1.3-myth-phase1` (`umtc` base denom) |
| Canonical genesis JSON and SHA-256 | `<GENESIS_URL>` / `<GENESIS_SHA256>` |
| Seed/persistent peer addresses | `<SEED_ID>@<SEED_IP>:26656` |
| RPC endpoint for tx/query | `<RPC_URL>` |
| Minimum gas-price config (both configured from Phase 1; ZYRA fee use gated until Phase 2) | `<MIN_GAS_PRICE_UMTC>,<MIN_GAS_PRICE_UZYRA>` |
| Bond denom | `umtc` |

Do not announce these placeholders as working public endpoints. The initial
validator set, founder allocation/bond split, Treasury governance process,
unbonding, commission, and slashing parameters must be approved in the public genesis
manifest. The genesis builder defaults to a 5% commission rate, 6% maximum, and
1-percentage-point maximum change; these can be explicitly set on the command line.

## Publish and install the updated Ubuntu package

The package installs `mythprotocold` at `/usr/bin/mythprotocold` and the validator
guide under `/usr/share/doc/mythprotocold/`.

### Publish from Windows PowerShell (no WSL shell required)

Requirements: Python 3, Git, GitHub CLI (`gh`) authenticated to GitHub, and a clean
working tree. From the repository root, the script pushes the current branch and
release tag; GitHub Actions runs tests, builds both Ubuntu architectures, and publishes
the prerelease:

```powershell
python release_mythchain.py v0.1.3-myth-phase1
```

The script asks for confirmation unless `--yes` is supplied and waits for the Actions
release job to finish. To rerun this workflow for a later release, use a new version
tag and commit the intended changes first.

### Install on the mini-PC

On Ubuntu, use `uname -m`: `x86_64` selects `amd64`; `aarch64` selects `arm64`.
Download, verify, and install the amd64 package:

```sh
VERSION=0.1.3-myth-phase1
TAG=v0.1.3-myth-phase1
wget "https://github.com/annabil-dev/ON-CHAIN-VALIDATOR/releases/download/${TAG}/mythprotocold_${VERSION}_amd64.deb"
wget "https://github.com/annabil-dev/ON-CHAIN-VALIDATOR/releases/download/${TAG}/mythprotocold_${VERSION}_amd64.deb.sha256"
sha256sum -c "mythprotocold_${VERSION}_amd64.deb.sha256"
sudo apt install "./mythprotocold_${VERSION}_amd64.deb"
mythprotocold version
```

For `aarch64`, replace `_amd64.deb` with `_arm64.deb` in the download, checksum, and
install commands. Installing the package alone does not create or connect to the
public chain.

## Stop and retire the old `umyth` state

Changing a denom in the binary does not convert balances or rewrite a genesis file.
Stop **every** validator running the old chain before preparing the new MTC genesis.
If a validator is managed by systemd, stop its actual unit; if it runs in a terminal,
stop that process with Ctrl+C. Back up the old genesis and private keys securely for
recovery, but do not reuse the old `umyth` genesis on MTC.

For the old local home, reset CometBFT data/WAL after stopping its node:

```sh
export OLD_MYTH_HOME="$HOME/.mythprotocol" # replace with the actual old --home
mythprotocold comet unsafe-reset-all --home "$OLD_MYTH_HOME"
```

This resets local CometBFT state; it does **not** convert `umyth` balances or create
the new MTC genesis. Retire that old home and use the fresh `MYTH_GENESIS_DIR/validator0`
created below. Repeat the reset/backup on each old validator, and do not restart any
old home. All validators joining the replacement network must use the same new MTC
genesis and a **new chain ID**. `myth-mainnet-2` is an example; confirm the exact ID
before generating or publishing genesis.

## Initialize and sync a full node

```sh
export MYTH_HOME="$HOME/.mythprotocol-mtc"
export MYTH_CHAIN_ID="<NEW_MTC_CHAIN_ID>" # choose a fresh ID; do not reuse the old umyth chain ID

mythprotocold init "<NODE_MONIKER>" --chain-id "$MYTH_CHAIN_ID" --home "$MYTH_HOME"
```

Add the released persistent peer/seed information to
`$MYTH_HOME/config/config.toml`. Configure RPC access according to the public
operator guide; do not expose unsafe RPC methods to the public internet. Set Phase 1
`minimum-gas-prices` in `$MYTH_HOME/config/app.toml` to the published MTC and ZYRA
prices. Phase 1 ante accepts MTC only; uzyra fees are rejected until governance
enables Phase 2 PoUW. Keeping both min prices configured avoids an app.toml restart
when Phase 2 is activated.

Start and verify synchronization:

```sh
mythprotocold start --home "$MYTH_HOME"
mythprotocold status --node "<RPC_URL>"
mythprotocold query staking validators --status bonded --node "<RPC_URL>"
```

A full node follows the chain but does not vote in consensus until it creates a
validator and is in the active validator set.

## Founder mini-PC: prepare the genesis validator

The founder's mini-PC is intended to be the first genesis validator. The new MTC
genesis needs a fresh chain ID (example `myth-mainnet-2`), public IP `114.10.44.157`,
a 1,000,000 MTC allocation, and an 800,000 MTC self-bond, leaving 200,000 MTC liquid.
Confirm the new chain ID before generating genesis.

```sh
export MYTH_GENESIS_DIR="$HOME/myth-mainnet-mtc-genesis"
export MYTH_CHAIN_ID="<NEW_MTC_CHAIN_ID>" # do not reuse the old umyth chain ID
export MYTH_PUBLIC_IP="114.10.44.157"
export FOUNDER_SELF_BOND="800000000000" # 800,000 MTC in umtc; integer base units, no denom suffix

if [ -e "$MYTH_GENESIS_DIR" ]; then
  echo "Refusing to overwrite existing genesis directory: $MYTH_GENESIS_DIR" >&2
  exit 1
fi

mythprotocold multi-node --v 1 \
  --output-dir "$MYTH_GENESIS_DIR" \
  --chain-id "$MYTH_CHAIN_ID" \
  --validators-stake-amount "$FOUNDER_SELF_BOND" \
  --starting-ip-address "$MYTH_PUBLIC_IP" \
  --commission-rate 0.05 \
  --commission-max-rate 0.06 \
  --commission-max-change-rate 0.01 \
  --keyring-backend file \
  --minimum-gas-prices "0.0001umtc,0.0001uzyra"
```

Run this interactively on the mini-PC; the encrypted `file` keyring asks you to set
and confirm its passphrase. This builder performs the node initialization, operator
key creation, MTC genesis allocation, gentx creation/signing, and genesis collection
as one coordinated operation. Do not also run separate `init`, `keys add`,
`add-genesis-account`, or `gentx` commands for this node; that would duplicate steps
or bypass the MTC Treasury/supply setup. In non-test keyring modes the builder does
not add fake `testtoken` balances or write the wallet mnemonic to plaintext.

The `multi-node --v 1` builder creates the validator/operator key with the encrypted
file keyring, node/consensus keys, MTC genesis allocation, 20M Community Pool, mint
cap, Phase 1 gas configuration, and a signed genesis gentx. It does not save a
plaintext mnemonic in file-keyring mode. The gentx bonds 800,000 of the founder's
1,000,000 MTC allocation; approximately 200,000 MTC remains liquid in the founder
account after the gentx is applied. Change the amount only if the public genesis
manifest is revised before finalization.

The generated genesis is at
`$MYTH_GENESIS_DIR/validator0/config/genesis.json`; the signed gentx is in
`$MYTH_GENESIS_DIR/validator0/config/gentx/`. Validate before sharing the public
genesis:

```sh
export MYTH_HOME="$MYTH_GENESIS_DIR/validator0"
mythprotocold genesis validate-genesis --home "$MYTH_HOME"
sha256sum "$MYTH_HOME/config/genesis.json"
```

The gentx and genesis contain public validator/account data but no private keys; send
only those files/checksums to the public launch review. Keep the keyring, node key,
and `priv_validator_key.json` on the mini-PC. After the genesis/checksum is approved,
start the founder node and advertise its P2P address:

```sh
mythprotocold start --home "$MYTH_HOME" \
  --p2p.external-address "${MYTH_PUBLIC_IP}:26656"
```

The first validator is a single point of failure until additional MTC-funded
validators join. Their approved genesis/start instructions and seed list must be
published before onboarding them.

## Join as a validator

Keep the validator consensus key on the validator host and back it up securely.
Never send its private key or the wallet recovery phrase to the project team. Fund a
separate operator account with the approved MTC self-bond and transaction-fee
balance. Get the consensus public key from the local daemon:

```sh
mythprotocold tendermint show-validator --home "$MYTH_HOME"
```

Create `validator.json` using that public key and an approved self-bond amount:

```json
{
  "pubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"<CONSENSUS_PUBKEY>"},
  "amount": "<SELF_BOND_UMTC>umtc",
  "moniker": "<NODE_MONIKER>",
  "identity": "",
  "website": "",
  "security": "",
  "details": "",
  "commission-rate": "<COMMISSION_RATE>",
  "commission-max-rate": "<COMMISSION_MAX_RATE>",
  "commission-max-change-rate": "<COMMISSION_MAX_CHANGE_RATE>",
  "min-self-delegation": "<MIN_SELF_DELEGATION>"
}
```

Submit the validator transaction. Replace fee and RPC placeholders with the published
Phase 1 values:

```sh
mythprotocold tx staking create-validator ./validator.json \
  --from <OPERATOR_KEY> --chain-id "$MYTH_CHAIN_ID" \
  --node "<RPC_URL>" --gas auto --gas-prices "<MIN_GAS_PRICE_UMTC>" --yes \
  --home "$MYTH_HOME"
```

Monitor validator status, signing activity, missed blocks, jailed status, and
delegations. Validators earn transaction-fee distributions; there is no ongoing
MTC inflation. Double-signing and downtime penalties follow the published slashing
parameters.

## Join as a delegator

A delegator does not need to run a validator node. After acquiring MTC and selecting
a bonded validator:

```sh
mythprotocold tx staking delegate <VALIDATOR_OPERATOR_ADDRESS> <AMOUNT>umtc \
  --from <DELEGATOR_KEY> --chain-id "$MYTH_CHAIN_ID" \
  --node "<RPC_URL>" --gas auto --gas-prices "<MIN_GAS_PRICE_UMTC>" --yes \
  --home "$MYTH_HOME"

mythprotocold query staking delegations <DELEGATOR_ADDRESS> --node "<RPC_URL>"
```

Unbonding takes the published unbonding period. Delegated stake shares validator
slashing risk; delegators should review validator uptime and commission before
delegating.

## Phase 2: enable ZYRA PoUW

The MTC-only chain must first pass the public stability/upgrade review. Then all
validators coordinate the approved binary/schema upgrade and governance activation
of ZYRA emissions. Phase 2 accepts fees in either `umtc` or `uzyra`; the MTC
staking denom remains unchanged. Public ZYRA emission parameters and the activation
height are published in the upgrade proposal before activation.
