package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgMintMortgageToken{}

func NewMsgMintMortgageToken(creator string, mortgageId string, amount uint64) *MsgMintMortgageToken {
	return &MsgMintMortgageToken{
		Creator:    creator,
		MortgageId: mortgageId,
		Amount:     amount,
	}
}

func (msg *MsgMintMortgageToken) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
