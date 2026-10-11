#!/usr/bin/env bash
# One-command MythChain joiner bootstrap (Linux/macOS).
# Usage: curl -fsSL <release-url>/bootstrap-joiner.sh | bash -s -- [version]
set -euo pipefail

VERSION="${1:-latest}"
REPO="${MYTHCHAIN_REPO:-annabil-dev/ON-CHAIN-VALIDATOR}"
if [[ "${VERSION}" == "latest" ]]; then
  TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -nE 's/.*"tag_name": "v?([^"]+)".*/\1/p')"
  VERSION="${TAG:-}"
fi
[[ "${VERSION}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo "Could not resolve a release version" >&2; exit 2; }
echo "Joining release ${VERSION}"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "${ARCH}" in x86_64|amd64) ARCH=amd64 ;; aarch64|arm64) ARCH=arm64 ;; *) echo "Unsupported architecture: ${ARCH}" >&2; exit 1 ;; esac
TGZ="mythprotocold-${VERSION}-${OS}-${ARCH}.tar.gz"
BASE="https://github.com/${REPO}/releases/download/v${VERSION}"
TMP="$(mktemp -d)"
trap 'rm -r -- "${TMP}"' EXIT
curl -fsSL --max-time 120 "${BASE}/${TGZ}" -o "${TMP}/${TGZ}"
curl -fsSL --max-time 20 "${BASE}/SHA256SUMS" -o "${TMP}/SHA256SUMS"
EXPECTED="$(awk -v f="${TGZ}" '$2==f || $2=="*"f {print $1}' "${TMP}/SHA256SUMS")"
[[ "${EXPECTED}" =~ ^[0-9a-f]{64}$ ]] || { echo 'Missing valid SHA-256' >&2; exit 1; }
if command -v sha256sum >/dev/null; then ACTUAL="$(sha256sum "${TMP}/${TGZ}" | awk '{print $1}')"; else ACTUAL="$(shasum -a 256 "${TMP}/${TGZ}" | awk '{print $1}')"; fi
[[ "${ACTUAL}" == "${EXPECTED}" ]] || { echo 'SHA-256 mismatch' >&2; exit 1; }
mkdir -p "${TMP}/payload"
tar -xzf "${TMP}/${TGZ}" -C "${TMP}/payload"
MYTHCHAIN_VERIFIED_ARCHIVE="${TMP}/${TGZ}" MYTHCHAIN_VERIFIED_CHECKSUMS="${TMP}/SHA256SUMS" MYTHCHAIN_INSTALL_ONLY=1 bash "${TMP}/payload/install_mythprotocold.sh" "${VERSION}" </dev/null
if [[ "${MYTHCHAIN_NO_START:-0}" != "1" ]]; then
  echo 'Starting Mythchain node (Ctrl+C to stop)'
  rm -r -- "${TMP}"
  trap - EXIT
  exec "${MYTHCHAIN_PREFIX:-$HOME/.local}/bin/mythprotocold" start --home "${MYTHCHAIN_HOME:-$HOME/.mythprotocol}"
fi
