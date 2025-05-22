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
  echo "Usage: $0 <address> <region> <value> <owners> <shares>"
  echo
  echo "Example:"
  echo "  $0 \"123 Main St\" dubai 1000000 \"addr1,addr2\" \"60,40\""
  exit 1
}

# Check if we have enough arguments
if [ $# -lt 5 ]; then
  echo -e "${RED}Error: Not enough arguments${NC}"
  usage
fi

PROPERTY_ADDRESS="$1"
REGION="$2"
VALUE="$3"
OWNERS="$4"
SHARES="$5"
FROM_KEY="ERES"
HOME_DIR="$HOME/.arda-poc"
CHAIN_ID="ardapoc"

echo -e "${BLUE}========== Property Registration Transaction ==========${NC}"
echo -e "${YELLOW}Property Address:${NC} $PROPERTY_ADDRESS"
echo -e "${YELLOW}Region:${NC} $REGION"
echo -e "${YELLOW}Value:${NC} $VALUE"
echo -e "${YELLOW}Owners:${NC} $OWNERS"
echo -e "${YELLOW}Shares:${NC} $SHARES"
echo -e "${YELLOW}From:${NC} $FROM_KEY"
echo -e "${YELLOW}Home:${NC} $HOME_DIR"
echo

# Step 1: Register the property
echo -e "${GREEN}[1/3] Registering property...${NC}"
  
REGISTER_CMD="arda-pocd tx property register-property \"$PROPERTY_ADDRESS\" $REGION $VALUE --owners $OWNERS --shares $SHARES --from $FROM_KEY --home $HOME_DIR -y --output json"
  
echo -e "${YELLOW}Running:${NC} $REGISTER_CMD"
  
TX_RESULT=$(eval $REGISTER_CMD 2>&1)
TX_EXIT_CODE=$?
  
if [ $TX_EXIT_CODE -ne 0 ]; then
  echo -e "${RED}Transaction failed:${NC}"
  echo "$TX_RESULT"
  exit 1
fi
  
echo -e "${GREEN}Property registration completed!${NC}"
echo -e "${BLUE}Transaction result:${NC}"
echo "$TX_RESULT" | jq . || echo "$TX_RESULT"

# Wait for the first transaction to be processed
echo -e "${YELLOW}Waiting for transaction to be processed...${NC}"
sleep 3

# Simple approach: hardcoded sequence for now
SEQUENCE_FLAG="--sequence 13"

# Step 2: Generate hash and signature using the Go function
echo -e "${GREEN}[2/3] Generating hash and signature for property data...${NC}"

# Create the message to hash (concatenate all property data)
MESSAGE="$PROPERTY_ADDRESS:$REGION:$VALUE:$OWNERS:$SHARES"

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