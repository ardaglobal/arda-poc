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
					Short:     "List all mortgage",
				},
				{
					RpcMethod:      "Mortgage",
					Use:            "show-mortgage [id]",
					Short:          "Shows a mortgage",
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
					Short:          "Create a new mortgage",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "index"}, {ProtoField: "lender"}, {ProtoField: "lendee"}, {ProtoField: "collateral"}, {ProtoField: "amount"}, {ProtoField: "interestRate"}, {ProtoField: "term"}},
				},
				{
					RpcMethod:      "UpdateMortgage",
					Use:            "update-mortgage [index] [lender] [lendee] [collateral] [amount] [interestRate] [term]",
					Short:          "Update mortgage",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "index"}, {ProtoField: "lender"}, {ProtoField: "lendee"}, {ProtoField: "collateral"}, {ProtoField: "amount"}, {ProtoField: "interestRate"}, {ProtoField: "term"}},
				},
				{
					RpcMethod:      "DeleteMortgage",
					Use:            "delete-mortgage [index]",
					Short:          "Delete mortgage",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "index"}},
				},
				{
					RpcMethod:      "MintMortgageToken",
					Use:            "mint-mortgage-token [mortgage_id] [amount]",
					Short:          "mint mortgage token",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "mortgage_id"}, {ProtoField: "amount"}},
				},
				{
					RpcMethod: "BurnMortgageToken",
					Use:       "burn-mortgage-token [mortgage_id] [amount]",
					Short:     "burn mortgage token",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "mortgage_id"}, {ProtoField: "amount"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
