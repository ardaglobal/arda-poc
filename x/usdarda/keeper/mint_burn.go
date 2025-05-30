package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	propertytypes "github.com/ardaglobal/arda-poc/x/property/types"
	usdtypes "github.com/ardaglobal/arda-poc/x/usdarda/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetMintInfo retrieves mint info for a property
func (k Keeper) GetMintInfo(ctx sdk.Context, propertyId string) (usdtypes.MintInfo, bool) {
	store := k.storeService.OpenKVStore(ctx)
	key := append([]byte(usdtypes.MintInfoKeyPrefix), []byte(propertyId)...)
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return usdtypes.MintInfo{}, false
	}
	var info usdtypes.MintInfo
	if err := json.Unmarshal(bz, &info); err != nil {
		panic(err)
	}
	return info, true
}

// setMintInfo stores mint info
func (k Keeper) setMintInfo(ctx sdk.Context, info usdtypes.MintInfo) {
	store := k.storeService.OpenKVStore(ctx)
	key := append([]byte(usdtypes.MintInfoKeyPrefix), []byte(info.PropertyId)...)
	bz, err := json.Marshal(&info)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
}

// deleteMintInfo removes mint info
func (k Keeper) deleteMintInfo(ctx sdk.Context, propertyId string) {
	store := k.storeService.OpenKVStore(ctx)
	key := append([]byte(usdtypes.MintInfoKeyPrefix), []byte(propertyId)...)
	store.Delete(key)
}

// Mint mints usdarda for the given property and amount, distributing to owners by share
func (k Keeper) Mint(ctx sdk.Context, property propertytypes.Property, amount uint64) error {
	info, _ := k.GetMintInfo(ctx, property.Index)
	if info.Minted-info.Burned+amount > property.Value {
		return fmt.Errorf("mint exceeds allowed limit")
	}
	denom := usdtypes.USDArdaDenom
	for i, owner := range property.Owners {
		if i >= len(property.Shares) {
			break
		}
		share := property.Shares[i]
		minted := amount * share / 100
		if minted == 0 {
			continue
		}
		addr, err := sdk.AccAddressFromBech32(owner)
		if err != nil {
			return err
		}
		coin := sdk.NewCoin(denom, math.NewInt(int64(minted)))
		if err := k.bankKeeper.MintCoins(ctx, usdtypes.ModuleName, sdk.NewCoins(coin)); err != nil {
			return err
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, usdtypes.ModuleName, addr, sdk.NewCoins(coin)); err != nil {
			return err
		}
	}
	info.PropertyId = property.Index
	info.Minted += amount
	k.setMintInfo(ctx, info)
	return nil
}

// MintToAddress mints amount of usdarda to a single address
func (k Keeper) MintToAddress(ctx sdk.Context, addr sdk.AccAddress, amount uint64) error {
	if amount == 0 {
		return nil
	}
	denom := usdtypes.USDArdaDenom
	coin := sdk.NewCoin(denom, math.NewInt(int64(amount)))
	if err := k.bankKeeper.MintCoins(ctx, usdtypes.ModuleName, sdk.NewCoins(coin)); err != nil {
		return err
	}
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, usdtypes.ModuleName, addr, sdk.NewCoins(coin))
}

// BurnFromAddress burns amount of usdarda from an address
func (k Keeper) BurnFromAddress(ctx sdk.Context, addr sdk.AccAddress, amount uint64) error {
	if amount == 0 {
		return nil
	}
	denom := usdtypes.USDArdaDenom
	coin := sdk.NewCoin(denom, math.NewInt(int64(amount)))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, usdtypes.ModuleName, sdk.NewCoins(coin)); err != nil {
		return err
	}
	return k.bankKeeper.BurnCoins(ctx, usdtypes.ModuleName, sdk.NewCoins(coin))
}

// Burn burns usdarda from an account for a property
func (k Keeper) Burn(ctx sdk.Context, property propertytypes.Property, burner sdk.AccAddress, amount uint64) error {
	info, found := k.GetMintInfo(ctx, property.Index)
	if !found || info.Minted-info.Burned < amount {
		return fmt.Errorf("insufficient minted amount")
	}
	denom := usdtypes.USDArdaDenom
	coin := sdk.NewCoin(denom, math.NewInt(int64(amount)))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, burner, usdtypes.ModuleName, sdk.NewCoins(coin)); err != nil {
		return err
	}
	if err := k.bankKeeper.BurnCoins(ctx, usdtypes.ModuleName, sdk.NewCoins(coin)); err != nil {
		return err
	}
	info.Burned += amount
	if info.Burned >= info.Minted {
		k.deleteMintInfo(ctx, property.Index)
	} else {
		k.setMintInfo(ctx, info)
	}
	return nil
}
