package app

import (
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	cosmosante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	opchildlanes "github.com/initia-labs/OPinit/x/opchild/lanes"
	initialanes "github.com/initia-labs/initia/app/lanes"

	blockabci "github.com/skip-mev/block-sdk/v2/abci"
	blockchecktx "github.com/skip-mev/block-sdk/v2/abci/checktx"
	signer_extraction "github.com/skip-mev/block-sdk/v2/adapters/signer_extraction_adapter"
	"github.com/skip-mev/block-sdk/v2/block"
	blockbase "github.com/skip-mev/block-sdk/v2/block/base"
	mevlane "github.com/skip-mev/block-sdk/v2/lanes/mev"

	appante "github.com/initia-labs/miniwasm/app/ante"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

func setupBlockSDK(
	app *MinitiaApp,
	mempoolMaxTxs int,
	wasmConfig wasmtypes.NodeConfig,
	txCounterStoreKey *storetypes.KVStoreKey,
) (
	mempool.Mempool,
	sdk.AnteHandler,
	blockchecktx.CheckTx,
	sdk.PrepareProposalHandler,
	sdk.ProcessProposalHandler,
	error,
) {

	// initialize and set the InitiaApp mempool. The current mempool will be the
	// x/auction module's mempool which will extract the top bid from the current block's auction
	// and insert the txs at the top of the block spots.
	signerExtractor := signer_extraction.NewDefaultAdapter()

	systemLane := initialanes.NewSystemLane(blockbase.LaneConfig{
		Logger:          app.Logger(),
		TxEncoder:       app.txConfig.TxEncoder(),
		TxDecoder:       app.txConfig.TxDecoder(),
		MaxBlockSpace:   math.LegacyMustNewDecFromStr("0.1"),
		MaxTxs:          1,
		SignerExtractor: signerExtractor,
	}, opchildlanes.SystemLaneMatchHandler())

	factory := mevlane.NewDefaultAuctionFactory(app.txConfig.TxDecoder(), signerExtractor)
	mevLane := initialanes.NewMEVLane(blockbase.LaneConfig{
		Logger:          app.Logger(),
		TxEncoder:       app.txConfig.TxEncoder(),
		TxDecoder:       app.txConfig.TxDecoder(),
		MaxBlockSpace:   math.LegacyMustNewDecFromStr("0.1"),
		MaxTxs:          100,
		SignerExtractor: signerExtractor,
	}, factory, factory.MatchHandler())

	freeLane := initialanes.NewFreeLane(blockbase.LaneConfig{
		Logger:          app.Logger(),
		TxEncoder:       app.txConfig.TxEncoder(),
		TxDecoder:       app.txConfig.TxDecoder(),
		MaxBlockSpace:   math.LegacyMustNewDecFromStr("0.1"),
		MaxTxs:          100,
		SignerExtractor: signerExtractor,
	}, opchildlanes.NewFreeLaneMatchHandler(app.ac, app.OPChildKeeper).MatchHandler())

	defaultLane := initialanes.NewDefaultLane(blockbase.LaneConfig{
		Logger:          app.Logger(),
		TxEncoder:       app.txConfig.TxEncoder(),
		TxDecoder:       app.txConfig.TxDecoder(),
		MaxBlockSpace:   math.LegacyMustNewDecFromStr("0.7"),
		MaxTxs:          mempoolMaxTxs,
		SignerExtractor: signerExtractor,
	})

	lanes := []block.Lane{systemLane, mevLane, freeLane, defaultLane}
	mempool, err := block.NewLanedMempool(app.Logger(), lanes)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	anteHandler, err := appante.NewAnteHandler(
		appante.HandlerOptions{
			HandlerOptions: cosmosante.HandlerOptions{
				AccountKeeper:   app.AccountKeeper,
				BankKeeper:      app.BankKeeper,
				FeegrantKeeper:  app.FeeGrantKeeper,
				SignModeHandler: app.txConfig.SignModeHandler(),
			},
			IBCkeeper:             app.IBCKeeper,
			Codec:                 app.appCodec,
			OPChildKeeper:         app.OPChildKeeper,
			TxEncoder:             app.txConfig.TxEncoder(),
			AuctionKeeper:         app.AuctionKeeper,
			MevLane:               mevLane,
			FreeLane:              freeLane,
			WasmKeeper:            app.WasmKeeper,
			WasmConfig:            &wasmConfig,
			TXCounterStoreService: runtime.NewKVStoreService(txCounterStoreKey),
		},
	)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	// set ante handler to lanes
	opt := []blockbase.LaneOption{
		blockbase.WithAnteHandler(anteHandler),
	}
	for _, lane := range lanes {
		if blane, ok := lane.(*blockbase.BaseLane); ok {
			blane.WithOptions(opt...)
		} else if mlane, ok := lane.(*mevlane.MEVLane); ok {
			mlane.WithOptions(opt...)
		}
	}

	mevCheckTx := blockchecktx.NewMEVCheckTxHandler(
		app.BaseApp,
		app.txConfig.TxDecoder(),
		mevLane,
		anteHandler,
		app.BaseApp.CheckTx,
	)
	checkTxHandler := blockchecktx.NewMempoolParityCheckTx(
		app.Logger(),
		mempool,
		app.txConfig.TxDecoder(),
		mevCheckTx.CheckTx(),
		app.BaseApp,
	)
	checkTx := checkTxHandler.CheckTx()

	proposalHandler := blockabci.New(
		app.Logger(),
		app.txConfig.TxDecoder(),
		app.txConfig.TxEncoder(),
		mempool,
		true,
	)

	prepareProposalHandler := proposalHandler.PrepareProposalHandler()
	processProposalHandler := proposalHandler.ProcessProposalHandler()

	return mempool, anteHandler, checkTx, prepareProposalHandler, processProposalHandler, nil
}
