package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	m "github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"

	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

const (
	NewMaxBytes = 5 * 1024 * 1024 // 5MB
)

// generate upgrade version from the current version (v999999.999999.999999 => v999999)
func generateUpgradeVersion() string {
	currentVersion := version.Version
	// if current version empty then override it with localnet version
	if currentVersion == "v" {
		currentVersion = "v999999.999999.999999"
	}
	parts := strings.Split(currentVersion, ".")
	// Needed for devnet
	if len(parts) == 1 {
		return currentVersion
	}
	if len(parts) != 3 {
		panic(fmt.Sprintf("Invalid version format: %s. Expected format: vX.Y.Z", currentVersion))
	}
	majorVersion := strings.TrimPrefix(parts[0], "v")
	minorVersion := parts[1]
	// required for testnet
	patchParts := strings.Split(parts[2], "-")
	rcVersion := ""
	if len(patchParts) > 1 {
		rcVersion = strings.Join(patchParts[1:], "-")
	}
	// testnet
	if rcVersion != "" {
		if minorVersion != "0" && minorVersion != "999999" {
			return fmt.Sprintf("v%s.%s-%s", majorVersion, minorVersion, rcVersion)
		}
		return fmt.Sprintf("v%s-%s", majorVersion, rcVersion)
	}
	if minorVersion != "0" && minorVersion != "999999" {
		return fmt.Sprintf("v%s.%s", majorVersion, parts[1])
	}
	return fmt.Sprintf("v%s", majorVersion)
}

func (app *ElysApp) setUpgradeHandler() {
	upgradeVersion := generateUpgradeVersion()
	app.Logger().Info("Current version", "version", version.Version)
	app.Logger().Info("Upgrade version", "version", upgradeVersion)
	
	// Create a special handler that ensures EVM params are set
	app.UpgradeKeeper.SetUpgradeHandler(
		upgradeVersion,
		func(goCtx context.Context, plan upgradetypes.Plan, vm m.VersionMap) (m.VersionMap, error) {
			ctx := sdk.UnwrapSDKContext(goCtx)
			app.Logger().Info("Running upgrade handler for " + upgradeVersion)
			
			// Force EVM parameter setting
			app.forceSetEVMParams(ctx)
			
			return app.runStandardUpgrade(ctx, plan, vm)
		},
	)
}

// forceSetEVMParams ensures EVM parameters are properly configured
func (app *ElysApp) forceSetEVMParams(ctx sdk.Context) {
	defer func() {
		if r := recover(); r != nil {
			app.Logger().Error("Failed to set EVM params", "error", r)
		}
	}()
	
	app.Logger().Info("Checking EVM parameters configuration")
	currentParams := app.EVMKeeper.GetParams(ctx)
	if currentParams.EvmDenom == "" {
		app.Logger().Info("Setting EVM parameters - evm_denom is empty")
		evmParams := evmtypes.DefaultParams()
		evmParams.EvmDenom = "uelys"
		app.EVMKeeper.SetParams(ctx, evmParams)
		app.Logger().Info("EVM parameters set successfully", "evm_denom", "uelys")
	} else {
		app.Logger().Info("EVM parameters already configured", "evm_denom", currentParams.EvmDenom)
	}
}

// runStandardUpgrade runs the standard upgrade logic
func (app *ElysApp) runStandardUpgrade(ctx sdk.Context, plan upgradetypes.Plan, vm m.VersionMap) (m.VersionMap, error) {
	// Initialize EVM module versions if not present (for upgrade from non-EVM chain)
	if _, ok := vm[evmtypes.ModuleName]; !ok {
		app.Logger().Info("Initializing EVM module version")
		vm[evmtypes.ModuleName] = 1
	}
	
	if _, ok := vm[feemarkettypes.ModuleName]; !ok {
		app.Logger().Info("Initializing FeeMarket module version") 
		vm[feemarkettypes.ModuleName] = 1
		
		// Set default FeeMarket params for new module
		feeMarketParams := feemarkettypes.DefaultParams()
		// Disable base fee mechanism initially to prevent nil pointer issues
		feeMarketParams.NoBaseFee = true
		feeMarketParams.EnableHeight = 1
		app.FeeMarketKeeper.SetParams(ctx, feeMarketParams)
		
		// Initialize base fee state to prevent nil pointer in EndBlock
		app.FeeMarketKeeper.SetBlockGasWanted(ctx, 0)
	}

	vm, vmErr := app.mm.RunMigrations(ctx, app.configurator, vm)

	// Skip native token wrapper setup for now - focus on fixing eth_getBalance first
	app.Logger().Info("Native token handling will be done via balance override")

	for _, profile := range app.AssetprofileKeeper.GetAllEntry(ctx) {
		if profile.DisplayName == "WBTC" || profile.DisplayName == "wBTC" {
			profile.DisplayName = "BTC"
		}
		if profile.DisplayName == "WETH" || profile.DisplayName == "wETH" {
			profile.DisplayName = "ETH"
		}
		app.AssetprofileKeeper.SetEntry(ctx, profile)
	}

	for _, assetInfo := range app.LegacyOracleKeepper.GetAllAssetInfo(ctx) {
		if assetInfo.Display == "WBTC" || assetInfo.Display == "wBTC" {
			assetInfo.Display = "BTC"
			assetInfo.BandTicker = "BTC"
			assetInfo.ElysTicker = "BTC"
		}
		if assetInfo.Display == "WETH" || assetInfo.Display == "wETH" {
			assetInfo.Display = "ETH"
			assetInfo.BandTicker = "ETH"
			assetInfo.ElysTicker = "ETH"
		}
		app.LegacyOracleKeepper.SetAssetInfo(ctx, assetInfo)
	}

	for _, price := range app.LegacyOracleKeepper.GetAllAssetPrice(ctx, "WBTC") {
		price.Asset = "BTC"
		app.LegacyOracleKeepper.SetPrice(ctx, price)
	}

	for _, price := range app.LegacyOracleKeepper.GetAllAssetPrice(ctx, "WETH") {
		price.Asset = "ETH"
		app.LegacyOracleKeepper.SetPrice(ctx, price)
	}

	oracleParams := app.OracleKeeper.GetParams(ctx)
	if len(oracleParams.MandatoryList) == 0 {
		err := app.ojoOracleMigration(ctx, plan.Height+1)
		if err != nil {
			return nil, err
		}
	}

	return vm, vmErr
}

