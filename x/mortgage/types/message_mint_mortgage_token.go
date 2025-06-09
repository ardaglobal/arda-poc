package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgMintMortgageToken{}

type MsgMintMortgageToken struct {
	Creator      string
	PropertyType string
	Amount       uint64
	Recipient    string
}

func NewMsgMintMortgageToken(creator, propertyType string, amount uint64, recipient string) *MsgMintMortgageToken {
	return &MsgMintMortgageToken{Creator: creator, PropertyType: propertyType, Amount: amount, Recipient: recipient}
}

func (msg *MsgMintMortgageToken) Route() string { return RouterKey }
func (msg *MsgMintMortgageToken) Type() string  { return "MintMortgageToken" }

func (msg *MsgMintMortgageToken) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgMintMortgageToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgMintMortgageToken) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid recipient address (%s)", err)
	}
	return nil
}
