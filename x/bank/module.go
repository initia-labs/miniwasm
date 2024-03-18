package bank

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
	types "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/initia-labs/miniwasm/x/bank/keeper"
)

const ConsensusVersion = 1

// AppModule implements an application module for the bank module.
type AppModule struct {
	Keeper *keeper.Keeper
	bank.AppModule
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper *keeper.Keeper, accountKeeper types.AccountKeeper, ss exported.Subspace) AppModule {
	return AppModule{
		Keeper:    keeper,
		AppModule: bank.NewAppModule(cdc, keeper.BaseKeeper, accountKeeper, ss),
	}
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.Keeper))
	types.RegisterQueryServer(cfg.QueryServer(), am.Keeper.BaseKeeper)
}

// ConsensusVersion implements ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }
