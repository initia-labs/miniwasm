package ante

import (
	corestoretypes "cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	txsigning "cosmossdk.io/x/tx/signing"
	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"

	opchildante "github.com/initia-labs/OPinit/x/opchild/ante"
	opchildkeeper "github.com/initia-labs/OPinit/x/opchild/keeper"
	"github.com/initia-labs/initia/app/ante/sigverify"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// HandlerOptions extends the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
	ante.HandlerOptions
	Codec         codec.BinaryCodec
	IBCkeeper     *ibckeeper.Keeper
	OPChildKeeper *opchildkeeper.Keeper

	// wasm ante options
	WasmKeeper            *wasmkeeper.Keeper
	WasmConfig            *wasmtypes.NodeConfig
	TXCounterStoreService corestoretypes.KVStoreService
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}
	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}
	if options.SignModeHandler == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}
	if options.WasmConfig == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "wasm config is required for ante builder")
	}
	if options.WasmKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "wasm keeper is required for ante builder")
	}
	if options.TXCounterStoreService == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "wasm store service is required for ante builder")
	}
	if options.OPChildKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "opchild keeper is required for ante builder")
	}
	if options.IBCkeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "IBC keeper is required for ante builder")
	}

	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = sigverify.DefaultSigVerificationGasConsumer
	}

	txFeeChecker := options.TxFeeChecker
	if txFeeChecker == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "tx fee checker is required for ante builder")
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		// NOTE - WASM simulation gas limit can affect other module messages.
		wasmkeeper.NewLimitSimulationGasDecorator(options.WasmConfig.SimulationGasLimit),
		wasmkeeper.NewCountTXDecorator(options.TXCounterStoreService),
		wasmkeeper.NewGasRegisterDecorator(options.WasmKeeper.GetGasRegister()),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, txFeeChecker),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		sigverify.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCkeeper),
		opchildante.NewRedundantBridgeDecorator(options.OPChildKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}

func CreateAnteHandlerForOPinit(ak ante.AccountKeeper, signModeHandler *txsigning.HandlerMap) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewSetPubKeyDecorator(ak),
		ante.NewValidateSigCountDecorator(ak),
		ante.NewSigGasConsumeDecorator(ak, sigverify.DefaultSigVerificationGasConsumer),
		sigverify.NewSigVerificationDecorator(ak, signModeHandler),
		ante.NewIncrementSequenceDecorator(ak),
	)
}

// NewDualAnteHandler returns an AnteHandler that routes to the minimal handler
// during CheckTx/ReCheckTx and to the full handler otherwise.
func NewDualAnteHandler(minimal, full sdk.AnteHandler) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		if ctx.IsCheckTx() || ctx.IsReCheckTx() {
			return minimal(ctx, tx, simulate)
		}
		return full(ctx, tx, simulate)
	}
}
