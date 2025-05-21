package keeper

import (
	"context"

	"github.com/ardaglobal/arda-poc/x/property/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) TransferShares(goCtx context.Context, msg *types.MsgTransferShares) (*types.MsgTransferSharesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgTransferSharesResponse{}, nil
}
