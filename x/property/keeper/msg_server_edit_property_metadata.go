package keeper

import (
	"context"
	"fmt"

	"github.com/ardaglobal/arda-poc/pkg/utils"
	ardatypes "github.com/ardaglobal/arda-poc/x/arda/types"
	"github.com/ardaglobal/arda-poc/x/property/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) EditPropertyMetadata(goCtx context.Context, msg *types.MsgEditPropertyMetadata) (*types.MsgEditPropertyMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	property, found := k.GetProperty(ctx, msg.PropertyId)
	if !found {
		return nil, fmt.Errorf("property not found: %s", msg.PropertyId)
	}

	property.PropertyName = msg.PropertyName
	property.PropertyType = msg.PropertyType
	property.ParcelNumber = msg.ParcelNumber
	property.Size_ = msg.Size
	property.ConstructionInformation = msg.ConstructionInformation
	property.ZoningClassification = msg.ZoningClassification
	property.OwnerInformation = msg.OwnerInformation
	property.TenantId = msg.TenantId
	property.UnitNumber = msg.UnitNumber

	k.SetProperty(ctx, property)

	hash, err := hashProperty(property)
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

	return &types.MsgEditPropertyMetadataResponse{}, nil
}
