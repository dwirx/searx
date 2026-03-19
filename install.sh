#!/usr/bin/env bash
set -euo pipefail

REPO="dwirx/searx"
BINARY_NAME="search"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"
ACTION="install"
KEEP_LIGHTPANDA=0
AUTO_PATH="${SEARX_AUTO_PATH:-1}"

print_help() {
    cat <<'EOF'
Usage: install.sh [OPTIONS]

Options:
  --install            Install latest release (default)
  --update             Update/reinstall to latest release
  --uninstall          Remove search binary (and Lightpanda by default)
  --keep-lightpanda    Keep Lightpanda during uninstall
  -h, --help           Show this help

Examples:
  curl -sSL https://github.com/dwirx/searx/releases/latest/download/install.sh | bash
  curl -sSL https://github.com/dwirx/searx/releases/latest/download/install.sh | bash -s -- --update
  curl -sSL https://github.com/dwirx/searx/releases/latest/download/install.sh | bash -s -- --uninstall

Environment variables:
  SEARX_INSTALL_DIR   Override install target directory
  SEARX_SKIP_SETUP=1  Skip automatic "search setup" after install/update
  SEARX_AUTO_PATH=0   Disable automatic PATH profile update
EOF
}

while [ $# -gt 0 ]; do
    case "$1" in
        --install) ACTION="install" ;;
        --update) ACTION="update" ;;
        --uninstall) ACTION="uninstall" ;;
        --keep-lightpanda) KEEP_LIGHTPANDA=1 ;;
        -h|--help)
            print_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            print_help
            exit 1
            ;;
    esac
    shift
done

lightpanda_path() {
    if [ -n "${SEARX_LIGHTPANDA_PATH:-}" ]; then
        printf '%s' "${SEARX_LIGHTPANDA_PATH}"
        return
    fi
    printf '%s' "${HOME}/.local/share/searx/lightpanda"
}

path_contains_dir() {
    dir="$1"
    case ":${PATH}:" in
        *":${dir}:"*) return 0 ;;
        *) return 1 ;;
    esac
}

append_path_profile() {
    profile_file="$1"
    dir="$2"
    line="export PATH=\"${dir}:\$PATH\""

    if [ ! -f "${profile_file}" ]; then
        printf '# Added by searx installer\n%s\n' "${line}" > "${profile_file}"
        echo "[✔] PATH entry added to ${profile_file}"
        return
    fi

    if grep -Fq "${dir}" "${profile_file}"; then
        return
    fi

    {
        printf '\n# Added by searx installer\n'
        printf '%s\n' "${line}"
    } >> "${profile_file}"
    echo "[✔] PATH entry added to ${profile_file}"
}

