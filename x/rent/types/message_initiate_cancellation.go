package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgInitiateCancellation{}

func NewMsgInitiateCancellation(creator string, leaseId string) *MsgInitiateCancellation {
	return &MsgInitiateCancellation{
		Creator: creator,
		LeaseId: leaseId,
	}
}

func (msg *MsgInitiateCancellation) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
