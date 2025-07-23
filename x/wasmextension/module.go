package wasmextension

import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/initia-labs/miniwasm/x/wasmextension/client/cli"
	"github.com/initia-labs/miniwasm/x/wasmextension/keeper"
	"github.com/initia-labs/miniwasm/x/wasmextension/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	opchildkeeper "github.com/initia-labs/OPinit/x/opchild/keeper"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the wasm module.
type AppModuleBasic struct{}

// RegisterGRPCGatewayRoutes implements module.AppModuleBasic.
func (b AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {
	panic("unimplemented")
}

func (b AppModuleBasic) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(amino)
}

// Name returns the wasm module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// GetTxCmd returns the root tx command for the wasm module.
func (b AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// RegisterInterfaces implements InterfaceModule
func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// ____________________________________________________________________________
var _ appmodule.AppModule = AppModule{}

// AppModule implements an application module for the wasm module.
type AppModule struct {
	AppModuleBasic
	cdc           codec.Codec
	keeper        *wasmkeeper.Keeper
	opchildKeeper *opchildkeeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec,
	keeper *wasmkeeper.Keeper,
	opchildKeeper *opchildkeeper.Keeper,
) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		cdc:            cdc,
		keeper:         keeper,
		opchildKeeper:  opchildKeeper,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() { // marker
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() { // marker
}

// ConsensusVersion is a sequence number for state-breaking change of the
// module. It should be incremented on each consensus-breaking change
// introduced by the module. To avoid wrong/empty versions, the initial version
// should be set to 1.
func (AppModule) ConsensusVersion() uint64 { return 4 }

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper, am.opchildKeeper))
}

// RegisterInvariants registers the wasm module invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}
