package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	propertytypes "github.com/ardaglobal/arda-poc/x/property/types"
)

// PropertyKeeper defines the expected interface for the Property module.
type PropertyKeeper interface {
	// GetProperty retrieves a property by its ID
	GetProperty(ctx sdk.Context, id string) (propertytypes.Property, bool)
	
	// SetProperty stores a property
	SetProperty(ctx sdk.Context, property propertytypes.Property)
	
	// GetAllProperties returns all properties in the store
	GetAllProperties(ctx sdk.Context) ([]propertytypes.Property, error)
	
	// ConvertPropertyOwnersToMap converts the owners and shares slices to a map
	ConvertPropertyOwnersToMap(property propertytypes.Property) map[string]uint64
	
	// UpdatePropertyFromOwnerMap updates property owners and shares from a map
	UpdatePropertyFromOwnerMap(property *propertytypes.Property, ownerMap map[string]uint64)
}

// AccountKeeper defines the expected interface for the Account module.
type AccountKeeper interface {
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI // only used for simulation
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins
	// SendCoins transfers coins from one account to another
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	// GetBalance returns the balance of a specific coin for an account
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	// Methods imported from bank should be defined here
}

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(context.Context, []byte, interface{})
	Set(context.Context, []byte, interface{})
}
