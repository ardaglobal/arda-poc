package keeper

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/ardaglobal/arda-poc/x/property/keeper"
	"github.com/ardaglobal/arda-poc/x/property/types"
)

func PropertyKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	bk := BankKeeperMock{}
	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		log.NewNopLogger(),
		bk,
		authority.String(),
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	if err := k.SetParams(ctx, types.DefaultParams()); err != nil {
		panic(err)
	}

	return k, ctx
}

// BankKeeperMock implements types.BankKeeper for tests.
// TODO: how can we remove this mock? What is the right way to handle this?
type BankKeeperMock struct{}

func (BankKeeperMock) SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	return sdk.Coins{}
}
func (BankKeeperMock) MintCoins(ctx context.Context, module string, amt sdk.Coins) error { return nil }
func (BankKeeperMock) SendCoinsFromModuleToAccount(ctx context.Context, sender string, recipient sdk.AccAddress, amt sdk.Coins) error {
	return nil
}
func (BankKeeperMock) SendCoinsFromAccountToModule(ctx context.Context, sender sdk.AccAddress, recipient string, amt sdk.Coins) error {
	return nil
}
func (BankKeeperMock) SendCoins(ctx context.Context, from sdk.AccAddress, to sdk.AccAddress, amt sdk.Coins) error {
	return nil
}
