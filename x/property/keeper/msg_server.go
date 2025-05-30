package keeper

import (
	ardamodulekeeper "github.com/ardaglobal/arda-poc/x/arda/keeper"
	"github.com/ardaglobal/arda-poc/x/property/types"
	usdardakeeper "github.com/ardaglobal/arda-poc/x/usdarda/keeper"
)

type msgServer struct {
	Keeper
	ardaKeeper    ardamodulekeeper.Keeper
	bankKeeper    types.BankKeeper
	usdardaKeeper usdardakeeper.Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper, ardaKeeper ardamodulekeeper.Keeper, bankKeeper types.BankKeeper, usdardaKeeper usdardakeeper.Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper, ardaKeeper: ardaKeeper, bankKeeper: bankKeeper, usdardaKeeper: usdardaKeeper}
}

var _ types.MsgServer = msgServer{}