// checkAndSetEVMParams ensures EVM parameters are set on startup
func (app *ElysApp) checkAndSetEVMParams() {
	// This runs at startup, so we don't have a context yet
	// We'll defer this check to the first BeginBlocker where we have a proper context
	app.Logger().Info("EVM parameter check scheduled for first block")
}

// ensureEVMParams ensures EVM parameters are set, called from BeginBlocker
func (app *ElysApp) ensureEVMParams(ctx sdk.Context) {
	app.Logger().Info("Checking EVM parameters in BeginBlocker", "height", ctx.BlockHeight())
	currentParams := app.EVMKeeper.GetParams(ctx)
	
	// Force correct EVM denomination for unified balance system
	if currentParams.EvmDenom != "uelys" {
		app.Logger().Info("Setting EVM parameters - correcting evm_denom", "old", currentParams.EvmDenom, "new", "uelys", "height", ctx.BlockHeight())
		evmParams := currentParams // Keep existing params
		evmParams.EvmDenom = "uelys" // Force correct denomination
		app.EVMKeeper.SetParams(ctx, evmParams)
		app.Logger().Info("EVM parameters corrected successfully in BeginBlocker", "evm_denom", "uelys", "height", ctx.BlockHeight())
	} else {
		app.Logger().Info("EVM parameters already configured correctly in BeginBlocker", "evm_denom", currentParams.EvmDenom, "height", ctx.BlockHeight())
	}
}

func (app *ElysApp) setUpgradeStore() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("Failed to read upgrade info from disk: %v", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	app.Logger().Debug("Upgrade info", "info", upgradeInfo)

	// Always check if we need to add EVM stores - even without formal upgrade
	storeUpgrades := storetypes.StoreUpgrades{
		Added: []string{evmtypes.StoreKey, feemarkettypes.StoreKey},
		//Renamed: []storetypes.StoreRename{},
		//Deleted: []string{ratelimittypes.StoreKey},
	}
	
	// Check if EVM stores exist, if not, we need to add them
	if app.needsEVMStoreUpgrade() {
		app.Logger().Info("Adding EVM stores to existing chain state")
		app.Logger().Info(fmt.Sprintf("Setting store loader with store upgrades: %+v\n", storeUpgrades))
		
		// Use last committed height for the upgrade
		lastHeight := app.LastBlockHeight()
		if lastHeight <= 0 {
			lastHeight = 1 // Genesis case
		}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(lastHeight, &storeUpgrades))
	} else if shouldLoadUpgradeStore(app, upgradeInfo) {
		app.Logger().Info(fmt.Sprintf("Setting store loader with height %d and store upgrades: %+v\n", upgradeInfo.Height, storeUpgrades))
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	} else {
		app.Logger().Debug("No need to load upgrade store.")
	}
}

// needsEVMStoreUpgrade checks if EVM stores need to be added to existing chain
func (app *ElysApp) needsEVMStoreUpgrade() bool {
	// Check if this is first run with EVM on existing chain
	// We use a simple file-based check
	
	// Get home directory
	home := os.Getenv("HOME")
	if home == "" {
		home = "."
	}
	
	evmInitFile := filepath.Join(home, ".elys", "evm_initialized")
	
	// Check if file exists
	if _, err := os.Stat(evmInitFile); err == nil {
		// File exists, EVM already initialized
		return false
	}
	
	// File doesn't exist and we have existing chain data - need EVM upgrade
	currentHeight := app.LastBlockHeight()
	needsUpgrade := currentHeight > 0
	
	if needsUpgrade {
		// Create the file to mark EVM as initialized
		os.WriteFile(evmInitFile, []byte("EVM stores initialized"), 0644)
	}
	
	return needsUpgrade
}

func shouldLoadUpgradeStore(app *ElysApp, upgradeInfo upgradetypes.Plan) bool {
	currentHeight := app.LastBlockHeight()
	app.Logger().Debug(fmt.Sprintf("Current block height: %d, Upgrade height: %d\n", currentHeight, upgradeInfo.Height))
	upgradeVersion := generateUpgradeVersion()
	app.Logger().Debug("Current version", "version", version.Version)
	app.Logger().Debug("Upgrade version", "version", upgradeVersion)
	return upgradeInfo.Name == upgradeVersion && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height)
}
