package simulation

import (
	"math/rand"

	"github.com/ardaglobal/arda-poc/x/rent/keeper"
	"github.com/ardaglobal/arda-poc/x/rent/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func SimulateMsgApproveCancellation(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgApproveCancellation{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the ApproveCancellation simulation

		return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "ApproveCancellation simulation not implemented"), nil, nil
	}
}
