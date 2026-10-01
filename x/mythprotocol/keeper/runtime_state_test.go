package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"mythprotocol/x/mythprotocol/types"
)

func TestEmissionStatePersistsAndRoundTripsThroughGenesis(t *testing.T) {
	f := initFixture(t)
	params := types.DefaultParams()
	params.EnablePouwEmissions = true
	require.NoError(t, f.keeper.Params.Set(f.ctx, params))
	minted, err := f.keeper.AccrueBlockEmission(f.ctx)
	require.NoError(t, err)
	require.EqualValues(t, 200_000, minted)

	genesis, err := f.keeper.ExportGenesis(f.ctx)
	require.NoError(t, err)
	state, err := types.DecodeRuntimeState(genesis.RuntimeStateJson)
	require.NoError(t, err)
	require.EqualValues(t, 200_000, state.Emission.RewardPoolUzyra)
	require.EqualValues(t, 200_000*types.EmissionFixedPointScale, state.Emission.CumulativeScaled)

	restored := initFixture(t)
	require.NoError(t, restored.keeper.InitGenesis(restored.ctx, *genesis))
	minted, err = restored.keeper.AccrueBlockEmission(restored.ctx)
	require.NoError(t, err)
	require.EqualValues(t, 200_000, minted)
	restoredState, err := restored.keeper.GetRuntimeState(restored.ctx)
	require.NoError(t, err)
	require.EqualValues(t, 400_000, restoredState.Emission.RewardPoolUzyra)
}

func TestGaslessAllowanceIsPerAccountRateLimitedAndEndsAtFirstReward(t *testing.T) {
	f := initFixture(t)
	address := "myth1bootstrap"
	allowed, err := f.keeper.ConsumeGaslessAllowance(f.ctx, address, 1_000)
	require.NoError(t, err)
	require.True(t, allowed)

	allowed, err = f.keeper.ConsumeGaslessAllowance(f.ctx, address, 1_059)
	require.NoError(t, err)
	require.False(t, allowed)
	allowed, err = f.keeper.ConsumeGaslessAllowance(f.ctx, address, 1_060)
	require.NoError(t, err)
	require.True(t, allowed)

	require.NoError(t, f.keeper.MarkFirstReward(f.ctx, address))
	allowed, err = f.keeper.ConsumeGaslessAllowance(f.ctx, address, 1_120)
	require.NoError(t, err)
	require.False(t, allowed)
	index, err := f.keeper.GaslessAccountIndex.Get(f.ctx)
	require.NoError(t, err)
	require.Contains(t, index, address)
	runtimeState, err := f.keeper.GetRuntimeState(f.ctx)
	require.NoError(t, err)
	require.Contains(t, runtimeState.GaslessAccounts, address)

	genesis, err := f.keeper.ExportGenesis(f.ctx)
	require.NoError(t, err)
	require.Contains(t, genesis.RuntimeStateJson, address)
	restored := initFixture(t)
	require.NoError(t, restored.keeper.InitGenesis(restored.ctx, *genesis))
	account, found, err := restored.keeper.GetGaslessAccount(restored.ctx, address)
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, account.ReceivedFirstReward)
}
