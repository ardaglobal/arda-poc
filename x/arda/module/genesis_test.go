package arda_test

import (
	"testing"

	arda "github.com/ardaglobal/arda-poc/x/arda/module"
	"github.com/ardaglobal/arda-poc/x/arda/types"

	keepertest "github.com/ardaglobal/arda-poc/testutil/keeper"
	"github.com/ardaglobal/arda-poc/testutil/nullify"

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
