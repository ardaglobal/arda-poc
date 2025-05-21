package keeper

import (
	"context"
	"fmt"

	"github.com/ardaglobal/arda-poc/x/property/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) TransferShares(goCtx context.Context, msg *types.MsgTransferShares) (*types.MsgTransferSharesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)


	// Retrieve property
	property, found := k.GetProperty(ctx, msg.PropertyId)
	if !found {
		return nil, fmt.Errorf("property not found: %s", msg.PropertyId)
	}

	// Validate input
	if len(msg.FromOwners) != len(msg.FromShares) {
		return nil, fmt.Errorf("mismatch in from_owners and from_shares")
	}
	if len(msg.ToOwners) != len(msg.ToShares) {
		return nil, fmt.Errorf("mismatch in to_owners and to_shares")
	}

	// Sum shares
	var totalFrom uint64
	for _, share := range msg.FromShares {
		totalFrom += share
	}

	var totalTo uint64
	for _, share := range msg.ToShares {
		totalTo += share
	}

	if totalFrom != totalTo {
		return nil, fmt.Errorf("total shares out (%d) must match shares in (%d)", totalTo, totalFrom)
	}


	ownerMap := k.ConvertPropertyOwnersToMap(property)
	
	// Deduct shares from from_owners
	for i, owner := range msg.FromOwners {
		currentShare := ownerMap[owner]
		if currentShare < msg.FromShares[i] {
			return nil, fmt.Errorf("owner %s does not have enough shares", owner)
		}
		ownerMap[owner] = currentShare - msg.FromShares[i]
		if ownerMap[owner] == 0 {
			delete(ownerMap, owner)
		}
	}

	// Add shares to to_owners
	for i, newOwner := range msg.ToOwners {
		ownerMap[newOwner] += msg.ToShares[i]
	}

	k.UpdatePropertyFromOwnerMap(&property, ownerMap)
	k.SetProperty(ctx, property)

	return &types.MsgTransferSharesResponse{}, nil
}
