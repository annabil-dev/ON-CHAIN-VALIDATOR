#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-dev}"
DEB_VERSION="${VERSION//-/.}"
OUT_DIR="${OUT_DIR:-release/${VERSION}}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

cd "${ROOT}"
mkdir -p "${OUT_DIR}"

for ARCH in amd64 arm64; do
  STAGE="$(mktemp -d)"
  trap 'rm -rf "${STAGE}"' EXIT
  CGO_ENABLED=0 GOOS=linux GOARCH="${ARCH}" go build \
    -trimpath \
    -mod=readonly \
    -ldflags "-s -w -X github.com/cosmos/cosmos-sdk/version.Version=${VERSION}" \
    -o "${STAGE}/mythprotocold" ./cmd/mythprotocold
  cp docs/MYTH_PUBLIC_VALIDATOR_ONBOARDING.md "${STAGE}/"
  cp docs/MYTH_TOKENOMICS_LAUNCH_PLAN.md "${STAGE}/"
  tar -C "${STAGE}" -czf "${OUT_DIR}/mythprotocold-${VERSION}-linux-${ARCH}.tar.gz" \
    mythprotocold MYTH_PUBLIC_VALIDATOR_ONBOARDING.md MYTH_TOKENOMICS_LAUNCH_PLAN.md

  DEB_ROOT="${STAGE}/deb-root"
  mkdir -p "${DEB_ROOT}/DEBIAN" "${DEB_ROOT}/usr/bin" "${DEB_ROOT}/usr/share/doc/mythprotocold"
  install -m 0755 "${STAGE}/mythprotocold" "${DEB_ROOT}/usr/bin/mythprotocold"
  install -m 0644 docs/MYTH_PUBLIC_VALIDATOR_ONBOARDING.md "${DEB_ROOT}/usr/share/doc/mythprotocold/validator-onboarding.md"
  install -m 0644 docs/MYTH_TOKENOMICS_LAUNCH_PLAN.md "${DEB_ROOT}/usr/share/doc/mythprotocold/tokenomics-launch-plan.md"
  cat > "${DEB_ROOT}/DEBIAN/control" <<CONTROL
Package: mythprotocold
Version: ${DEB_VERSION}
Section: net
Priority: optional
Architecture: ${ARCH}
Maintainer: Mythchain maintainers
Description: Mythchain full node and validator daemon
 Provides the Mythchain node binary. Chain ID, genesis, peers, and fee policy
 are published separately for each network.
CONTROL
  dpkg-deb --build --root-owner-group "${DEB_ROOT}" \
    "${OUT_DIR}/mythprotocold_${VERSION}_${ARCH}.deb"
  rm -rf "${STAGE}"
  trap - EXIT
done

(
	cd "${OUT_DIR}"
	for ARCH in amd64 arm64; do
		DEB="mythprotocold_${VERSION}_${ARCH}.deb"
		sha256sum "${DEB}" > "${DEB}.sha256"
	done
	sha256sum mythprotocold-"${VERSION}"-linux-*.tar.gz mythprotocold_"${VERSION}"_*.deb mythprotocold_"${VERSION}"_*.deb.sha256 > SHA256SUMS
)
printf 'Built Mythchain validator tarballs and Ubuntu .deb packages for linux/amd64 and linux/arm64 in %s\n' "${OUT_DIR}"
printf 'Genesis, chain ID, seed peers, RPCs, and fee prices must be published separately after rollout approval.\n'
