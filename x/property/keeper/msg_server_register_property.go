package keeper

import (
	"context"
	"fmt"
	"strings"

	"github.com/ardaglobal/arda-poc/x/property/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RegisterProperty(goCtx context.Context, msg *types.MsgRegisterProperty) (*types.MsgRegisterPropertyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Use address as deterministic property ID
	id := strings.ToLower(strings.TrimSpace(msg.Address))

	// Prevent duplicate registration
	if _, found := k.GetProperty(ctx, id); found {
		return nil, fmt.Errorf("property already exists: %s", id)
	}

	// Validate ownership sum = 100
	var total uint64
	for _, share := range msg.Owners {
		total += share
	}
	if total != 100 {
		return nil, fmt.Errorf("ownership shares must total 100, got %d", total)
	}

	// Create and store property
	property := types.Property{
		Index:   id,
		Address: msg.Address,
		Region:  msg.Region,
		Value:   msg.Value,
		Owners:  msg.Owners,
	}
	k.SetProperty(ctx, property)

	return &types.MsgRegisterPropertyResponse{}, nil
}
