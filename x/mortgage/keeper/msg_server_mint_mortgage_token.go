package keeper

import (
	"context"

	"github.com/ardaglobal/arda-poc/x/mortgage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) MintMortgageToken(goCtx context.Context, msg *types.MsgMintMortgageToken) (*types.MsgMintMortgageTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgMintMortgageTokenResponse{}, nil
}
