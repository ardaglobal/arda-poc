package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgTransferShares{}

func NewMsgTransferShares(creator string, propertyId string, fromOwners []string, fromShares []uint64, toOwners []string, toShares []uint64) *MsgTransferShares {
	return &MsgTransferShares{
		Creator:    creator,
		PropertyId: propertyId,
		FromOwners: fromOwners,
		FromShares: fromShares,
		ToOwners:   toOwners,
		ToShares:   toShares,
	}
}

func (msg *MsgTransferShares) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
