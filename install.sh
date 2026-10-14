#!/usr/bin/env bash
# ==============================================================================
# mxping installer script
# Installs the latest mxping binary from GitHub Releases
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/yashbaddi/mxping/main/install.sh | bash
# ==============================================================================

set -euo pipefail

REPO="yashbaddi/mxping"
BINARY="mxping"
INSTALL_DIR="/usr/local/bin"

# Terminal formatting
BOLD="$(tput bold 2>/dev/null || echo '')"
GREEN="$(tput setaf 2 2>/dev/null || echo '')"
YELLOW="$(tput setaf 3 2>/dev/null || echo '')"
RED="$(tput setaf 1 2>/dev/null || echo '')"
RESET="$(tput sgr0 2>/dev/null || echo '')"

info() {
    printf "${GREEN}==>${RESET} ${BOLD}%s${RESET}\n" "$1"
}

warn() {
    printf "${YELLOW}Warning:${RESET} %s\n" "$1"
}

error() {
    printf "${RED}Error:${RESET} %s\n" "$1" >&2
    exit 1
}

# 1. Detect Operating System
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
    linux*)  OS="linux" ;;
    darwin*) OS="darwin" ;;
    *)       error "Unsupported operating system: $OS. Please install manually or build from source." ;;
esac

# 2. Detect CPU Architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)            error "Unsupported architecture: $ARCH." ;;
esac

info "Detected platform: ${OS}/${ARCH}"

# 3. Determine Version to Download
if [ -z "${VERSION:-}" ]; then
    info "Fetching latest release version from GitHub..."
    LATEST_JSON="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null || true)"
    if [ -n "$LATEST_JSON" ]; then
        VERSION="$(echo "$LATEST_JSON" | grep '"tag_name":' | head -n1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
    fi
fi

if [ -z "${VERSION:-}" ]; then
    warn "Could not resolve latest release via GitHub API. Falling back to default."
    VERSION="v0.1.0"
fi

# Strip leading 'v' for archive naming matching goreleaser template
CLEAN_VERSION="${VERSION#v}"
TARBALL="${BINARY}_${CLEAN_VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}"

info "Downloading ${BINARY} ${VERSION}..."
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

if ! curl -fsSL "$DOWNLOAD_URL" -o "${TMP_DIR}/${TARBALL}"; then
    error "Failed to download from ${DOWNLOAD_URL}. Check if release ${VERSION} exists."
fi

# 4. Extract archive
info "Extracting archive..."
tar -xzf "${TMP_DIR}/${TARBALL}" -C "$TMP_DIR"

if [ ! -f "${TMP_DIR}/${BINARY}" ]; then
    error "Binary ${BINARY} was not found in archive."
fi

# 5. Determine installation destination
TARGET_DIR="$INSTALL_DIR"
if [ ! -w "$TARGET_DIR" ]; then
    # Try using ~/.local/bin if /usr/local/bin is not writable and sudo is not available
    if command -v sudo >/dev/null 2>&1 && [ -t 0 ]; then
        info "Elevating permissions to install to ${TARGET_DIR}..."
        sudo mv "${TMP_DIR}/${BINARY}" "${TARGET_DIR}/${BINARY}"
        sudo chmod +x "${TARGET_DIR}/${BINARY}"
    else
        TARGET_DIR="${HOME}/.local/bin"
        mkdir -p "$TARGET_DIR"
        mv "${TMP_DIR}/${BINARY}" "${TARGET_DIR}/${BINARY}"
        chmod +x "${TARGET_DIR}/${BINARY}"
        warn "Installed to ${TARGET_DIR} because ${INSTALL_DIR} requires root permissions."
        if [[ ":$PATH:" != *":${TARGET_DIR}:"* ]]; then
            warn "Please add ${TARGET_DIR} to your PATH (e.g., export PATH=\"\$HOME/.local/bin:\$PATH\")."
        fi
    fi
else
    mv "${TMP_DIR}/${BINARY}" "${TARGET_DIR}/${BINARY}"
    chmod +x "${TARGET_DIR}/${BINARY}"
fi

info "✓ ${BINARY} successfully installed to ${TARGET_DIR}/${BINARY}!"

# 6. Verify installation
if command -v "$BINARY" >/dev/null 2>&1; then
    printf "\n"
    "$BINARY" -version || true
    printf "\n${GREEN}Run 'mxping --help' to get started.${RESET}\n"
fi

