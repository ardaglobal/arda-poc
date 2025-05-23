package types

import (
	"context"

	propertytypes "github.com/ardaglobal/arda-poc/x/property/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AccountKeeper interface {
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
}

type BankKeeper interface {
	MintCoins(ctx context.Context, module string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, module string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

type PropertyKeeper interface {
	GetProperty(ctx sdk.Context, id string) (propertytypes.Property, bool)
}
