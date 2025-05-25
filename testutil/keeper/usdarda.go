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

	propertykeeper "github.com/ardaglobal/arda-poc/x/property/keeper"
	propertytypes "github.com/ardaglobal/arda-poc/x/property/types"
	usdkeeper "github.com/ardaglobal/arda-poc/x/usdarda/keeper"
	usdtypes "github.com/ardaglobal/arda-poc/x/usdarda/types"
)

type MockBankKeeper struct {
	balances map[string]sdk.Coins
}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{balances: make(map[string]sdk.Coins)}
}

func (m *MockBankKeeper) MintCoins(ctx context.Context, module string, amt sdk.Coins) error {
	m.balances[module] = m.balances[module].Add(amt...)
	return nil
}

func (m *MockBankKeeper) BurnCoins(ctx context.Context, module string, amt sdk.Coins) error {
	m.balances[module] = m.balances[module].Sub(amt)
	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, module string, addr sdk.AccAddress, amt sdk.Coins) error {
	m.balances[module] = m.balances[module].Sub(amt)
	key := addr.String()
	m.balances[key] = m.balances[key].Add(amt...)
	return nil
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, addr sdk.AccAddress, module string, amt sdk.Coins) error {
	key := addr.String()
	m.balances[key] = m.balances[key].Sub(amt)
	m.balances[module] = m.balances[module].Add(amt...)
	return nil
}

func UsdArdaKeeper(t testing.TB) (usdkeeper.Keeper, propertykeeper.Keeper, *MockBankKeeper, sdk.Context) {
	usdKey := storetypes.NewKVStoreKey(usdtypes.StoreKey)
	propertyKey := storetypes.NewKVStoreKey(propertytypes.StoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	ms.MountStoreWithDB(usdKey, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(propertyKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, ms.LoadLatestVersion())

	reg := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(reg)
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	propKeeper := propertykeeper.NewKeeper(cdc, runtime.NewKVStoreService(propertyKey), log.NewNopLogger(), authority.String())
	bankKeeper := NewMockBankKeeper()
	k := usdkeeper.NewKeeper(cdc, runtime.NewKVStoreService(usdKey), log.NewNopLogger(), bankKeeper, propKeeper, authority.String())

	ctx := sdk.NewContext(ms, cmtproto.Header{}, false, log.NewNopLogger())
	require.NoError(t, propKeeper.SetParams(ctx, propertytypes.DefaultParams()))
	require.NoError(t, k.SetParams(ctx, usdtypes.DefaultParams()))

	return k, propKeeper, bankKeeper, ctx
}
