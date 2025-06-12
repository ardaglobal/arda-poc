package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/ardaglobal/arda-poc/x/mortgage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) RepayMortgage(goCtx context.Context, msg *types.MsgRepayMortgage) (*types.MsgRepayMortgageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	mortgage, isFound := k.GetMortgage(ctx, msg.MortgageId)
	if !isFound {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "mortgage not found")
	}

	if mortgage.Status != types.APPROVED {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "mortgage is not active")
	}

	lender, err := sdk.AccAddressFromBech32(mortgage.Lender)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid lender address (%s)", err)
	}

	lendee, err := sdk.AccAddressFromBech32(mortgage.Lendee)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid lendee address (%s)", err)
	}

	if msg.Creator != mortgage.Lendee {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "repayment can only be made by the lendee")
	}

	if msg.Amount > mortgage.OutstandingAmount {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "repayment amount exceeds outstanding balance")
	}

	denom := "usdarda"
	coins := sdk.NewCoins(sdk.NewInt64Coin(denom, int64(msg.Amount)))
	if err := k.bankKeeper.SendCoins(ctx, lendee, lender, coins); err != nil {
		return nil, errorsmod.Wrap(err, "failed to send funds from lendee to lender")
	}

	mortgage.OutstandingAmount -= msg.Amount

	if mortgage.OutstandingAmount == 0 {
		mortgage.Status = types.PAID

		markerDenom := types.MortgageMarkerDenom(mortgage.Collateral, mortgage.Index)
		markerCoin := sdk.NewCoins(sdk.NewInt64Coin(markerDenom, 1))
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, lendee, types.ModuleName, markerCoin); err != nil {
			return nil, errorsmod.Wrap(err, "failed to send marker token from lendee to module")
		}
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, markerCoin); err != nil {
			return nil, errorsmod.Wrap(err, "failed to burn marker token")
		}
	}

	k.SetMortgage(ctx, mortgage)

	return &types.MsgRepayMortgageResponse{}, nil
}
