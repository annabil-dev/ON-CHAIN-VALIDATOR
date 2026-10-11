#!/usr/bin/env bash
set -euo pipefail
shopt -s nocasematch # bash 3.2 compatibility for macOS interactive prompts

REPO="${MYTHCHAIN_REPO:-annabil-dev/ON-CHAIN-VALIDATOR}"
VERSION="${1:-}"
PREFIX="${MYTHCHAIN_PREFIX:-${HOME}/.local}"
NODE_HOME="${MYTHCHAIN_HOME:-${HOME}/.mythprotocol}"

usage() {
  echo "Usage: $0 VERSION [PREFIX] (example: $0 1.2.3 ~/.local, or $0 latest)" >&2
  exit 2
}
[[ -n "${VERSION}" ]] || usage
if [[ "${VERSION}" == "latest" ]]; then
  TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -nE 's/.*"tag_name": "v?([^"]+)".*/\1/p')"
  VERSION="${TAG:-}"
fi
[[ "${VERSION}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || usage
if [[ -n "${2:-}" ]]; then PREFIX="$2"; fi

# Stop safely before upgrading a running node; never touch its chain database.
INSTALLED_BIN="${PREFIX}/bin/mythprotocold"
if [[ -f "${INSTALLED_BIN}" ]] && command -v pgrep >/dev/null 2>&1; then
  while read -r pid; do
    [[ -n "${pid}" ]] || continue
    command_line="$(ps -p "${pid}" -o args= 2>/dev/null || true)"
    case "${command_line}" in
      "${INSTALLED_BIN} start"*|"${INSTALLED_BIN} run"*)
        echo "Node is already running (PID ${pid}). Stop it normally with Ctrl+C before upgrading." >&2
        echo "Existing node database will NOT be modified." >&2
        exit 1 ;;
    esac
  done < <(pgrep -x mythprotocold || true)
fi
if [[ -f "${NODE_HOME}/config/genesis.json" ]]; then
  expected_genesis="0e9c91fbfdef9e42f27a3f64326ad5b4ef0afb45e086b4c6c945ba8a85d6e3cf"
  current_genesis="$(if command -v sha256sum >/dev/null 2>&1; then sha256sum "${NODE_HOME}/config/genesis.json"; else shasum -a 256 "${NODE_HOME}/config/genesis.json"; fi | awk '{print $1}')"
  [[ "${current_genesis}" == "${expected_genesis}" ]] || {
    echo "Existing node genesis is not the official testnet genesis; upgrade cancelled." >&2
    exit 1
  }
fi

for dependency in curl tar awk; do
  command -v "${dependency}" >/dev/null 2>&1 || { echo "Missing dependency: ${dependency}" >&2; exit 1; }
done
if command -v sha256sum >/dev/null 2>&1; then
  hash_file() { sha256sum "$1" | awk '{print $1}'; }
elif command -v shasum >/dev/null 2>&1; then
  hash_file() { shasum -a 256 "$1" | awk '{print $1}'; }
else
  echo "Missing dependency: sha256sum or shasum" >&2
  exit 1
fi
command -v install >/dev/null 2>&1 || { echo "Missing dependency: install" >&2; exit 1; }

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "${OS}" in linux|darwin) ;; *) echo "Unsupported operating system: ${OS}" >&2; exit 1 ;; esac
case "${ARCH}" in x86_64|amd64) ARCH=amd64 ;; aarch64|arm64) ARCH=arm64 ;; *) echo "Unsupported architecture: ${ARCH}" >&2; exit 1 ;; esac
ARCHIVE="mythprotocold-${VERSION}-${OS}-${ARCH}.tar.gz"
BASE="${MYTHCHAIN_RELEASE_BASE_URL:-https://github.com/${REPO}/releases/download/v${VERSION}}"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

