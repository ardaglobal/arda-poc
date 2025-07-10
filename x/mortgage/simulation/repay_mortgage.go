package simulation

import (
	"math/rand"

	"github.com/ardaglobal/arda-poc/x/mortgage/keeper"
	"github.com/ardaglobal/arda-poc/x/mortgage/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func SimulateMsgRepayMortgage(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgRepayMortgage{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the RepayMortgage simulation

		return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "RepayMortgage simulation not implemented"), nil, nil
	}
}
