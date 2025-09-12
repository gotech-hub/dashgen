# Variables
BINARY_NAME=dashgen
VERSION?=dev
GIT_COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildTime=${BUILD_TIME}"

# Build directories
BUILD_DIR=build
DIST_DIR=dist

# Platforms
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build build-cli clean test lint install release help

# Default target
all: clean build

# Build for current platform
build:
	@echo "Building ${BINARY_NAME} for current platform..."
	@mkdir -p ${BUILD_DIR}
	go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} ./cmd/dashgen

# Build CLI tool (alias for build)
build-cli: build

# Build for all platforms
build-all: clean
	@echo "Building ${BINARY_NAME} for all platforms..."
	@mkdir -p ${DIST_DIR}
	@for platform in ${PLATFORMS}; do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		OUTPUT_NAME=${BINARY_NAME}-$$OS-$$ARCH; \
		if [ "$$OS" = "windows" ]; then \
			OUTPUT_NAME=$$OUTPUT_NAME.exe; \
		fi; \
		echo "Building for $$OS/$$ARCH..."; \
		GOOS=$$OS GOARCH=$$ARCH CGO_ENABLED=0 go build ${LDFLAGS} -o ${DIST_DIR}/$$OUTPUT_NAME ./cmd/dashgen; \
	done

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf ${BUILD_DIR} ${DIST_DIR}

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Install binary to system
install: build
	@echo "Installing ${BINARY_NAME} to /usr/local/bin..."
	sudo cp ${BUILD_DIR}/${BINARY_NAME} /usr/local/bin/

# Create release
release: build-all
	@echo "Creating release archives..."
	@mkdir -p ${DIST_DIR}/archives
	@for platform in ${PLATFORMS}; do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		BINARY_NAME_PLATFORM=${BINARY_NAME}-$$OS-$$ARCH; \
		if [ "$$OS" = "windows" ]; then \
			BINARY_NAME_PLATFORM=$$BINARY_NAME_PLATFORM.exe; \
		fi; \
		ARCHIVE_NAME=${BINARY_NAME}-${VERSION}-$$OS-$$ARCH; \
		if [ "$$OS" = "windows" ]; then \
			zip -j ${DIST_DIR}/archives/$$ARCHIVE_NAME.zip ${DIST_DIR}/$$BINARY_NAME_PLATFORM README.md LICENSE; \
		else \
			tar -czf ${DIST_DIR}/archives/$$ARCHIVE_NAME.tar.gz -C ${DIST_DIR} $$BINARY_NAME_PLATFORM -C ../ README.md LICENSE; \
		fi; \
		echo "Created archive: $$ARCHIVE_NAME"; \
	done

# Development build with race detection
dev:
	@echo "Building development version with race detection..."
	@mkdir -p ${BUILD_DIR}
	go build -race ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-dev ./cmd/dashgen

# Run example
example: build
	@echo "Running example..."
	@mkdir -p example/model/user
	@echo 'package user\n\nimport "time"\n\n// @entity db:users\ntype User struct {\n\tID string `json:"id" bson:"_id"`\n\tName string `json:"name" bson:"name"`\n\tCreatedAt time.Time `json:"created_at" bson:"created_at"`\n}' > example/model/user/data.go
	./${BUILD_DIR}/${BINARY_NAME} --root=example --module=github.com/example/myapp --dry
	@rm -rf example

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build for current platform"
	@echo "  build-all  - Build for all platforms"
	@echo "  build-cli  - Build CLI tool (alias for build)"
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run tests"
	@echo "  lint       - Run linter"
	@echo "  install    - Install binary to system"
	@echo "  release    - Create release archives"
	@echo "  dev        - Build development version"
	@echo "  example    - Run example"
	@echo "  help       - Show this help"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION    - Version to build (default: dev)"
	@echo "  Example: make build VERSION=v1.0.0"
