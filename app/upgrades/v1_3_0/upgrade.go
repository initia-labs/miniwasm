package v1_3_0

import (
	"context"

	tmos "github.com/cometbft/cometbft/libs/os"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/initia-labs/miniwasm/app/upgrades"
)

const upgradeName = "v1.3.0"

// RegisterUpgradeHandlers returns upgrade handlers
func RegisterUpgradeHandlers(app upgrades.MinitiaApp) {
	// apply store upgrade only if this upgrade is scheduled at a height
	if upgradeInfo, err := app.GetUpgradeKeeper().ReadUpgradeInfoFromDisk(); err == nil {
		if upgradeInfo.Name == upgradeName && !app.GetUpgradeKeeper().IsSkipHeight(upgradeInfo.Height) {
			// capability + feeibc were removed in ibc-go v10, crisis was removed in
			// cosmos-sdk v0.53.
			storeUpgrades := storetypes.StoreUpgrades{
				Deleted: []string{"auction", "capability", "crisis", "feeibc"},
			}

			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
		}
	} else {
		tmos.Exit(err.Error())
	}

	app.GetUpgradeKeeper().SetUpgradeHandler(
		upgradeName,
		func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
			return app.GetModuleManager().RunMigrations(ctx, app.GetConfigurator(), vm)
		},
	)
}
