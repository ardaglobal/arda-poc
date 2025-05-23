package keeper

import (
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ardaglobal/arda-poc/x/usdarda/types"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	logger       log.Logger

	bankKeeper     types.BankKeeper
	propertyKeeper types.PropertyKeeper

	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	bankKeeper types.BankKeeper,
	propertyKeeper types.PropertyKeeper,
	authority string,
) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return Keeper{
		cdc:            cdc,
		storeService:   storeService,
		logger:         logger,
		bankKeeper:     bankKeeper,
		propertyKeeper: propertyKeeper,
		authority:      authority,
	}
}

func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if err == nil && bz != nil {
		k.cdc.MustUnmarshal(bz, &params)
	}
	return params
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	return store.Set(types.ParamsKey, bz)
}

func (k Keeper) storeKey(propertyID string) []byte {
	return append(types.KeyPrefix(types.KeyPrefixUsdArda), []byte(propertyID)...)
}

func (k Keeper) GetRecord(ctx sdk.Context, propertyID string) (types.UsdArdaRecord, bool) {
	kv := k.storeService.OpenKVStore(ctx)
	bz, err := kv.Get(k.storeKey(propertyID))
	if err != nil || bz == nil {
		return types.UsdArdaRecord{}, false
	}
	var rec types.UsdArdaRecord
	k.cdc.MustUnmarshal(bz, &rec)
	return rec, true
}

func (k Keeper) setRecord(ctx sdk.Context, rec types.UsdArdaRecord) {
	kv := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&rec)
	kv.Set(k.storeKey(rec.PropertyId), bz)
}

func (k Keeper) deleteRecord(ctx sdk.Context, propertyID string) {
	kv := k.storeService.OpenKVStore(ctx)
	kv.Delete(k.storeKey(propertyID))
}

func (k Keeper) Mint(ctx sdk.Context, propertyID string, amount uint64, addr sdk.AccAddress) error {
	property, found := k.propertyKeeper.GetProperty(ctx, propertyID)
	if !found {
		return fmt.Errorf("property %s not found", propertyID)
	}

	rec, _ := k.GetRecord(ctx, propertyID)
	max := property.Value * 80 / 100
	if rec.Minted+amount > max {
		return fmt.Errorf("mint exceeds 80%% of property value")
	}

	coin := sdk.NewCoin(types.ModuleName, sdk.NewIntFromUint64(amount))
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(coin)); err != nil {
		return err
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(coin)); err != nil {
		return err
	}

	rec.PropertyId = propertyID
	rec.Minted += amount
	k.setRecord(ctx, rec)
	return nil
}

func (k Keeper) Burn(ctx sdk.Context, propertyID string, amount uint64, addr sdk.AccAddress) error {
	rec, found := k.GetRecord(ctx, propertyID)
	if !found || rec.Minted < amount {
		return fmt.Errorf("insufficient minted amount")
	}

	coin := sdk.NewCoin(types.ModuleName, sdk.NewIntFromUint64(amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, sdk.NewCoins(coin)); err != nil {
		return err
	}
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(coin)); err != nil {
		return err
	}

	rec.Minted -= amount
	if rec.Minted == 0 {
		k.deleteRecord(ctx, propertyID)
	} else {
		k.setRecord(ctx, rec)
	}
	return nil
}
