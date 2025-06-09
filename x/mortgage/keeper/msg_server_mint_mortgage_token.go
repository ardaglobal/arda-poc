package keeper

import (
	"context"

	"github.com/ardaglobal/arda-poc/x/mortgage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) MsgMintMortgageToken(goCtx context.Context, msg *types.MsgMintMortgageToken) (*types.MsgMintMortgageTokenResponse, error) {
	return k.MintMortgageToken(goCtx, msg)
}

func (k msgServer) MintMortgageToken(goCtx context.Context, msg *types.MsgMintMortgageToken) (*types.MsgMintMortgageTokenResponse, error) {
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

	// Mint the tokens
	err := k.Mint(ctx, property, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgMintMortgageTokenResponse{}, nil
}
