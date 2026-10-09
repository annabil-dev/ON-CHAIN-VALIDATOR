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
	GovModuleName = "gov"

	MaxTaskLeaseBlocks uint64 = 10000

	// MTC tokenomics
	// 1 MTC = 1,000,000 uMTC
	MTCDecimals uint64 = 1_000_000

	// Fixed maximum supply: 21 million MTC
	MTCMaxSupply uint64 = 21_000_000

	// Genesis validator allocation
	// Total genesis validator allocation = 1,000,000 MTC
	MTCGenesisValidatorAllocation uint64 = 1_000_000

	// Initial validator self bonded stake
	MTCGenesisBondedAllocation uint64 = 800_000

	// Initial validator liquid wallet balance
	MTCGenesisLiquidAllocation uint64 = 200_000

	// Remaining genesis treasury allocation
	MTCTreasuryAllocation uint64 = 20_000_000
)

// ParamsKey is the prefix to retrieve all Params
var ParamsKey = collections.NewPrefix("p_mythprotocol")
var TaskLeasesPrefix = collections.NewPrefix("task_lease")
var RuntimeStateKey = collections.NewPrefix("runtime_state")
var GaslessAccountsPrefix = collections.NewPrefix("gasless_account")
var GaslessAccountIndexKey = collections.NewPrefix("gasless_index")
