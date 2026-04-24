package ante

import (
	"math/big"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	feemarketkeeper "github.com/cosmos/evm/x/feemarket/keeper"
)

// NewDynamicFeeChecker returns a TxFeeChecker that calculates fees for EVM transactions
// based on gas price and gas limit, while falling back to the default fee checker for Cosmos transactions
func NewDynamicFeeChecker(feeMarketKeeper feemarketkeeper.Keeper) authante.TxFeeChecker {
	return func(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
		// Check if this is an EVM transaction
		for _, msg := range tx.GetMsgs() {
			if ethTx, ok := msg.(*evmtypes.MsgEthereumTx); ok {
				// Calculate fee for EVM transaction
				return calculateEVMFee(ctx, ethTx, feeMarketKeeper)
			}
		}
		
		// Fall back to default fee checker for non-EVM transactions
		return CheckTxFeeWithValidatorMinGasPrices(ctx, tx)
	}
}

// calculateEVMFee calculates the fee for an EVM transaction
func calculateEVMFee(ctx sdk.Context, ethTx *evmtypes.MsgEthereumTx, feeMarketKeeper feemarketkeeper.Keeper) (sdk.Coins, int64, error) {
	// Get the gas price from the transaction
	gasPrice := ethTx.AsTransaction().GasPrice()
	if gasPrice == nil {
		gasPrice = big.NewInt(0)
	}
	
	// Get gas limit
	gasLimit := ethTx.AsTransaction().Gas()
	
	// Calculate total fee: gasPrice * gasLimit
	gasLimitBig := big.NewInt(int64(gasLimit))
	feeAmount := new(big.Int).Mul(gasPrice, gasLimitBig)
	
	// Create fee coin in uelys
	feeCoin := sdk.NewCoin("uelys", math.NewIntFromBigInt(feeAmount))
	fees := sdk.NewCoins(feeCoin)
	
	return fees, int64(gasLimit), nil
}