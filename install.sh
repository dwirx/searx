#!/usr/bin/env bash
set -euo pipefail

REPO="dwirx/searx"
BINARY_NAME="search"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"

for cmd in curl uname mktemp; do
    if ! command -v "${cmd}" >/dev/null 2>&1; then
        echo "Missing required command: ${cmd}" >&2
        exit 1
    fi
done

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "${OS}" in
    linux|darwin) ;;
    *)
        echo "Unsupported OS: ${OS} (supported: linux, darwin)" >&2
        exit 1
        ;;
esac

case "${ARCH}" in
    x86_64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="aarch64" ;;
    *)
        echo "Unsupported architecture: ${ARCH}" >&2
        exit 1
        ;;
esac

echo "Detecting latest ${BINARY_NAME} release..."
LATEST_TAG="$(curl -fsSL "${API_URL}" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')"
if [ -z "${LATEST_TAG}" ]; then
    echo "Could not detect latest release tag from ${API_URL}" >&2
    exit 1
fi

INSTALLED_VERSION=""
if command -v "${BINARY_NAME}" >/dev/null 2>&1; then
    INSTALLED_VERSION="$("${BINARY_NAME}" --version 2>/dev/null || true)"
fi

SEARCH_BIN_PATH="$(command -v "${BINARY_NAME}" 2>/dev/null || true)"
if [ "${INSTALLED_VERSION}" = "${LATEST_TAG}" ] && [ -n "${SEARCH_BIN_PATH}" ]; then
    echo "${BINARY_NAME} ${LATEST_TAG} is already installed at ${SEARCH_BIN_PATH}."
else
    ASSET_NAME="${BINARY_NAME}-${ARCH}-${OS}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${ASSET_NAME}"
    TMP_DIR="$(mktemp -d)"
    TMP_BIN="${TMP_DIR}/${BINARY_NAME}"

    trap 'rm -rf "${TMP_DIR}"' EXIT

    echo "Downloading ${ASSET_NAME}..."
    curl -fL "${DOWNLOAD_URL}" -o "${TMP_BIN}"
    chmod +x "${TMP_BIN}"

    INSTALL_DIR=""
    if [ -w "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
        install -m 0755 "${TMP_BIN}" "${INSTALL_DIR}/${BINARY_NAME}"
    elif command -v sudo >/dev/null 2>&1; then
        INSTALL_DIR="/usr/local/bin"
        sudo install -m 0755 "${TMP_BIN}" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        INSTALL_DIR="${HOME}/.local/bin"
        mkdir -p "${INSTALL_DIR}"
        install -m 0755 "${TMP_BIN}" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    SEARCH_BIN_PATH="${INSTALL_DIR}/${BINARY_NAME}"
    echo "[✔] Installed ${BINARY_NAME} ${LATEST_TAG} to ${SEARCH_BIN_PATH}"

    if ! command -v "${BINARY_NAME}" >/dev/null 2>&1; then
        echo "[!] ${BINARY_NAME} is not in PATH yet."
        echo "    Add this to your shell profile:"
        echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
    fi
fi

if [ -n "${SEARCH_BIN_PATH}" ]; then
    echo "Checking Lightpanda setup..."
    if "${SEARCH_BIN_PATH}" setup; then
        echo "[✔] Lightpanda setup checked."
    else
        echo "[!] Lightpanda setup failed. You can retry manually:"
        echo "    ${SEARCH_BIN_PATH} setup"
    fi
fi

echo
echo "[✔] Installation complete!"
echo "Try:"
echo "  ${BINARY_NAME} --version"
echo "  ${BINARY_NAME} -read \"https://www.nytimes.com/2026/03/17/world/middleeast/iran-war-israel-middle-east-recap.html\" -save"
