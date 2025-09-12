# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release preparation
- Multi-platform build support
- Docker support
- GitHub Actions CI/CD
- Installation scripts

## [v1.0.0] - TBD

### Added
- **Core Features**
  - Generate repository layer with MongoDB integration
  - Generate API handlers with RESTful endpoints
  - Generate action services for business logic
  - Generate client SDK for service-to-service communication
  - Generate main.go with complete application setup
  - Support for soft delete functionality

- **CLI Features**
  - Support for `--root`, `--module`, `--model` flags
  - Dry run mode with `--dry` flag
  - Force overwrite with `--force` flag
  - Version information with `--version` flag

- **Code Generation**
  - Template-based code generation
  - Type-safe Go code generation
  - Automatic import management
  - Entity detection from `@entity` comments
  - Custom database collection naming

- **Infrastructure**
  - Cross-platform binaries (Linux, macOS, Windows)
  - Docker image support
  - GitHub Actions CI/CD pipeline
  - Automated releases
  - Installation scripts

### Technical Details
- **Supported Go versions**: 1.21+
- **Supported platforms**: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- **Dependencies**: Minimal external dependencies
- **Architecture**: Clean, modular codebase with separation of concerns

### Documentation
- Comprehensive README with examples
- CLI help documentation
- Installation guides for multiple methods
- Usage examples and best practices

---

## Release Process

To create a new release:

1. Update version in relevant files
2. Update this CHANGELOG.md
3. Run the release script:
   ```bash
   ./scripts/release.sh v1.0.0
   ```
4. GitHub Actions will automatically build and publish the release

## Migration Guide

### From Development to v1.0.0
- No breaking changes expected
- First stable release
