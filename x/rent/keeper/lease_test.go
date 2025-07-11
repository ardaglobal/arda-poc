package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/ardaglobal/arda-poc/testutil/keeper"
	"github.com/ardaglobal/arda-poc/testutil/nullify"
	"github.com/ardaglobal/arda-poc/x/rent/keeper"
	"github.com/ardaglobal/arda-poc/x/rent/types"
	"github.com/stretchr/testify/require"
)

func createNLease(keeper keeper.Keeper, ctx context.Context, n int) []types.Lease {
	items := make([]types.Lease, n)
	for i := range items {
		items[i].Id = keeper.AppendLease(ctx, items[i])
	}
	return items
}

func TestLeaseGet(t *testing.T) {
	keeper, ctx := keepertest.RentKeeper(t)
	items := createNLease(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetLease(ctx, item.Id)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&got),
		)
	}
}

func TestLeaseRemove(t *testing.T) {
	keeper, ctx := keepertest.RentKeeper(t)
	items := createNLease(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveLease(ctx, item.Id)
		_, found := keeper.GetLease(ctx, item.Id)
		require.False(t, found)
	}
}

func TestLeaseGetAll(t *testing.T) {
	keeper, ctx := keepertest.RentKeeper(t)
	items := createNLease(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllLease(ctx)),
	)
}

func TestLeaseCount(t *testing.T) {
	keeper, ctx := keepertest.RentKeeper(t)
	items := createNLease(keeper, ctx, 10)
	count := uint64(len(items))
	require.Equal(t, count, keeper.GetLeaseCount(ctx))
}