if [[ -n "${MYTHCHAIN_VERIFIED_ARCHIVE:-}" || -n "${MYTHCHAIN_VERIFIED_CHECKSUMS:-}" ]]; then
  [[ -f "${MYTHCHAIN_VERIFIED_ARCHIVE:-}" && -f "${MYTHCHAIN_VERIFIED_CHECKSUMS:-}" ]] || {
    echo "Both verified archive and checksum files are required" >&2
    exit 1
  }
  [[ "$(basename "${MYTHCHAIN_VERIFIED_ARCHIVE}")" == "${ARCHIVE}" ]] || {
    echo "Unexpected verified archive name" >&2
    exit 1
  }
  cp "${MYTHCHAIN_VERIFIED_ARCHIVE}" "${TMP}/${ARCHIVE}"
  cp "${MYTHCHAIN_VERIFIED_CHECKSUMS}" "${TMP}/SHA256SUMS"
else
  curl -fsSL --max-time 120 "${BASE}/${ARCHIVE}" -o "${TMP}/${ARCHIVE}"
  curl -fsSL --max-time 20 "${BASE}/SHA256SUMS" -o "${TMP}/SHA256SUMS"
fi
EXPECTED="$(awk -v file="${ARCHIVE}" '$2 == file || $2 == "*" file {print $1}' "${TMP}/SHA256SUMS")"
[[ "${EXPECTED}" =~ ^[0-9a-f]{64}$ ]] || { echo "No valid checksum entry for ${ARCHIVE}" >&2; exit 1; }
ACTUAL="$(hash_file "${TMP}/${ARCHIVE}")"
[[ "${EXPECTED}" == "${ACTUAL}" ]] || { echo "Checksum mismatch for ${ARCHIVE}" >&2; exit 1; }

mkdir -p "${TMP}/extract" "${PREFIX}/bin" "${PREFIX}/share/mythprotocold/release" "${PREFIX}/share/doc/mythprotocold"
tar -xzf "${TMP}/${ARCHIVE}" -C "${TMP}/extract"
install -m 0755 "${TMP}/extract/mythprotocold" "${PREFIX}/bin/mythprotocold"
if [[ ! -e "${PREFIX}/bin/mythprotocol" && ! -L "${PREFIX}/bin/mythprotocol" ]]; then
  ln -s mythprotocold "${PREFIX}/bin/mythprotocol"
fi
install -m 0644 "${TMP}/extract/release/genesis.json" "${PREFIX}/share/mythprotocold/release/genesis.json"
install -m 0644 "${TMP}/extract/release/genesis.sha256" "${PREFIX}/share/mythprotocold/release/genesis.sha256"
for GUIDE in PRODUCTION_USER_GUIDE.md PRODUCTION_VALIDATOR_GUIDE.md PRODUCTION_FOUNDER_GUIDE.md SECURITY_POLICY.md; do
  install -m 0644 "${TMP}/extract/${GUIDE}" "${PREFIX}/share/doc/mythprotocold/${GUIDE}"
done

BIN="${PREFIX}/bin/mythprotocold"
GENESIS="${PREFIX}/share/mythprotocold/release/genesis.json"
echo "Installed Mythchain ${VERSION} at ${BIN}"
"${BIN}" version

if [[ -f "${NODE_HOME}/config/genesis.json" ]]; then
  echo "Existing node home retained at ${NODE_HOME}; no initialization or join necessary."
  exit 0
fi

fetch_manifest() {
  MANIFEST_JSON="$(curl -fsSL --max-time 15 https://mythchain.pages.dev/network.json)" || return 1
  if command -v python3 >/dev/null 2>&1; then
    VALUES="$(printf '%s' "$MANIFEST_JSON" | python3 -c 'import json,sys; x=json.load(sys.stdin); print(x["chain_id"]); print(",".join(x["persistent_peers"])); print(x["genesis_sha256"])')" || return 1
  elif command -v jq >/dev/null 2>&1; then
    VALUES="$(printf '%s' "$MANIFEST_JSON" | jq -er '.chain_id, (.persistent_peers|join(",")), .genesis_sha256')" || return 1
  else
    echo "JSON parser missing: install python3 or jq first" >&2
    return 1
  fi
  MANIFEST_CHAIN="$(printf '%s\n' "$VALUES" | sed -n '1p')"
  MANIFEST_PEERS="$(printf '%s\n' "$VALUES" | sed -n '2p')"
  MANIFEST_SHA="$(printf '%s\n' "$VALUES" | sed -n '3p')"
  [[ "$MANIFEST_CHAIN" == "mythchain-testnet-v2" && "$MANIFEST_SHA" == "0e9c91fbfdef9e42f27a3f64326ad5b4ef0afb45e086b4c6c945ba8a85d6e3cf" ]] || return 1
  [[ "$MANIFEST_PEERS" =~ ^[0-9a-f]{40}@[a-zA-Z0-9.-]+:[0-9]{2,5}(,[0-9a-f]{40}@[a-zA-Z0-9.-]+:[0-9]{2,5})*$ ]] || return 1
}

