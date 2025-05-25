#!/bin/sh
set -e
HOME_DIR="$HOME/.arda-poc"

while [ ! -f /genesis/genesis.json ]; do
  echo "Waiting for genesis file..."
  sleep 1
done

if [ ! -f "$HOME_DIR/config/genesis.json" ]; then
  arda-pocd init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR"
  mkdir -p "$HOME_DIR/config"
  cp /genesis/genesis.json "$HOME_DIR/config/genesis.json"
fi

exec arda-pocd start --home "$HOME_DIR" \
  --p2p.seeds "$SEEDS" \
  --rpc.laddr tcp://0.0.0.0:26657 \
  --p2p.laddr tcp://0.0.0.0:26656 \
  --minimum-gas-prices 0.001uarda
