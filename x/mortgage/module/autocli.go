package mortgage

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	modulev1 "github.com/ardaglobal/arda-poc/api/ardapoc/mortgage"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: modulev1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the module",
				},
				{
					RpcMethod: "MortgageAll",
					Use:       "list-mortgage",
					Short:     "List all Mortgage",
				},
				{
					RpcMethod:      "Mortgage",
					Use:            "show-mortgage [id]",
					Short:          "Shows a Mortgage",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "index"}},
				},
				// this line is used by ignite scaffolding # autocli/query
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              modulev1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true, // only required if you want to use the custom command
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
				{
					RpcMethod:      "CreateMortgage",
					Use:            "create-mortgage [index] [lender] [lendee] [collateral] [amount] [interestRate] [term]",
					Short:          "Create a new Mortgage",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "index"}, {ProtoField: "lender"}, {ProtoField: "lendee"}, {ProtoField: "collateral"}, {ProtoField: "amount"}, {ProtoField: "interestRate"}, {ProtoField: "term"}},
				},
				{
					RpcMethod:      "UpdateMortgage",
					Use:            "update-mortgage [index] [lender] [lendee] [collateral] [amount] [interestRate] [term]",
					Short:          "Update Mortgage",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "index"}, {ProtoField: "lender"}, {ProtoField: "lendee"}, {ProtoField: "collateral"}, {ProtoField: "amount"}, {ProtoField: "interestRate"}, {ProtoField: "term"}},
				},
				{
					RpcMethod:      "DeleteMortgage",
					Use:            "delete-mortgage [index]",
					Short:          "Delete Mortgage",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "index"}},
				},
				{
					RpcMethod:      "RepayMortgage",
					Use:            "repay-mortgage [mortgage-id] [amount]",
					Short:          "Send a repay-mortgage tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "mortgageId"}, {ProtoField: "amount"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
