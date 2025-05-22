#!/bin/bash

set -e  # Exit on error
set -o pipefail

# Colors for debug output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script usage
usage() {
  echo "Usage: $0 <property_address> <from_owner> <from_shares> <to_owner> <to_shares>"
  echo
  echo "Example:"
  echo "  $0 \"123 Main St\" ERES 30 bob 30"
  exit 1
}

# Check if we have enough arguments
if [ $# -lt 5 ]; then
  echo -e "${RED}Error: Not enough arguments${NC}"
  usage
fi

PROPERTY_ADDRESS="$1"
FROM_OWNER="$2"
FROM_SHARES="$3"
TO_OWNER="$4"
TO_SHARES="$5"
FROM_KEY="ERES"
HOME_DIR="$HOME/.arda-poc"
REGION="dubai"  # Default region, could be parameterized

echo -e "${BLUE}========== Property Share Transfer Transaction ==========${NC}"
echo -e "${YELLOW}Property Address:${NC} $PROPERTY_ADDRESS"
echo -e "${YELLOW}From Owner:${NC} $FROM_OWNER"
echo -e "${YELLOW}From Shares:${NC} $FROM_SHARES"
echo -e "${YELLOW}To Owner:${NC} $TO_OWNER"
echo -e "${YELLOW}To Shares:${NC} $TO_SHARES"
echo -e "${YELLOW}Signer Key:${NC} $FROM_KEY"
echo -e "${YELLOW}Home:${NC} $HOME_DIR"
echo -e "${YELLOW}Region:${NC} $REGION"
echo

# Step 1: Transfer property shares
echo -e "${GREEN}[1/3] Transferring property shares...${NC}"
  
TRANSFER_CMD="arda-pocd tx property transfer-shares \"$PROPERTY_ADDRESS\" $FROM_OWNER $FROM_SHARES $TO_OWNER $TO_SHARES --from $FROM_KEY --home $HOME_DIR -y --output json"
  
echo -e "${YELLOW}Running:${NC} $TRANSFER_CMD"
  
TX_RESULT=$(eval $TRANSFER_CMD 2>&1)
TX_EXIT_CODE=$?
  
if [ $TX_EXIT_CODE -ne 0 ]; then
  echo -e "${RED}Transaction failed:${NC}"
  echo "$TX_RESULT"
  exit 1
fi
  
echo -e "${GREEN}Share transfer completed!${NC}"
echo -e "${BLUE}Transaction result:${NC}"
echo "$TX_RESULT" | jq . || echo "$TX_RESULT"

# Wait for the first transaction to be processed
echo -e "${YELLOW}Waiting for transaction to be processed...${NC}"
sleep 3

# Simple approach: hardcoded sequence for next transaction
# This will need to be updated if the account's sequence changes
SEQUENCE_FLAG="--sequence 14"

# Step 2: Generate hash and signature using the Go function
echo -e "${GREEN}[2/3] Generating hash and signature for transfer data...${NC}"

# Create the message to hash (concatenate all transfer data)
MESSAGE="$PROPERTY_ADDRESS:$FROM_OWNER:$FROM_SHARES:$TO_OWNER:$TO_SHARES"

# Define key file location
KEY_FILE="$HOME_DIR/config/priv_validator_key.json"
if [ ! -f "$KEY_FILE" ]; then
  echo -e "${RED}Private key file not found: $KEY_FILE${NC}"
  exit 1
fi

# Use the Go function to generate hash and signature
echo -e "${YELLOW}Calling Go hash and signature generator...${NC}"

# Create a temporary Go file
TEMP_GO_FILE=$(mktemp)
TEMP_GO_FILE="${TEMP_GO_FILE}.go"

cat > "$TEMP_GO_FILE" << EOF
package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

type KeyJSON struct {
	PrivKey struct {
		Type  string \`json:"type"\`
		Value string \`json:"value"\`
	} \`json:"priv_key"\`
}

func main() {
	keyFile := "$KEY_FILE"
	message := "$MESSAGE"
	
	file, err := os.ReadFile(keyFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read %s: %v\n", keyFile, err)
		os.Exit(1)
	}

	var key KeyJSON
	if err := json.Unmarshal(file, &key); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse key.json: %v\n", err)
		os.Exit(1)
	}

	privBytes, err := base64.StdEncoding.DecodeString(key.PrivKey.Value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode base64 private key: %v\n", err)
		os.Exit(1)
	}
	if len(privBytes) != 64 {
		fmt.Fprintf(os.Stderr, "expected 64-byte Ed25519 private key, got %d bytes\n", len(privBytes))
		os.Exit(1)
	}

	privKey := ed25519.NewKeyFromSeed(privBytes[:32])
	hash := sha256.Sum256([]byte(message))
	hashHex := hex.EncodeToString(hash[:])
	signature := ed25519.Sign(privKey, hash[:])
	sigHex := hex.EncodeToString(signature)

	// Output in format that can be easily parsed by shell script
	fmt.Printf("%s:%s\n", hashHex, sigHex)
}
EOF

# Run the Go program
HASH_SIG_RESULT=$(go run "$TEMP_GO_FILE")
GO_EXIT_CODE=$?

# Clean up temporary file
rm -f "$TEMP_GO_FILE"

if [ $GO_EXIT_CODE -ne 0 ]; then
  echo -e "${RED}Failed to generate hash and signature:${NC}"
  echo "$HASH_SIG_RESULT"
  exit 1
fi

# Parse the output to get hash and signature
IFS=":" read -r HASH_HEX SIGNATURE_HEX <<< "$HASH_SIG_RESULT"

echo -e "${GREEN}Hash generated:${NC} $HASH_HEX"
echo -e "${GREEN}Signature generated:${NC} $SIGNATURE_HEX"

# Step 3: Submit the hash
echo -e "${GREEN}[3/3] Submitting hash...${NC}"

SUBMIT_CMD="arda-pocd tx arda submit-hash $REGION $HASH_HEX $SIGNATURE_HEX --from $FROM_KEY --home $HOME_DIR $SEQUENCE_FLAG -y --output json"
echo -e "${YELLOW}Running:${NC} $SUBMIT_CMD"

SUBMIT_RESULT=$(eval $SUBMIT_CMD 2>&1)
SUBMIT_EXIT_CODE=$?

if [ $SUBMIT_EXIT_CODE -ne 0 ]; then
  echo -e "${RED}Hash submission failed:${NC}"
  echo "$SUBMIT_RESULT"
  exit 1
fi

echo -e "${GREEN}Hash submission completed!${NC}"
echo -e "${BLUE}Transaction result:${NC}"
echo "$SUBMIT_RESULT" | jq . || echo "$SUBMIT_RESULT"

echo -e "${BLUE}========== Transaction Process Complete ==========${NC}" 