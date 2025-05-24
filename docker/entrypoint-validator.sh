#!/bin/sh
set -e
HOME_DIR="$HOME/.arda-poc"

if [ ! -f "$HOME_DIR/config/genesis.json" ]; then
  arda-pocd init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR"
  arda-pocd config keyring-backend test --home "$HOME_DIR"
  arda-pocd keys add "$MONIKER" --keyring-backend test --home "$HOME_DIR" --algo secp256k1 --yes
  arda-pocd add-genesis-account $(arda-pocd --home "$HOME_DIR" keys show "$MONIKER" -a --keyring-backend test) 1000000000uarda --home "$HOME_DIR"
  arda-pocd gentx "$MONIKER" 100000000uarda --chain-id "$CHAIN_ID" --home "$HOME_DIR" --keyring-backend test
  arda-pocd collect-gentxs --home "$HOME_DIR"
  arda-pocd validate-genesis --home "$HOME_DIR"
  mkdir -p /genesis
  cp "$HOME_DIR/config/genesis.json" /genesis/genesis.json
fi

exec arda-pocd start --home "$HOME_DIR" --rpc.laddr tcp://0.0.0.0:26657 --p2p.laddr tcp://0.0.0.0:26656
