package keeper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ardaglobal/arda-poc/pkg/utils"
	ardatypes "github.com/ardaglobal/arda-poc/x/arda/types"
	"github.com/ardaglobal/arda-poc/x/property/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) TransferShares(goCtx context.Context, msg *types.MsgTransferShares) (_ *types.MsgTransferSharesResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Retrieve property
	property, found := k.GetProperty(ctx, msg.PropertyId)
	if !found {
		return nil, fmt.Errorf("property not found: %s", msg.PropertyId)
	}

	// If error, submit bad hash to arda module to show an invalid hash
	defer func() {
		if err != nil {
			badSig := strings.Repeat("00", 64)
			nilHash, _ := hashProperty(types.Property{})
			_, _ = k.ardaKeeper.SubmitHash(goCtx, ardatypes.NewMsgSubmitHash(msg.Creator, property.Region, nilHash, badSig))
		}
	}()

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
	fromSharesAndOwners := make([]string, 0, len(msg.FromOwners))
	for i, owner := range msg.FromOwners {
		fromSharesAndOwners = append(fromSharesAndOwners, fmt.Sprintf("%s:%d", owner, msg.FromShares[i]))
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
	toSharesAndOwners := make([]string, 0, len(msg.ToOwners))
	for i, newOwner := range msg.ToOwners {
		toSharesAndOwners = append(toSharesAndOwners, fmt.Sprintf("%s:%d", newOwner, msg.ToShares[i]))
		ownerMap[newOwner] += msg.ToShares[i]
	}

	k.UpdatePropertyFromOwnerMap(&property, ownerMap)

	// Append transfer history
	transfer := &types.Transfer{
		From:      strings.Join(fromSharesAndOwners, ","),
		To:        strings.Join(toSharesAndOwners, ","),
		Timestamp: ctx.BlockTime().UTC().Format(time.RFC3339),
	}
	property.Transfers = append(property.Transfers, transfer)

	k.SetProperty(ctx, property)

	hash, err := hashTransfer(msg, property)
	if err != nil {
		return nil, err
	}
	hashHex, sigHex, err := utils.GenerateHashAndSignature(utils.DefaultKeyFile, hash)
	if err != nil {
		return nil, err
	}
	_, err = k.ardaKeeper.SubmitHash(goCtx, ardatypes.NewMsgSubmitHash(msg.Creator, property.Region, hashHex, sigHex))
	if err != nil {
		return nil, err
	}

	return &types.MsgTransferSharesResponse{}, nil
}
