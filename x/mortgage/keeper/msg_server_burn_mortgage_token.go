package keeper

import (
	"context"

	"github.com/ardaglobal/arda-poc/x/mortgage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) MsgBurnMortgageToken(goCtx context.Context, msg *types.MsgBurnMortgageToken) (*types.MsgBurnMortgageTokenResponse, error) {
	return k.BurnMortgageToken(goCtx, msg)
}

func (k msgServer) BurnMortgageToken(goCtx context.Context, msg *types.MsgBurnMortgageToken) (*types.MsgBurnMortgageTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the mortgage
	mortgage, found := k.GetMortgage(ctx, msg.MortgageId)
	if !found {
		return nil, types.ErrMortgageNotFound
	}

	// Get the collateral property
	property, found := k.propertyKeeper.GetProperty(ctx, mortgage.Collateral)
	if !found {
		return nil, types.ErrPropertyNotFound
	}

	// Get the burner address
	burner, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	// Burn the tokens
	err = k.Burn(ctx, property, burner, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgBurnMortgageTokenResponse{}, nil
}
