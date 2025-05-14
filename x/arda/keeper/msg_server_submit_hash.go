package keeper

import (
	"context"

	"arda/x/arda/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) SubmitHash(goCtx context.Context, msg *types.MsgSubmitHash) (*types.MsgSubmitHashResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgSubmitHashResponse{}, nil
}
