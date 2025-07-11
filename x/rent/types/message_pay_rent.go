package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgPayRent{}

func NewMsgPayRent(creator string, leaseId string) *MsgPayRent {
	return &MsgPayRent{
		Creator: creator,
		LeaseId: leaseId,
	}
}

func (msg *MsgPayRent) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
