package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateLease{}

func NewMsgCreateLease(creator string, propertyId string, tenant string, rentAmount uint64, rentDueDate string, status string, timePeriod uint64, paymentsOutstanding string, termLength string, recurringStatus bool, cancellationPending bool, cancellationInitiator string, cancellationDeadline uint64, lastPaymentBlock uint64, paymentTerms string, cancellationRequirements string) *MsgCreateLease {
	return &MsgCreateLease{
		Creator:                  creator,
		PropertyId:               propertyId,
		Tenant:                   tenant,
		RentAmount:               rentAmount,
		RentDueDate:              rentDueDate,
		Status:                   status,
		TimePeriod:               timePeriod,
		PaymentsOutstanding:      paymentsOutstanding,
		TermLength:               termLength,
		RecurringStatus:          recurringStatus,
		CancellationPending:      cancellationPending,
		CancellationInitiator:    cancellationInitiator,
		CancellationDeadline:     cancellationDeadline,
		LastPaymentBlock:         lastPaymentBlock,
		PaymentTerms:             paymentTerms,
		CancellationRequirements: cancellationRequirements,
	}
}

func (msg *MsgCreateLease) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateLease{}

func NewMsgUpdateLease(creator string, id uint64, propertyId string, tenant string, rentAmount uint64, rentDueDate string, status string, timePeriod uint64, paymentsOutstanding string, termLength string, recurringStatus bool, cancellationPending bool, cancellationInitiator string, cancellationDeadline uint64, lastPaymentBlock uint64, paymentTerms string, cancellationRequirements string) *MsgUpdateLease {
	return &MsgUpdateLease{
		Id:                       id,
		Creator:                  creator,
		PropertyId:               propertyId,
		Tenant:                   tenant,
		RentAmount:               rentAmount,
		RentDueDate:              rentDueDate,
		Status:                   status,
		TimePeriod:               timePeriod,
		PaymentsOutstanding:      paymentsOutstanding,
		TermLength:               termLength,
		RecurringStatus:          recurringStatus,
		CancellationPending:      cancellationPending,
		CancellationInitiator:    cancellationInitiator,
		CancellationDeadline:     cancellationDeadline,
		LastPaymentBlock:         lastPaymentBlock,
		PaymentTerms:             paymentTerms,
		CancellationRequirements: cancellationRequirements,
	}
}

func (msg *MsgUpdateLease) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgDeleteLease{}

func NewMsgDeleteLease(creator string, id uint64) *MsgDeleteLease {
	return &MsgDeleteLease{
		Id:      id,
		Creator: creator,
	}
}

func (msg *MsgDeleteLease) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
