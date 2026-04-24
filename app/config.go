package app

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

const (
	// Keep Cosmos standards (not Ethereum)
	ChainID      = "elys-local-dev"
	BaseDenom    = "uelys"
	DisplayDenom = "elys"
	CoinType     = 118 // Keep Cosmos coin type (not 60)
	Decimals     = 6   // Keep Cosmos decimals (not 18)
)

// EVMOptionsFn defines a function type for setting app options specifically for
// the app. The function should receive the chainID and return an error if any.
type EVMOptionsFn func(string) error

// NoOpEVMOptions is a no-op function that can be used when the app does not
// need any specific configuration.
func NoOpEVMOptions(_ string) error {
	return nil
}

var sealed = false

// ChainsCoinInfo is a map of the chain id and its corresponding EvmCoinInfo
// that allows initializing the app with different coin info based on the chain id
var ChainsCoinInfo = map[string]evmtypes.EvmCoinInfo{
	"elys-local-dev": {
		Denom:         BaseDenom,
		ExtendedDenom: BaseDenom + "extended", // Extended denom for fractional precision
		DisplayDenom:  DisplayDenom,
		Decimals:      Decimals,
	},
	"elys": {
		Denom:         BaseDenom,
		ExtendedDenom: BaseDenom + "extended",
		DisplayDenom:  DisplayDenom,
		Decimals:      Decimals,
	},
}

// EVMAppOptions configures EVM to read native ELYS from Cosmos bank
func EVMAppOptions(chainID string) error {
	if sealed {
		fmt.Printf("🔄 EVMAppOptions: already sealed, skipping\n")
		return nil
	}

	fmt.Printf("🚀 EVMAppOptions: configuring EVM for chain %s\n", chainID)

	if chainID == "" {
		chainID = ChainID
	}

	id := strings.Split(chainID, "-")[0]
	coinInfo, found := ChainsCoinInfo[id]
	if !found {
		coinInfo, found = ChainsCoinInfo[chainID]
		if !found {
			return fmt.Errorf("❌ unknown chain id: %s, available: %+v", chainID, ChainsCoinInfo)
		}
	}

	fmt.Printf("✅ Found coin info: %+v\n", coinInfo)

	// Set the denom info for the chain (crucial for EVM to read native tokens)
	if err := setBaseDenom(coinInfo); err != nil {
		return fmt.Errorf("❌ failed to set base denom: %w", err)
	}

	baseDenom, err := sdk.GetBaseDenom()
	if err != nil {
		return fmt.Errorf("❌ failed to get base denom: %w", err)
	}

	fmt.Printf("✅ Base denom set to: %s\n", baseDenom)

	// Configure EVM to understand our native token
	// Map chain ID to numeric EVM chain ID
	evmChainID := uint64(40004) // Local dev chain ID
	if chainID == "elys" {
		evmChainID = 40001 // Mainnet
	}
	
	ethCfg := evmtypes.DefaultChainConfig(evmChainID)
	fmt.Printf("✅ EVM chain config created for: %s (EVM ID: %d)\n", chainID, evmChainID)

	fmt.Printf("🔧 Configuring EVM with coin info: Denom=%s, ExtendedDenom=%s, DisplayDenom=%s, Decimals=%d\n", 
		coinInfo.Denom, coinInfo.ExtendedDenom, coinInfo.DisplayDenom, coinInfo.Decimals)
	
	err = evmtypes.NewEVMConfigurator().
		WithChainConfig(ethCfg).
		// CRITICAL: Tell EVM how to read native ELYS (6 decimals, not 18)
		WithEVMCoinInfo(coinInfo).
		Configure()
	if err != nil {
		return fmt.Errorf("❌ failed to configure EVM: %w", err)
	}

	fmt.Printf("🎉 EVM successfully configured for native %s with %d decimals\n", baseDenom, coinInfo.Decimals)

	sealed = true
	return nil
}

// setBaseDenom registers the display denom and base denom and sets the
// base denom for the chain.
func setBaseDenom(ci evmtypes.EvmCoinInfo) error {
	fmt.Printf("🔧 Registering display denom: %s\n", ci.DisplayDenom)
	if err := sdk.RegisterDenom(ci.DisplayDenom, math.LegacyOneDec()); err != nil {
		return fmt.Errorf("failed to register display denom %s: %w", ci.DisplayDenom, err)
	}

	fmt.Printf("🔧 Registering base denom: %s with %d decimals\n", ci.Denom, ci.Decimals)
	// sdk.RegisterDenom will automatically overwrite the base denom when the
	// new setBaseDenom() are lower than the current base denom's units.
	if err := sdk.RegisterDenom(ci.Denom, math.LegacyNewDecWithPrec(1, int64(ci.Decimals))); err != nil {
		return fmt.Errorf("failed to register base denom %s: %w", ci.Denom, err)
	}

	// CRITICAL: Set this as the chain's base denomination
	fmt.Printf("🔧 Setting chain base denom to: %s\n", ci.Denom)
	if err := sdk.SetBaseDenom(ci.Denom); err != nil {
		return fmt.Errorf("failed to set base denom %s: %w", ci.Denom, err)
	}

	return nil
}