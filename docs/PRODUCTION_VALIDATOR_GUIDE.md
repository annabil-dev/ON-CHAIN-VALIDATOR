# Mythchain Validator Guide

## Prerequisites

- A supported stable Mythchain release and the official network manifest.
- A Linux host with a stable public IP, reliable storage, time synchronization, and
  firewall rules allowing P2P traffic on the published port.
- MTC acquired independently of the node software for self-bond and transaction fees.

Check the release signature/checksum, chain ID, genesis hash, minimum gas prices,
unbonding and slashing policy, and bootstrap peer IDs before joining. Placeholder
network values in development documents are not production endpoints.

## Initialize and join

Run `init-node`, then `join` using the peer values from the signed network manifest.
Keep the generated `config/node_key.json`, `config/priv_validator_key.json`, and
`data/priv_validator_state.json` in the same protected node home. Back them up using
an encrypted, access-controlled procedure; never copy a validator key into a public
release archive or source repository.

Set a non-empty `minimum-gas-prices` value in `config/app.toml` according to the
official manifest. Configure RPC and REST listeners for the intended network
exposure, and disable unsafe RPC methods on public interfaces.

## Run as a service

Use a dedicated unprivileged OS account and a systemd unit similar to:

```ini
[Unit]
Description=Mythchain validator
After=network-online.target
Wants=network-online.target

[Service]
User=mythchain
Group=mythchain
ExecStart=/usr/bin/mythprotocold start --home /var/lib/mythchain
Restart=on-failure
RestartSec=5
LimitNOFILE=65536
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/mythchain

[Install]
WantedBy=multi-user.target
```

Review the sandbox paths against local logs/sockets before enabling the service.
Use a coordinated rolling restart to preserve quorum. Monitor block height, peer
count, signing status, disk space, and upgrade announcements.

## Validator key operations

Use the SDK keyring for operator account transactions. Consensus signing keys are
separate from operator wallet keys. For production deployments, consider remote
signing/HSM support and independently tested backups; do not use the local test
keyring backend outside disposable networks.
