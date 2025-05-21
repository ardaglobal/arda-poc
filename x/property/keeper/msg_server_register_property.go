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
	if p, found := k.GetProperty(ctx, id); found {
		return nil, fmt.Errorf("property already exists: %s: %v", id, p)
	}

	if len(msg.Owners) != len(msg.Shares) {
		return nil, fmt.Errorf("owners and shares length mismatch")
	}

	var total uint64
	for i, share := range msg.Shares {
		if share == 0 {
			continue
		}
		total += share
		if total < share { // overflow check
			return nil, fmt.Errorf("ownership share overflow")
		}
		if i >= len(msg.Owners) {
			break
		}
	}

	// if total != 100 {
	// 	return nil, fmt.Errorf("ownership shares must total 100, got %d", total)
	// }

	// Create and store property
	property := types.Property{
		Index:   id,
		Address: msg.Address,
		Region:  msg.Region,
		Value:   msg.Value,
		Owners:  msg.Owners,
		Shares:  msg.Shares,
	}
	k.SetProperty(ctx, property)

	return &types.MsgRegisterPropertyResponse{}, nil
}
