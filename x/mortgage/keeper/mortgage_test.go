package keeper_test

import (
	"context"
	"strconv"
	"testing"

	keepertest "github.com/ardaglobal/arda-poc/testutil/keeper"
	"github.com/ardaglobal/arda-poc/testutil/nullify"
	"github.com/ardaglobal/arda-poc/x/mortgage/keeper"
	"github.com/ardaglobal/arda-poc/x/mortgage/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNMortgage(keeper keeper.Keeper, ctx context.Context, n int) []types.Mortgage {
	items := make([]types.Mortgage, n)
	for i := range items {
		items[i].Index = strconv.Itoa(i)

		keeper.SetMortgage(ctx, items[i])
	}
	return items
}

func TestMortgageGet(t *testing.T) {
	keeper, ctx := keepertest.MortgageKeeper(t)
	items := createNMortgage(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetMortgage(ctx,
			item.Index,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestMortgageRemove(t *testing.T) {
	keeper, ctx := keepertest.MortgageKeeper(t)
	items := createNMortgage(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveMortgage(ctx,
			item.Index,
		)
		_, found := keeper.GetMortgage(ctx,
			item.Index,
		)
		require.False(t, found)
	}
}

func TestMortgageGetAll(t *testing.T) {
	keeper, ctx := keepertest.MortgageKeeper(t)
	items := createNMortgage(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllMortgage(ctx)),
	)
}
