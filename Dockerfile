# Build stage
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildTime=${BUILD_TIME}" \
    -o dashgen ./cmd/dashgen

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S dashgen && \
    adduser -u 1001 -S dashgen -G dashgen

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/dashgen /usr/local/bin/dashgen

# Copy additional files
COPY README.md LICENSE ./

# Change ownership
RUN chown -R dashgen:dashgen /app

# Switch to non-root user
USER dashgen

# Set entrypoint
ENTRYPOINT ["dashgen"]

# Default command
CMD ["--help"]
