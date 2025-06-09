package types

import (
	"context"

	"github.com/ardaglobal/arda-poc/x/property/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected interface for the Account module.
type AccountKeeper interface {
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI // only used for simulation
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

// PropertyKeeper defines the expected interface for the Property module.
type PropertyKeeper interface {
	GetProperty(ctx sdk.Context, id string) (types.Property, bool)
}

// USDArdaKeeper defines the expected interface for the USDArda module.
type USDArdaKeeper interface {
	Mint(ctx sdk.Context, property types.Property, amount uint64) error
	Burn(ctx sdk.Context, property types.Property, burner sdk.AccAddress, amount uint64) error
}

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(context.Context, []byte, interface{})
	Set(context.Context, []byte, interface{})
}
