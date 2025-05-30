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

var defaultKeyFile = "priv_validator_key.json"

type KeyJSON struct {
	PrivKey struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"priv_key"`
}

// GenerateHashAndSignature creates a hash of the provided message and signs it with the private key
// Returns the hex-encoded hash and signature
func generateHashAndSignature(keyFile string, message string) (hashHex string, sigHex string, err error) {
	hash := sha256.Sum256([]byte(message))
	sigHex, err = signHash(keyFile, hash[:])
	if err != nil {
		return "", "", fmt.Errorf("failed to sign hash: %w", err)
	}

	hashHex = hex.EncodeToString(hash[:])

	fmt.Printf("🔐 Signing complete for message: '%s'\n", message)
	fmt.Printf("🔑 Key file used: %s\n", keyFile)
	fmt.Printf("📝 Hash: %s\n", hashHex)
	fmt.Printf("✅ Signature: %s\n\n", sigHex)

	return hashHex, sigHex, nil
}

func signHash(keyFile string, hash []byte) (sigHex string, err error) {

	file, err := os.ReadFile(keyFile)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", keyFile, err)
	}

	var key KeyJSON
	if err := json.Unmarshal(file, &key); err != nil {
		return "", fmt.Errorf("failed to parse key.json: %w", err)
	}

	privBytes, err := base64.StdEncoding.DecodeString(key.PrivKey.Value)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 private key: %w", err)
	}
	if len(privBytes) != 64 {
		return "", fmt.Errorf("expected 64-byte Ed25519 private key, got %d bytes", len(privBytes))
	}

	privKey := ed25519.NewKeyFromSeed(privBytes[:32])
	signature := ed25519.Sign(privKey, hash[:])
	sigHex = hex.EncodeToString(signature)

	return sigHex, nil
}

// Helper function for command line usage
func generateAndPrintSubmitCommand() { //nolint:unused
	message := "Hello Dubai!"

	hashHex, sigHex, err := generateHashAndSignature(defaultKeyFile, message)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("🔐 Here's your ardad tx command:\n\n")
	fmt.Printf("arda-pocd tx arda submit-hash dubai \\\n")
	fmt.Printf("    %s \\\n", hashHex)
	fmt.Printf("    %s \\\n", sigHex)
	fmt.Printf("    --from ERES -y\n\n")
}
