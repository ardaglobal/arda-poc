package keeper

import (
	"context"

	"arda/x/property/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RegisterProperty(goCtx context.Context, msg *types.MsgRegisterProperty) (*types.MsgRegisterPropertyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgRegisterPropertyResponse{}, nil
}
