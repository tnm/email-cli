#!/usr/bin/env bash
set -euo pipefail

REPO_MODULE="github.com/tnm/email-cli"
VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

usage() {
  cat <<'EOF'
Usage:
  ./install.sh [--version <version>] [--dir <install_dir>]

Env:
  VERSION=latest            Go module version to install (tag/commit/latest)
  INSTALL_DIR=~/.local/bin  Install directory (uses GOBIN)

Examples:
  ./install.sh
  ./install.sh --version v0.1.0
  INSTALL_DIR=/usr/local/bin ./install.sh
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    --version)
      VERSION="${2:-}"
      shift 2
      ;;
    --dir)
      INSTALL_DIR="${2:-}"
      shift 2
      ;;
    *)
      echo "Unknown arg: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "${VERSION}" ]]; then
  echo "Missing --version value" >&2
  exit 2
fi

if [[ -z "${INSTALL_DIR}" ]]; then
  echo "Missing --dir value" >&2
  exit 2
fi

if ! command -v go >/dev/null 2>&1; then
  echo "go not found in PATH. Install Go, then re-run." >&2
  echo "https://go.dev/dl/" >&2
  exit 1
fi

mkdir -p "${INSTALL_DIR}"

if [[ ! -w "${INSTALL_DIR}" ]]; then
  echo "Install dir is not writable: ${INSTALL_DIR}" >&2
  echo "Try: INSTALL_DIR=\$HOME/.local/bin ./install.sh" >&2
  exit 1
fi

echo "Installing ${REPO_MODULE}@${VERSION} -> ${INSTALL_DIR}"
GOBIN="${INSTALL_DIR}" go install "${REPO_MODULE}@${VERSION}"

if [[ ":${PATH}:" != *":${INSTALL_DIR}:"* ]]; then
  cat <<EOF

NOTE: ${INSTALL_DIR} is not on your PATH.
Add this to your shell rc (e.g. ~/.zshrc):
  export PATH="${INSTALL_DIR}:\$PATH"
EOF
fi

echo "Installed: ${INSTALL_DIR}/email-cli"

