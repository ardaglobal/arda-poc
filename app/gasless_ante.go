package app

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GaslessAnteHandler returns an AnteHandler that skips fee and gas checks by
// setting an infinite gas meter on the context.
func GaslessAnteHandler() sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return ctx.WithGasMeter(storetypes.NewInfiniteGasMeter()), nil
	}
}
