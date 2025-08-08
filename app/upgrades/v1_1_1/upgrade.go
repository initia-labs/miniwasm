package v1_1_1

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/initia-labs/miniwasm/app/upgrades"
)

const upgradeName = "v1.1.1"

// RegisterUpgradeHandlers returns upgrade handlers
func RegisterUpgradeHandlers(app upgrades.MinitiaApp) {
	app.GetUpgradeKeeper().SetUpgradeHandler(
		upgradeName,
		func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
			return app.GetModuleManager().RunMigrations(ctx, app.GetConfigurator(), vm)
		},
	)
}
