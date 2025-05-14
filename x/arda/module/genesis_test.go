package arda_test

import (
	"testing"

	keepertest "arda/testutil/keeper"
	"arda/testutil/nullify"
	arda "arda/x/arda/module"
	"arda/x/arda/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.ArdaKeeper(t)
	arda.InitGenesis(ctx, k, genesisState)
	got := arda.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
