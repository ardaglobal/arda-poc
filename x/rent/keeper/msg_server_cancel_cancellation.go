package keeper

import (
	"context"

	"github.com/ardaglobal/arda-poc/x/rent/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CancelCancellation(goCtx context.Context, msg *types.MsgCancelCancellation) (*types.MsgCancelCancellationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgCancelCancellationResponse{}, nil
}
