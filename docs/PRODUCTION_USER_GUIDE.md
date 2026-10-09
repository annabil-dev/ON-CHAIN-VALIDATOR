# Mythchain User Guide

## Before joining

Get the current stable release tag, official chain ID, canonical genesis, checksum,
and at least one persistent peer or seed from the official release page. Do not use
values copied from an old release or a local testnet.

## Install and initialize

### Linux or macOS

```sh
curl -fsSL -o install_mythprotocold.sh \
  https://raw.githubusercontent.com/annabil-dev/ON-CHAIN-VALIDATOR/v1.2.3/install_mythprotocold.sh
bash install_mythprotocold.sh 1.2.3
```

The installer verifies the SHA-256 release manifest before installing. It can then
run the init and join wizard. For a non-interactive setup:

```sh
mythprotocold init-node --home "$HOME/.mythprotocol" --genesis /path/to/genesis.json
mythprotocold join --home "$HOME/.mythprotocol" \
  --chain-id "$MYTH_CHAIN_ID" \
  --persistent-peers "$MYTH_PERSISTENT_PEERS" \
  --seeds "$MYTH_SEEDS"
mythprotocold start --home "$HOME/.mythprotocol"
```

### Windows PowerShell

Download `install_mythprotocold.ps1` from the matching source tag and run:

```powershell
Set-ExecutionPolicy -Scope Process Bypass
.\install_mythprotocold.ps1 -Version 1.2.3
```

The default binary location is `%LOCALAPPDATA%\Programs\MythChain`. Add that
directory to `PATH` to invoke `mythprotocold` from a new terminal.

## Verify node status

```sh
mythprotocold status --node tcp://127.0.0.1:26657
mythprotocold query bank balances "$MYTH_ADDRESS" --node tcp://127.0.0.1:26657
```

Never install a genesis from an unverified mirror. The daemon checks the pinned
official genesis checksum before it initializes or starts a release node.
