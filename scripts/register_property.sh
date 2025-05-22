#!/bin/bash

set -e  # Exit on error
set -o pipefail

# Colors for debug output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
CHAIN_ID="ardapoc"
GAS="200000"
HOME_DIR="$HOME/.arda-poc"
NODE="http://localhost:1317"

# Script usage
usage() {
  echo "Usage: $0 [options] <address> <region> <value> <owner1,owner2,...> <share1,share2,...> <from_key>"
  echo
  echo "Options:"
  echo "  -h, --help         Show this help message"
  echo "  -c, --chain-id     Chain ID (default: ardapoc)"
  echo "  -g, --gas          Gas limit (default: 200000)"
  echo "  --home             Home directory (default: ~/.arda-poc)"
  echo "  -n, --node         Node address (default: http://localhost:1317)"
  echo
  echo "Example:"
  echo "  $0 \"123 Main St\" dubai 1000000 \"addr1,addr2\" \"60,40\" sender_key"
  exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -h|--help)
      usage
      ;;
    -c|--chain-id)
      CHAIN_ID="$2"
      shift 2
      ;;
    -g|--gas)
      GAS="$2"
      shift 2
      ;;
    --home)
      HOME_DIR="$2"
      shift 2
      ;;
    -n|--node)
      NODE="$2"
      shift 2
      ;;
    *)
      break
      ;;
  esac
done

