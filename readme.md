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

## MYTH public validator rollout (draft)

The MYTH-first PoS launch sequence, validator/delegator commands, and required
public genesis/seed/RPC release fields are documented in
[`docs/MYTH_PUBLIC_VALIDATOR_ONBOARDING.md`](docs/MYTH_PUBLIC_VALIDATOR_ONBOARDING.md).
The current token roles, MYTH/ZYRA supplies, reward rules, phase gates, and local
validation results are summarized in [`docs/MYTH_TOKENOMICS_LAUNCH_PLAN.md`](docs/MYTH_TOKENOMICS_LAUNCH_PLAN.md).
The `v0.1.0-myth-phase1` development prerelease has Linux amd64/arm64 Ubuntu packages.
Public chain ID, genesis hash, seed peers, and RPC endpoints are still pending.
`build_validator_release.sh` prepares local packages; a package release alone does
not launch or configure the public chain.

## Release
To release a new version of your blockchain, create and push a new tag with `v` prefix. A new draft release with the configured targets will be created.

```
git tag v0.1
git push origin v0.1
```

After a draft release is created, make your final changes from the release page and publish it.

### Install
To install the latest version of your blockchain node's binary, execute the following command on your machine:

```
curl https://get.ignite.com/username/mythprotocol@latest! | sudo bash
```
`username/mythprotocol` should match the `username` and `repo_name` of the Github repository to which the source code was pushed. Learn more about [the install process](https://github.com/ignite/installer).

## Learn more

- [Ignite CLI](https://ignite.com/cli)
- [Tutorials](https://docs.ignite.com/guide)
- [Ignite CLI docs](https://docs.ignite.com)
- [Cosmos SDK docs](https://docs.cosmos.network)
