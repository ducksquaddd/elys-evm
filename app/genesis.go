package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

// The genesis state of the blockchain is represented here as a map of raw json
// messages key'd by a identifier string.
// The identifier is used to determine which module genesis information belongs
// to so it may be appropriately routed during init chain.
// Within this application default genesis information is retrieved from
// the ModuleBasicManager which populates json from each BasicModule
// object provided to it during init.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(app *ElysApp, cdc codec.JSONCodec) GenesisState {
	genesis := app.ModuleBasics.DefaultGenesis(cdc)

	// Add EVM genesis configuration with proper defaults
	evmGenState := evmtypes.DefaultGenesisState()
	// Enable static precompiles - include our custom ELYS precompiles!
	evmGenState.Params.ActiveStaticPrecompiles = []string{
		// Standard Ethereum precompiles (Berlin fork)
		"0x0000000000000000000000000000000000000001", // ecRecover
		"0x0000000000000000000000000000000000000002", // sha256hash
		"0x0000000000000000000000000000000000000003", // ripemd160hash
		"0x0000000000000000000000000000000000000004", // dataCopy
		"0x0000000000000000000000000000000000000005", // bigModExp
		"0x0000000000000000000000000000000000000006", // bn256Add
		"0x0000000000000000000000000000000000000007", // bn256ScalarMul
		"0x0000000000000000000000000000000000000008", // bn256Pairing
		"0x0000000000000000000000000000000000000009", // blake2F
		// Our custom precompiles
		"0x0000000000000000000000000000000000000804", // Universal Bank Precompile
		"0x0000000000000000000000000000000000000806", // Wrapper Factory Precompile
	}
	genesis[evmtypes.ModuleName] = cdc.MustMarshalJSON(evmGenState)

	// Add FeeMarket genesis configuration with proper defaults
	feeMarketGenState := feemarkettypes.DefaultGenesisState()
	// Disable base fee mechanism initially to prevent nil pointer issues
	// This can be enabled later via governance when EVM is fully operational
	feeMarketGenState.Params.NoBaseFee = true
	feeMarketGenState.Params.EnableHeight = 1
	feeMarketGenState.BlockGas = 0
	genesis[feemarkettypes.ModuleName] = cdc.MustMarshalJSON(feeMarketGenState)

	return genesis
}
