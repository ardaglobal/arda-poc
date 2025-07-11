package rent

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/ardaglobal/arda-poc/testutil/sample"
	rentsimulation "github.com/ardaglobal/arda-poc/x/rent/simulation"
	"github.com/ardaglobal/arda-poc/x/rent/types"
)

// avoid unused import issue
var (
	_ = rentsimulation.FindAccount
	_ = rand.Rand{}
	_ = sample.AccAddress
	_ = sdk.AccAddress{}
	_ = simulation.MsgEntryKind
)

const (
	opWeightMsgCreateLease = "op_weight_msg_lease"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreateLease int = 100

	opWeightMsgUpdateLease = "op_weight_msg_lease"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateLease int = 100

	opWeightMsgDeleteLease = "op_weight_msg_lease"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDeleteLease int = 100

	opWeightMsgPayRent = "op_weight_msg_pay_rent"
	// TODO: Determine the simulation weight value
	defaultWeightMsgPayRent int = 100

	opWeightMsgInitiateCancellation = "op_weight_msg_initiate_cancellation"
	// TODO: Determine the simulation weight value
	defaultWeightMsgInitiateCancellation int = 100

	opWeightMsgApproveCancellation = "op_weight_msg_approve_cancellation"
	// TODO: Determine the simulation weight value
	defaultWeightMsgApproveCancellation int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	rentGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		LeaseList: []types.Lease{
			{
				Id:      0,
				Creator: sample.AccAddress(),
			},
			{
				Id:      1,
				Creator: sample.AccAddress(),
			},
		},
		LeaseCount: 2,
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&rentGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgCreateLease int
	simState.AppParams.GetOrGenerate(opWeightMsgCreateLease, &weightMsgCreateLease, nil,
		func(_ *rand.Rand) {
			weightMsgCreateLease = defaultWeightMsgCreateLease
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateLease,
		rentsimulation.SimulateMsgCreateLease(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateLease int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateLease, &weightMsgUpdateLease, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateLease = defaultWeightMsgUpdateLease
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateLease,
		rentsimulation.SimulateMsgUpdateLease(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgDeleteLease int
	simState.AppParams.GetOrGenerate(opWeightMsgDeleteLease, &weightMsgDeleteLease, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteLease = defaultWeightMsgDeleteLease
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteLease,
		rentsimulation.SimulateMsgDeleteLease(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgPayRent int
	simState.AppParams.GetOrGenerate(opWeightMsgPayRent, &weightMsgPayRent, nil,
		func(_ *rand.Rand) {
			weightMsgPayRent = defaultWeightMsgPayRent
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgPayRent,
		rentsimulation.SimulateMsgPayRent(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgInitiateCancellation int
	simState.AppParams.GetOrGenerate(opWeightMsgInitiateCancellation, &weightMsgInitiateCancellation, nil,
		func(_ *rand.Rand) {
			weightMsgInitiateCancellation = defaultWeightMsgInitiateCancellation
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgInitiateCancellation,
		rentsimulation.SimulateMsgInitiateCancellation(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgApproveCancellation int
	simState.AppParams.GetOrGenerate(opWeightMsgApproveCancellation, &weightMsgApproveCancellation, nil,
		func(_ *rand.Rand) {
			weightMsgApproveCancellation = defaultWeightMsgApproveCancellation
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgApproveCancellation,
		rentsimulation.SimulateMsgApproveCancellation(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			opWeightMsgCreateLease,
			defaultWeightMsgCreateLease,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				rentsimulation.SimulateMsgCreateLease(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgUpdateLease,
			defaultWeightMsgUpdateLease,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				rentsimulation.SimulateMsgUpdateLease(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgDeleteLease,
			defaultWeightMsgDeleteLease,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				rentsimulation.SimulateMsgDeleteLease(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgPayRent,
			defaultWeightMsgPayRent,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				rentsimulation.SimulateMsgPayRent(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgInitiateCancellation,
			defaultWeightMsgInitiateCancellation,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				rentsimulation.SimulateMsgInitiateCancellation(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgApproveCancellation,
			defaultWeightMsgApproveCancellation,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				rentsimulation.SimulateMsgApproveCancellation(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		// this line is used by starport scaffolding # simapp/module/OpMsg
	}
}
