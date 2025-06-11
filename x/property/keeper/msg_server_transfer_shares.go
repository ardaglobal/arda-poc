package keeper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/ardaglobal/arda-poc/pkg/utils"
	ardatypes "github.com/ardaglobal/arda-poc/x/arda/types"
	"github.com/ardaglobal/arda-poc/x/property/types"
	usdtypes "github.com/ardaglobal/arda-poc/x/usdarda/types"
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
		return nil, fmt.Errorf("total shares out (%d) must match shares in (%d)", totalFrom, totalTo)
	}

	ownerMap := k.ConvertPropertyOwnersToMap(property)
	denom := types.PropertyShareDenom(property.Index)
	usdDenom := usdtypes.USDArdaDenom

	// Ensure all recipients have sufficient USDArda to cover the purchase
	for i, newOwner := range msg.ToOwners {
		addr, err := sdk.AccAddressFromBech32(newOwner)
		if err != nil {
			return nil, err
		}
		usdAmt := property.Value * msg.ToShares[i] / 100
		balance := k.bankKeeper.SpendableCoins(ctx, addr).AmountOf(usdDenom)
		if balance.LT(math.NewIntFromUint64(usdAmt)) {
			return nil, fmt.Errorf("owner %s has insufficient usdarda", newOwner)
		}
	}

	// Deduct shares from from_owners and move share tokens to the module account
	fromSharesAndOwners := make([]string, 0, len(msg.FromOwners))
	// NOTE: this is not atomic
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

		addr, err := sdk.AccAddressFromBech32(owner)
		if err != nil {
			return nil, err
		}
		coin := sdk.NewCoin(denom, math.NewInt(int64(msg.FromShares[i])))
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, sdk.NewCoins(coin)); err != nil {
			return nil, err
		}
	}

	// Add shares to to_owners, distributing share tokens from the module account
	toSharesAndOwners := make([]string, 0, len(msg.ToOwners))
	// NOTE: this is not atomic
	for i, newOwner := range msg.ToOwners {
		toSharesAndOwners = append(toSharesAndOwners, fmt.Sprintf("%s:%d", newOwner, msg.ToShares[i]))
		ownerMap[newOwner] += msg.ToShares[i]

		addr, err := sdk.AccAddressFromBech32(newOwner)
		if err != nil {
			return nil, err
		}
		coin := sdk.NewCoin(denom, math.NewInt(int64(msg.ToShares[i])))
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(coin)); err != nil {
			return nil, err
		}
	}

	// Move USDArda from buyers to sellers via the module account
	for i, newOwner := range msg.ToOwners {
		addr, err := sdk.AccAddressFromBech32(newOwner)
		if err != nil {
			return nil, err
		}
		usdAmt := property.Value * msg.ToShares[i] / 100
		coin := sdk.NewCoin(usdDenom, math.NewInt(int64(usdAmt)))
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, sdk.NewCoins(coin)); err != nil {
			return nil, err
		}
	}
	for i, owner := range msg.FromOwners {
		addr, err := sdk.AccAddressFromBech32(owner)
		if err != nil {
			return nil, err
		}
		usdAmt := property.Value * msg.FromShares[i] / 100
		coin := sdk.NewCoin(usdDenom, math.NewInt(int64(usdAmt)))
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(coin)); err != nil {
			return nil, err
		}
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
