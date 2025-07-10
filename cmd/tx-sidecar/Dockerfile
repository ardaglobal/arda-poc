
# Stage 1: Build the application
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
# We are building from the root of the project, so we need to specify the path to the main package
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/tx-sidecar ./cmd/tx-sidecar

# Stage 2: Create the final image
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/tx-sidecar .

# The swagger documentation seems to be served by the application, so we copy it as well.
COPY ./cmd/tx-sidecar/docs ./docs
COPY config.yml .

# The application uses a local sqlite database, so we create the directory for it.
# This directory can be mounted as a volume in production.
RUN mkdir -p /app/local_data

# Expose the port the application runs on
EXPOSE 8080

# Run the application
ENTRYPOINT ["./tx-sidecar"] 