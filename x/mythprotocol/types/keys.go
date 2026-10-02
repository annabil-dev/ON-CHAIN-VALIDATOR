package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName     = "mythprotocol"
	PoUWModuleName = "pouw"
	ZYRADenom      = "uzyra"
	MTCDenom       = "umtc"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// GovModuleName duplicates the gov module's name to avoid a dependency with x/gov.
	// It should be synced with the gov module's name if it is ever changed.
	// See: https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/x/gov/types/keys.go#L9
	GovModuleName = "gov"

	MaxTaskLeaseBlocks     uint64 = 10000
	MTCDecimals            uint64 = 1_000_000
	MTCMaxSupply           uint64 = 21_000_000
	MTCValidatorAllocation uint64 = 1_000_000
	MTCTreasuryAllocation  uint64 = 20_000_000
)

// ParamsKey is the prefix to retrieve all Params
var ParamsKey = collections.NewPrefix("p_mythprotocol")
var TaskLeasesPrefix = collections.NewPrefix("task_lease")
var RuntimeStateKey = collections.NewPrefix("runtime_state")
var GaslessAccountsPrefix = collections.NewPrefix("gasless_account")
var GaslessAccountIndexKey = collections.NewPrefix("gasless_index")
