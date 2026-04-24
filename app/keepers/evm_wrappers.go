package keepers

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

// EVMAccountKeeper wraps the cosmos-sdk auth keeper to satisfy cosmos-evm interfaces
type EVMAccountKeeper struct {
	authkeeper.AccountKeeper
}

// NewEVMAccountKeeper creates a new EVM account keeper wrapper
func NewEVMAccountKeeper(ak authkeeper.AccountKeeper) *EVMAccountKeeper {
	return &EVMAccountKeeper{
		AccountKeeper: ak,
	}
}

// Verify EVMAccountKeeper implements the required interface
var _ evmtypes.AccountKeeper = (*EVMAccountKeeper)(nil)

// RemoveExpiredUnorderedNonces removes expired unordered nonces - stub implementation
func (ak *EVMAccountKeeper) RemoveExpiredUnorderedNonces(ctx sdk.Context) error {
	// For now, this is a no-op since the underlying cosmos-sdk account keeper
	// doesn't support unordered transactions. This feature is specific to cosmos-evm.
	return nil
}

// TryAddUnorderedNonce tries to add an unordered nonce - stub implementation
func (ak *EVMAccountKeeper) TryAddUnorderedNonce(ctx sdk.Context, sender []byte, timestamp time.Time) error {
	// For now, this is a no-op since the underlying cosmos-sdk account keeper
	// doesn't support unordered transactions. This feature is specific to cosmos-evm.
	return nil
}

// UnorderedTransactionsEnabled returns whether unordered transactions are enabled
func (ak *EVMAccountKeeper) UnorderedTransactionsEnabled() bool {
	// For now, unordered transactions are disabled since we're using the standard
	// cosmos-sdk account keeper which doesn't support this feature.
	return false
}