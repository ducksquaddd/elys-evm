package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/elys-network/elys/v6/x/assetprofile/types"
)

// CreateWrapper creates and registers an ERC-20 wrapper contract for a Cosmos denom
// This combines factory deployment and asset profile registration in one governance action
func (k msgServer) CreateWrapper(goCtx context.Context, msg *types.MsgCreateWrapper) (*types.MsgCreateWrapperResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority (must be governance)
	if k.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	// Check if wrapper already exists for this denom
	entry, found := k.GetEntry(ctx, msg.BaseDenom)
	if found && entry.Address != "" {
		return nil, errorsmod.Wrapf(types.ErrWrapperAlreadyExists, "wrapper already exists for denom %s: %s", msg.BaseDenom, entry.Address)
	}

	// Call the factory precompile to deploy the wrapper contract
	wrapperAddress, err := k.deployWrapperViaFactory(ctx, msg.BaseDenom, msg.DisplayName, msg.DisplaySymbol, uint8(msg.Decimals))
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrWrapperDeploymentFailed, "failed to deploy wrapper: %v", err)
	}

	// Create or update asset profile entry
	if !found {
		entry = types.Entry{
			BaseDenom: msg.BaseDenom,
			Denom:     msg.BaseDenom,
		}
	}

	// Update with wrapper information
	entry.Address = wrapperAddress
	entry.DisplayName = msg.DisplayName
	entry.DisplaySymbol = msg.DisplaySymbol
	entry.Decimals = msg.Decimals
	entry.CommitEnabled = true
	entry.WithdrawEnabled = true

	// Save the updated entry
	k.SetEntry(ctx, entry)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"wrapper_created",
			sdk.NewAttribute("base_denom", msg.BaseDenom),
			sdk.NewAttribute("wrapper_address", wrapperAddress),
			sdk.NewAttribute("name", msg.DisplayName),
			sdk.NewAttribute("symbol", msg.DisplaySymbol),
		),
	)

	return &types.MsgCreateWrapperResponse{
		WrapperAddress: wrapperAddress,
	}, nil
}

// deployWrapperViaFactory calls the wrapper factory precompile to deploy a new wrapper contract
func (k Keeper) deployWrapperViaFactory(ctx sdk.Context, denom, name, symbol string, decimals uint8) (string, error) {
	// This would need to interact with the EVM module to call the factory precompile
	// For now, we'll generate a deterministic address similar to what the factory would do

	// Import the factory precompile logic
	factoryAddr := "0x0000000000000000000000000000000000000806"

	// Generate deterministic wrapper address using CREATE2 logic
	// This should match the factory precompile's logic
	salt := generateSaltForDenom(denom)
	bytecode := getWrapperBytecode(denom, name, symbol, decimals)

	// Calculate CREATE2 address
	wrapperAddr := calculateCreate2Address(factoryAddr, salt, bytecode)

	// TODO: Actually deploy the contract via EVM module
	// For now, return the calculated address
	return wrapperAddr, nil
}

// Helper functions (these should match the factory precompile implementation)
func generateSaltForDenom(denom string) string {
	// Simple salt generation - should match factory precompile
	return denom + "_wrapper"
}

func getWrapperBytecode(denom, name, symbol string, decimals uint8) []byte {
	// This should return the actual wrapper contract bytecode
	// For now, return a placeholder
	return []byte("wrapper_bytecode_placeholder")
}

func calculateCreate2Address(factory, salt string, bytecode []byte) string {
	// This should calculate the actual CREATE2 address
	// For now, return a deterministic placeholder
	return "0x742d35Cc6634C0532925a3b8D4C9db96C4b5Da5e"
}
