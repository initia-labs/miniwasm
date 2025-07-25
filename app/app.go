package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/version"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"

	// initia imports
	"github.com/initia-labs/initia/app/params"
	cryptocodec "github.com/initia-labs/initia/crypto/codec"

	// skip imports
	blockchecktx "github.com/skip-mev/block-sdk/v2/abci/checktx"
	"github.com/skip-mev/block-sdk/v2/block"
	blockservice "github.com/skip-mev/block-sdk/v2/block/service"

	// CosmWasm imports
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	// local imports
	"github.com/initia-labs/miniwasm/app/keepers"
	"github.com/initia-labs/miniwasm/app/upgrades/v1_1_0"

	// kvindexer
	kvindexermodule "github.com/initia-labs/kvindexer/x/kvindexer"
	kvindexerkeeper "github.com/initia-labs/kvindexer/x/kvindexer/keeper"

	// unnamed import of statik for swagger UI support
	_ "github.com/initia-labs/miniwasm/client/docs/statik"
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string
)

var (
	_ servertypes.Application = (*MinitiaApp)(nil)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		tmos.Exit(err.Error())
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+AppName)
}

// MinitiaApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type MinitiaApp struct {
	*baseapp.BaseApp
	keepers.AppKeepers

	// address codecs
	ac, vc, cc address.Codec

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	// the module manager
	ModuleManager      *module.Manager
	BasicModuleManager module.BasicManager

	// the configurator
	configurator module.Configurator

	// Override of BaseApp's CheckTx
	checkTxHandler blockchecktx.CheckTx

	// indexer keeper for graceful shutdown
	kvIndexerKeeper *kvindexerkeeper.Keeper
	// indexer module for grpc-gateway registration
	kvIndexerModule *kvindexermodule.AppModuleBasic
}

