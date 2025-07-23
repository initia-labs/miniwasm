package wasmextension

import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/initia-labs/miniwasm/x/wasmextension/keeper"
	"github.com/initia-labs/miniwasm/x/wasmextension/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the wasm module.
type AppModuleBasic struct{}

// RegisterGRPCGatewayRoutes implements module.AppModuleBasic.
func (b AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {

}

func (b AppModuleBasic) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(amino)
}

// Name returns the wasm module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
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
	cdc       codec.Codec
	keeper    *wasmkeeper.Keeper
	authority string
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec,
	keeper *wasmkeeper.Keeper,
	authority string,
) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		cdc:            cdc,
		keeper:         keeper,
		authority:      authority,
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
func (AppModule) ConsensusVersion() uint64 { return 1 }

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper, am.authority))
}

// RegisterInvariants registers the wasm module invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}
