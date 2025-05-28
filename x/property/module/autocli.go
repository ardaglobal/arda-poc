package property

import (
	"fmt"
	"strconv"
	"strings"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	modulev1 "github.com/ardaglobal/arda-poc/api/ardapoc/property"
	"github.com/ardaglobal/arda-poc/x/property/types"
	"github.com/cosmos/cosmos-sdk/client/tx"
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
				printProperty(clientCtx, prop)

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

			printProperty(clientCtx, res.Property)

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func printProperty(clientCtx client.Context, prop *types.Property) {
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

		// Try to get human-readable name for the address
		ownerName, err := getNameFromAddress(clientCtx, prop.Owners[i])
		if err != nil {
			ownerName = prop.Owners[i] // Use address if name not found
		}

		fmt.Printf("    - %s (%s) / %s\n", ownerName, prop.Owners[i], ownerShare)
	}

	fmt.Println("  Transfers:")
	for _, transfer := range prop.Transfers {
		// Try to get human-readable names for the addresses
		fromName, err := getNameFromAddress(clientCtx, transfer.From)
		if err != nil {
			fromName = transfer.From
		}

		toName, err := getNameFromAddress(clientCtx, transfer.To)
		if err != nil {
			toName = transfer.To
		}

		fmt.Printf("    - From: %s (%s); To: %s (%s); Timestamp: %s\n",
			fromName, transfer.From, toName, transfer.To, transfer.Timestamp)
	}
}

// GetTxCmd returns the custom-formatted property transaction commands
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "property",
		Short:                      "Transaction commands for the property module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdRegisterProperty(),
		GetCmdTransferShares(),
	)

	return cmd
}

// getAddressFromName converts a keyring name to its address
func getAddressFromName(clientCtx client.Context, name string) (string, error) {
	// If input is the length of an address and starts with "arda", assume it's already an address
	if len(name) == 44 && strings.HasPrefix(name, "arda") {
		return name, nil
	}
	record, err := clientCtx.Keyring.Key(name)
	if err != nil {
		return "", fmt.Errorf("failed to get address for %s: %s", name, err)
	}
	addr, err := record.GetAddress()
	if err != nil {
		return "", fmt.Errorf("failed to get address for %s: %s", name, err)
	}
	return addr.String(), nil
}

// getNameFromAddress attempts to find a key name in the keyring given an address
func getNameFromAddress(clientCtx client.Context, address string) (string, error) {
	records, err := clientCtx.Keyring.List()
	if err != nil {
		return "", fmt.Errorf("failed to list keyring entries: %s", err)
	}

	for _, record := range records {
		addr, err := record.GetAddress()
		if err != nil {
			continue // Skip records that can't get address
		}
		if addr.String() == address {
			return record.Name, nil
		}
	}

	return "", fmt.Errorf("no key found for address %s", address)
}

// GetCmdRegisterProperty implements a custom-formatted register-property command
func GetCmdRegisterProperty() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-property [address] [region] [value]",
		Short: "Register a new property with address, region, and value",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse value as uint64
			value, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid value: %s", err)
			}

			// Get owners and shares from flags
			ownerNames, err := cmd.Flags().GetStringSlice("owners")
			if err != nil {
				return err
			}

			sharesStr, err := cmd.Flags().GetStringSlice("shares")
			if err != nil {
				return err
			}

			// Convert shares to uint64
			shares := make([]uint64, len(sharesStr))
			for i, share := range sharesStr {
				shareUint, err := strconv.ParseUint(share, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid share value at position %d: %s", i, err)
				}
				shares[i] = shareUint
			}

			// Validate owners and shares have same length
			if len(ownerNames) != len(shares) {
				return fmt.Errorf("number of owners (%d) must match number of shares (%d)", len(ownerNames), len(shares))
			}

			// Convert owner names to addresses
			ownerAddresses := make([]string, len(ownerNames))
			for i, name := range ownerNames {
				addr, err := getAddressFromName(clientCtx, name)
				if err != nil {
					return err
				}
				ownerAddresses[i] = addr
			}

			msg := types.NewMsgRegisterProperty(
				clientCtx.GetFromAddress().String(),
				strings.TrimSpace(args[0]), // address
				strings.TrimSpace(args[1]), // region
				value,
				ownerAddresses,
				shares,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().StringSlice("owners", []string{}, "Comma-separated list of owner names from the keyring")
	cmd.Flags().StringSlice("shares", []string{}, "Comma-separated list of shares (must match number of owners)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdTransferShares implements a custom-formatted transfer-shares command
func GetCmdTransferShares() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-shares [property-id] [from-owners] [from-shares] [to-owners] [to-shares]",
		Short: "Transfer property shares between owners",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse from-owners and from-shares as arrays
			fromOwnerNames := strings.Split(args[1], ",")
			fromShares := strings.Split(args[2], ",")
			toOwnerNames := strings.Split(args[3], ",")
			toShares := strings.Split(args[4], ",")

			// Convert share strings to uint64 slices
			fromSharesUint := make([]uint64, len(fromShares))
			for i, share := range fromShares {
				shareUint, err := strconv.ParseUint(share, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid from-share value at position %d: %s", i, err)
				}
				fromSharesUint[i] = shareUint
			}

			toSharesUint := make([]uint64, len(toShares))
			for i, share := range toShares {
				shareUint, err := strconv.ParseUint(share, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid to-share value at position %d: %s", i, err)
				}
				toSharesUint[i] = shareUint
			}

			// Convert from-owner names to addresses
			fromOwnerAddresses := make([]string, len(fromOwnerNames))
			for i, name := range fromOwnerNames {
				addr, err := getAddressFromName(clientCtx, name)
				if err != nil {
					return fmt.Errorf("failed to get address for from-owner: %s", err)
				}
				fromOwnerAddresses[i] = addr
			}

			// Convert to-owner names to addresses
			toOwnerAddresses := make([]string, len(toOwnerNames))
			for i, name := range toOwnerNames {
				addr, err := getAddressFromName(clientCtx, name)
				if err != nil {
					return fmt.Errorf("failed to get address for to-owner: %s", err)
				}
				toOwnerAddresses[i] = addr
			}

			msg := types.NewMsgTransferShares(
				clientCtx.GetFromAddress().String(),
				strings.TrimSpace(args[0]), // property-id
				fromOwnerAddresses,
				fromSharesUint,
				toOwnerAddresses,
				toSharesUint,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
