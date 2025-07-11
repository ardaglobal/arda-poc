package keeper

import (
	"context"

	"github.com/ardaglobal/arda-poc/x/rent/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) PayRent(goCtx context.Context, msg *types.MsgPayRent) (*types.MsgPayRentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgPayRentResponse{}, nil
}