configure_path_if_needed() {
    install_dir="$1"
    [ "${AUTO_PATH}" = "0" ] && return
    path_contains_dir "${install_dir}" && return

    case "${install_dir}" in
        "${HOME}"/*) ;;
        *) return ;;
    esac

    shell_name="$(basename "${SHELL:-}")"
    case "${shell_name}" in
        zsh)
            append_path_profile "${HOME}/.zshrc" "${install_dir}"
            ;;
        bash)
            append_path_profile "${HOME}/.bashrc" "${install_dir}"
            ;;
        *)
            append_path_profile "${HOME}/.profile" "${install_dir}"
            ;;
    esac
}

remove_file() {
    path="$1"
    if rm -f "${path}" 2>/dev/null; then
        echo "[✔] Removed: ${path}"
        return 0
    fi

    if command -v sudo >/dev/null 2>&1 && sudo rm -f "${path}" 2>/dev/null; then
        echo "[✔] Removed: ${path}"
        return 0
    fi

    echo "[!] Could not remove ${path} (permission denied)." >&2
    return 1
}

run_uninstall() {
    removed_any=0
    seen="|"
    cmd_candidate="$(command -v "${BINARY_NAME}" 2>/dev/null || true)"
    candidates=""

    if [ -n "${SEARX_INSTALL_DIR:-}" ]; then
        custom_candidate="${SEARX_INSTALL_DIR}/${BINARY_NAME}"
        candidates="${custom_candidate}"
        if [ -n "${cmd_candidate}" ] && [ "${cmd_candidate}" = "${custom_candidate}" ]; then
            candidates="${cmd_candidate} ${candidates}"
        fi
    else
        candidates="${cmd_candidate} /usr/local/bin/${BINARY_NAME} ${HOME}/.local/bin/${BINARY_NAME}"
    fi

    for candidate in ${candidates}; do
        [ -z "${candidate}" ] && continue
        case "${seen}" in
            *"|${candidate}|"*) continue ;;
        esac
        seen="${seen}${candidate}|"

        if [ -e "${candidate}" ]; then
            if remove_file "${candidate}"; then
                removed_any=1
            fi
        fi
    done

    lp_path="$(lightpanda_path)"
    if [ "${KEEP_LIGHTPANDA}" -eq 0 ] && [ -e "${lp_path}" ]; then
        if remove_file "${lp_path}"; then
            removed_any=1
            remove_file "${lp_path}.tag" || true
            rmdir "$(dirname "${lp_path}")" 2>/dev/null || true
            rmdir "${HOME}/.local/share/searx" 2>/dev/null || true
        fi
    fi

    if [ "${removed_any}" -eq 0 ]; then
        echo "[i] Nothing to uninstall."
    else
        echo "[✔] Uninstall complete."
    fi

    if [ "${KEEP_LIGHTPANDA}" -eq 1 ]; then
        echo "[i] Lightpanda is kept at: ${lp_path}"
    fi
}

if [ "${ACTION}" = "uninstall" ]; then
    run_uninstall
    exit 0
fi

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

if [ "${ACTION}" = "update" ]; then
    echo "Updating ${BINARY_NAME} to ${LATEST_TAG}..."
fi

INSTALLED_VERSION=""
SEARCH_BIN_PATH=""
if [ -n "${SEARX_INSTALL_DIR:-}" ]; then
    if [ -x "${SEARX_INSTALL_DIR}/${BINARY_NAME}" ]; then
        SEARCH_BIN_PATH="${SEARX_INSTALL_DIR}/${BINARY_NAME}"
    fi
else
    SEARCH_BIN_PATH="$(command -v "${BINARY_NAME}" 2>/dev/null || true)"
fi
if [ -n "${SEARCH_BIN_PATH}" ]; then
    INSTALLED_VERSION="$("${SEARCH_BIN_PATH}" --version 2>/dev/null || true)"
fi

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

    install_binary() {
        dest_dir="$1"
        if mkdir -p "${dest_dir}" 2>/dev/null && install -m 0755 "${TMP_BIN}" "${dest_dir}/${BINARY_NAME}" 2>/dev/null; then
            return 0
        fi
        if command -v sudo >/dev/null 2>&1 && sudo mkdir -p "${dest_dir}" && sudo install -m 0755 "${TMP_BIN}" "${dest_dir}/${BINARY_NAME}"; then
            return 0
        fi
        return 1
    }

    INSTALL_DIR=""
    PRIMARY_DIR=""
    if [ -n "${SEARX_INSTALL_DIR:-}" ]; then
        PRIMARY_DIR="${SEARX_INSTALL_DIR}"
    elif [ -n "${SEARCH_BIN_PATH}" ]; then
        PRIMARY_DIR="$(dirname "${SEARCH_BIN_PATH}")"
    elif [ -w "/usr/local/bin" ] || command -v sudo >/dev/null 2>&1; then
        PRIMARY_DIR="/usr/local/bin"
    else
        PRIMARY_DIR="${HOME}/.local/bin"
    fi

    if install_binary "${PRIMARY_DIR}"; then
        INSTALL_DIR="${PRIMARY_DIR}"
    else
        FALLBACK_DIR="${HOME}/.local/bin"
        if [ "${PRIMARY_DIR}" != "${FALLBACK_DIR}" ] && install_binary "${FALLBACK_DIR}"; then
            INSTALL_DIR="${FALLBACK_DIR}"
        else
            echo "[!] Failed to install ${BINARY_NAME} to ${PRIMARY_DIR} (and fallback ${FALLBACK_DIR})." >&2
            exit 1
        fi
    fi

    SEARCH_BIN_PATH="${INSTALL_DIR}/${BINARY_NAME}"
    echo "[✔] Installed ${BINARY_NAME} ${LATEST_TAG} to ${SEARCH_BIN_PATH}"

    if ! command -v "${BINARY_NAME}" >/dev/null 2>&1; then
        configure_path_if_needed "${INSTALL_DIR}"
        if ! command -v "${BINARY_NAME}" >/dev/null 2>&1; then
            echo "[!] ${BINARY_NAME} is not in PATH yet."
            echo "    Add this to your shell profile:"
            echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
        fi
    fi
fi

if [ "${SEARX_SKIP_SETUP:-0}" = "1" ]; then
    echo "[i] Skipping Lightpanda setup because SEARX_SKIP_SETUP=1."
elif [ -n "${SEARCH_BIN_PATH}" ]; then
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
echo
echo "Update:"
echo "  curl -sSL https://github.com/dwirx/searx/releases/latest/download/install.sh | bash -s -- --update"
echo "Uninstall:"
echo "  curl -sSL https://github.com/dwirx/searx/releases/latest/download/install.sh | bash -s -- --uninstall"
