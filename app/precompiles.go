package app

import (
	"fmt"
	"maps"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	// Our custom precompiles
	"github.com/elys-network/elys/v6/app/precompiles"
	assetprofilekeeper "github.com/elys-network/elys/v6/x/assetprofile/keeper"
)

// NewAvailableStaticPrecompiles returns our custom ELYS precompiles
// This follows the cosmos/evm pattern but focuses on our specific needs
//
// NOTE: this should only be used during initialization of the Keeper.
func NewAvailableStaticPrecompiles(
	bankKeeper bankkeeper.Keeper,
	assetProfileKeeper assetprofilekeeper.Keeper,
) map[common.Address]vm.PrecompiledContract {
	// Clone the mapping from the latest EVM fork.
	staticPrecompiles := maps.Clone(vm.PrecompiledContractsBerlin)

	// === CUSTOM ELYS PRECOMPILES ===
	// Our custom bank precompile for universal bank operations
	bankPrecompile, err := precompiles.NewBankPrecompile(bankKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate bank precompile: %w", err))
	}

	// Our custom factory precompile for CREATE2 wrapper deployment
	factoryPrecompile := precompiles.NewWrapperFactoryPrecompile(assetProfileKeeper)

	// Register our custom precompiles
	staticPrecompiles[bankPrecompile.Address()] = bankPrecompile
	staticPrecompiles[factoryPrecompile.Address()] = factoryPrecompile

	return staticPrecompiles
}
