package app

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GaslessAnteHandler returns an AnteHandler that wraps the existing handler and
// skips fee and gas checks by setting an infinite gas meter on the context.
func GaslessAnteHandler(next sdk.AnteHandler) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
		return next(ctx, tx, simulate) // keep sig-checks, fee deduction, etc.
	}
}
