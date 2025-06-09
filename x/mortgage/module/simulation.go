package mortgage

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/ardaglobal/arda-poc/testutil/sample"
	mortgagesimulation "github.com/ardaglobal/arda-poc/x/mortgage/simulation"
	"github.com/ardaglobal/arda-poc/x/mortgage/types"
)

// avoid unused import issue
var (
	_ = mortgagesimulation.FindAccount
	_ = rand.Rand{}
	_ = sample.AccAddress
	_ = sdk.AccAddress{}
	_ = simulation.MsgEntryKind
)

const (
	opWeightMsgCreateMortgage = "op_weight_msg_mortgage"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreateMortgage int = 100

	opWeightMsgUpdateMortgage = "op_weight_msg_mortgage"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateMortgage int = 100

	opWeightMsgDeleteMortgage = "op_weight_msg_mortgage"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDeleteMortgage int = 100

	opWeightMsgMintMortgageToken = "op_weight_msg_mint_mortgage_token"
	// TODO: Determine the simulation weight value
	defaultWeightMsgMintMortgageToken int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	mortgageGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		MortgageList: []types.Mortgage{
			{
				Creator: sample.AccAddress(),
				Index:   "0",
			},
			{
				Creator: sample.AccAddress(),
				Index:   "1",
			},
		},
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&mortgageGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgCreateMortgage int
	simState.AppParams.GetOrGenerate(opWeightMsgCreateMortgage, &weightMsgCreateMortgage, nil,
		func(_ *rand.Rand) {
			weightMsgCreateMortgage = defaultWeightMsgCreateMortgage
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateMortgage,
		mortgagesimulation.SimulateMsgCreateMortgage(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateMortgage int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateMortgage, &weightMsgUpdateMortgage, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateMortgage = defaultWeightMsgUpdateMortgage
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateMortgage,
		mortgagesimulation.SimulateMsgUpdateMortgage(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgDeleteMortgage int
	simState.AppParams.GetOrGenerate(opWeightMsgDeleteMortgage, &weightMsgDeleteMortgage, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteMortgage = defaultWeightMsgDeleteMortgage
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteMortgage,
		mortgagesimulation.SimulateMsgDeleteMortgage(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgMintMortgageToken int
	simState.AppParams.GetOrGenerate(opWeightMsgMintMortgageToken, &weightMsgMintMortgageToken, nil,
		func(_ *rand.Rand) {
			weightMsgMintMortgageToken = defaultWeightMsgMintMortgageToken
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgMintMortgageToken,
		mortgagesimulation.SimulateMsgMintMortgageToken(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			opWeightMsgCreateMortgage,
			defaultWeightMsgCreateMortgage,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				mortgagesimulation.SimulateMsgCreateMortgage(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgUpdateMortgage,
			defaultWeightMsgUpdateMortgage,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				mortgagesimulation.SimulateMsgUpdateMortgage(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgDeleteMortgage,
			defaultWeightMsgDeleteMortgage,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				mortgagesimulation.SimulateMsgDeleteMortgage(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgMintMortgageToken,
			defaultWeightMsgMintMortgageToken,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				mortgagesimulation.SimulateMsgMintMortgageToken(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		// this line is used by starport scaffolding # simapp/module/OpMsg
	}
}
