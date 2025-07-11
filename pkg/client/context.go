package client

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ardaglobal/arda-poc/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

// NewClientContext creates a new client context for the sidecar.
func NewClientContext() (client.Context, error) {
	// a custom home directory can be specified via environment variable
	homeDir, ok := os.LookupEnv("ARDA_HOME")
	if !ok {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return client.Context{}, fmt.Errorf("failed to get user home dir: %w", err)
		}
		homeDir = filepath.Join(userHome, "."+app.Name)
	}

	encodingConfig := app.MakeEncodingConfig()
	interfaceRegistry := encodingConfig.InterfaceRegistry
	txConfig := encodingConfig.TxConfig
	appCodec := encodingConfig.Codec

	clientCtx := client.Context{}.
		WithCodec(appCodec).
		WithInterfaceRegistry(interfaceRegistry).
		WithTxConfig(txConfig).
		WithHomeDir(homeDir).
		WithKeyringDir(homeDir).
		WithViper("") // Use an empty prefix for viper to avoid conflicts

	// Read the client config from the app's home directory
	clientCtx, err := config.ReadFromClientConfig(clientCtx)
	if err != nil {
		return client.Context{}, err
	}

	// Create a keyring
	kr, err := keyring.New(app.Name, "test", clientCtx.HomeDir, nil, clientCtx.Codec)
	if err != nil {
		return client.Context{}, err
	}

	clientCtx = clientCtx.WithKeyring(kr)

	return clientCtx, nil
}
