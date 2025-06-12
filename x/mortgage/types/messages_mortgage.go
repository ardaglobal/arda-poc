package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateMortgage{}

func NewMsgCreateMortgage(
	creator string,
	index string,
	lender string,
	lendee string,
	collateral string,
	amount uint64,
	interestRate string,
	term string,

) *MsgCreateMortgage {
	return &MsgCreateMortgage{
		Creator:      creator,
		Index:        index,
		Lender:       lender,
		Lendee:       lendee,
		Collateral:   collateral,
		Amount:       amount,
		InterestRate: interestRate,
		Term:         term,
	}
}

func (msg *MsgCreateMortgage) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateMortgage{}

func NewMsgUpdateMortgage(
	creator string,
	index string,
	lender string,
	lendee string,
	collateral string,
	amount uint64,
	interestRate string,
	term string,

) *MsgUpdateMortgage {
	return &MsgUpdateMortgage{
		Creator:      creator,
		Index:        index,
		Lender:       lender,
		Lendee:       lendee,
		Collateral:   collateral,
		Amount:       amount,
		InterestRate: interestRate,
		Term:         term,
	}
}

func (msg *MsgUpdateMortgage) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgDeleteMortgage{}

func NewMsgDeleteMortgage(
	creator string,
	index string,

) *MsgDeleteMortgage {
	return &MsgDeleteMortgage{
		Creator: creator,
		Index:   index,
	}
}

func (msg *MsgDeleteMortgage) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
