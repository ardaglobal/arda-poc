package keeper

import (
	"context"
	"encoding/binary"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/ardaglobal/arda-poc/x/rent/types"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// GetLeaseCount get the total number of lease
func (k Keeper) GetLeaseCount(ctx context.Context) uint64 {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	byteKey := types.KeyPrefix(types.LeaseCountKey)
	bz := store.Get(byteKey)

	// Count doesn't exist: no element
	if bz == nil {
		return 0
	}

	// Parse bytes
	return binary.BigEndian.Uint64(bz)
}

// SetLeaseCount set the total number of lease
func (k Keeper) SetLeaseCount(ctx context.Context, count uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	byteKey := types.KeyPrefix(types.LeaseCountKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(byteKey, bz)
}

// AppendLease appends a lease in the store with a new id and update the count
func (k Keeper) AppendLease(
	ctx context.Context,
	lease types.Lease,
) uint64 {
	// Create the lease
	count := k.GetLeaseCount(ctx)

	// Set the ID of the appended value
	lease.Id = count

	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.LeaseKey))
	appendedValue := k.cdc.MustMarshal(&lease)
	store.Set(GetLeaseIDBytes(lease.Id), appendedValue)

	// Update lease count
	k.SetLeaseCount(ctx, count+1)

	return count
}

// SetLease set a specific lease in the store
func (k Keeper) SetLease(ctx context.Context, lease types.Lease) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.LeaseKey))
	b := k.cdc.MustMarshal(&lease)
	store.Set(GetLeaseIDBytes(lease.Id), b)
}

// GetLease returns a lease from its id
func (k Keeper) GetLease(ctx context.Context, id uint64) (val types.Lease, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.LeaseKey))
	b := store.Get(GetLeaseIDBytes(id))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveLease removes a lease from the store
func (k Keeper) RemoveLease(ctx context.Context, id uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.LeaseKey))
	store.Delete(GetLeaseIDBytes(id))
}

// GetAllLease returns all lease
func (k Keeper) GetAllLease(ctx context.Context) (list []types.Lease) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.LeaseKey))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Lease
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetLeaseIDBytes returns the byte representation of the ID
func GetLeaseIDBytes(id uint64) []byte {
	bz := types.KeyPrefix(types.LeaseKey)
	bz = append(bz, []byte("/")...)
	bz = binary.BigEndian.AppendUint64(bz, id)
	return bz
}
