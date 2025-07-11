package rent

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	modulev1 "github.com/ardaglobal/arda-poc/api/ardapoc/rent"
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
					RpcMethod: "LeaseAll",
					Use:       "list-lease",
					Short:     "List all Lease",
				},
				{
					RpcMethod:      "Lease",
					Use:            "show-lease [id]",
					Short:          "Shows a Lease by id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
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
					RpcMethod:      "CreateLease",
					Use:            "create-lease [propertyId] [tenant] [rentAmount] [rentDueDate] [status] [timePeriod] [paymentsOutstanding] [termLength] [recurringStatus] [cancellationPending] [cancellationInitiator] [cancellationDeadline] [lastPaymentBlock] [paymentTerms] [cancellationRequirements]",
					Short:          "Create Lease",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "propertyId"}, {ProtoField: "tenant"}, {ProtoField: "rentAmount"}, {ProtoField: "rentDueDate"}, {ProtoField: "status"}, {ProtoField: "timePeriod"}, {ProtoField: "paymentsOutstanding"}, {ProtoField: "termLength"}, {ProtoField: "recurringStatus"}, {ProtoField: "cancellationPending"}, {ProtoField: "cancellationInitiator"}, {ProtoField: "cancellationDeadline"}, {ProtoField: "lastPaymentBlock"}, {ProtoField: "paymentTerms"}, {ProtoField: "cancellationRequirements"}},
				},
				{
					RpcMethod:      "UpdateLease",
					Use:            "update-lease [id] [propertyId] [tenant] [rentAmount] [rentDueDate] [status] [timePeriod] [paymentsOutstanding] [termLength] [recurringStatus] [cancellationPending] [cancellationInitiator] [cancellationDeadline] [lastPaymentBlock] [paymentTerms] [cancellationRequirements]",
					Short:          "Update Lease",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}, {ProtoField: "propertyId"}, {ProtoField: "tenant"}, {ProtoField: "rentAmount"}, {ProtoField: "rentDueDate"}, {ProtoField: "status"}, {ProtoField: "timePeriod"}, {ProtoField: "paymentsOutstanding"}, {ProtoField: "termLength"}, {ProtoField: "recurringStatus"}, {ProtoField: "cancellationPending"}, {ProtoField: "cancellationInitiator"}, {ProtoField: "cancellationDeadline"}, {ProtoField: "lastPaymentBlock"}, {ProtoField: "paymentTerms"}, {ProtoField: "cancellationRequirements"}},
				},
				{
					RpcMethod:      "DeleteLease",
					Use:            "delete-lease [id]",
					Short:          "Delete Lease",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "PayRent",
					Use:            "pay-rent [lease-id]",
					Short:          "Send a PayRent tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "leaseId"}},
				},
				{
					RpcMethod:      "InitiateCancellation",
					Use:            "initiate-cancellation [lease-id]",
					Short:          "Send a InitiateCancellation tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "leaseId"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
