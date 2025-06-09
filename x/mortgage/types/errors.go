package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/mortgage module sentinel errors
var (
	ErrInvalidSigner = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrSample        = sdkerrors.Register(ModuleName, 1101, "sample error")
	// ErrInvalidMortgage is the error for invalid mortgage
	ErrInvalidMortgage = sdkerrors.Register(ModuleName, 1, "invalid mortgage")
	// ErrMortgageNotFound is the error for mortgage not found
	ErrMortgageNotFound = sdkerrors.Register(ModuleName, 2, "mortgage not found")
	// ErrPropertyNotFound is the error for property not found
	ErrPropertyNotFound = sdkerrors.Register(ModuleName, 3, "property not found")
)