join_network() {
  local chain_id="$1" peers="$2" seeds="$3"
  local JOIN_ARGS=(join --home "${NODE_HOME}" --chain-id "${chain_id}")
  [[ -z "${peers}" ]] || JOIN_ARGS+=(--persistent-peers "${peers}")
  [[ -z "${seeds}" ]] || JOIN_ARGS+=(--seeds "${seeds}")
  "${BIN}" "${JOIN_ARGS[@]}"
}

if [[ -t 0 ]]; then
  read -r -p "Initialize this node now? [y/N] " answer
  if [[ "${answer}" != y && "${answer}" != yes ]]; then
    echo "Skipped. When ready:"
    echo "  ${BIN} init-node --home \"${NODE_HOME}\" --genesis \"${GENESIS}\""
    exit 0
  fi
  read -r -p "Node home [${NODE_HOME}]: " input
  NODE_HOME="${input:-${NODE_HOME}}"
  "${BIN}" init-node --home "${NODE_HOME}" --genesis "${GENESIS}"
  read -r -p "Join a network now? [y/N] " answer
  if [[ "${answer}" != y && "${answer}" != yes ]]; then
    echo "Next step:"
    echo "  ${BIN} join --home \"${NODE_HOME}\" --chain-id CHAIN_ID --persistent-peers PEERS"
    exit 0
  fi
  fetch_manifest
  [[ -n "${MANIFEST_CHAIN:-}" ]] && echo "Official testnet: ${MANIFEST_CHAIN} (press Enter to accept)"
  read -r -p "Chain ID [${MANIFEST_CHAIN:-}]: " CHAIN_ID
  [[ -z "${CHAIN_ID}" ]] && CHAIN_ID="${MANIFEST_CHAIN:-}"
  read -r -p "Persistent peer(s) [${MANIFEST_PEERS:-}]: " PEERS
  [[ -z "${PEERS}" ]] && PEERS="${MANIFEST_PEERS:-}"
  read -r -p "Seed(s), optional (ID@host:26656): " SEEDS
  join_network "${CHAIN_ID}" "${PEERS}" "${SEEDS:-}"
  read -r -p "Start the node in this terminal now? [y/N] " answer
  if [[ "${answer}" == y || "${answer}" == yes ]]; then
    exec "${BIN}" start --home "${NODE_HOME}"
  fi
  echo "Next step:"
  echo "  ${BIN} start --home \"${NODE_HOME}\""
  exit 0
fi

# Non-interactive (piped) install: run the full join automatically using the
# live manifest, then start the node in the foreground. Ctrl+C stops it.
fetch_manifest
if [[ -z "${MANIFEST_CHAIN:-}" || -z "${MANIFEST_PEERS:-}" ]]; then
  echo "Could not fetch the network manifest; run these steps manually:" >&2
  echo "  ${BIN} init-node --home \"${NODE_HOME}\" --genesis \"${GENESIS}\"" >&2
  echo "  ${BIN} join --home \"${NODE_HOME}\" --chain-id CHAIN_ID --persistent-peers PEERS" >&2
  echo "  ${BIN} start --home \"${NODE_HOME}\"" >&2
  exit 1
fi
echo "Joining ${MANIFEST_CHAIN} automatically (Ctrl+C to stop the node later)."
"${BIN}" init-node --home "${NODE_HOME}" --genesis "${GENESIS}"
join_network "${MANIFEST_CHAIN}" "${MANIFEST_PEERS}" ""
if [[ "${MYTHCHAIN_INSTALL_ONLY:-0}" == "1" ]]; then
  echo "Node ready. Launch it using: ${PREFIX}/bin/mythprotocol start"
  exit 0
fi
exec "${BIN}" start --home "${NODE_HOME}"
