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

		LeaseList: []types.Lease{
			{
				Id: 0,
			},
			{
				Id: 1,
			},
		},
		LeaseCount: 2,
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.RentKeeper(t)
	rent.InitGenesis(ctx, k, genesisState)
	got := rent.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.LeaseList, got.LeaseList)
	require.Equal(t, genesisState.LeaseCount, got.LeaseCount)
	// this line is used by starport scaffolding # genesis/test/assert
}
