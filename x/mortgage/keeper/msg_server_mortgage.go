package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/ardaglobal/arda-poc/x/mortgage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CreateMortgage(goCtx context.Context, msg *types.MsgCreateMortgage) (*types.MsgCreateMortgageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, isFound := k.GetMortgage(
		ctx,
		msg.Index,
	)
	if isFound {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "index already set")
	}

	var mortgage = types.Mortgage{
		Creator:           msg.Creator,
		Index:             msg.Index,
		Lender:            msg.Lender,
		Lendee:            msg.Lendee,
		Collateral:        msg.Collateral,
		Amount:            msg.Amount,
		InterestRate:      msg.InterestRate,
		Term:              msg.Term,
		Status:            types.APPROVED,
		OutstandingAmount: msg.Amount,
	}

	lender, err := sdk.AccAddressFromBech32(mortgage.Lender)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid lender address (%s)", err)
	}

	lendee, err := sdk.AccAddressFromBech32(mortgage.Lendee)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid lendee address (%s)", err)
	}

	if msg.Creator != mortgage.Lender {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "mortgage can only be created by the lender")
	}

	denom := "usdarda"
	coins := sdk.NewCoins(sdk.NewInt64Coin(denom, int64(mortgage.Amount)))
	if err := k.bankKeeper.SendCoins(ctx, lender, lendee, coins); err != nil {
		return nil, errorsmod.Wrap(err, "failed to send funds from lender to lendee")
	}

	markerDenom := fmt.Sprintf("mortgage/%s/%s", mortgage.Collateral, mortgage.Index)
	markerCoin := sdk.NewCoins(sdk.NewInt64Coin(markerDenom, 1))
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, markerCoin); err != nil {
		return nil, errorsmod.Wrap(err, "failed to mint marker token")
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lendee, markerCoin); err != nil {
		return nil, errorsmod.Wrap(err, "failed to send marker token to lendee")
	}

	k.SetMortgage(
		ctx,
		mortgage,
	)

	return &types.MsgCreateMortgageResponse{}, nil
}

func (k msgServer) UpdateMortgage(goCtx context.Context, msg *types.MsgUpdateMortgage) (*types.MsgUpdateMortgageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valFound, isFound := k.GetMortgage(
		ctx,
		msg.Index,
	)
	if !isFound {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "index not set")
	}

	if msg.Creator != valFound.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	var mortgage = types.Mortgage{
		Creator:      msg.Creator,
		Index:        msg.Index,
		Lender:       msg.Lender,
		Lendee:       msg.Lendee,
		Collateral:   msg.Collateral,
		Amount:       msg.Amount,
		InterestRate: msg.InterestRate,
		Term:         msg.Term,
	}

	k.SetMortgage(ctx, mortgage)

	return &types.MsgUpdateMortgageResponse{}, nil
}

func (k msgServer) DeleteMortgage(goCtx context.Context, msg *types.MsgDeleteMortgage) (*types.MsgDeleteMortgageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valFound, isFound := k.GetMortgage(
		ctx,
		msg.Index,
	)
	if !isFound {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "index not set")
	}

	if msg.Creator != valFound.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	k.RemoveMortgage(
		ctx,
		msg.Index,
	)

	return &types.MsgDeleteMortgageResponse{}, nil
}
