package app

import (
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgrades "github.com/public-awesome/stargaze/v14/app/upgrades"
	mainnetupgradesv14 "github.com/public-awesome/stargaze/v14/app/upgrades/mainnet/v14"
)

var Upgrades = []upgrades.Upgrade{
	mainnetupgradesv14.Upgrade,
}

func (app App) RegisterUpgradeHandlers(configurator module.Configurator) {
	for _, u := range Upgrades {
		app.Keepers.UpgradeKeeper.SetUpgradeHandler(
			u.UpgradeName,
			u.CreateUpgradeHandler(app.ModuleManager, configurator, app.Keepers),
		)
	}

	upgradeInfo, err := app.Keepers.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if app.Keepers.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	for _, u := range Upgrades {
		u := u
		if upgradeInfo.Name == u.UpgradeName {
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &u.StoreUpgrades))
		}
	}
}
