package usdarda_test

import (
	"testing"

	keepertest "github.com/ardaglobal/arda-poc/testutil/keeper"
	"github.com/ardaglobal/arda-poc/testutil/nullify"
	usdarda "github.com/ardaglobal/arda-poc/x/usdarda/module"
	"github.com/ardaglobal/arda-poc/x/usdarda/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.UsdardaKeeper(t)
	usdarda.InitGenesis(ctx, k, genesisState)
	got := usdarda.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
