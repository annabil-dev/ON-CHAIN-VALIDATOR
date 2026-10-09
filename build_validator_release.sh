#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-}"
OUT_DIR="${OUT_DIR:-release/${VERSION}}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMMIT="${COMMIT:-$(git -C "${ROOT}" rev-parse HEAD)}"

if [[ ! "${VERSION}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "VERSION must be a stable semantic version without a leading v (e.g. 1.2.3)." >&2
  exit 2
fi

cd "${ROOT}"
mkdir -p "${OUT_DIR}"
GENESIS_SHA256="$(tr -d '\r\n' < release/genesis.sha256)"
ACTUAL_GENESIS_SHA256="$(sha256sum release/genesis.json | awk '{print $1}')"
if [[ ! "${GENESIS_SHA256}" =~ ^[0-9a-f]{64}$ || "${ACTUAL_GENESIS_SHA256}" != "${GENESIS_SHA256}" ]]; then
  echo "release/genesis.json does not match release/genesis.sha256; refusing release build." >&2
  exit 1
fi

TARGETS=("linux amd64" "linux arm64" "darwin amd64" "darwin arm64" "windows amd64" "windows arm64")
for target in "${TARGETS[@]}"; do
  read -r GOOS GOARCH <<< "${target}"
  STAGE="$(mktemp -d)"
  trap 'rm -rf "${STAGE}"' EXIT
  BINARY="mythprotocold"
  if [[ "${GOOS}" == windows ]]; then BINARY="${BINARY}.exe"; fi

  CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" go build \
    -trimpath -mod=readonly \
    -ldflags "-s -w -X github.com/cosmos/cosmos-sdk/version.Version=${VERSION} -X github.com/cosmos/cosmos-sdk/version.Commit=${COMMIT} -X mythprotocol/cmd/mythprotocold/cmd.officialGenesisSHA256=${GENESIS_SHA256}" \
    -o "${STAGE}/${BINARY}" ./cmd/mythprotocold

  mkdir -p "${STAGE}/release"
  cp release/genesis.json release/genesis.sha256 "${STAGE}/release/"
  cp docs/MYTH_PUBLIC_VALIDATOR_ONBOARDING.md docs/MYTH_TOKENOMICS_LAUNCH_PLAN.md \
    docs/PRODUCTION_USER_GUIDE.md docs/PRODUCTION_VALIDATOR_GUIDE.md \
    docs/PRODUCTION_FOUNDER_GUIDE.md docs/SECURITY_POLICY.md "${STAGE}/"
  cp install_mythprotocold.sh install_mythprotocold.ps1 "${STAGE}/"

  if [[ "${GOOS}" == windows ]]; then
    python3 - "${STAGE}" "${OUT_DIR}/mythprotocold-${VERSION}-windows-${GOARCH}.zip" <<'PY'
import pathlib, sys, zipfile
root, archive = pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2])
with zipfile.ZipFile(archive, "w", compression=zipfile.ZIP_DEFLATED) as zf:
    for path in root.rglob("*"):
        if path.is_file():
            zf.write(path, path.relative_to(root).as_posix())
PY
  else
    tar -C "${STAGE}" -czf "${OUT_DIR}/mythprotocold-${VERSION}-${GOOS}-${GOARCH}.tar.gz" .
  fi

  if [[ "${GOOS}" == linux ]]; then
    DEB_ROOT="${STAGE}/deb-root"
    mkdir -p "${DEB_ROOT}/DEBIAN" "${DEB_ROOT}/usr/bin" "${DEB_ROOT}/usr/share/mythprotocold/release" "${DEB_ROOT}/usr/share/doc/mythprotocold"
    install -m 0755 "${STAGE}/${BINARY}" "${DEB_ROOT}/usr/bin/mythprotocold"
    install -m 0644 release/genesis.json release/genesis.sha256 "${DEB_ROOT}/usr/share/mythprotocold/release/"
    install -m 0644 docs/MYTH_PUBLIC_VALIDATOR_ONBOARDING.md "${DEB_ROOT}/usr/share/doc/mythprotocold/validator-onboarding.md"
    install -m 0644 docs/MYTH_TOKENOMICS_LAUNCH_PLAN.md "${DEB_ROOT}/usr/share/doc/mythprotocold/tokenomics-launch-plan.md"
    install -m 0644 docs/PRODUCTION_USER_GUIDE.md "${DEB_ROOT}/usr/share/doc/mythprotocold/user-guide.md"
    install -m 0644 docs/PRODUCTION_VALIDATOR_GUIDE.md "${DEB_ROOT}/usr/share/doc/mythprotocold/validator-guide.md"
    install -m 0644 docs/PRODUCTION_FOUNDER_GUIDE.md "${DEB_ROOT}/usr/share/doc/mythprotocold/founder-guide.md"
    install -m 0644 docs/SECURITY_POLICY.md "${DEB_ROOT}/usr/share/doc/mythprotocold/security-policy.md"
    cat > "${DEB_ROOT}/DEBIAN/control" <<CONTROL
Package: mythprotocold
Version: ${VERSION}
Section: net
Priority: optional
Architecture: ${GOARCH}
Maintainer: Mythchain maintainers
Description: Mythchain full node and validator daemon
 Provides mythprotocold, the pinned release genesis, and operator documentation.
CONTROL
    dpkg-deb --build --root-owner-group "${DEB_ROOT}" "${OUT_DIR}/mythprotocold_${VERSION}_${GOARCH}.deb"
  fi

  rm -rf "${STAGE}"
  trap - EXIT
done

(
  cd "${OUT_DIR}"
  for DEB in mythprotocold_"${VERSION}"_*.deb; do
    [[ -f "${DEB}" ]] || continue
    sha256sum "${DEB}" > "${DEB}.sha256"
  done
  sha256sum mythprotocold-* mythprotocold_* > SHA256SUMS
)

printf 'Pinned genesis SHA-256: %s\n' "${GENESIS_SHA256}"
