package keeper_test

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/core/address"
	math "cosmossdk.io/math"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"mythprotocol/x/mythprotocol/keeper"
	module "mythprotocol/x/mythprotocol/module"
	"mythprotocol/x/mythprotocol/types"
)

type fixture struct {
	ctx          context.Context
	keeper       keeper.Keeper
	addressCodec address.Codec
	bank         *mockBankKeeper
}

type mockBankKeeper struct {
	balances map[string]map[string]uint64
}

func newMockBankKeeper() *mockBankKeeper {
	return &mockBankKeeper{balances: make(map[string]map[string]uint64)}
}

func (b *mockBankKeeper) add(address, denom string, amount uint64) {
	if b.balances[address] == nil {
		b.balances[address] = make(map[string]uint64)
	}
	b.balances[address][denom] += amount
}

func (b *mockBankKeeper) SpendableCoins(_ context.Context, address sdk.AccAddress) sdk.Coins {
	coins := sdk.NewCoins()
	for denom, amount := range b.balances[string(address)] {
		if amount > 0 {
			coins = coins.Add(sdk.NewCoin(denom, math.NewIntFromUint64(amount)))
		}
	}
	return coins
}

func (b *mockBankKeeper) MintCoins(_ context.Context, module string, coins sdk.Coins) error {
	for _, coin := range coins {
		b.add(module, coin.Denom, coin.Amount.Uint64())
	}
	return nil
}

func (b *mockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, module string, recipient sdk.AccAddress, coins sdk.Coins) error {
	for _, coin := range coins {
		amount := coin.Amount.Uint64()
		if b.balances[module][coin.Denom] < amount {
			return fmt.Errorf("insufficient mock module balance")
		}
		b.balances[module][coin.Denom] -= amount
		b.add(string(recipient), coin.Denom, amount)
	}
	return nil
}

func initFixture(t *testing.T) *fixture {
	t.Helper()

	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModule{})
	addressCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	storeService := runtime.NewKVStoreService(storeKey)
	ctx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test")).Ctx

	authority := authtypes.NewModuleAddress(types.GovModuleName)
	bank := newMockBankKeeper()

	k := keeper.NewKeeper(
		storeService,
		encCfg.Codec,
		addressCodec,
		authority,
		bank,
	)

	// Initialize params
	if err := k.Params.Set(ctx, types.DefaultParams()); err != nil {
		t.Fatalf("failed to set params: %v", err)
	}

	return &fixture{
		ctx:          ctx,
		keeper:       k,
		addressCodec: addressCodec,
		bank:         bank,
	}
}
