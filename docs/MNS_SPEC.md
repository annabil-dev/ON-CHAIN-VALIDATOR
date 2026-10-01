# Myth Name Service (MNS) — Specification v0.1

Status: design baseline for a local prototype; **not** an on-chain protocol or
public DNS namespace. The `.myth` suffix is usable only by applications or
machines configured with an MNS-aware resolver.

## 1. Scope and rollout

MNS provides canonical human-readable names for Mythchain ecosystem services.
The implementation is staged:

1. **v0 local prototype:** static local mappings and a resolver; no ownership,
   blockchain writes, registration fees, or claims of decentralization.
2. **v1 registry:** deterministic on-chain ownership and records in a separate
   Cosmos SDK `x/mns` module, after the schema and name rules are reviewed.
3. **Later:** hierarchical delegation, chain-backed resolver/cache, and ZYRA
   integration. These are out of scope for v0.

MNS records are aliases/discovery metadata. A name never replaces wallet,
validator, peer-key, or judge-signature verification.

## 2. Canonical name format

- Input is trimmed of surrounding ASCII whitespace and lowercased.
- Only ASCII labels are accepted in v0/v1. Unicode and IDNA conversion are not
  performed; this avoids confusable-name and normalization ambiguity.
- A name must end in `.myth` and contain at least one label before the suffix.
- Each label is 1–63 characters, matches `[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?`,
  and therefore cannot begin or end with `-`.
- Total canonical name length is at most 253 characters, excluding a final DNS
  root dot. Empty labels, repeated dots, ports, paths, control characters, and
  embedded whitespace are rejected.
- Name hashes for a future chain registry are SHA-256 over the UTF-8 bytes of
  the canonical ASCII name, with no trailing dot.

Examples accepted: `mythchain.myth`, `rpc.mythchain.myth`,
`miner01.zyra.mythchain.myth`.

## 3. Record schema

Each name maps to zero or more typed records. A record has:

```text
type: one of A, AAAA, RPC, API, P2P, WALLET, VALIDATOR, CID, TEXT, SERVICE
value: UTF-8 string, 1–1024 bytes
ttl: integer seconds, 0–86400
```

Limits: at most 32 records per name and at most 16 records of one type. Record
types are case-insensitive on input and stored uppercase. Duplicate `(type,
value)` entries are rejected. Values are not implicitly treated as trusted
URLs or identities.

Type-specific validation:

- `A` / `AAAA`: parse as a literal IPv4 / IPv6 address; hostnames are not
  accepted as address values.
- `RPC` / `API`: absolute `http` or `https` URL with a hostname, no credentials,
  no fragment, and at most 2048 characters. Resolution does not fetch it.
- `P2P`: non-empty peer/multiaddress text, at most 512 characters. Cryptographic
  peer identity must still be checked by the consuming protocol.
- `WALLET` / `VALIDATOR`: non-empty chain-address string, at most 128 chars;
  chain-specific address verification belongs to the chain/client layer.
- `CID`: `sha256:<64 lowercase hex>` or a syntactically valid CIDv1. This is a
  locator/commitment only; MNS does not fetch or validate artifact contents.
- `TEXT` / `SERVICE`: printable non-control text, at most 256 characters.

## 4. Ownership and state-transition rules (v1 target)

- A name is unique and has exactly one owner address at a time.
- Only the owner may set/delete records or transfer the name.
- Transfer must include a valid destination address and is atomic: after the
  transaction, the prior owner has no authority.
- Registration, record mutation, and transfer are authenticated by the Cosmos
  transaction signer; no external DNS/HTTP lookup is permitted in a keeper.
- No parent/child delegation or controller role is included in v1. A child name
  is registered independently until a separately specified delegation version.
- Expiry and fees are disabled for the first testnet registry. No name becomes
  available through wall-clock expiry.
- The v0 reserved set is fixed in the local config. For v1, reserved names must
  be explicit genesis state with a recorded owner/governance authority; there
  is no implicit privileged registration path.

Proposed reserved names for review: `mythchain.myth`, `zyra.mythchain.myth`,
`rpc.mythchain.myth`, `explorer.mythchain.myth`, `faucet.mythchain.myth`,
`seed.mythchain.myth`, and `governance.mythchain.myth`.

## 5. Query and resolver behavior

The logical query returns the canonical name, owner, ordered records, and the
state height/revision when the source is a chain. Resolution never follows an
RPC/API record automatically. DNS mapping is restricted to `A` and `AAAA`
records; other MNS records are available to an MNS-aware application/CLI.

For a future chain-backed resolver:

- chain state is authoritative; local cache entries are disposable;
- cache lifetime is `min(record.ttl, configured_max_ttl)` and negative answers
  use a separately bounded negative TTL;
- RPC failure is an error/temporary failure, not permission to invent or retain
  a stale positive result beyond its TTL;
- upstream DNS forwarding applies only to names outside `.myth`;
- response parsing and cache keys use canonical names; unsolicited or mismatched
  RPC responses are rejected.

The v0 local prototype may use static A/AAAA mappings and must label them as
local configuration, not chain state.

## 6. Determinism and security requirements

- State transitions must depend only on the signed transaction and current
  deterministic chain state. No network calls, local clock, OS resolver, or
  map iteration order may affect consensus state.
- Enforce name, record-count, value-size, and TTL limits before writing state.
- Reject malformed addresses/URLs and unknown record types.
- Reject case variants as duplicate names after canonicalization.
- Reject all non-ASCII names in the initial version; revisit IDNA only with a
  pinned normalization standard and confusable-name policy.
- Resolver cache data never grants write authority. Name-to-key binding must be
  checked through signatures/chain identity independently.
- Do not expose a general HTTP proxy: records are data, not instructions to
  fetch resources.

## 7. Versioned transaction/query surface

Planned v1 transactions:

- `MsgRegisterName(creator, name)`
- `MsgSetRecord(creator, name, type, value, ttl)`
- `MsgDeleteRecord(creator, name, type, value)`
- `MsgTransferName(creator, name, new_owner)`

Planned queries:

- `Name(name)` — full owner and record state
- `Resolve(name)` — canonical name and records
- `NamesByOwner(owner, pagination)` — owner index

This surface does not include delegation, controllers, renewal, auction, freeze,
or burn. Those require a separately reviewed state machine and migration plan.

## 8. Acceptance criteria

### v0 local prototype

- Canonicalization and invalid-name cases have unit tests.
- Static local `mythchain.myth`, `zyra.mythchain.myth`, and `rpc.mythchain.myth`
  mappings can be queried without changing system DNS settings.
- Ordinary DNS lookups are unaffected when using the documented local resolver
  configuration; no installer silently edits OS resolver settings.
- Documentation clearly states that v0 is centralized local config.

### v1 chain registry (not implemented by this specification)

- Duplicate/case-variant registration and unauthorized mutation are rejected.
- Owner transfer and genesis import/export are covered by tests.
- Two or more validators return identical state after commit and restart.
- Keeper tests prove no external I/O or nondeterministic behavior is used.

## 9. Explicit non-goals

This spec does not claim `.myth` is a public ICANN TLD, provide browser-wide
resolution, specify token economics, prove endpoint availability, authenticate
service operators from names alone, or make the MNS registry decentralized in
v0.
