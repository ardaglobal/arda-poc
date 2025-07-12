package keeper

import (
	"context"

	"github.com/ardaglobal/arda-poc/x/rent/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) ApproveCancellation(goCtx context.Context, msg *types.MsgApproveCancellation) (*types.MsgApproveCancellationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgApproveCancellationResponse{}, nil
}
