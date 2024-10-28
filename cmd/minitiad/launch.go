package main

import (
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/initia-labs/OPinit/contrib/launchtools"
	"github.com/initia-labs/OPinit/contrib/launchtools/steps"

	"github.com/initia-labs/initia/app/params"
	minitiaapp "github.com/initia-labs/miniwasm/app"

	"github.com/initia-labs/miniwasm/contrib"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// DefaultLaunchStepFactories is a list of default launch step factories.
var DefaultLaunchStepFactories = []launchtools.LauncherStepFuncFactory[*launchtools.Config]{
	steps.InitializeConfig,
	steps.InitializeRPCHelpers,

	// Initialize genesis
	steps.InitializeGenesis,

	// Add system keys to the keyring
	steps.InitializeKeyring,

	// Run the app
	steps.RunApp,

	// MINIWASM: Store/Instantiate cw721 and ics721 contracts
	StoreAndInstantiateNFTContracts,

	// Establish IBC channels for fungible and NFT transfer
	// MINIWASM: Use wasm contract addresses for srcPort, dstPort, channelVersion

	steps.EstablishIBCChannelsWithNFTTransfer(func() (string, string, string) {
		return "wasm." + wasmkeeper.BuildContractAddressClassic(2, 1).String(),
			"nft-transfer",
			"ics721-1"
	}),

	// Create OP Bridge, using open channel states
	steps.InitializeOpBridge,

	// Set bridge info and update clients
	steps.SetBridgeInfo,

	// Get the L1 and L2 heights
	steps.GetL1Height,
	steps.GetL2Height,

	// Cleanup
	steps.StopApp,
}

func LaunchCommand(ac *appCreator, enc params.EncodingConfig, mbm module.BasicManager) *cobra.Command {
	return launchtools.LaunchCmd(
		ac,
		func(denom string) map[string]json.RawMessage {
			return minitiaapp.NewDefaultGenesisState(enc.Codec, mbm, denom)
		},
		DefaultLaunchStepFactories,
	)
}

// StoreAndInstantiateNFTContracts stores and instantiates cw721 and ics721 contracts
func StoreAndInstantiateNFTContracts(input *launchtools.Config) launchtools.LauncherStepFunc {
	return func(ctx launchtools.Launcher) error {
		ctx.Logger().Info("Storing and instantiating cw721 and ics721 contracts")
		fs := contrib.FS()
		cw721, err := fs.ReadFile("wasm/cw721_base.wasm")
		//cw721, err := os.ReadFile("contrib/wasm/cw721_base.wasm")
		if err != nil {
			return errors.Wrapf(err, "failed to read cw721_base.wasm")
		}

		ics721, err := fs.ReadFile("wasm/ics721_base.wasm")
		if err != nil {
			return errors.Wrapf(err, "failed to read ics721_base.wasm")
		}

		msgs := []sdk.Msg{
			&wasmtypes.MsgStoreCode{
				Sender:                input.SystemKeys.Validator.L2Address,
				WASMByteCode:          cw721,
				InstantiatePermission: nil,
			},
			&wasmtypes.MsgStoreCode{
				Sender:                input.SystemKeys.Validator.L2Address,
				WASMByteCode:          ics721,
				InstantiatePermission: nil,
			},
			&wasmtypes.MsgInstantiateContract{
				Sender: input.SystemKeys.Validator.L2Address,
				Admin:  input.SystemKeys.Validator.L2Address,
				CodeID: 2,
				Label:  "ics721",
				Msg:    []byte(`{"cw721_base_code_id":1}`),
				Funds:  nil,
			},
		}

		for i, msg := range msgs {
			ctx.Logger().Info(
				"Broadcasting tx to store and instantiate cw721 and ics721 contracts",
				"step", i+1,
			)

			res, err := ctx.GetRPCHelperL2().BroadcastTxAndWait(
				input.SystemKeys.Validator.L2Address,
				input.SystemKeys.Validator.Mnemonic,
				10000000,
				sdk.NewCoins(),
				msg,
			)

			if err != nil {
				return errors.Wrapf(err, "failed to store and instantiate nft contracts")
			}

			ctx.Logger().Info(
				"Successfully stored and instantiated cw721 and ics721 contracts",
				"tx_hash", res.Hash,
			)
		}

		return nil
	}
}
