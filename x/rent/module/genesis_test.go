package rent_test

import (
	"testing"

	keepertest "github.com/ardaglobal/arda-poc/testutil/keeper"
	"github.com/ardaglobal/arda-poc/testutil/nullify"
	rent "github.com/ardaglobal/arda-poc/x/rent/module"
	"github.com/ardaglobal/arda-poc/x/rent/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.RentKeeper(t)
	rent.InitGenesis(ctx, k, genesisState)
	got := rent.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
