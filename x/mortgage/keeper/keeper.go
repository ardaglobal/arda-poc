package keeper

import (
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ardaglobal/arda-poc/x/mortgage/types"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	logger       log.Logger

	bankKeeper types.BankKeeper
	authority  string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	bankKeeper types.BankKeeper,
	authority string,
) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}
	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		logger:       logger,
		bankKeeper:   bankKeeper,
		authority:    authority,
	}
}

func (k Keeper) GetAuthority() string { return k.authority }

func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) SetMortgage(ctx sdk.Context, m types.Mortgage) {
	store := k.storeService.OpenKVStore(ctx)
	key := append([]byte(types.KeyPrefixMortgage), []byte(m.Index)...)
	store.Set(key, k.cdc.MustMarshal(&m))
}

func (k Keeper) GetMortgage(ctx sdk.Context, id string) (types.Mortgage, bool) {
	store := k.storeService.OpenKVStore(ctx)
	key := append([]byte(types.KeyPrefixMortgage), []byte(id)...)
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return types.Mortgage{}, false
	}
	var m types.Mortgage
	k.cdc.MustUnmarshal(bz, &m)
	return m, true
}

func (k Keeper) GetAllMortgages(ctx sdk.Context) ([]types.Mortgage, error) {
	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.KeyPrefixMortgage)
	iter, err := store.Iterator(prefix, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	res := []types.Mortgage{}
	for ; iter.Valid(); iter.Next() {
		var m types.Mortgage
		k.cdc.MustUnmarshal(iter.Value(), &m)
		res = append(res, m)
	}
	return res, nil
}
