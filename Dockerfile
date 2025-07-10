# Use the official Go image as base
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install git, ca-certificates, and build tools
RUN apk add --no-cache git ca-certificates build-base

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code (protobuf files are already generated and committed)
COPY . .

# Run go mod tidy to ensure dependencies are clean
RUN go mod tidy

# Build the blockchain application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X github.com/cosmos/cosmos-sdk/version.Name=arda-poc -X github.com/cosmos/cosmos-sdk/version.AppName=arda-pocd -X github.com/cosmos/cosmos-sdk/version.Version=$(git rev-parse --abbrev-ref HEAD)-$(git log -1 --format='%H') -X github.com/cosmos/cosmos-sdk/version.Commit=$(git log -1 --format='%H')" -o arda-pocd ./cmd/arda-pocd

# Build the sidecar application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.Version=$(git rev-parse --abbrev-ref HEAD)-$(git log -1 --format='%H') -X main.Commit=$(git log -1 --format='%H')" -o tx-sidecar ./cmd/tx-sidecar

# Development stage with Ignite CLI and Air
FROM golang:1.24-alpine AS dev

# Install development dependencies
RUN apk add --no-cache git ca-certificates build-base curl bash

# Install Ignite CLI
RUN curl https://get.ignite.com/cli@v28.11.0! | bash

# Install Air for hot reloading
RUN go install github.com/air-verse/air@latest

# Set working directory
WORKDIR /app

# Copy the entire source code for development
COPY . .

# Create necessary directories
RUN mkdir -p cmd/tx-sidecar/local_data

# Expose ports
EXPOSE 26656 26657 1317 8080

# Default command for development (will be overridden by docker-compose)
CMD ["sh", "-c", "echo 'Development container ready. Use docker-compose to start services.'"]

# Production stage
FROM alpine:latest AS production

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binaries from builder stage
COPY --from=builder /app/arda-pocd .
COPY --from=builder /app/tx-sidecar .

# Copy configuration files
COPY --from=builder /app/config.yml .

# Create necessary directories
RUN mkdir -p cmd/tx-sidecar/local_data && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 26656 26657 1317 8080

# Default command (can be overridden)
CMD ["./arda-pocd", "start", "--home", "/app/.arda-poc", "--api.enable=true", "--api.address=0.0.0.0:1317"]