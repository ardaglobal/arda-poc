package keeper

import (
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	propertytypes "github.com/ardaglobal/arda-poc/x/property/types"
	"github.com/ardaglobal/arda-poc/x/rent/types"
)

type PropertyKeeper interface {
	GetProperty(ctx sdk.Context, id string) (propertytypes.Property, bool)
}

type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, from sdk.AccAddress, toModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, fromModule string, to sdk.AccAddress, amt sdk.Coins) error
}

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	logger       log.Logger

	bankKeeper     BankKeeper
	propertyKeeper PropertyKeeper
	authority      string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	bankKeeper BankKeeper,
	propertyKeeper PropertyKeeper,
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

func (k Keeper) GetLease(ctx sdk.Context, id string) (types.Lease, bool) {
	store := k.storeService.OpenKVStore(ctx)
	key := append(types.KeyPrefix(types.KeyPrefixLease), []byte(id)...)
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return types.Lease{}, false
	}
	var lease types.Lease
	k.cdc.MustUnmarshal(bz, &lease)
	return lease, true
}

func (k Keeper) SetLease(ctx sdk.Context, lease types.Lease) {
	store := k.storeService.OpenKVStore(ctx)
	key := append(types.KeyPrefix(types.KeyPrefixLease), []byte(lease.Id)...)
	store.Set(key, k.cdc.MustMarshal(&lease))
}

func (k Keeper) PayRent(ctx sdk.Context, tenant sdk.AccAddress, leaseId string, amount sdk.Coins) error {
	lease, found := k.GetLease(ctx, leaseId)
	if !found {
		return fmt.Errorf("lease %s not found", leaseId)
	}
	if tenant.String() != lease.Tenant {
		return fmt.Errorf("only tenant may pay rent")
	}
	// send coins from tenant to module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, tenant, types.ModuleName, amount); err != nil {
		return err
	}
	property, found := k.propertyKeeper.GetProperty(ctx, lease.PropertyId)
	if !found {
		return fmt.Errorf("property %s not found", lease.PropertyId)
	}
	var totalShares uint64
	for _, s := range property.Shares {
		totalShares += s
	}
	for i, owner := range property.Owners {
		share := property.Shares[i]
		ownerAcc, err := sdk.AccAddressFromBech32(owner)
		if err != nil {
			return err
		}
		portion := sdk.NewDecFromInt(amount.AmountOf(amount.GetDenomByIndex(0))).MulInt64(int64(share)).QuoInt64(int64(totalShares)).TruncateInt()
		coins := sdk.NewCoins(sdk.NewCoin(amount.GetDenomByIndex(0), portion))
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, ownerAcc, coins); err != nil {
			return err
		}
	}
	return nil
}
