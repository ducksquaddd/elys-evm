package precompiles

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	cmn "github.com/cosmos/evm/precompiles/common"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

const (
	// BankPrecompileAddress is the address of the bank precompile
	BankPrecompileAddress = "0x0000000000000000000000000000000000000804"

	// Method names for general bank operations
	GetBalanceMethod     = "getBalance"
	GetAllBalancesMethod = "getAllBalances"
	TransferMethod       = "transfer"
	GetSupplyMethod      = "getSupply"
	GetAllSupplyMethod   = "getAllSupply"
)

var (
	// Common errors
	ErrExecutionReverted = fmt.Errorf("execution reverted")
	ErrWriteProtection   = fmt.Errorf("write protection")
)

// BankPrecompile provides universal access to bank keeper operations
// This is a general bank interface that can handle any denom, not token-specific
type BankPrecompile struct {
	cmn.Precompile
	bankKeeper bankkeeper.Keeper
}

// NewBankPrecompile creates a new general bank precompile instance
func NewBankPrecompile(bankKeeper bankkeeper.Keeper) (*BankPrecompile, error) {
	// Create ABI for general bank operations
	abiStr := `[
		{
			"inputs": [{"name": "account", "type": "address"}, {"name": "denom", "type": "string"}],
			"name": "getBalance",
			"outputs": [{"name": "", "type": "uint256"}],
			"stateMutability": "view",
			"type": "function"
		},
		{
			"inputs": [{"name": "account", "type": "address"}],
			"name": "getAllBalances",
			"outputs": [{"name": "denoms", "type": "string[]"}, {"name": "amounts", "type": "uint256[]"}],
			"stateMutability": "view",
			"type": "function"
		},
		{
			"inputs": [{"name": "to", "type": "address"}, {"name": "denom", "type": "string"}, {"name": "amount", "type": "uint256"}],
			"name": "transfer", 
			"outputs": [{"name": "", "type": "bool"}],
			"stateMutability": "nonpayable",
			"type": "function"
		},
		{
			"inputs": [{"name": "denom", "type": "string"}],
			"name": "getSupply",
			"outputs": [{"name": "", "type": "uint256"}],
			"stateMutability": "view",
			"type": "function"
		},
		{
			"inputs": [],
			"name": "getAllSupply",
			"outputs": [{"name": "denoms", "type": "string[]"}, {"name": "amounts", "type": "uint256[]"}],
			"stateMutability": "view",
			"type": "function"
		}
	]`

	parsedABI, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	p := &BankPrecompile{
		Precompile: cmn.Precompile{
			ABI: parsedABI,
		},
		bankKeeper: bankKeeper,
	}

	// Set the precompile address
	p.SetAddress(common.HexToAddress(BankPrecompileAddress))

	return p, nil
}

// RequiredGas calculates the precompiled contract's base gas rate
func (p BankPrecompile) RequiredGas(input []byte) uint64 {
	// NOTE: This check avoid panicking when trying to decode the method ID
	if len(input) < 4 {
		return 0
	}

	methodID := input[:4]
	method, err := p.MethodById(methodID)
	if err != nil {
		return 0
	}

	switch method.Name {
	case GetBalanceMethod:
		return 20_000
	case GetAllBalancesMethod:
		return 30_000
	case TransferMethod:
		return 50_000
	case GetSupplyMethod:
		return 20_000
	case GetAllSupplyMethod:
		return 40_000
	}

	return 0
}

// Run executes the precompiled contract using official cosmos/evm pattern
func (p BankPrecompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	// Use official cosmos/evm context extraction and method dispatch
	ctx, _, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// Handle gas errors gracefully
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

	// Get the caller address from the contract (this is msg.sender)
	caller := contract.Caller()

	switch method.Name {
	case GetBalanceMethod:
		bz, err = p.getBalance(ctx, method, args)
	case GetAllBalancesMethod:
		bz, err = p.getAllBalances(ctx, method, args)
	case TransferMethod:
		if readOnly {
			return nil, ErrWriteProtection
		}
		bz, err = p.transfer(ctx, method, args, caller)
	case GetSupplyMethod:
		bz, err = p.getSupply(ctx, method, args)
	case GetAllSupplyMethod:
		bz, err = p.getAllSupply(ctx, method, args)
	default:
		return nil, fmt.Errorf(cmn.ErrUnknownMethod, method.Name)
	}

	return bz, err
}

// IsTransaction checks if the given method name corresponds to a transaction or query
func (p BankPrecompile) IsTransaction(method *abi.Method) bool {
	switch method.Name {
	case TransferMethod:
		return true
	default:
		return false
	}
}

// getBalance returns the balance of a user for a specific denom
func (p BankPrecompile) getBalance(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("invalid number of arguments; expected 2, got %d", len(args))
	}

	user, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid user address type")
	}

	denom, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid denom type")
	}

	// Convert Ethereum address to Cosmos address
	cosmosAddr, err := convertEthToCosmos(user)
	if err != nil {
		return nil, fmt.Errorf("address conversion failed: %w", err)
	}

	fmt.Printf("🔍 getBalance Debug:\n")
	fmt.Printf("   Ethereum Address: %s\n", user.Hex())
	fmt.Printf("   Cosmos Address: %s\n", cosmosAddr.String())
	fmt.Printf("   Querying denom: %s\n", denom)

	// Get balance from bank keeper for the specified denom
	balance := p.bankKeeper.GetBalance(ctx, cosmosAddr, denom)

	fmt.Printf("   Balance found: %s\n", balance.String())

	// Pack the result
	result, err := method.Outputs.Pack(balance.Amount.BigInt())
	if err != nil {
		return nil, fmt.Errorf("failed to pack balance result: %w", err)
	}

	return result, nil
}

