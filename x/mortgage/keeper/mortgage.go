package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/ardaglobal/arda-poc/x/mortgage/types"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// SetMortgage set a specific mortgage in the store from its index
func (k Keeper) SetMortgage(ctx context.Context, mortgage types.Mortgage) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.MortgageKeyPrefix))
	b := k.cdc.MustMarshal(&mortgage)
	store.Set(types.MortgageKey(
		mortgage.Index,
	), b)
}

// GetMortgage returns a mortgage from its index
func (k Keeper) GetMortgage(
	ctx context.Context,
	index string,

) (val types.Mortgage, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.MortgageKeyPrefix))

	b := store.Get(types.MortgageKey(
		index,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveMortgage removes a mortgage from the store
func (k Keeper) RemoveMortgage(
	ctx context.Context,
	index string,

) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.MortgageKeyPrefix))
	store.Delete(types.MortgageKey(
		index,
	))
}

// GetAllMortgage returns all mortgage
func (k Keeper) GetAllMortgage(ctx context.Context) (list []types.Mortgage) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.MortgageKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Mortgage
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
