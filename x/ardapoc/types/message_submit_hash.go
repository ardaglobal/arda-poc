package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgSubmitHash{}

func NewMsgSubmitHash(creator string, region string, hash string, signature string) *MsgSubmitHash {
	return &MsgSubmitHash{
		Creator:   creator,
		Region:    region,
		Hash:      hash,
		Signature: signature,
	}
}

func (msg *MsgSubmitHash) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
