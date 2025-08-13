#!/bin/bash

# Work CLI Installation Script
# Apple Silicon (macOS ARM64) only

set -e

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Check for supported platform
if [ "$OS" != "darwin" ] || [ "$ARCH" != "arm64" ]; then
    echo "Error: Work CLI only supports Apple Silicon (macOS ARM64)"
    echo "Current platform: ${OS}-${ARCH}"
    exit 1
fi

# GitHub repository
REPO="jesses-code-adventures/work"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/latest/work"

echo "Installing Work CLI for Apple Silicon..."
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