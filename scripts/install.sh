#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Configuration
REPO="gotech-hub/dashgen"
BINARY_NAME="dashgen"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        log_error "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case $OS in
    linux)
        PLATFORM="linux"
        ;;
    darwin)
        PLATFORM="darwin"
        ;;
    *)
        log_error "Unsupported operating system: $OS"
        exit 1
        ;;
esac

# Get latest version if not specified
if [ -z "$VERSION" ]; then
    log_info "Fetching latest version..."
    VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        log_error "Failed to fetch latest version"
        exit 1
    fi
fi

log_info "Installing DashGen $VERSION for $PLATFORM/$ARCH"

# Construct download URL
BINARY_NAME_PLATFORM="${BINARY_NAME}-${PLATFORM}-${ARCH}"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$BINARY_NAME_PLATFORM"

log_info "Downloading from: $DOWNLOAD_URL"

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download binary
if ! curl -L -o "$TMP_DIR/$BINARY_NAME" "$DOWNLOAD_URL"; then
    log_error "Failed to download binary"
    exit 1
fi

# Make binary executable
chmod +x "$TMP_DIR/$BINARY_NAME"

# Test the binary
log_info "Testing binary..."
if ! "$TMP_DIR/$BINARY_NAME" --version >/dev/null 2>&1; then
    log_error "Binary test failed"
    exit 1
fi

# Install binary
log_info "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
else
    sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
fi

# Verify installation
if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    log_success "DashGen installed successfully!"
    log_info "Version: $($BINARY_NAME --version)"
    log_info "Location: $(which $BINARY_NAME)"
    echo
    log_info "Usage:"
    echo "  $BINARY_NAME --help"
    echo "  $BINARY_NAME --root=. --module=github.com/yourorg/yourapp"
else
    log_error "Installation failed - binary not found in PATH"
    exit 1
fi
