package precompiles

import (
	"crypto/sha256"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ABI method signatures for factory precompile
var (
	// Factory precompile methods
	createWrapperSig = crypto.Keccak256([]byte("createWrapper(string,string,string,uint8)"))[:4]
)

// Method selectors for factory precompile
func getCreateWrapperSelector() string {
	return string(createWrapperSig)
}

// ABI types for encoding/decoding
var (
	addressType, _ = abi.NewType("address", "", nil)
	stringType, _  = abi.NewType("string", "", nil)
	uint256Type, _ = abi.NewType("uint256", "", nil)
	uint8Type, _   = abi.NewType("uint8", "", nil)
	boolType, _    = abi.NewType("bool", "", nil)
)

// Unpack createWrapper arguments: createWrapper(string denom, string name, string symbol, uint8 decimals)
func unpackCreateWrapperArgs(input []byte) (*CreateWrapperArgs, error) {
	args := abi.Arguments{
		{Type: stringType},
		{Type: stringType},
		{Type: stringType},
		{Type: uint8Type},
	}

	unpacked, err := args.Unpack(input)
	if err != nil {
		return nil, err
	}

	return &CreateWrapperArgs{
		Denom:    unpacked[0].(string),
		Name:     unpacked[1].(string),
		Symbol:   unpacked[2].(string),
		Decimals: unpacked[3].(uint8),
	}, nil
}

// Pack address result
func packAddress(addr common.Address) []byte {
	return common.LeftPadBytes(addr.Bytes(), 32)
}

// Encode constructor parameters for wrapper contract
func encodeConstructorParams(denom, name, symbol string, decimals uint8) []byte {
	// Create constructor ABI
	constructor := abi.Arguments{
		{Type: stringType},
		{Type: stringType},
		{Type: stringType},
		{Type: uint8Type},
	}

	// Encode parameters
	packed, err := constructor.Pack(denom, name, symbol, decimals)
	if err != nil {
		return []byte{}
	}

	return packed
}

// BankWrapper contract bytecode (simplified for demo)
// In production, this would be the actual compiled Solidity bytecode
func getWrapperContractBytecode(denom, name, symbol string, decimals uint8) []byte {
	// This is a placeholder - in a real implementation, you would:
	// 1. Have the compiled BankWrapper.sol bytecode
	// 2. Append the encoded constructor parameters
	// 3. Return the complete deployment bytecode

	baseBytecode := []byte{
		// EVM bytecode for BankWrapper contract would go here
		// For now, returning a minimal valid contract bytecode
		0x60, 0x80, 0x60, 0x40, 0x52, 0x34, 0x80, 0x15, 0x61, 0x00, 0x10, 0x57, 0x60, 0x00, 0x80, 0xfd,
		0x5b, 0x50, 0x60, 0x40, 0x51, 0x61, 0x01, 0x00, 0x38, 0x03, 0x80, 0x61, 0x01, 0x00, 0x83, 0x39,
		0x81, 0x01, 0x60, 0x40, 0x52, 0x39, 0x60, 0x00, 0xf3, 0xfe,
	}

	// Append constructor parameters
	constructorParams := encodeConstructorParams(denom, name, symbol, decimals)
	return append(baseBytecode, constructorParams...)
}

// Generate deterministic salt for CREATE2
func generateSalt(denom string) [32]byte {
	hash := sha256.Sum256([]byte(denom))
	return hash
}
