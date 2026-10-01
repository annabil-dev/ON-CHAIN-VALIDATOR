package types

import "fmt"

const UZYRA = uint64(1_000_000)

const (
	TaskCategoryLight     = "light"
	TaskCategoryMedium    = "medium"
	TaskCategoryHeavy     = "heavy"
	TaskCategoryVeryHeavy = "very_heavy"
)

type RewardMultiplier struct {
	Numerator   uint64
	Denominator uint64
}

type RewardSplit struct {
	Miner     uint64
	JudgePool uint64
}

func BaseReward(category string) (uint64, error) {
	switch category {
	case TaskCategoryLight:
		return 100_000, nil // 0.10 ZYRA
	case TaskCategoryMedium:
		return 250_000, nil // 0.25 ZYRA
	case TaskCategoryHeavy:
		return 500_000, nil // 0.50 ZYRA
	case TaskCategoryVeryHeavy:
		return 1_000_000, nil // 1.00 ZYRA
	default:
		return 0, fmt.Errorf("unknown task category %q", category)
	}
}

// ScoreRewardMultiplier returns the direct total multiplier for a verified score.
// Scores at or below 75% receive 1x; tiers above 75% receive 1.5x, 1.75x, or 2x.
func ScoreRewardMultiplier(passedWeight, totalWeight uint64) (RewardMultiplier, error) {
	if totalWeight == 0 || totalWeight > MaxAcceptanceWeightTotal || passedWeight > totalWeight {
		return RewardMultiplier{}, fmt.Errorf("invalid weighted score %d/%d", passedWeight, totalWeight)
	}
	switch {
	case passedWeight*100 <= totalWeight*75:
		return RewardMultiplier{Numerator: 1, Denominator: 1}, nil
	case passedWeight*100 < totalWeight*80:
		return RewardMultiplier{Numerator: 3, Denominator: 2}, nil
	case passedWeight*100 < totalWeight*90:
		return RewardMultiplier{Numerator: 7, Denominator: 4}, nil
	default:
		return RewardMultiplier{Numerator: 2, Denominator: 1}, nil
	}
}

func CalculateTaskReward(category string, passedWeight, totalWeight uint64) (uint64, error) {
	base, err := BaseReward(category)
	if err != nil {
		return 0, err
	}
	multiplier, err := ScoreRewardMultiplier(passedWeight, totalWeight)
	if err != nil {
		return 0, err
	}
	return base * multiplier.Numerator / multiplier.Denominator, nil
}

func SplitTaskReward(total uint64) RewardSplit {
	miner := total/100*60 + total%100*60/100
	return RewardSplit{Miner: miner, JudgePool: total - miner}
}
