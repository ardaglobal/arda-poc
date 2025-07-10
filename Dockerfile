# Stage 1: Build the application
FROM golang:1.24-alpine AS builder

# Set build arguments
ARG APPNAME=arda-poc
ARG VERSION=latest
ARG COMMIT=unknown

WORKDIR /src

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application with version information
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X github.com/cosmos/cosmos-sdk/version.Name=${APPNAME} \
    -X github.com/cosmos/cosmos-sdk/version.AppName=${APPNAME}d \
    -X github.com/cosmos/cosmos-sdk/version.Version=${VERSION} \
    -X github.com/cosmos/cosmos-sdk/version.Commit=${COMMIT}" \
    -o /app/arda-pocd ./cmd/arda-pocd

# Stage 2: Create the final image
FROM alpine:latest

# Create a non-root user
RUN addgroup -S appuser && adduser -S -G appuser appuser

# Set user and workdir
USER appuser
WORKDIR /home/appuser

# Copy the built binary from the builder stage
COPY --from=builder /app/arda-pocd /usr/bin/arda-pocd

# Expose ports
# 26657: Tendermint RPC
# 1313: API server
EXPOSE 26657 1313

# Set the entrypoint and default command
ENTRYPOINT ["arda-pocd"]
CMD ["start"] 