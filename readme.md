# mythprotocol
**mythprotocol** is a blockchain built using Cosmos SDK and Tendermint and created with [Ignite CLI](https://ignite.com/cli).

## Get started

```
ignite cosmos chain serve
```

`serve` command installs dependencies, builds, initializes, and starts your blockchain in development.

### Configure

Your blockchain in development can be configured with `config.yml`. To learn more, see the [Ignite CLI docs](https://docs.ignite.com).

### Web Frontend

Additionally, Ignite CLI offers a frontend scaffolding feature (based on Vue) to help you quickly build a web frontend for your blockchain:

Use: `ignite cosmos scaffold vue`
This command can be run within your scaffolded blockchain project.


For more information see the [monorepo for Ignite front-end development](https://github.com/ignite/web).

## MTC base-denom migration and validator rollout

The MTC (`umtc`) PoS launch sequence, validator/delegator commands, and required
genesis/seed/RPC release fields are documented in
[`docs/MYTH_PUBLIC_VALIDATOR_ONBOARDING.md`](docs/MYTH_PUBLIC_VALIDATOR_ONBOARDING.md).
The current token roles, MTC/ZYRA supplies, reward rules, phase gates, and local
validation results are summarized in [`docs/MYTH_TOKENOMICS_LAUNCH_PLAN.md`](docs/MYTH_TOKENOMICS_LAUNCH_PLAN.md).
The published `v0.1.3-myth-phase1` prerelease changes the bond/base-fee denom from
`umyth` to `umtc`. It requires a fresh genesis and new chain ID; the previous `umyth`
genesis and checksum cannot be reused. The package release alone does not migrate or
launch a chain.

## Release
Stable releases use semantic-version tags. The workflow runs dependency verification,
unit tests, `go vet`, builds Linux/macOS/Windows amd64/arm64 archives and Ubuntu
packages, and publishes SHA-256 manifests.

```sh
python release_mythchain.py v1.2.3
```

The release helper requires a clean worktree and authenticated GitHub CLI; it asks for
confirmation before pushing the branch and tag. Do not tag until the canonical
production genesis and network manifest are approved.

### Install
Use the versioned installers for release verification and init/join setup:

```sh
bash install_mythprotocold.sh 1.2.3
```

Windows PowerShell installer: `install_mythprotocold.ps1 -Version 1.2.3`.

### Production documentation

- [User guide](docs/PRODUCTION_USER_GUIDE.md)
- [Validator guide](docs/PRODUCTION_VALIDATOR_GUIDE.md)
- [Founder guide](docs/PRODUCTION_FOUNDER_GUIDE.md)
- [Production genesis ceremony checklist](docs/PRODUCTION_GENESIS_CEREMONY_CHECKLIST.md)
- [Security policy](docs/SECURITY_POLICY.md)
- [Treasury audit and tokenomics](docs/MYTH_TOKENOMICS_LAUNCH_PLAN.md)
- [Per-OS Founder/Joiner package layout](release-packages/README.md)

## Learn more

- [Ignite CLI](https://ignite.com/cli)
- [Tutorials](https://docs.ignite.com/guide)
- [Ignite CLI docs](https://docs.ignite.com)
- [Cosmos SDK docs](https://docs.cosmos.network)
