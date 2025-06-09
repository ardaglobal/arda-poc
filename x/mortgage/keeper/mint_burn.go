package keeper

import (
	"encoding/json"

	"github.com/ardaglobal/arda-poc/x/mortgage/types"
	propertytypes "github.com/ardaglobal/arda-poc/x/property/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetMortgageMintInfo retrieves mint info for a mortgage
func (k Keeper) GetMortgageMintInfo(ctx sdk.Context, mortgageId string) (types.MortgageMintInfo, bool) {
	store := k.storeService.OpenKVStore(ctx)
	key := append([]byte(types.MintInfoKeyPrefix), []byte(mortgageId)...)
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return types.MortgageMintInfo{}, false
	}
	var info types.MortgageMintInfo
	if err := json.Unmarshal(bz, &info); err != nil {
		panic(err)
	}
	return info, true
}

// setMortgageMintInfo stores mint info
func (k Keeper) setMortgageMintInfo(ctx sdk.Context, info types.MortgageMintInfo) {
	store := k.storeService.OpenKVStore(ctx)
	key := append([]byte(types.MintInfoKeyPrefix), []byte(info.MortgageId)...)
	bz, err := json.Marshal(&info)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
}

// deleteMortgageMintInfo removes mint info
func (k Keeper) deleteMortgageMintInfo(ctx sdk.Context, mortgageId string) {
	store := k.storeService.OpenKVStore(ctx)
	key := append([]byte(types.MintInfoKeyPrefix), []byte(mortgageId)...)
	store.Delete(key)
}

// Mint mints usdarda for the given property and amount, distributing to owners by share
func (k Keeper) Mint(ctx sdk.Context, property propertytypes.Property, amount uint64) error {
	return k.usdardaKeeper.Mint(ctx, property, amount)
}

// Burn burns usdarda from an account for a property
func (k Keeper) Burn(ctx sdk.Context, property propertytypes.Property, burner sdk.AccAddress, amount uint64) error {
	return k.usdardaKeeper.Burn(ctx, property, burner, amount)
} 