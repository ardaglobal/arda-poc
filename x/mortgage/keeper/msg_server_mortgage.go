package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/ardaglobal/arda-poc/x/mortgage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CreateMortgage(goCtx context.Context, msg *types.MsgCreateMortgage) (*types.MsgCreateMortgageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value already exists
	_, isFound := k.GetMortgage(
		ctx,
		msg.Index,
	)
	if isFound {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "index already set")
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

	k.SetMortgage(
		ctx,
		mortgage,
	)
	return &types.MsgCreateMortgageResponse{}, nil
}

func (k msgServer) UpdateMortgage(goCtx context.Context, msg *types.MsgUpdateMortgage) (*types.MsgUpdateMortgageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	valFound, isFound := k.GetMortgage(
		ctx,
		msg.Index,
	)
	if !isFound {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "index not set")
	}

	// Checks if the msg creator is the same as the current owner
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

	// Check if the value exists
	valFound, isFound := k.GetMortgage(
		ctx,
		msg.Index,
	)
	if !isFound {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "index not set")
	}

	// Checks if the msg creator is the same as the current owner
	if msg.Creator != valFound.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	k.RemoveMortgage(
		ctx,
		msg.Index,
	)

	return &types.MsgDeleteMortgageResponse{}, nil
}
