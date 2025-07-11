package main

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ardaglobal/arda-poc/testutil/network"
	propertytypes "github.com/ardaglobal/arda-poc/x/property/types"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type E2ETestSuite struct {
	suite.Suite
	network *network.Network
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")
	s.network = network.New(s.T(), network.DefaultConfig())
	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

func (s *E2ETestSuite) TestRegisterProperty() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	// Define the property to register
	propertyAddress := "123 Main Street"
	propertyRegion := "e2e-test-region"
	propertyValue := uint64(1000000)
	ownerAddress := val.Address.String()

	// Create the RegisterProperty message
	msg := propertytypes.NewMsgRegisterProperty(
		val.Address.String(),
		propertyAddress,
		propertyRegion,
		propertyValue,
		[]string{ownerAddress},
		[]uint64{100},
	)
	s.Require().NoError(msg.ValidateBasic(), "message validation failed")

	// --- This block replicates the core logic of the sidecar's handler ---

	// Create a new TxFactory
	txf := tx.Factory{}.
		WithChainID(s.network.Config.ChainID).
		WithKeybase(val.ClientCtx.Keyring).
		WithTxConfig(clientCtx.TxConfig).
		WithGas(200000).
		WithFees("2uarda")

	// Get account number and sequence
	authClient := authtypes.NewQueryClient(clientCtx)
	acc, err := authClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: val.Address.String()})
	s.Require().NoError(err)
	
	var accI types.AccountI
	s.Require().NoError(clientCtx.InterfaceRegistry.UnpackAny(acc.Account, &accI))
	
	baseAcc, ok := accI.(*authtypes.BaseAccount)
	s.Require().True(ok, "account is not a BaseAccount")
	
	txf = txf.WithAccountNumber(baseAcc.AccountNumber).WithSequence(baseAcc.Sequence)

	// Build and sign the transaction
	txb, err := txf.BuildUnsignedTx(msg)
	s.Require().NoError(err)
	s.Require().NoError(tx.Sign(context.Background(), txf, val.Moniker, txb, true))

	// Broadcast the transaction
	txBytes, err := clientCtx.TxConfig.TxEncoder()(txb.GetTx())
	s.Require().NoError(err)
	txClient := txtypes.NewServiceClient(clientCtx)
	res, err := txClient.BroadcastTx(
		context.Background(),
		&txtypes.BroadcastTxRequest{
			Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	s.Require().NoError(err)
	s.Require().Equal(uint32(0), res.TxResponse.Code, "transaction failed with code: %d, raw_log: %s", res.TxResponse.Code, res.TxResponse.RawLog)
	// --- End of sidecar logic replication ---

	// Wait for the next block to ensure the transaction is committed
	s.Require().NoError(s.network.WaitForNextBlock())

	// Get the transaction details to verify it was successful
	txResp, err := txClient.GetTx(context.Background(), &txtypes.GetTxRequest{Hash: res.TxResponse.TxHash})
	s.Require().NoError(err)
	s.T().Logf("Transaction committed in block %d with code %d", txResp.TxResponse.Height, txResp.TxResponse.Code)
	s.Require().Equal(uint32(0), txResp.TxResponse.Code, "transaction failed in block with code: %d, raw_log: %s", txResp.TxResponse.Code, txResp.TxResponse.RawLog)

	// Verify the property was created
	queryClient := propertytypes.NewQueryClient(clientCtx)
	expectedIndex := strings.ToLower(strings.TrimSpace(propertyAddress))
	s.T().Logf("Querying for property with index: %s", expectedIndex)
	
	queryResp, err := queryClient.Property(context.Background(), &propertytypes.QueryGetPropertyRequest{
		Index: expectedIndex,
	})
	s.Require().NoError(err)
	s.Require().Equal(propertyAddress, queryResp.Property.Address)
	s.Require().Equal(propertyRegion, queryResp.Property.Region)
	s.Require().Equal(propertyValue, queryResp.Property.Value)

	s.T().Log("Property registration test completed successfully")
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
