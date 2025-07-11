package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/rent module sentinel errors
var (
	ErrInvalidSigner               = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrSample                      = sdkerrors.Register(ModuleName, 1101, "sample error")
	ErrInvalidLeaseID              = sdkerrors.Register(ModuleName, 1102, "invalid lease ID")
	ErrLeaseNotFound               = sdkerrors.Register(ModuleName, 1103, "lease not found")
	ErrUnauthorized                = sdkerrors.Register(ModuleName, 1104, "unauthorized")
	ErrInvalidAddress              = sdkerrors.Register(ModuleName, 1105, "invalid address")
	ErrInvalidOutstandingPayments  = sdkerrors.Register(ModuleName, 1106, "invalid outstanding payments")
	ErrInsufficientFunds           = sdkerrors.Register(ModuleName, 1107, "insufficient funds")
	ErrPropertyNotFound            = sdkerrors.Register(ModuleName, 1108, "property not found")
	ErrNoPropertyOwners            = sdkerrors.Register(ModuleName, 1109, "no property owners")
	ErrInvalidShares               = sdkerrors.Register(ModuleName, 1110, "invalid shares")
	ErrDisbursementFailed          = sdkerrors.Register(ModuleName, 1111, "disbursement failed")
	ErrSerializationFailed         = sdkerrors.Register(ModuleName, 1112, "serialization failed")
)
