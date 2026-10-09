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
curl -fsSL "https://github.com/${REPO}/releases/download/v${VERSION}/${TGZ}" -o "${TGZ}"
tar -xzf "${TGZ}"
bash install_mythprotocold.sh "${VERSION}"
