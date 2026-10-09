# Mythchain OS / Role Package Layout

The platform-first layout requested for operator deliveries is:

```text
release-packages/
├── linux/
│   ├── founder/INSTALL.md
│   └── joiner/INSTALL.md
├── windows/
│   ├── founder/INSTALL.md
│   └── joiner/INSTALL.md
├── debian/
│   ├── founder/INSTALL.md
│   └── joiner/INSTALL.md
└── macos/
    ├── founder/INSTALL.md
    └── joiner/INSTALL.md
```

Each OS has separate role instructions and uses the matching release artifacts:

| OS folder | Package format | Architectures |
|---|---|---|
| `linux` | `.tar.gz` | amd64, arm64 |
| `windows` | `.zip` | amd64, arm64 |
| `debian` | `.deb` | amd64, arm64 |
| `macos` | `.tar.gz` | amd64, arm64 |

Right now these folders contain only the `INSTALL.md` guides. At release time each
folder will also hold its role bundle (binary + genesis + checksum + installer).
The Founder and Joiner role does not imply different binaries: both roles install
the same exact release build and canonical genesis. The Founder package must not
include any private key, mnemonic, or pre-generated validator identity.

Joiner flow is wizard-first: run the installer script, follow the
`init-node` → `join` → `start` prompts with the manifest values. The manual
commands in each guide are fallback only. Founder path targets Ubuntu Linux.

## Release readiness

- [x] Stable cross-platform binary artifacts and checksums are built.
- [x] Linux/macOS and Windows install wizard scripts exist.
- [x] Separate Founder and Joiner install guides exist for all four OS choices.
- [ ] Publish role-specific bundles only after founder genesis creation is implemented
  and the production genesis/network manifest is approved.
- [ ] macOS stable package notarization/signing is not configured yet.

**Do not run the founder install instructions as a production ceremony yet.** The
current `founder-init` creates candidates but the final network manifest/release is
not approved; see
[`docs/PRODUCTION_GENESIS_CEREMONY_CHECKLIST.md`](../docs/PRODUCTION_GENESIS_CEREMONY_CHECKLIST.md).