# Check if we have enough arguments
if [ $# -lt 6 ]; then
  echo -e "${RED}Error: Not enough arguments${NC}"
  usage
fi

PROPERTY_ADDRESS="$1"
REGION="$2"
VALUE="$3"
OWNERS="$4"
SHARES="$5"
FROM_KEY="$6"

# Temporary files
TX_FILE="tx_register_property.json"
SIGNED_TX="signed_tx_register_property.json"

echo -e "${BLUE}========== Property Registration Transaction ==========${NC}"
echo -e "${YELLOW}Property Address:${NC} $PROPERTY_ADDRESS"
echo -e "${YELLOW}Region:${NC} $REGION"
echo -e "${YELLOW}Value:${NC} $VALUE"
echo -e "${YELLOW}Owners:${NC} $OWNERS"
echo -e "${YELLOW}Shares:${NC} $SHARES"
echo -e "${YELLOW}From:${NC} $FROM_KEY"
echo -e "${YELLOW}Chain ID:${NC} $CHAIN_ID"
echo -e "${YELLOW}Gas Limit:${NC} $GAS"
echo -e "${YELLOW}Home:${NC} $HOME_DIR"
echo -e "${YELLOW}Node:${NC} $NODE"
echo

# Step 1: Try the direct REST endpoint method first
echo -e "${GREEN}[1/3] Attempting direct REST endpoint method...${NC}"

# Extract creator address if FROM_KEY is a name
if [[ "$FROM_KEY" != arda* ]]; then
  echo -e "${YELLOW}Resolving address for key ${FROM_KEY}...${NC}"
  FROM_ADDRESS=$(arda-pocd keys show "$FROM_KEY" -a --home "$HOME_DIR")
  if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to get address for key ${FROM_KEY}${NC}"
    exit 1
  fi
  echo -e "${GREEN}Resolved address: ${NC}$FROM_ADDRESS"
else
  FROM_ADDRESS=$FROM_KEY
fi

# Process owners: convert to array if comma-separated
if [[ $OWNERS == *","* ]]; then
  # It's a comma-separated list
  OWNERS_ARRAY=$(echo $OWNERS | tr ',' ' ' | xargs -n1 | jq -R -s -c 'split("\n") | map(select(length > 0))')
else
  # It's a single value
  OWNERS_ARRAY="[\"$OWNERS\"]"
fi

# Process shares: convert to array if comma-separated
if [[ $SHARES == *","* ]]; then
  # It's a comma-separated list
  SHARES_ARRAY=$(echo $SHARES | tr ',' ' ' | xargs -n1 | jq -R -s -c 'split("\n") | map(select(length > 0))')
else
  # It's a single value
  SHARES_ARRAY="[\"$SHARES\"]"
fi

REST_PAYLOAD=$(cat <<EOF
{
  "creator": "$FROM_ADDRESS",
  "address": "$PROPERTY_ADDRESS",
  "region": "$REGION",
  "value": "$VALUE",
  "owners": $OWNERS_ARRAY,
  "shares": $SHARES_ARRAY
}
EOF
)

echo -e "${BLUE}REST payload:${NC}"
echo "$REST_PAYLOAD" | jq .
echo -e "${YELLOW}Sending direct REST request to ${NC}$NODE/cosmonaut/arda/property/register"

DIRECT_RESULT=$(curl -s -X POST "$NODE/cosmonaut/arda/property/register" \
  -H "Content-Type: application/json" \
  -d "$REST_PAYLOAD")

echo -e "${BLUE}Direct REST Response:${NC}"
echo "$DIRECT_RESULT" | jq . || echo "$DIRECT_RESULT"

# Check if the direct method worked or returned "Not Implemented"
if [[ "$DIRECT_RESULT" == *"Not Implemented"* ]]; then
  echo -e "${YELLOW}Direct REST endpoint not implemented, trying CLI broadcast method...${NC}"
  
  # Step 2: Generate and sign the transaction in one step
  echo -e "${GREEN}[2/3] Generating and signing transaction...${NC}"
  
  BROADCAST_CMD="arda-pocd tx property register-property \"$PROPERTY_ADDRESS\" $REGION $VALUE --owners $OWNERS --shares $SHARES --from $FROM_KEY --chain-id $CHAIN_ID --home $HOME_DIR --yes --output json"
  
  echo -e "${YELLOW}Running:${NC} $BROADCAST_CMD"
  
  TX_RESULT=$(eval $BROADCAST_CMD 2>&1)
  TX_EXIT_CODE=$?
  
  if [ $TX_EXIT_CODE -ne 0 ]; then
    echo -e "${RED}Transaction failed:${NC}"
    echo "$TX_RESULT"
    exit 1
  fi
  
  echo -e "${GREEN}Transaction completed!${NC}"
  echo -e "${BLUE}Transaction result:${NC}"
  echo "$TX_RESULT" | jq . || echo "$TX_RESULT"
  
  # Extract txhash if available
  TXHASH=$(echo "$TX_RESULT" | jq -r '.txhash' 2>/dev/null)
  if [[ -n "$TXHASH" && "$TXHASH" != "null" ]]; then
    echo -e "${GREEN}Transaction successful!${NC}"
    echo -e "${YELLOW}Transaction Hash:${NC} $TXHASH"
    
    # Step 3: Query the transaction
    echo -e "${GREEN}[3/3] Querying transaction status...${NC}"
    echo -e "${YELLOW}Waiting 5 seconds for transaction to be included in a block...${NC}"
    sleep 5
    
    QUERY_CMD="arda-pocd query tx $TXHASH --output json"
    echo -e "${YELLOW}Running:${NC} $QUERY_CMD"
    
    TX_STATUS=$(eval $QUERY_CMD 2>&1)
    QUERY_EXIT_CODE=$?
    
    if [ $QUERY_EXIT_CODE -ne 0 ]; then
      echo -e "${YELLOW}Could not query transaction status yet:${NC}"
      echo "$TX_STATUS"
    else
      echo -e "${GREEN}Transaction status:${NC}"
      echo "$TX_STATUS" | jq . || echo "$TX_STATUS"
    fi
  else
    echo -e "${RED}Transaction failed or txhash not found in response${NC}"
  fi
else
  echo -e "${GREEN}Direct REST endpoint succeeded!${NC}"
fi

echo -e "${BLUE}========== Transaction Process Complete ==========${NC}" 