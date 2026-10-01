package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"mythprotocol/x/mythprotocol/types"
)

func TestScoreRewardMultiplier(t *testing.T) {
	tests := []struct {
		name            string
		passed, total   uint64
		wantNumerator   uint64
		wantDenominator uint64
	}{
		{name: "below passing", passed: 74, total: 100, wantNumerator: 1, wantDenominator: 1},
		{name: "exactly passing", passed: 75, total: 100, wantNumerator: 1, wantDenominator: 1},
		{name: "first multiplier tier", passed: 76, total: 100, wantNumerator: 3, wantDenominator: 2},
		{name: "second multiplier tier", passed: 80, total: 100, wantNumerator: 7, wantDenominator: 4},
		{name: "third multiplier tier", passed: 90, total: 100, wantNumerator: 2, wantDenominator: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := types.ScoreRewardMultiplier(tc.passed, tc.total)
			require.NoError(t, err)
			require.Equal(t, tc.wantNumerator, got.Numerator)
			require.Equal(t, tc.wantDenominator, got.Denominator)
		})
	}
}

func TestCalculateRewardAndSplit(t *testing.T) {
	total, err := types.CalculateTaskReward(types.TaskCategoryMedium, 85, 100)
	require.NoError(t, err)
	require.EqualValues(t, 437_500, total) // 0.4375 ZYRA

	split := types.SplitTaskReward(total)
	require.EqualValues(t, 262_500, split.Miner)
	require.EqualValues(t, 175_000, split.JudgePool)
	require.EqualValues(t, total, split.Miner+split.JudgePool)
}

func TestRewardRejectsInvalidCategoryAndScore(t *testing.T) {
	_, err := types.CalculateTaskReward("unknown", 100, 100)
	require.Error(t, err)
	_, err = types.CalculateTaskReward(types.TaskCategoryLight, 101, 100)
	require.Error(t, err)
	_, err = types.CalculateTaskReward(types.TaskCategoryLight, 0, 0)
	require.Error(t, err)
}
