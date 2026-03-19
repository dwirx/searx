#!/bin/bash
set -e

# Configuration
REPO="dwirx/searx-cli" # Change this to your actual username/repo
BINARY_NAME="search"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="aarch64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

if [ "$OS" != "linux" ]; then
    echo "This script only supports Linux."
    exit 1
fi

echo "Detecting latest version..."
LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
    echo "Could not find latest release. Please check the repo URL."
    exit 1
fi

DOWNLOAD_URL="https://github.com/$(echo $REPO)/releases/download/$LATEST_TAG/$(echo $BINARY_NAME)-$ARCH-linux"

echo "Downloading $BINARY_NAME $LATEST_TAG for $ARCH-linux..."
curl -L -o $BINARY_NAME "$DOWNLOAD_URL"
chmod +x $BINARY_NAME

# Install to /usr/local/bin if possible, otherwise keep in current dir
if [ -w "/usr/local/bin" ]; then
    mv $BINARY_NAME /usr/local/bin/
    echo "[✔] $BINARY_NAME installed to /usr/local/bin/$BINARY_NAME"
else
    echo "[!] /usr/local/bin is not writable. $BINARY_NAME is in the current directory."
    echo "You can move it manually: sudo mv $BINARY_NAME /usr/local/bin/"
fi

# Run setup to ensure lightpanda is ready
if command -v $BINARY_NAME >/dev/null 2>&1; then
    $BINARY_NAME setup
else
    ./$BINARY_NAME setup
fi

echo "[✔] Installation complete! Try running: $BINARY_NAME \"golang\""
