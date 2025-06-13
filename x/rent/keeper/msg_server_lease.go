package keeper

import (
	"context"

	"github.com/ardaglobal/arda-poc/x/rent/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CreateLease(goCtx context.Context, msg *types.MsgCreateLease) (*types.MsgCreateLeaseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lease := types.Lease{
		Id:          msg.PropertyId + msg.Tenant, // simplistic ID generation
		PropertyId:  msg.PropertyId,
		Tenant:      msg.Tenant,
		RentAmount:  msg.RentAmount,
		RentDueDate: msg.RentDueDate,
		Status:      msg.Status,
	}
	k.Keeper.SetLease(ctx, lease)
	return &types.MsgCreateLeaseResponse{}, nil
}
