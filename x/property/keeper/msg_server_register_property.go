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

	// Validate ownership sum = 100
	var total uint64
	fmt.Println("Debug - RegisterProperty - msg:", msg)
	fmt.Println("Debug - RegisterProperty - owners:", msg.Owners)

	if msg.Owners == nil {
		fmt.Println("Debug - RegisterProperty - owners is nil!")
		msg.Owners = []string{}
	}

	if msg.Shares == nil {
		fmt.Println("Debug - RegisterProperty - shares is nil!")
		msg.Shares = []uint64{}
	}

	for i, owner := range msg.Owners {
		fmt.Printf("Debug - Owner: %s, Share: %d\n", owner, msg.Shares[i])
		total += msg.Shares[i]
	}
	fmt.Printf("Debug - Total ownership shares: %d\n", total)

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

	fmt.Println("Debug - Property saved with owners:", property.Owners)

	return &types.MsgRegisterPropertyResponse{}, nil
}
