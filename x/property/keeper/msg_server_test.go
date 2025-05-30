package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ardaglobal/arda-poc/x/property/keeper"

	keepertest "github.com/ardaglobal/arda-poc/testutil/keeper"

	"github.com/ardaglobal/arda-poc/x/property/types"
)

func setupMsgServer(t testing.TB) (keeper.Keeper, types.MsgServer, context.Context) {
	pk, ctx := keepertest.PropertyKeeper(t)
	ak, _ := keepertest.ArdaKeeper(t)
	bk := keepertest.BankKeeperMock{}
	uk, _ := keepertest.UsdardaKeeper(t)
	return pk, keeper.NewMsgServerImpl(pk, ak, bk, uk), ctx
}

func TestMsgServer(t *testing.T) {
	k, ms, ctx := setupMsgServer(t)
	require.NotNil(t, ms)
	require.NotNil(t, ctx)
	require.NotEmpty(t, k)
}
