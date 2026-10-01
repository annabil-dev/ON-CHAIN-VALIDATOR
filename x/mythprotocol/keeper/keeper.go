package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"mythprotocol/x/mythprotocol/types"
)

type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec
	bankKeeper   types.BankKeeper
	// Address capable of executing a MsgUpdateParams message.
	// Typically, this should be the x/gov module account.
	authority []byte

	Schema              collections.Schema
	Params              collections.Item[types.Params]
	TaskLeases          collections.Map[string, types.TaskLease]
	RuntimeState        collections.Item[string]
	GaslessAccounts     collections.Map[string, string]
	GaslessAccountIndex collections.Item[string]
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,
	bankKeeper types.BankKeeper,

) Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address %s: %s", authority, err))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService: storeService,
		cdc:          cdc,
		addressCodec: addressCodec,
		authority:    authority,
		bankKeeper:   bankKeeper,

		Params:              collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		TaskLeases:          collections.NewMap(sb, types.TaskLeasesPrefix, "task_leases", collections.StringKey, codec.CollValue[types.TaskLease](cdc)),
		RuntimeState:        collections.NewItem(sb, types.RuntimeStateKey, "runtime_state", collections.StringValue),
		GaslessAccounts:     collections.NewMap(sb, types.GaslessAccountsPrefix, "gasless_accounts", collections.StringKey, collections.StringValue),
		GaslessAccountIndex: collections.NewItem(sb, types.GaslessAccountIndexKey, "gasless_account_index", collections.StringValue),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() []byte {
	return k.authority
}
