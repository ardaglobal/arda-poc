package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgRepayMortgage{}

func NewMsgRepayMortgage(creator string, mortgageId string, amount uint64) *MsgRepayMortgage {
	return &MsgRepayMortgage{
		Creator:    creator,
		MortgageId: mortgageId,
		Amount:     amount,
	}
}

func (msg *MsgRepayMortgage) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
