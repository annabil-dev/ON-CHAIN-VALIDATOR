# MYTH Public Validator and Delegator Onboarding

**Status: validator-binary prerelease available; public network not launched.** The
`v0.1.1-myth-phase1` GitHub prerelease contains amd64/arm64 Ubuntu packages. No public
chain ID, genesis file/checksum, seed address, or RPC endpoint has been published.
Do not use local `.testnets` files or local validator keys for a public network.

## MYTH and ZYRA roles

- **MYTH (`umyth`)** is the proof-of-stake bond/security token. MYTH has a 21M hard
  cap, fully allocated at genesis: 1M to the founder's mini-PC genesis-validator account and
  20M locked in the distribution Community Pool/Treasury. MYTH is not inflated after
  genesis. At launch, validator/delegator rewards come from transaction fees; any
  later Treasury distribution requires on-chain governance.
- **ZYRA (`uzyra`)** is the separate PoUW task-reward token. It starts at zero and
  remains inactive during the MYTH-only stabilization phase. A coordinated upgrade
  enables its emissions after the base chain passes its stability gate.
- Phase 1 network fees are paid in MYTH. After ZYRA activation, the chain accepts
  either MYTH or ZYRA for fees. Gasless access is limited to one transaction per
  minute while the account has zero balance in all currently accepted fee denoms;
  a positive balance in either accepted fee denom revokes gasless permanently.

MYTH is sold/distributed by the founder outside the validator software. Validator
software does not custody keys or sell tokens. A prospective operator first obtains
MYTH, funds an account, and then self-bonds through the standard staking transaction.
The initial public genesis allocation is planned for the founder's mini-PC
genesis-validator account (1M MYTH); future operators acquire MYTH from the founder
and join through staking. The 20M Community Pool cannot be spent without on-chain
governance approval.

## Release artifacts required before public launch

The official release page must publish and sign-off all of the following:

| Artifact/configuration | Public launch value |
|---|---|
| Chain ID | `<MYTH_CHAIN_ID>` |
| Binary package release | `v0.1.1-myth-phase1` (development prerelease) |
| Canonical genesis JSON and SHA-256 | `<GENESIS_URL>` / `<GENESIS_SHA256>` |
| Seed/persistent peer addresses | `<SEED_ID>@<SEED_IP>:26656` |
| RPC endpoint for tx/query | `<RPC_URL>` |
| Minimum gas-price config (both configured from Phase 1; ZYRA fee use gated until Phase 2) | `<MIN_GAS_PRICE_UMYTH>,<MIN_GAS_PRICE_UZYRA>` |
| Bond denom | `umyth` |

Do not announce these placeholders as working public endpoints. The initial
validator set, exact founder self-bond, Treasury governance process, unbonding,
commission, and slashing parameters must be approved in the public genesis manifest.

## Install a released Ubuntu package

The release folder contains `.deb` packages and `SHA256SUMS`. The package installs
`mythprotocold` at `/usr/bin/mythprotocold` and the validator guide under
`/usr/share/doc/mythprotocold/`.

### Build the package from the Windows source tree with WSL

From Windows PowerShell, build both Linux architectures into the WSL cache:

```powershell
wsl.exe -d Ubuntu --cd '/mnt/d/Semester 5/AI/mythchain/mythprotocol' -e bash -c 'export PATH=/home/mythchain/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.7.linux-amd64/bin:/usr/local/bin:/usr/bin:/bin; VERSION=0.1.1-dev-myth-phase1 OUT_DIR=/home/mythchain/.cache/myth-validator-release-myth-phase1-v0.1.1 bash build_validator_release.sh'
```

The output folder contains amd64/arm64 tarballs, `.deb` packages, and per-package
checksums. It is available in Windows Explorer at:
`\\wsl.localhost\Ubuntu\home\mythchain\.cache\myth-validator-release-myth-phase1-v0.1.1`.
For a mini-PC, transfer just the matching `.deb` and its `.sha256` file with `scp` or
USB. On Ubuntu, use `uname -m`: `x86_64` selects `amd64`; `aarch64` selects `arm64`.
Verify and install the package:

```sh
VERSION=0.1.1-dev-myth-phase1
sha256sum -c "mythprotocold_${VERSION}_amd64.deb.sha256" # use _arm64.deb on aarch64
sudo apt install "./mythprotocold_${VERSION}_amd64.deb" # use _arm64.deb on aarch64
mythprotocold version
```

The binary is available from the GitHub release. For amd64:

```sh
VERSION=0.1.1-dev-myth-phase1
TAG=v0.1.1-myth-phase1
wget "https://github.com/annabil-dev/ON-CHAIN-VALIDATOR/releases/download/${TAG}/mythprotocold_${VERSION}_amd64.deb"
wget "https://github.com/annabil-dev/ON-CHAIN-VALIDATOR/releases/download/${TAG}/mythprotocold_${VERSION}_amd64.deb.sha256"
sha256sum -c "mythprotocold_${VERSION}_amd64.deb.sha256"
sudo apt install "./mythprotocold_${VERSION}_amd64.deb"
```

