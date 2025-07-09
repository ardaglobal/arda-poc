# Use the official Go image as base
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and ca-certificates (needed for go mod download)
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

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
CMD ["./arda-pocd", "start"]