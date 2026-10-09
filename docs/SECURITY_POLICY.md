# Mythchain Security Policy

## Reporting

Do not publish exploit details or private keys in public issues. Until a dedicated
security contact is published, report suspected vulnerabilities privately to the
repository maintainers through GitHub's private vulnerability reporting. Include the
affected version, reproduction steps, impact, and a proposed mitigation. Do not send
seed phrases or production key material.

## Key custody

- Founder wallet, validator consensus key, node P2P key, and governance control are
  distinct roles and must not share a secret.
- The protocol treasury is held by the distribution module/Community Pool and is
  authorized through governance, not a founder wallet private key.
- Never commit private keys, mnemonics, keyring directories, validator state, or
  production `.env` files.
- Use encrypted, access-controlled, tested backups and least-privilege OS accounts.

## Genesis and release integrity

- Verify the published SHA-256 through a trusted independent channel before launch.
- A release must keep its linked checksum, packaged genesis, chain ID, and public
  manifest consistent.
- Stable builds run tests and static analysis and publish checksums for every asset.
- A genesis change means a new network/chain ID; it is not an in-place upgrade.

## Operations

Restrict RPC and P2P exposure to documented ports, disable unsafe RPC endpoints,
monitor validator signing and peer health, and use coordinated upgrades that retain
consensus quorum. Security reviews and mainnet readiness are required before any
production launch announcement.
