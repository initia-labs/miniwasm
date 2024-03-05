package keepers

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
	types "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// AppModule implements an application module for the bank module.
type AppModule struct {
	Keeper *BaseKeeper
	bank.AppModule
}

// NewAppModule creates a new AppModule object
func NewBankAppModule(cdc codec.Codec, keeper *BaseKeeper, accountKeeper types.AccountKeeper, ss exported.Subspace) AppModule {
	return AppModule{
		Keeper:    keeper,
		AppModule: bank.NewAppModule(cdc, keeper.BaseKeeper, accountKeeper, ss),
	}
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), NewBankMsgServerImpl(am.Keeper))
	types.RegisterQueryServer(cfg.QueryServer(), am.Keeper.BaseKeeper)
}
