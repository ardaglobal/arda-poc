package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgRegisterProperty{}

func NewMsgRegisterProperty(creator string, address string, region string, value uint64, owners []string, shares []uint64) *MsgRegisterProperty {
	return &MsgRegisterProperty{
		Creator: creator,
		Address: address,
		Region:  region,
		Value:   value,
		Owners:  owners,
		Shares:  shares,
	}
}

func (msg *MsgRegisterProperty) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
