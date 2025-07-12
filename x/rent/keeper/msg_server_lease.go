package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/ardaglobal/arda-poc/x/rent/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CreateLease(goCtx context.Context, msg *types.MsgCreateLease) (*types.MsgCreateLeaseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var lease = types.Lease{
		Creator:                  msg.Creator,
		PropertyId:               msg.PropertyId,
		Tenant:                   msg.Tenant,
		RentAmount:               msg.RentAmount,
		RentDueDate:              msg.RentDueDate,
		Status:                   msg.Status,
		TimePeriod:               msg.TimePeriod,
		PaymentsOutstanding:      msg.PaymentsOutstanding,
		TermLength:               msg.TermLength,
		RecurringStatus:          msg.RecurringStatus,
		CancellationPending:      msg.CancellationPending,
		CancellationInitiator:    msg.CancellationInitiator,
		CancellationDeadline:     msg.CancellationDeadline,
		LastPaymentBlock:         msg.LastPaymentBlock,
		PaymentTerms:             msg.PaymentTerms,
		CancellationRequirements: msg.CancellationRequirements,
	}

	id := k.AppendLease(
		ctx,
		lease,
	)

	return &types.MsgCreateLeaseResponse{
		Id: id,
	}, nil
}

func (k msgServer) UpdateLease(goCtx context.Context, msg *types.MsgUpdateLease) (*types.MsgUpdateLeaseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var lease = types.Lease{
		Creator:                  msg.Creator,
		Id:                       msg.Id,
		PropertyId:               msg.PropertyId,
		Tenant:                   msg.Tenant,
		RentAmount:               msg.RentAmount,
		RentDueDate:              msg.RentDueDate,
		Status:                   msg.Status,
		TimePeriod:               msg.TimePeriod,
		PaymentsOutstanding:      msg.PaymentsOutstanding,
		TermLength:               msg.TermLength,
		RecurringStatus:          msg.RecurringStatus,
		CancellationPending:      msg.CancellationPending,
		CancellationInitiator:    msg.CancellationInitiator,
		CancellationDeadline:     msg.CancellationDeadline,
		LastPaymentBlock:         msg.LastPaymentBlock,
		PaymentTerms:             msg.PaymentTerms,
		CancellationRequirements: msg.CancellationRequirements,
	}

	// Checks that the element exists
	val, found := k.GetLease(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}

	// Checks if the msg creator is the same as the current owner
	if msg.Creator != val.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	k.SetLease(ctx, lease)

	return &types.MsgUpdateLeaseResponse{}, nil
}

func (k msgServer) DeleteLease(goCtx context.Context, msg *types.MsgDeleteLease) (*types.MsgDeleteLeaseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Checks that the element exists
	val, found := k.GetLease(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}

	// Checks if the msg creator is the same as the current owner
	if msg.Creator != val.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	k.RemoveLease(ctx, msg.Id)

	return &types.MsgDeleteLeaseResponse{}, nil
}
