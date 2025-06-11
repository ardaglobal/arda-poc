package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgEditPropertyMetadata{}

func NewMsgEditPropertyMetadata(creator, propertyId, name, ptype, parcel, size, construction, zoning, ownerInfo, tenantId, unit string) *MsgEditPropertyMetadata {
	return &MsgEditPropertyMetadata{
		Creator:                 creator,
		PropertyId:              propertyId,
		PropertyName:            name,
		PropertyType:            ptype,
		ParcelNumber:            parcel,
		ParcelSize:              size,
		ConstructionInformation: construction,
		ZoningClassification:    zoning,
		OwnerInformation:        ownerInfo,
		TenantId:                tenantId,
		UnitNumber:              unit,
	}
}

func (msg *MsgEditPropertyMetadata) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if msg.PropertyId == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "property id cannot be empty")
	}
	return nil
}
