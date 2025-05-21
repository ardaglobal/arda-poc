package keeper

import (
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ardaglobal/arda-poc/x/property/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,

) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    authority,
		logger:       logger,
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetProperty(ctx sdk.Context, id string) (types.Property, bool) {
	kvStore := k.storeService.OpenKVStore(ctx)

	// Use prefixed key for getting property
	propertyKey := types.KeyPrefix(types.KeyPrefixProperty)
	propertyKey = append(propertyKey, []byte(id)...)

	bz, err := kvStore.Get(propertyKey)
	if err != nil {
		return types.Property{}, false
	}

	if bz == nil {
		return types.Property{}, false
	}

	var property types.Property
	k.cdc.MustUnmarshal(bz, &property)
	return property, true
}

func (k Keeper) SetProperty(ctx sdk.Context, property types.Property) {
	kvStore := k.storeService.OpenKVStore(ctx)

	// Use prefixed key for storing property
	propertyKey := types.KeyPrefix(types.KeyPrefixProperty)
	propertyKey = append(propertyKey, []byte(property.Index)...)

	kvStore.Set(propertyKey, k.cdc.MustMarshal(&property))
}

// GetAllProperties returns all properties in the store
func (k Keeper) GetAllProperties(ctx sdk.Context) ([]types.Property, error) {
	kvStore := k.storeService.OpenKVStore(ctx)
	propertyPrefix := types.KeyPrefix(types.KeyPrefixProperty)

	// Get an iterator over all property keys with the prefix
	iterator, err := kvStore.Iterator(propertyPrefix, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get iterator: %w", err)
	}
	defer iterator.Close()

	properties := []types.Property{}

	// Iterate over all keys
	for ; iterator.Valid(); iterator.Next() {
		// Ensure the key has the proper format (prefix + property ID)
		key := iterator.Key()
		if len(key) <= len(propertyPrefix) {
			continue
		}

		value := iterator.Value()
		var property types.Property
		k.cdc.MustUnmarshal(value, &property)
		properties = append(properties, property)
	}

	return properties, nil
}
