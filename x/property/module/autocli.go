package property

import (
	"fmt"
	"strings"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	modulev1 "github.com/ardaglobal/arda-poc/api/ardapoc/property"
	"github.com/ardaglobal/arda-poc/x/property/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:              modulev1.Query_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true, // Use custom commands defined in GetQueryCmd
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the module",
				},
				{
					RpcMethod: "PropertyAll",
					Use:       "property-all",
					Short:     "Query all properties",
					Long:      "Query all registered properties with formatted display of owners and shares",
				},
				{
					RpcMethod:      "Property",
					Use:            "property [index]",
					Short:          "Query a single property by index",
					Long:           "Query a registered property by its index with formatted display of owners and shares",
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
					RpcMethod:      "RegisterProperty",
					Use:            "register-property [address] [region] [value]",
					Short:          "Send a register-property tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}, {ProtoField: "region"}, {ProtoField: "value"}},
				},
				{
					RpcMethod:      "TransferShares",
					Use:            "transfer-shares [property-id] [from-owners] [from-shares] [to-owners] [to-shares]",
					Short:          "Send a transfer-shares tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "propertyId"}, {ProtoField: "fromOwners"}, {ProtoField: "fromShares"}, {ProtoField: "toOwners"}, {ProtoField: "toShares"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}

// GetQueryCmd returns the custom-formatted property query commands
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "property",
		Short:                      "Querying commands for the property module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryProperties(),
		GetCmdQueryProperty(),
	)

	return cmd
}

// GetCmdQueryProperties implements a custom-formatted property-all command
func GetCmdQueryProperties() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "property-all",
		Short: "Query all properties with formatted output",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			params := &types.QueryAllPropertyRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.PropertyAll(cmd.Context(), params)
			if err != nil {
				return err
			}

			// Custom formatted output
			fmt.Fprintln(cmd.OutOrStdout(), "Properties:")
			for _, prop := range res.Properties {
				printProperty(prop)

				fmt.Fprintln(cmd.OutOrStdout(), "") // Empty line between properties
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "properties")

	return cmd
}

// GetCmdQueryProperty implements a custom-formatted property query command
func GetCmdQueryProperty() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "property [index]",
		Short: "Query a single property by index with formatted output",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			params := &types.QueryGetPropertyRequest{
				Index: strings.TrimSpace(args[0]),
			}

			res, err := queryClient.Property(cmd.Context(), params)
			if err != nil {
				return err
			}

			printProperty(res.Property)

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func printProperty(prop *types.Property) {
	fmt.Printf("Property:\n")
	fmt.Printf("  Index: %s\n", prop.Index)
	fmt.Printf("  Address: %s\n", prop.Address)
	fmt.Printf("  Region: %s\n", prop.Region)
	fmt.Printf("  Value: %v\n", prop.Value)
	fmt.Println("  Owners / Shares:")

	// Display owners and shares together
	for i := 0; i < len(prop.Owners); i++ {
		ownerShare := "unknown"
		if i < len(prop.Shares) {
			ownerShare = fmt.Sprintf("%d", prop.Shares[i])
		}
		fmt.Printf("    - %s / %s\n", prop.Owners[i], ownerShare)
	}

	fmt.Println("  Transfers:")
	for _, transfer := range prop.Transfers {
		fmt.Printf("    - From: %s; To: %s; Timestamp: %s\n",
			transfer.From, transfer.To, transfer.Timestamp)
	}
}
