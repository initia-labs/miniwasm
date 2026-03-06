package upgrades

import (
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	opchildkeeper "github.com/initia-labs/OPinit/x/opchild/keeper"
)

type MinitiaApp interface {
	GetAccountKeeper() *authkeeper.AccountKeeper
	GetWasmKeeper() *wasmkeeper.Keeper
	GetUpgradeKeeper() *upgradekeeper.Keeper
	GetOPChildKeeper() *opchildkeeper.Keeper

	GetConfigurator() module.Configurator
	GetModuleManager() *module.Manager
	SetStoreLoader(loader baseapp.StoreLoader)
}
