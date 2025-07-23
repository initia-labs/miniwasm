package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	"github.com/initia-labs/miniwasm/x/wasmextension/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

const (
	flagInstantiateByEverybody    = "instantiate-everybody"
	flagInstantiateNobody         = "instantiate-nobody"
	flagInstantiateByAddress      = "instantiate-only-address"
	flagInstantiateByAnyOfAddress = "instantiate-anyof-addresses"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Wasm extension transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
		SilenceUsage:               true,
	}
	txCmd.AddCommand(
		StoreCodeAdminCmd(),
	)
	return txCmd
}

// StoreCodeAdminCmd will upload code to be reused.
func StoreCodeAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "store-admin [wasm file]",
		Short:   "Upload a wasm binary with admin permission",
		Aliases: []string{"upload", "st", "s"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg, err := parseStoreCodeArgs(args[0], clientCtx.GetFromAddress().String(), cmd.Flags())
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
		SilenceUsage: true,
	}

	addInstantiatePermissionFlags(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// Prepares MsgStoreCode object from flags with gzipped wasm byte code field
func parseStoreCodeArgs(file, sender string, flags *flag.FlagSet) (types.MsgStoreCodeAdmin, error) {
	wasm, err := os.ReadFile(file)
	if err != nil {
		return types.MsgStoreCodeAdmin{}, err
	}

	// gzip the wasm file
	if ioutils.IsWasm(wasm) {
		wasm, err = ioutils.GzipIt(wasm)

		if err != nil {
			return types.MsgStoreCodeAdmin{}, err
		}
	} else if !ioutils.IsGzip(wasm) {
		return types.MsgStoreCodeAdmin{}, fmt.Errorf("invalid input file. Use wasm binary or gzip")
	}

	perm, err := parseAccessConfigFlags(flags)
	if err != nil {
		return types.MsgStoreCodeAdmin{}, err
	}

	msg := types.MsgStoreCodeAdmin{
		Sender:                sender,
		WASMByteCode:          wasm,
		InstantiatePermission: perm,
	}
	return msg, msg.ValidateBasic()
}

func parseAccessConfigFlags(flags *flag.FlagSet) (*wasmtypes.AccessConfig, error) {
	addrs, err := flags.GetStringSlice(flagInstantiateByAnyOfAddress)
	if err != nil {
		return nil, fmt.Errorf("flag any of: %s", err)
	}
	if len(addrs) != 0 {
		acceptedAddrs := make([]sdk.AccAddress, len(addrs))
		for i, v := range addrs {
			acceptedAddrs[i], err = sdk.AccAddressFromBech32(v)
			if err != nil {
				return nil, fmt.Errorf("parse %q: %w", v, err)
			}
		}
		x := wasmtypes.AccessTypeAnyOfAddresses.With(acceptedAddrs...)
		return &x, nil
	}

	onlyAddrStr, err := flags.GetString(flagInstantiateByAddress)
	if err != nil {
		return nil, fmt.Errorf("instantiate by address: %s", err)
	}
	if onlyAddrStr != "" {
		return nil, fmt.Errorf("not supported anymore. Use: %s", flagInstantiateByAnyOfAddress)
	}
	everybodyStr, err := flags.GetString(flagInstantiateByEverybody)
	if err != nil {
		return nil, fmt.Errorf("instantiate by everybody: %s", err)
	}
	if everybodyStr != "" {
		ok, err := strconv.ParseBool(everybodyStr)
		if err != nil {
			return nil, fmt.Errorf("boolean value expected for instantiate by everybody: %s", err)
		}
		if ok {
			return &wasmtypes.AllowEverybody, nil
		}
	}

	nobodyStr, err := flags.GetString(flagInstantiateNobody)
	if err != nil {
		return nil, fmt.Errorf("instantiate by nobody: %s", err)
	}
	if nobodyStr != "" {
		ok, err := strconv.ParseBool(nobodyStr)
		if err != nil {
			return nil, fmt.Errorf("boolean value expected for instantiate by nobody: %s", err)
		}
		if ok {
			return &wasmtypes.AllowNobody, nil
		}
	}
	return nil, nil
}

func addInstantiatePermissionFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagInstantiateByEverybody, "", "Everybody can instantiate a contract from the code, optional")
	cmd.Flags().String(flagInstantiateNobody, "", "Nobody except the governance process can instantiate a contract from the code, optional")
	cmd.Flags().String(flagInstantiateByAddress, "", fmt.Sprintf("Removed: use %s instead", flagInstantiateByAnyOfAddress))
	cmd.Flags().StringSlice(flagInstantiateByAnyOfAddress, []string{}, "Any of the addresses can instantiate a contract from the code, optional")
}
