#!/usr/bin/env bash
set -e

# Seed the volume on first run
if [ ! -d "$ARDA_HOME" ]; then
  echo "[entrypoint] Seeding chain state in $ARDA_HOME"
  mkdir -p "$(dirname "$ARDA_HOME")"
  cp -r /template/.arda-poc "$ARDA_HOME"
fi

exec arda-pocd start --home "$ARDA_HOME" "$@"