// NewMinitiaApp returns a reference to an initialized Initia.
func NewMinitiaApp(
	logger log.Logger,
	db dbm.DB,
	kvindexerDB dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	wasmOpts []wasmkeeper.Option,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *MinitiaApp {
	// load the configs
	mempoolMaxTxs := cast.ToInt(appOpts.Get(server.FlagMempoolMaxTxs))
	queryGasLimit := cast.ToInt(appOpts.Get(server.FlagQueryGasLimit))

	logger.Info("mempool max txs", "max_txs", mempoolMaxTxs)
	logger.Info("query gas limit", "gas_limit", queryGasLimit)

	encodingConfig := params.MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	cryptocodec.RegisterLegacyAminoCodec(encodingConfig.Amino)
	cryptocodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	appCodec := encodingConfig.Codec
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry
	txConfig := encodingConfig.TxConfig

	bApp := baseapp.NewBaseApp(AppName, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	// app opts
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))
	invCheckPeriod := cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod))
	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	app := &MinitiaApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,

		// codecs
		ac: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		vc: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		cc: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}

	i := 0
	moduleAddrs := make([]sdk.AccAddress, len(maccPerms))
	for name := range maccPerms {
		moduleAddrs[i] = authtypes.NewModuleAddress(name)
		i += 1
	}

	moduleAccountAddresses := app.ModuleAccountAddrs()
	blockedModuleAccountAddrs := app.BlockedModuleAccountAddrs(moduleAccountAddresses)

	wasmConfig, err := wasm.ReadNodeConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	// Setup keepers
	app.AppKeepers = keepers.NewAppKeeper(
		app.ac, app.vc, app.cc,
		appCodec,
		txConfig,
		bApp,
		legacyAmino,
		maccPerms,
		blockedModuleAccountAddrs,
		skipUpgradeHeights,
		homePath,
		invCheckPeriod,
		logger,
		wasmConfig,
		wasmOpts,
		appOpts,
	)

	/****  Module Options ****/

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.ModuleManager = module.NewManager(appModules(app, skipGenesisInvariants)...)

	// BasicModuleManager defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration and genesis verification.
	// By default it is composed of all the module from the module manager.
	// Additionally, app module basics can be overwritten by passing them as argument.
	app.BasicModuleManager = newBasicManagerFromManager(app)

	// NOTE: upgrade module is required to be prioritized
	app.ModuleManager.SetOrderPreBlockers(
		upgradetypes.ModuleName,
	)

	// set order of module operations
	app.ModuleManager.SetOrderBeginBlockers(orderBeginBlockers()...)
	app.ModuleManager.SetOrderEndBlockers(orderEndBlockers()...)
	genesisModuleOrder := orderInitBlockers()
	app.ModuleManager.SetOrderInitGenesis(genesisModuleOrder...)
	app.ModuleManager.SetOrderExportGenesis(genesisModuleOrder...)

	// register invariants for crisis module
	app.ModuleManager.RegisterInvariants(app.CrisisKeeper)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	err = app.ModuleManager.RegisterServices(app.configurator)
	if err != nil {
		tmos.Exit(err.Error())
	}

	// setup indexer
	if kvIndexerKeeper, kvIndexerModule, streamingManager, err := setupIndexer(app, appOpts, encodingConfig, kvindexerDB); err != nil {
		tmos.Exit(err.Error())
	} else if kvIndexerKeeper != nil && kvIndexerModule != nil && streamingManager != nil {
		// register kvindexer keeper and module, and register services
		app.SetKVIndexer(kvIndexerKeeper, kvIndexerModule)

		// override base-app's streaming manager
		app.SetStreamingManager(*streamingManager)
	}

	// Only register upgrade handlers when loading the latest version of the app.
	// This optimization skips unnecessary handler registration during app initialization.
	//
	// The cosmos upgrade handler attempts to create ${HOME}/.minitia/data to check for upgrade info,
	// but this isn't required during initial encoding config setup.
	if loadLatest {
		v1_1_0.RegisterUpgradeHandlers(app)
	}

	// register executor change plans for later use
	err = app.RegisterExecutorChangePlans()
	if err != nil {
		tmos.Exit(err.Error())
	}

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.ModuleManager.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		tmos.Exit(err.Error())
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// initialize stores
	app.MountKVStores(app.GetKVStoreKey())
	app.MountTransientStores(app.GetTransientStoreKey())
	app.MountMemoryStores(app.GetMemoryStoreKey())

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.setPostHandler()
	app.SetEndBlocker(app.EndBlocker)

	// setup BlockSDK
	mempool, anteHandler, checkTx, prepareProposalHandler, processProposalHandler, err := setupBlockSDK(app, mempoolMaxTxs, wasmConfig, app.GetKVStoreKey()[wasmtypes.StoreKey])
	if err != nil {
		tmos.Exit(err.Error())
	}

	// override base-app's mempool
	app.SetMempool(mempool)

	// override base-app's ante handler
	app.SetAnteHandler(anteHandler)

	// override base-app's ProcessProposal + PrepareProposal
	app.SetPrepareProposal(prepareProposalHandler)
	app.SetProcessProposal(processProposalHandler)

	// override base-app's CheckTx
	app.SetCheckTx(checkTx)

	// At startup, after all modules have been registered, check that all proto
	// annotations are correct.
	protoFiles, err := proto.MergedRegistry()
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = msgservice.ValidateProtoAnnotations(protoFiles)
	if err != nil {
		errMsg := ""

		// ignore injective proto annotations comes from github.com/cosoms/relayer
		for _, s := range strings.Split(err.Error(), "\n") {
			if strings.Contains(s, "injective") {
				continue
			}

			errMsg += s + "\n"
		}

		if errMsg != "" {
			// Once we switch to using protoreflect-based antehandlers, we might
			// want to panic here instead of logging a warning.
			fmt.Fprintln(os.Stderr, errMsg)
		}
	}

	// must be before Loading version
	// requires the snapshot store to be created and registered as a BaseAppOption
	// see cmd/wasmd/root.go: 206 - 214 approx
	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), app.WasmKeeper),
		)
		if err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %s", err))
		}
	}

	// Load the latest state from disk if necessary, and initialize the base-app. From this point on
	// no more modifications to the base-app can be made
	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}

		ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})

		// Initialize pinned codes in wasmvm as they are not persisted there
		if err := app.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
			tmos.Exit(fmt.Sprintf("failed initialize pinned codes %s", err))
		}
	}

	return app
}

