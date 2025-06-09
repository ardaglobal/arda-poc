package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateMortgage{}

type MsgCreateMortgage struct {
	Creator      string
	Lender       string
	Lendee       string
	Collateral   string
	Amount       uint64
	InterestRate string
	Term         string
}

func NewMsgCreateMortgage(creator, lender, lendee, collateral string, amount uint64, rate, term string) *MsgCreateMortgage {
	return &MsgCreateMortgage{
		Creator:      creator,
		Lender:       lender,
		Lendee:       lendee,
		Collateral:   collateral,
		Amount:       amount,
		InterestRate: rate,
		Term:         term,
	}
}

func (msg *MsgCreateMortgage) Route() string { return RouterKey }
func (msg *MsgCreateMortgage) Type() string  { return "CreateMortgage" }

func (msg *MsgCreateMortgage) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateMortgage) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateMortgage) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
