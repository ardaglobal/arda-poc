package client

import (
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/ardaglobal/arda-poc/app"
)

// NewClientContext creates a new client context, reading configuration from the
// application's home directory. This is a reusable function for creating
// a consistent client environment for tools and services.
func NewClientContext() (client.Context, error) {
	// Use the new MakeEncodingConfig to get the necessary configs
	encodingConfig := app.MakeEncodingConfig()
	interfaceRegistry := encodingConfig.InterfaceRegistry
	txConfig := encodingConfig.TxConfig
	appCodec := encodingConfig.Codec

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return client.Context{}, err
	}
	defaultNodeHome := filepath.Join(userHomeDir, "."+app.Name)

	clientCtx := client.Context{}.
		WithCodec(appCodec).
		WithInterfaceRegistry(interfaceRegistry).
		WithTxConfig(txConfig).
		WithHomeDir(defaultNodeHome).
		WithViper("") // Use an empty prefix for viper to avoid conflicts

	// Read the client config from the app's home directory
	clientCtx, err = config.ReadFromClientConfig(clientCtx)
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