// CheckTx will check the transaction with the provided checkTxHandler. We override the default
// handler so that we can verify bid transactions before they are inserted into the mempool.
// With the POB CheckTx, we can verify the bid transaction and all of the bundled transactions
// before inserting the bid transaction into the mempool.
func (app *MinitiaApp) CheckTx(req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	return app.checkTxHandler(req)
}

// SetCheckTx sets the checkTxHandler for the app.
func (app *MinitiaApp) SetCheckTx(handler blockchecktx.CheckTx) {
	app.checkTxHandler = handler
}

func (app *MinitiaApp) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		tmos.Exit(err.Error())
	}

	app.SetPostHandler(postHandler)
}

// SetKVIndexer sets the kvindexer keeper and module for the app and registers the services.
func (app *MinitiaApp) SetKVIndexer(kvIndexerKeeper *kvindexerkeeper.Keeper, kvIndexerModule *kvindexermodule.AppModuleBasic) {
	app.kvIndexerKeeper = kvIndexerKeeper
	app.kvIndexerModule = kvIndexerModule
	app.kvIndexerModule.RegisterServices(app.configurator)
}

// Name returns the name of the App
func (app *MinitiaApp) Name() string { return app.BaseApp.Name() }

// PreBlocker application updates every pre block
func (app *MinitiaApp) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.ModuleManager.PreBlock(ctx)
}

// BeginBlocker application updates every begin block
func (app *MinitiaApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *MinitiaApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

// InitChainer application update at chain initialization
func (app *MinitiaApp) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		tmos.Exit(err.Error())
	}
	if err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap()); err != nil {
		tmos.Exit(err.Error())
	}
	return app.ModuleManager.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *MinitiaApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *MinitiaApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		addrStr, _ := app.ac.BytesToString(authtypes.NewModuleAddress(acc).Bytes())
		modAccAddrs[addrStr] = true
	}

	return modAccAddrs
}

// BlockedModuleAccountAddrs returns all the app's blocked module account
// addresses.
func (app *MinitiaApp) BlockedModuleAccountAddrs(modAccAddrs map[string]bool) map[string]bool {
	modules := []string{}

	// remove module accounts that are ALLOWED to received funds
	for _, module := range modules {
		moduleAddr, _ := app.ac.BytesToString(authtypes.NewModuleAddress(module).Bytes())
		delete(modAccAddrs, moduleAddr)
	}

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *MinitiaApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns Initia's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *MinitiaApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns Initia's InterfaceRegistry
func (app *MinitiaApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *MinitiaApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx

	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new tendermint queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register the Block SDK mempool API routes.
	blockservice.RegisterGRPCGatewayRoutes(apiSvr.ClientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	app.BasicModuleManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for indexer module.
	if app.kvIndexerModule != nil {
		app.kvIndexerModule.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	}

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(apiSvr.Router)
	}
}

// Simulate customize gas simulation to add fee deduction gas amount.
func (app *MinitiaApp) Simulate(txBytes []byte) (sdk.GasInfo, *sdk.Result, error) {
	gasInfo, result, err := app.BaseApp.Simulate(txBytes)
	gasInfo.GasUsed += FeeDeductionGasAmount
	return gasInfo, result, err
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *MinitiaApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(
		app.BaseApp.GRPCQueryRouter(), clientCtx,
		app.Simulate, app.interfaceRegistry,
	)

	// Register the Block SDK mempool transaction service.
	mempool, ok := app.Mempool().(block.Mempool)
	if !ok {
		panic("mempool is not a block.Mempool")
	}

	blockservice.RegisterMempoolService(app.GRPCQueryRouter(), mempool)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *MinitiaApp) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry, app.Query,
	)
}

func (app *MinitiaApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		tmos.Exit(err.Error())
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// Close closes the underlying baseapp, the oracle service, and the prometheus server if required.
// This method blocks on the closure of both the prometheus server, and the oracle-service
func (app *MinitiaApp) Close() error {
	if app.kvIndexerKeeper != nil {
		if err := app.kvIndexerKeeper.Close(); err != nil {
			return err
		}
	}

	if err := app.BaseApp.Close(); err != nil {
		return err
	}

	return nil
}