// getAllBalances returns all balances for a user
func (p BankPrecompile) getAllBalances(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of arguments; expected 1, got %d", len(args))
	}

	user, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid user address type")
	}

	// Convert Ethereum address to Cosmos address
	cosmosAddr, err := convertEthToCosmos(user)
	if err != nil {
		return nil, fmt.Errorf("address conversion failed: %w", err)
	}

	fmt.Printf("🔍 getAllBalances Debug:\n")
	fmt.Printf("   Ethereum Address: %s\n", user.Hex())
	fmt.Printf("   Cosmos Address: %s\n", cosmosAddr.String())

	// Get all balances from bank keeper
	balances := p.bankKeeper.GetAllBalances(ctx, cosmosAddr)

	// Prepare arrays for return
	denoms := make([]string, len(balances))
	amounts := make([]*big.Int, len(balances))

	for i, balance := range balances {
		denoms[i] = balance.Denom
		amounts[i] = balance.Amount.BigInt()
	}

	fmt.Printf("   Found %d balances\n", len(balances))

	// Pack the result
	result, err := method.Outputs.Pack(denoms, amounts)
	if err != nil {
		return nil, fmt.Errorf("failed to pack balances result: %w", err)
	}

	return result, nil
}

// transfer sends tokens from msg.sender to another address
func (p BankPrecompile) transfer(ctx sdk.Context, method *abi.Method, args []interface{}, caller common.Address) ([]byte, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("invalid number of arguments; expected 3, got %d", len(args))
	}

	to, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid to address type")
	}

	denom, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid denom type")
	}

	amount, ok := args[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid amount type")
	}

	// Convert addresses
	fromAddr, err := convertEthToCosmos(caller)
	if err != nil {
		return nil, fmt.Errorf("from address conversion failed: %w", err)
	}

	toAddr, err := convertEthToCosmos(to)
	if err != nil {
		return nil, fmt.Errorf("to address conversion failed: %w", err)
	}

	fmt.Printf("🔄 Transfer Debug:\n")
	fmt.Printf("   From: %s → %s\n", caller.Hex(), fromAddr.String())
	fmt.Printf("   To: %s → %s\n", to.Hex(), toAddr.String())
	fmt.Printf("   Denom: %s\n", denom)
	fmt.Printf("   Amount: %s\n", amount.String())

	// Create coin and send
	coin := sdk.NewCoin(denom, math.NewIntFromBigInt(amount))
	coins := sdk.NewCoins(coin)

	// Execute the transfer
	err = p.bankKeeper.SendCoins(ctx, fromAddr, toAddr, coins)
	if err != nil {
		fmt.Printf("❌ Transfer failed: %s\n", err.Error())
		return nil, fmt.Errorf("transfer failed: %w", err)
	}

	fmt.Printf("✅ Transfer successful: %s → %s (%s)\n",
		fromAddr.String(), toAddr.String(), coin.String())

	return method.Outputs.Pack(true)
}

// getSupply returns the total supply of a specific denom
func (p BankPrecompile) getSupply(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of arguments; expected 1, got %d", len(args))
	}

	denom, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid denom type")
	}

	fmt.Printf("🔍 getSupply Debug:\n")
	fmt.Printf("   Querying denom: %s\n", denom)

	// Get supply from bank keeper
	supply := p.bankKeeper.GetSupply(ctx, denom)

	fmt.Printf("   Supply found: %s\n", supply.String())

	return method.Outputs.Pack(supply.Amount.BigInt())
}

// getAllSupply returns the total supply of all denoms
func (p BankPrecompile) getAllSupply(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("invalid number of arguments; expected 0, got %d", len(args))
	}

	fmt.Printf("🔍 getAllSupply Debug\n")

	// Get all supplies from bank keeper
	supplies, _, err := p.bankKeeper.GetPaginatedTotalSupply(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get total supply: %w", err)
	}

	// Prepare arrays for return
	denoms := make([]string, len(supplies))
	amounts := make([]*big.Int, len(supplies))

	for i, supply := range supplies {
		denoms[i] = supply.Denom
		amounts[i] = supply.Amount.BigInt()
	}

	fmt.Printf("   Found %d supplies\n", len(supplies))

	return method.Outputs.Pack(denoms, amounts)
}

// convertEthToCosmos converts an Ethereum address to a Cosmos address using the Cosmos EVM approach
// This is the same method used in the official Cosmos EVM bech32 precompile
// If the address is already in Cosmos format, it skips conversion
func convertEthToCosmos(evmAddr common.Address) (sdk.AccAddress, error) {
	b := evmAddr.Bytes()
	if len(b) != 20 {
		return nil, fmt.Errorf(
			"invalid address length: expected 20, got %d", len(b),
		)
	}
	// Bech32 prefixes must already be set via sdk.GetConfig().SetBech32Prefix…
	return sdk.AccAddress(b), nil
}

// isLikelyCosmosAddress checks if an address is already in Cosmos format
// It converts the bytes to bech32 and checks if it starts with the expected prefix
func isLikelyCosmosAddress(addressBytes []byte) bool {
	expectedPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix() // Should be "elys"

	// Convert bytes to bech32 address
	bech32Addr, err := sdk.Bech32ifyAddressBytes(expectedPrefix, addressBytes)
	if err != nil {
		// If we can't create a valid bech32 address, it's not a Cosmos address
		return false
	}

	// Check if the resulting bech32 address starts with our expected prefix
	expectedStart := expectedPrefix + "1"
	if !strings.HasPrefix(bech32Addr, expectedStart) {
		return false
	}

	fmt.Printf("   🔍 Detected Cosmos address: %s\n", bech32Addr)
	return true
}
