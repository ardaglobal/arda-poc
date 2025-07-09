# Use the official Go image as base
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install git, ca-certificates, protobuf compiler, and curl
RUN apk add --no-cache git ca-certificates protobuf curl

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Install Ignite CLI (needed for proto generation)
RUN curl https://get.ignite.com/cli@v28.10.0 | bash

# Install proto dependencies (equivalent to make proto-deps)
RUN go install github.com/bufbuild/buf/cmd/buf@v1.50.0 && \
    go install github.com/cosmos/gogoproto/protoc-gen-gogo@latest && \
    go install github.com/cosmos/cosmos-proto/cmd/protoc-gen-go-pulsar@latest && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1 && \
    go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.16.0 && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.20.0 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Copy source code
COPY . .

# Generate protobuf files (equivalent to make proto-gen)
RUN ignite generate proto-go --yes

# Run go mod tidy (equivalent to setup-dev)
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X github.com/cosmos/cosmos-sdk/version.Name=arda-poc -X github.com/cosmos/cosmos-sdk/version.AppName=arda-pocd -X github.com/cosmos/cosmos-sdk/version.Version=$(git rev-parse --abbrev-ref HEAD)-$(git log -1 --format='%H') -X github.com/cosmos/cosmos-sdk/version.Commit=$(git log -1 --format='%H')" -o arda-pocd ./cmd/arda-pocd

# Use a minimal alpine image for the final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/arda-pocd .

# Change ownership to non-root user
RUN chown appuser:appgroup /app/arda-pocd

# Switch to non-root user
USER appuser

# Expose port (adjust if needed)
EXPOSE 26656 26657 1317

# Run the application
CMD ["./arda-pocd", "start", "--home", "/app/.arda-poc", "--api.enable=true", "--api.address=0.0.0.0:1317"]