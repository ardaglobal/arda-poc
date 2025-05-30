package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ardaglobal/arda-poc/testutil/keeper"
	"github.com/ardaglobal/arda-poc/testutil/sample"
	propertykeeper "github.com/ardaglobal/arda-poc/x/property/keeper"
	"github.com/ardaglobal/arda-poc/x/property/types"
)

func setupMsgServerRegister(t testing.TB) (propertykeeper.Keeper, types.MsgServer, context.Context) {
	pk, ctx := keeper.PropertyKeeper(t)
	ak, _ := keeper.ArdaKeeper(t)
	bk := keeper.BankKeeperMock{}
	uk, _ := keeper.UsdardaKeeper(t)
	return pk, propertykeeper.NewMsgServerImpl(pk, ak, bk, uk), ctx
}

func TestRegisterPropertyLengthMismatch(t *testing.T) {
	_, ms, ctx := setupMsgServerRegister(t)

	msg := &types.MsgRegisterProperty{
		Creator: sample.AccAddress(),
		Address: "addr1",
		Region:  "dubai",
		Value:   100,
		Owners:  []string{"owner1", "owner2"},
		Shares:  []uint64{100},
	}

	_, err := ms.RegisterProperty(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "owners and shares length mismatch")
}
