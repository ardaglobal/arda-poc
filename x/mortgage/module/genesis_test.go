package mortgage_test

import (
	"testing"

	keepertest "github.com/ardaglobal/arda-poc/testutil/keeper"
	"github.com/ardaglobal/arda-poc/testutil/nullify"
	mortgage "github.com/ardaglobal/arda-poc/x/mortgage/module"
	"github.com/ardaglobal/arda-poc/x/mortgage/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		MortgageList: []types.Mortgage{
			{
				Index: "0",
			},
			{
				Index: "1",
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.MortgageKeeper(t)
	mortgage.InitGenesis(ctx, k, genesisState)
	got := mortgage.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.MortgageList, got.MortgageList)
	// this line is used by starport scaffolding # genesis/test/assert
}
