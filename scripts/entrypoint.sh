#!/usr/bin/env bash
set -e

# Seed the volume on first run
if [ ! -d "$ARDA_HOME" ]; then
  echo "[entrypoint] Seeding chain state in $ARDA_HOME"
  mkdir -p "$(dirname "$ARDA_HOME")"
  cp -r /template/.arda-poc "$ARDA_HOME"
fi

# Update gRPC address to listen on all interfaces and enable gRPC
# The sed command is removed in favor of command-line flags for robustness.

exec arda-pocd start --home "$ARDA_HOME" --grpc.address "0.0.0.0:9090" "$@"
