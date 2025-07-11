#!/usr/bin/env bash
set -e

# Seed the volume on first run
if [ ! -d "$ARDA_HOME" ]; then
  echo "[entrypoint] Seeding chain state in $ARDA_HOME"
  mkdir -p "$(dirname "$ARDA_HOME")"
  cp -r /template/.arda-poc "$ARDA_HOME"
fi

# Update gRPC address to listen on all interfaces
# TODO: should create app.toml and config.toml in repo for consistency and then copy them to data dir
sed -i -e 's/grpc.address = "localhost:9090"/grpc.address = "0.0.0.0:9090"/' "$ARDA_HOME/config/app.toml"

exec arda-pocd start --home "$ARDA_HOME" "$@"
