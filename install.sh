#!/bin/bash

# Work CLI Installation Script
# This script detects the platform and installs the appropriate binary

set -e

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture names
case $ARCH in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Set binary name based on OS
BINARY_NAME="work-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
fi

# GitHub repository (update this to match your actual repo)
REPO="jesses-code-adventures/work"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/latest/${BINARY_NAME}"

echo "Installing Work CLI..."
echo "Platform: ${OS}-${ARCH}"
echo "Downloading from: ${DOWNLOAD_URL}"

# Download binary
curl -L "${DOWNLOAD_URL}" -o work
if [ $? -ne 0 ]; then
    echo "Failed to download binary"
    exit 1
fi

# Make executable
chmod +x work

# Install to system
if [ -w "/usr/local/bin" ]; then
    mv work /usr/local/bin/work
    echo "Work CLI installed to /usr/local/bin/work"
elif [ -w "$HOME/.local/bin" ]; then
    mkdir -p "$HOME/.local/bin"
    mv work "$HOME/.local/bin/work"
    echo "Work CLI installed to $HOME/.local/bin/work"
    echo "Make sure $HOME/.local/bin is in your PATH"
else
    echo "Cannot write to /usr/local/bin or $HOME/.local/bin"
    echo "Please run with sudo or move the 'work' binary to a directory in your PATH"
    exit 1
fi

echo "Installation complete! Try running: work --help"