package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgBurnMortgageToken{}

func NewMsgBurnMortgageToken(creator string, usdToken string, amount uint64, owner string) *MsgBurnMortgageToken {
	return &MsgBurnMortgageToken{
		Creator:  creator,
		UsdToken: usdToken,
		Amount:   amount,
		Owner:    owner,
	}
}

func (msg *MsgBurnMortgageToken) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
