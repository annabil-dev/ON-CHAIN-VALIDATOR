package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"mythprotocol/x/mythprotocol/types"
)

func TestAccruePoUWEmissionUsesStepDownAndCarry(t *testing.T) {
	state, minted, err := types.AccruePoUWEmission(types.PoUWEmissionState{})
	require.NoError(t, err)
	require.EqualValues(t, 200_000, minted)
	require.EqualValues(t, 200_000*types.EmissionFixedPointScale, state.CumulativeScaled)
	require.Zero(t, state.FractionalCarry)

	state.CumulativeScaled = types.EmissionStepThresholdUzyra*types.EmissionFixedPointScale - 1
	state.RewardPoolUzyra = state.CumulativeScaled / types.EmissionFixedPointScale
	state.FractionalCarry = state.CumulativeScaled % types.EmissionFixedPointScale
	state, minted, err = types.AccruePoUWEmission(state)
	require.NoError(t, err)
	require.EqualValues(t, 200_000, minted)
	require.Greater(t, state.CumulativeScaled, types.EmissionStepThresholdUzyra*types.EmissionFixedPointScale)

	state, minted, err = types.AccruePoUWEmission(state)
	require.NoError(t, err)
	require.EqualValues(t, 150_000, minted)
}

func TestAccruePoUWEmissionStopsAtHardCap(t *testing.T) {
	state := types.PoUWEmissionState{
		CumulativeScaled: types.EmissionSupplyCapUzyra*types.EmissionFixedPointScale - 1,
		RewardPoolUzyra:  (types.EmissionSupplyCapUzyra*types.EmissionFixedPointScale - 1) / types.EmissionFixedPointScale,
		FractionalCarry:  types.EmissionFixedPointScale - 1,
	}
	state, minted, err := types.AccruePoUWEmission(state)
	require.NoError(t, err)
	require.Equal(t, types.EmissionSupplyCapUzyra*types.EmissionFixedPointScale, state.CumulativeScaled)
	require.EqualValues(t, 1, minted)
	state, minted, err = types.AccruePoUWEmission(state)
	require.NoError(t, err)
	require.Zero(t, minted)
}

func TestAccruePoUWEmissionCarriesFractionalUzyra(t *testing.T) {
	stepFourStart := 12_000_000 * types.UZYRA * types.EmissionFixedPointScale
	state := types.PoUWEmissionState{
		CumulativeScaled: stepFourStart,
		RewardPoolUzyra:  12_000_000 * types.UZYRA,
	}
	state, minted, err := types.AccruePoUWEmission(state)
	require.NoError(t, err)
	require.EqualValues(t, 63_281, minted)
	require.EqualValues(t, types.EmissionFixedPointScale/4, state.FractionalCarry)

	state, minted, err = types.AccruePoUWEmission(state)
	require.NoError(t, err)
	require.EqualValues(t, 63_281, minted)
	require.EqualValues(t, types.EmissionFixedPointScale/2, state.FractionalCarry)
	state, minted, err = types.AccruePoUWEmission(state)
	require.NoError(t, err)
	require.EqualValues(t, 63_281, minted)
	require.EqualValues(t, types.EmissionFixedPointScale*3/4, state.FractionalCarry)
	state, minted, err = types.AccruePoUWEmission(state)
	require.NoError(t, err)
	require.EqualValues(t, 63_282, minted)
	require.Zero(t, state.FractionalCarry)
}
