package scripts

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
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"priv_key"`
}

// Helper to generate hash and signature for a message
func GenerateHashAndSignature() (hashHex string, sigHex string, err error) {
	keyFile := "priv_validator_key.json"
	message := "Hello Dubai!"

	file, err := os.ReadFile(keyFile)
	if err != nil {
		return "", "", fmt.Errorf("failed to read %s: %w", keyFile, err)
	}

	var key KeyJSON
	if err := json.Unmarshal(file, &key); err != nil {
		return "", "", fmt.Errorf("failed to parse key.json: %w", err)
	}

	privBytes, err := base64.StdEncoding.DecodeString(key.PrivKey.Value)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode base64 private key: %w", err)
	}
	if len(privBytes) != 64 {
		return "", "", fmt.Errorf("expected 64-byte Ed25519 private key, got %d bytes", len(privBytes))
	}

	privKey := ed25519.NewKeyFromSeed(privBytes[:32])
	hash := sha256.Sum256([]byte(message))
	hashHex = hex.EncodeToString(hash[:])
	signature := ed25519.Sign(privKey, hash[:])
	sigHex = hex.EncodeToString(signature)

	fmt.Printf("🔐 Signing complete. Here's your ardad tx command:\n\n")
	fmt.Printf("arda-pocd tx arda submit-hash dubai \\\n")
	fmt.Printf("    %s \\\n", hashHex)
	fmt.Printf("    %s \\\n", sigHex)
	fmt.Printf("    --from ERES -y\n\n")

	return hashHex, sigHex, nil
}
