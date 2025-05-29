package keeper

import (
	ardamodulekeeper "github.com/ardaglobal/arda-poc/x/arda/keeper"
	"github.com/ardaglobal/arda-poc/x/property/types"
	usdkeeper "github.com/ardaglobal/arda-poc/x/usdarda/keeper"
)

type msgServer struct {
	Keeper
	ardaKeeper ardamodulekeeper.Keeper
	bankKeeper types.BankKeeper
	usdKeeper  usdkeeper.Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper, ardaKeeper ardamodulekeeper.Keeper, bankKeeper types.BankKeeper, usdKeeper usdkeeper.Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper, ardaKeeper: ardaKeeper, bankKeeper: bankKeeper, usdKeeper: usdKeeper}
}

var _ types.MsgServer = msgServer{}
