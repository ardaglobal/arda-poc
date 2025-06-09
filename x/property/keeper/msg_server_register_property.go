package keeper

import (
	"context"
	"fmt"
	"strings"

	"cosmossdk.io/math"
	"github.com/ardaglobal/arda-poc/pkg/utils"
	ardatypes "github.com/ardaglobal/arda-poc/x/arda/types"
	"github.com/ardaglobal/arda-poc/x/property/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RegisterProperty(goCtx context.Context, msg *types.MsgRegisterProperty) (_ *types.MsgRegisterPropertyResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Use address as deterministic property ID
	id := strings.ToLower(strings.TrimSpace(msg.Address))

	// If error, submit bad hash to arda module to show an invalid hash
	defer func() {
		if err != nil {
			badSig := strings.Repeat("00", 64)
			nilHash, _ := hashProperty(types.Property{})
			_, _ = k.ardaKeeper.SubmitHash(goCtx, ardatypes.NewMsgSubmitHash(msg.Creator, msg.Region, nilHash, badSig))
		}
	}()

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
		Shares:  msg.Shares,
	}
	k.SetProperty(ctx, property)

	// Hash property info and submit to arda module
	hash, err := hashProperty(property)
	if err != nil {
		return nil, err
	}
	hashHex, sigHex, err := utils.GenerateHashAndSignature(utils.DefaultKeyFile, hash)
	if err != nil {
		return nil, err
	}
	_, err = k.ardaKeeper.SubmitHash(goCtx, ardatypes.NewMsgSubmitHash(msg.Creator, msg.Region, hashHex, sigHex))
	if err != nil {
		return nil, err
	}

	// Mint property share tokens to owners using x/bank
	denom := types.PropertyShareDenom(id)
	for i, owner := range msg.Owners {
		if i >= len(msg.Shares) {
			break
		}
		share := msg.Shares[i]
		if share == 0 {
			continue
		}
		addr, err := sdk.AccAddressFromBech32(owner)
		if err != nil {
			return nil, err
		}
		coin := sdk.NewCoin(denom, math.NewInt(int64(share)))
		if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(coin)); err != nil {
			return nil, err
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(coin)); err != nil {
			return nil, err
		}
	}

	// Mint USDArda tokens equivalent to property value to owners
	if err := k.usdardaKeeper.Mint(ctx, property, property.Value); err != nil {
		return nil, err
	}

	return &types.MsgRegisterPropertyResponse{}, nil
}
