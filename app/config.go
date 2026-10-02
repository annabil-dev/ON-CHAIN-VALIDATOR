package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	mythtypes "mythprotocol/x/mythprotocol/types"
)

func init() {
	// MTC is the L1 bond and base-fee token; Mythchain remains the network name.
	sdk.DefaultBondDenom = mythtypes.MTCDenom

	// Set address prefixes
	accountPubKeyPrefix := AccountAddressPrefix + "pub"
	validatorAddressPrefix := AccountAddressPrefix + "valoper"
	validatorPubKeyPrefix := AccountAddressPrefix + "valoperpub"
	consNodeAddressPrefix := AccountAddressPrefix + "valcons"
	consNodePubKeyPrefix := AccountAddressPrefix + "valconspub"

	// Set and seal config
	config := sdk.GetConfig()
	config.SetCoinType(ChainCoinType)
	config.SetBech32PrefixForAccount(AccountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)
	config.Seal()
}
