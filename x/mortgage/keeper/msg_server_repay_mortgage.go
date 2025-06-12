package keeper

import (
	"context"

	"github.com/ardaglobal/arda-poc/x/mortgage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RepayMortgage(goCtx context.Context, msg *types.MsgRepayMortgage) (*types.MsgRepayMortgageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgRepayMortgageResponse{}, nil
}
