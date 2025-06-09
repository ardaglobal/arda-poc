package keeper

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ardaglobal/arda-poc/x/mortgage/types"
)

func (k msgServer) BurnMortgageToken(goCtx context.Context, msg *types.MsgBurnMortgageToken) (*types.MsgBurnMortgageTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	coin := sdk.NewCoin("usdarda", math.NewIntFromUint64(msg.Amount))
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, sdk.NewCoins(coin)); err != nil {
		return nil, err
	}
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(coin)); err != nil {
		return nil, err
	}
	return &types.MsgBurnMortgageTokenResponse{}, nil
}
