# Stage 1: Build environment
FROM golang:1.24-alpine AS builder

# Set ignite version, can be overridden at build time
ARG IGNITE_VERSION=v28.10.0

# Install build-base, git, curl and bash for ignite installation
RUN apk add --no-cache build-base git curl bash

# Install ignite
RUN curl -L https://get.ignite.com/cli@${IGNITE_VERSION}! | bash

# Copy all the source code
WORKDIR /src
COPY . .

# Build the app with ignite to cache dependencies and verify it works.
RUN ignite chain build
RUN ignite chain init

# Move the initialized data to a template location
RUN mv /src/.arda-poc /src/arda-poc-template

# Stage 2: Final image for development
# We use a golang image because ignite needs go to build and run the app.
FROM golang:1.24-alpine

# Install curl for healthcheck and git, bash (needed by ignite)
RUN apk add --no-cache curl git bash

# Copy ignite and the full project source from the builder stage
COPY --from=builder /usr/local/bin/ignite /usr/local/bin/ignite
COPY --from=builder /src /app

WORKDIR /app

# Expose ports used by ignite serve
# 26657: Tendermint RPC
# 1317: API Server
# 9090: gRPC
# 4500: Faucet
EXPOSE 26657 1317 9090 4500

# The command will be provided by docker-compose. By default, it will
# initialize the chain from config.yml and start it.
ENTRYPOINT ["arda-pocd"]
CMD ["start"] 