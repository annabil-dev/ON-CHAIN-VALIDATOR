package app

import (
	"bytes"
	"fmt"

	math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"mythprotocol/x/mythprotocol/types"
)

func (app *App) txFeeChecker(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, fmt.Errorf("transaction does not implement FeeTx")
	}
	fees := feeTx.GetFee()
	if err := fees.Validate(); err != nil {
		return nil, 0, fmt.Errorf("invalid transaction fees: %w", err)
	}
	// Genesis gentxs are executed internally at height zero and are not user-paid txs.
	if ctx.BlockHeight() == 0 {
		return fees, 0, nil
	}
	signers, signerAddresses, err := app.transactionSigners(ctx, tx)
	if err != nil {
		return nil, 0, err
	}

	if len(fees) == 0 {
		if len(feeTx.FeeGranter()) != 0 {
			return nil, 0, fmt.Errorf("zero-fee transactions cannot use a fee granter")
		}
		feePayer := feeTx.FeePayer()
		if len(feePayer) > 0 && !bytes.Equal(feePayer, signers[0]) {
			return nil, 0, fmt.Errorf("gasless fee payer must be the first transaction signer")
		}
		for _, address := range signerAddresses {
			allowed, err := app.MythprotocolKeeper.ConsumeGaslessAllowance(ctx, address, ctx.BlockTime().Unix())
			if err != nil {
				return nil, 0, fmt.Errorf("failed to check gasless allowance: %w", err)
			}
			if !allowed {
				return nil, 0, fmt.Errorf("gas fee required: account balance is positive, gasless was revoked, or the one-per-minute limit applies")
			}
		}
		return sdk.NewCoins(), 0, nil
	}
	if len(fees) != 1 || (fees[0].Denom != types.MTCDenom && fees[0].Denom != types.ZYRADenom) {
		return nil, 0, fmt.Errorf("gas fees must be paid in %s or %s", types.MTCDenom, types.ZYRADenom)
	}
	if fees[0].Denom == types.ZYRADenom {
		params, err := app.MythprotocolKeeper.Params.Get(ctx)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to read PoUW activation parameters: %w", err)
		}
		if !feeDenomAllowed(params.EnablePouwEmissions, fees[0].Denom) {
			return nil, 0, fmt.Errorf("uzyra gas fees are disabled until PoUW activation")
		}
	}

	gas := feeTx.GetGas()
	gasDec := math.LegacyNewDecFromInt(math.NewIntFromUint64(gas))
	requiredFees := ctx.MinGasPrices().MulDec(gasDec)
	providedFees := sdk.NewDecCoinsFromCoins(fees...)
	minimumMet := requiredFees.IsZero()
	for _, required := range requiredFees {
		if !providedFees.AmountOf(required.Denom).LT(required.Amount) {
			minimumMet = true
			break
		}
	}
	if !minimumMet {
		return nil, 0, fmt.Errorf("insufficient fees: provided %s, required at least one of %s", fees, requiredFees)
	}
	return fees, 0, nil
}

func feeDenomAllowed(pouwEnabled bool, denom string) bool {
	return denom == types.MTCDenom || (pouwEnabled && denom == types.ZYRADenom)
}

func (app *App) transactionSigners(ctx sdk.Context, tx sdk.Tx) ([][]byte, []string, error) {
	signedTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return nil, nil, fmt.Errorf("transaction does not expose signers")
	}
	signers, err := signedTx.GetSigners()
	if err != nil {
		return nil, nil, err
	}
	if len(signers) == 0 {
		return nil, nil, fmt.Errorf("transaction has no signer")
	}
	addresses := make([]string, 0, len(signers))
	for _, signer := range signers {
		address, err := app.AuthKeeper.AddressCodec().BytesToString(signer)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid transaction signer: %w", err)
		}
		addresses = append(addresses, address)
		balances := app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(signer))
		if balances.AmountOf(types.MTCDenom).IsPositive() || balances.AmountOf(types.ZYRADenom).IsPositive() {
			if err := app.MythprotocolKeeper.RevokeGasless(ctx, address); err != nil {
				return nil, nil, fmt.Errorf("failed to revoke gasless access for funded account: %w", err)
			}
		}
	}
	return signers, addresses, nil
}

var _ authante.TxFeeChecker = (*App)(nil).txFeeChecker