Use `_arm64.deb` for `aarch64`. This is a development binary package: installing it
does not create or connect to the public chain. Wait for the official genesis, chain
ID, seed peers, and RPC endpoint before operating it as a public validator.

## Initialize and sync a full node

```sh
export MYTH_HOME="$HOME/.mythprotocol"
export MYTH_CHAIN_ID="<MYTH_CHAIN_ID>"

mythprotocold init "<NODE_MONIKER>" --chain-id "$MYTH_CHAIN_ID" --home "$MYTH_HOME"
```

Add the released persistent peer/seed information to
`$MYTH_HOME/config/config.toml`. Configure RPC access according to the public
operator guide; do not expose unsafe RPC methods to the public internet. Set Phase 1
`minimum-gas-prices` in `$MYTH_HOME/config/app.toml` to the published MYTH and ZYRA
prices. Phase 1 ante accepts MYTH only; uzyra fees are rejected until governance
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

The founder's mini-PC is intended to be the first genesis validator. Once the final
public chain ID and public IP are chosen, initialize the one-validator genesis on
that mini-PC with the purpose-built MYTH genesis builder:

```sh
export MYTH_GENESIS_DIR="$HOME/myth-genesis"
export MYTH_CHAIN_ID="<FINAL_PUBLIC_CHAIN_ID>" # immutable once genesis is published
export MYTH_PUBLIC_IP="<FOUNDER_MINI_PC_PUBLIC_IP>"
export FOUNDER_SELF_BOND="100000000umyth" # example: 100 MYTH from the founder's 1M allocation

mythprotocold multi-node --v 1 \
  --output-dir "$MYTH_GENESIS_DIR" \
  --chain-id "$MYTH_CHAIN_ID" \
  --validators-stake-amount "$FOUNDER_SELF_BOND" \
  --starting-ip-address "$MYTH_PUBLIC_IP" \
  --keyring-backend file \
  --minimum-gas-prices "0.0001umyth,0.0001uzyra"
```

Run this interactively on the mini-PC; the encrypted `file` keyring asks you to set
and confirm its passphrase. In non-test keyring modes the builder does not add fake
`testtoken` balances or write the wallet mnemonic to a plaintext `key_seed.json`.

The `multi-node --v 1` builder creates the validator/operator key with the encrypted
file keyring, node/consensus keys, MYTH genesis allocation, 20M Community Pool, mint
cap, Phase 1 gas configuration, and a signed genesis gentx. It does not save a
plaintext mnemonic in file-keyring mode. The example self-bond is only 100 MYTH; the
founder's full 1M MYTH remains allocated to the founder account, with the rest liquid
for later sale. Change the self-bond amount only before finalizing genesis.

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

The first validator is a single point of failure until additional MYTH-funded
validators join. Their approved genesis/start instructions and seed list must be
published before onboarding them.

## Join as a validator

Keep the validator consensus key on the validator host and back it up securely.
Never send its private key or the wallet recovery phrase to the project team. Fund a
separate operator account with the approved MYTH self-bond and transaction-fee
balance. Get the consensus public key from the local daemon:

```sh
mythprotocold tendermint show-validator --home "$MYTH_HOME"
```

Create `validator.json` using that public key and an approved self-bond amount:

```json
{
  "pubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"<CONSENSUS_PUBKEY>"},
  "amount": "<SELF_BOND_UMYTH>umyth",
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
  --node "<RPC_URL>" --gas auto --gas-prices "<MIN_GAS_PRICE_UMYTH>" --yes \
  --home "$MYTH_HOME"
```

Monitor validator status, signing activity, missed blocks, jailed status, and
delegations. Validators earn transaction-fee distributions; there is no ongoing
MYTH inflation. Double-signing and downtime penalties follow the published slashing
parameters.

## Join as a delegator

A delegator does not need to run a validator node. After acquiring MYTH and selecting
a bonded validator:

```sh
mythprotocold tx staking delegate <VALIDATOR_OPERATOR_ADDRESS> <AMOUNT>umyth \
  --from <DELEGATOR_KEY> --chain-id "$MYTH_CHAIN_ID" \
  --node "<RPC_URL>" --gas auto --gas-prices "<MIN_GAS_PRICE_UMYTH>" --yes \
  --home "$MYTH_HOME"

mythprotocold query staking delegations <DELEGATOR_ADDRESS> --node "<RPC_URL>"
```

Unbonding takes the published unbonding period. Delegated stake shares validator
slashing risk; delegators should review validator uptime and commission before
delegating.

## Phase 2: enable ZYRA PoUW

The MYTH-only chain must first pass the public stability/upgrade review. Then all
validators coordinate the approved binary/schema upgrade and governance activation
of ZYRA emissions. Phase 2 accepts fees in either `umyth` or `uzyra`; the MYTH
staking denom remains unchanged. Public ZYRA emission parameters and the activation
height are published in the upgrade proposal before activation.
