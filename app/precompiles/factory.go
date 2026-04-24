package precompiles

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/cosmos/evm/precompiles/common"
	assetprofilekeeper "github.com/elys-network/elys/v6/x/assetprofile/keeper"
	assetprofiletypes "github.com/elys-network/elys/v6/x/assetprofile/types"
)

const (
	// WrapperFactoryAddress is the address of the wrapper factory precompile
	WrapperFactoryAddress = "0x0000000000000000000000000000000000000806"
)

// WrapperFactoryPrecompile creates ERC-20 wrapper contracts using CREATE2
type WrapperFactoryPrecompile struct {
	cmn.Precompile
	assetProfileKeeper assetprofilekeeper.Keeper
}

// NewWrapperFactoryPrecompile creates a new wrapper factory precompile
func NewWrapperFactoryPrecompile(apKeeper assetprofilekeeper.Keeper) (*WrapperFactoryPrecompile, error) {
	// Define the ABI for the factory precompile
	abiStr := `[
		{
			"inputs": [
				{"name": "denom", "type": "string"},
				{"name": "name", "type": "string"},
				{"name": "symbol", "type": "string"},
				{"name": "decimals", "type": "uint8"}
			],
			"name": "createWrapper",
			"outputs": [{"name": "", "type": "address"}],
			"stateMutability": "nonpayable",
			"type": "function"
		},
		{
			"inputs": [{"name": "denom", "type": "string"}],
			"name": "getWrapper",
			"outputs": [{"name": "", "type": "address"}],
			"stateMutability": "view",
			"type": "function"
		}
	]`

	parsedABI, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	p := &WrapperFactoryPrecompile{
		Precompile: cmn.Precompile{
			ABI: parsedABI,
		},
		assetProfileKeeper: apKeeper,
	}

	// Set the precompile address
	p.SetAddress(common.HexToAddress(WrapperFactoryAddress))

	return p, nil
}

// Address returns the precompile address
func (p *WrapperFactoryPrecompile) Address() common.Address {
	return common.HexToAddress(WrapperFactoryAddress)
}

// RequiredGas returns the gas required for this precompile
func (p *WrapperFactoryPrecompile) RequiredGas(input []byte) uint64 {
	return 200_000 // Higher gas for contract creation
}

// Run executes the precompile
func (p *WrapperFactoryPrecompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) ([]byte, error) {
	input := contract.Input
	if readOnly {
		return nil, ErrWriteProtection
	}

	if len(input) < 4 {
		return nil, ErrExecutionReverted
	}

	// Get method selector (first 4 bytes)
	methodID := input[:4]

	switch string(methodID) {
	case getCreateWrapperSelector():
		return p.createWrapper(evm, input[4:])
	default:
		return nil, ErrExecutionReverted
	}
}

// createWrapper registers a wrapper contract for a denom (simplified version)
// createWrapper(string denom, string name, string symbol, uint8 decimals) returns (address)
func (p *WrapperFactoryPrecompile) createWrapper(_ *vm.EVM, input []byte) ([]byte, error) {
	// Parse ABI-encoded arguments
	args, err := unpackCreateWrapperArgs(input)
	if err != nil {
		return nil, err
	}

	denom := args.Denom
	name := args.Name
	symbol := args.Symbol
	decimals := args.Decimals

	// Generate deterministic wrapper address using CREATE2 logic
	salt := generateSalt(denom)
	factoryAddr := p.Address()

	// Calculate what the wrapper address would be
	// In a real implementation, this would be the actual deployed contract address
	bytecode := getWrapperContractBytecode(denom, name, symbol, decimals)
	wrapperAddr := crypto.CreateAddress2(factoryAddr, salt, crypto.Keccak256(bytecode))

	// Update asset profile with wrapper address
	// Extract SDK context from EVM
	ctx := context.Background() // FIXME: Extract actual context from EVM
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err = p.updateAssetProfile(sdkCtx, denom, wrapperAddr.Hex(), name, symbol, decimals)
	if err != nil {
		return nil, err
	}

	// Return the calculated wrapper address
	return packAddress(wrapperAddr), nil
}

// Helper types
type CreateWrapperArgs struct {
	Denom    string
	Name     string
	Symbol   string
	Decimals uint8
}

// Update asset profile with wrapper contract address
func (p *WrapperFactoryPrecompile) updateAssetProfile(ctx sdk.Context, denom, wrapperAddr, name, symbol string, decimals uint8) error {
	// Get existing entry or create new one
	entry, found := p.assetProfileKeeper.GetEntry(ctx, denom)
	if !found {
		entry = assetprofiletypes.Entry{
			BaseDenom: denom,
		}
	}

	// Update with wrapper contract information
	entry.Address = wrapperAddr
	entry.DisplayName = name
	entry.DisplaySymbol = symbol
	entry.Decimals = uint64(decimals)

	// Save updated entry
	p.assetProfileKeeper.SetEntry(ctx, entry)

	return nil
}
