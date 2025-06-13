package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ardaglobal/arda-poc/x/rent/types"
)

func (k msgServer) PayRent(goCtx context.Context, msg *types.MsgPayRent) (*types.MsgPayRentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tenant, err := sdk.AccAddressFromBech32(msg.Tenant)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid tenant address (%s)", err)
	}

	coins := sdk.NewCoins(sdk.NewInt64Coin("usdarda", int64(msg.Amount)))

	if err := k.Keeper.PayRent(ctx, tenant, msg.LeaseId, coins); err != nil {
		return nil, err
	}

	return &types.MsgPayRentResponse{}, nil
}
