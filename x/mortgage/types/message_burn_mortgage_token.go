package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgBurnMortgageToken{}

type MsgBurnMortgageToken struct {
	Creator  string
	UsdToken string
	Amount   uint64
	Owner    string
}

func NewMsgBurnMortgageToken(creator, usdToken string, amount uint64, owner string) *MsgBurnMortgageToken {
	return &MsgBurnMortgageToken{Creator: creator, UsdToken: usdToken, Amount: amount, Owner: owner}
}

func (msg *MsgBurnMortgageToken) Route() string { return RouterKey }
func (msg *MsgBurnMortgageToken) Type() string  { return "BurnMortgageToken" }

func (msg *MsgBurnMortgageToken) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgBurnMortgageToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgBurnMortgageToken) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
	}
	return nil
}
