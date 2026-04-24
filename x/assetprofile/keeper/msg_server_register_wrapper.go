package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/elys-network/elys/v6/x/assetprofile/types"
)

// RegisterWrapper registers an ERC-20 wrapper contract for a Cosmos denom
// This is typically called after deploying a wrapper contract via the factory
func (k msgServer) RegisterWrapper(goCtx context.Context, msg *types.MsgUpdateEntry) (*types.MsgUpdateEntryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority (must be governance)
	if k.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	// Validate wrapper address is a valid Ethereum address
	if !common.IsHexAddress(msg.Address) {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid wrapper address: %s", msg.Address)
	}

	// Get existing entry or create new one
	entry, found := k.GetEntry(ctx, msg.BaseDenom)
	if !found {
		// Create new entry for wrapper registration
		entry = types.Entry{
			BaseDenom:   msg.BaseDenom,
			Denom:       msg.Denom,
			Decimals:    msg.Decimals,
			DisplayName: msg.DisplayName,
			DisplaySymbol: msg.DisplaySymbol,
			Address:     msg.Address, // This will store the wrapper contract address
		}
	} else {
		// Update existing entry with wrapper information
		entry.Address = msg.Address
		entry.DisplayName = msg.DisplayName
		entry.DisplaySymbol = msg.DisplaySymbol
		entry.Decimals = msg.Decimals
	}

	// Save the updated entry
	k.SetEntry(ctx, entry)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"wrapper_registered",
			sdk.NewAttribute("base_denom", msg.BaseDenom),
			sdk.NewAttribute("wrapper_address", msg.Address),
			sdk.NewAttribute("name", msg.DisplayName),
			sdk.NewAttribute("symbol", msg.DisplaySymbol),
		),
	)

	return &types.MsgUpdateEntryResponse{}, nil
}

// ValidateWrapperRegistration validates a wrapper registration request
func (k Keeper) ValidateWrapperRegistration(ctx sdk.Context, baseDenom, wrapperAddress, name, symbol string, decimals uint64) error {
	// Validate base denom
	if baseDenom == "" {
		return errorsmod.Wrap(types.ErrInvalidBaseDenom, "base denom cannot be empty")
	}

	// Validate wrapper address
	if !common.IsHexAddress(wrapperAddress) {
		return errorsmod.Wrapf(types.ErrInvalidAddress, "invalid wrapper address: %s", wrapperAddress)
	}

	// Validate decimals
	if decimals < 6 || decimals > 18 {
		return types.ErrDecimalsInvalid
	}

	// Validate name and symbol
	if name == "" {
		return errorsmod.Wrap(types.ErrInvalidRequest, "name cannot be empty")
	}
	if symbol == "" {
		return errorsmod.Wrap(types.ErrInvalidRequest, "symbol cannot be empty")
	}

	// Check if wrapper already exists for this denom
	entry, found := k.GetEntry(ctx, baseDenom)
	if found && entry.Address != "" {
		return errorsmod.Wrapf(types.ErrWrapperAlreadyExists, "wrapper already exists for denom %s: %s", baseDenom, entry.Address)
	}

	return nil
}

// Helper function to register a wrapper for governance proposals
func (k Keeper) RegisterWrapperForDenom(ctx sdk.Context, baseDenom, wrapperAddress, name, symbol string, decimals uint64) error {
	// Validate the registration
	if err := k.ValidateWrapperRegistration(ctx, baseDenom, wrapperAddress, name, symbol, decimals); err != nil {
		return err
	}

	// Get existing entry or create new
	entry, found := k.GetEntry(ctx, baseDenom)
	if !found {
		entry = types.Entry{
			BaseDenom: baseDenom,
			Denom:     baseDenom, // Use base denom as denom for new entries
		}
	}

	// Update with wrapper information
	entry.Address = wrapperAddress
	entry.DisplayName = name
	entry.DisplaySymbol = symbol
	entry.Decimals = decimals

	// Save entry
	k.SetEntry(ctx, entry)

	return nil
}