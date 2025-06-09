package keeper

import (
	"context"

	"github.com/ardaglobal/arda-poc/x/mortgage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) BurnMortgageToken(goCtx context.Context, msg *types.MsgBurnMortgageToken) (*types.MsgBurnMortgageTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgBurnMortgageTokenResponse{}, nil
}
