package keeper

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ardaglobal/arda-poc/x/mortgage/types"
)

func (k msgServer) MintMortgageToken(goCtx context.Context, msg *types.MsgMintMortgageToken) (*types.MsgMintMortgageTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	coin := sdk.NewCoin("mortgagetoken", math.NewIntFromUint64(msg.Amount))
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(coin)); err != nil {
		return nil, err
	}
	recipient, err := sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil {
		return nil, err
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, sdk.NewCoins(coin)); err != nil {
		return nil, err
	}
	return &types.MsgMintMortgageTokenResponse{}, nil
}
