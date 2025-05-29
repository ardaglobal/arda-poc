package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/ardaglobal/arda-poc/testutil/keeper"
	propertytypes "github.com/ardaglobal/arda-poc/x/property/types"
)

func setupProperty(t *testing.T, pk propertytypes.PropertyKeeper, ctx sdk.Context) propertytypes.Property {
	prop := propertytypes.Property{
		Index:   "prop1",
		Address: "addr",
		Region:  "region",
		Value:   1000,
		Owners:  []string{"owner"},
		Shares:  []uint64{100},
	}
	pk.SetProperty(ctx, prop)
	return prop
}

func TestMintBurnFlows(t *testing.T) {
	k, propK, _, ctx := keepertest.UsdArdaKeeper(t)
	prop := setupProperty(t, propK, ctx)
	addr := sdk.AccAddress([]byte("addr1____________"))

	// mint 80% at once
	err := k.Mint(ctx, prop.Index, prop.Value*80/100, addr)
	require.NoError(t, err)
	rec, found := k.GetRecord(ctx, prop.Index)
	require.True(t, found)
	require.Equal(t, uint64(800), rec.Minted)

	// burn all to unlock
	err = k.Burn(ctx, prop.Index, 800, addr)
	require.NoError(t, err)
	_, found = k.GetRecord(ctx, prop.Index)
	require.False(t, found)

	// multiple mints
	err = k.Mint(ctx, prop.Index, 300, addr)
	require.NoError(t, err)
	err = k.Mint(ctx, prop.Index, 500, addr)
	require.NoError(t, err)
	rec, _ = k.GetRecord(ctx, prop.Index)
	require.Equal(t, uint64(800), rec.Minted)

	// multiple burns
	err = k.Burn(ctx, prop.Index, 200, addr)
	require.NoError(t, err)
	err = k.Burn(ctx, prop.Index, 600, addr)
	require.NoError(t, err)
	_, found = k.GetRecord(ctx, prop.Index)
	require.False(t, found)
}
