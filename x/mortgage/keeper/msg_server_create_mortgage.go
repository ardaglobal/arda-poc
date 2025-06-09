package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ardaglobal/arda-poc/x/mortgage/types"
)

func (k msgServer) CreateMortgage(goCtx context.Context, msg *types.MsgCreateMortgage) (*types.MsgCreateMortgageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	id := msg.Collateral
	if id == "" {
		id = msg.Lendee + msg.Term
	}
	_, found := k.GetMortgage(ctx, id)
	if found {
		return nil, fmt.Errorf("mortgage already exists: %s", id)
	}

	mortgage := types.Mortgage{
		Index:        id,
		Lender:       msg.Lender,
		Lendee:       msg.Lendee,
		Collateral:   msg.Collateral,
		Amount:       msg.Amount,
		InterestRate: msg.InterestRate,
		Term:         msg.Term,
	}
	k.SetMortgage(ctx, mortgage)
	return &types.MsgCreateMortgageResponse{}, nil
}
