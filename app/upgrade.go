package app

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

const upgradeName = "0.6.0"

// RegisterUpgradeHandlers returns upgrade handlers
func (app *MinitiaApp) RegisterUpgradeHandlers(cfg module.Configurator) {
	app.UpgradeKeeper.SetUpgradeHandler(upgradeName,
		func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			return fromVM, nil
		},
	)
}
