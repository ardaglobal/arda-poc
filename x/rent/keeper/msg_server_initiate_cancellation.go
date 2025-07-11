package keeper

import (
	"context"

	"github.com/ardaglobal/arda-poc/x/rent/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) InitiateCancellation(goCtx context.Context, msg *types.MsgInitiateCancellation) (*types.MsgInitiateCancellationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgInitiateCancellationResponse{}, nil
}
