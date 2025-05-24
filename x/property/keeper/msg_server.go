package keeper

import (
	ardamodulekeeper "github.com/ardaglobal/arda-poc/x/arda/keeper"
	"github.com/ardaglobal/arda-poc/x/property/types"
)

type msgServer struct {
	Keeper
	ardaKeeper ardamodulekeeper.Keeper
	bankKeeper types.BankKeeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper, ardaKeeper ardamodulekeeper.Keeper, bankKeeper types.BankKeeper) types.MsgServer {
	return &msgServer{Keeper: keeper, ardaKeeper: ardaKeeper, bankKeeper: bankKeeper}
}

var _ types.MsgServer = msgServer{}
