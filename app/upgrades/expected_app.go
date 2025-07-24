package upgrades

import (
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"

	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
)

type MinitiaApp interface {
	GetAccountKeeper() *authkeeper.AccountKeeper
	GetWasmKeeper() *wasmkeeper.Keeper
	GetUpgradeKeeper() *upgradekeeper.Keeper

	GetConfigurator() module.Configurator
	GetModuleManager() *module.Manager
}
