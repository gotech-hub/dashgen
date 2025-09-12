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

# Check if version is provided
if [ -z "$1" ]; then
    log_error "Usage: $0 <version>"
    log_info "Example: $0 v1.0.0"
    exit 1
fi

VERSION=$1

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    log_error "Version must be in format vX.Y.Z (e.g., v1.0.0)"
    exit 1
fi

log_info "Starting release process for version: $VERSION"

# Check if we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    log_warning "You are not on the main branch (current: $CURRENT_BRANCH)"
    read -p "Do you want to continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Release cancelled"
        exit 1
    fi
fi

# Check if working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    log_error "Working directory is not clean. Please commit or stash your changes."
    git status --short
    exit 1
fi

# Check if tag already exists
if git tag -l | grep -q "^$VERSION$"; then
    log_error "Tag $VERSION already exists"
    exit 1
fi

# Update version in files if needed
log_info "Updating version in files..."

# You can add version updates here if needed
# For example, update version in a version file or README

# Run tests
log_info "Running tests..."
if ! make test; then
    log_error "Tests failed"
    exit 1
fi

# Build for all platforms
log_info "Building for all platforms..."
if ! make build-all VERSION=$VERSION; then
    log_error "Build failed"
    exit 1
fi

# Create and push tag
log_info "Creating and pushing tag: $VERSION"
git tag -a "$VERSION" -m "Release $VERSION"

log_info "Pushing tag to origin..."
git push origin "$VERSION"

log_success "Tag $VERSION has been created and pushed!"
log_info "GitHub Actions will automatically create the release."
log_info "Check the progress at: https://github.com/gotech-hub/dashgen/actions"

# Clean up build artifacts
log_info "Cleaning up build artifacts..."
make clean

log_success "Release process completed successfully!"
log_info "The release will be available at: https://github.com/gotech-hub/dashgen/releases/tag/$VERSION"
