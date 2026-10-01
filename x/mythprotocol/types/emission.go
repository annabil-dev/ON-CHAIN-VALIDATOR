package types

import "fmt"

const (
	EmissionFixedPointScale    = uint64(4096) // fractional parts of one uzyra
	EmissionStepThresholdUzyra = uint64(3_000_000) * UZYRA
	EmissionSupplyCapUzyra     = uint64(21_000_000) * UZYRA
	emissionBaseRateUzyra      = uint64(200_000) // 0.2 ZYRA per block
	emissionMaxStep            = uint64(6)
)

type PoUWEmissionState struct {
	CumulativeScaled uint64 `json:"cumulative_scaled"`
	RewardPoolUzyra  uint64 `json:"reward_pool_uzyra"`
	FractionalCarry  uint64 `json:"fractional_carry"`
}

func (state PoUWEmissionState) Validate() error {
	capScaled := EmissionSupplyCapUzyra * EmissionFixedPointScale
	if state.CumulativeScaled > capScaled {
		return fmt.Errorf("cumulative emission exceeds supply cap")
	}
	if state.FractionalCarry >= EmissionFixedPointScale || state.FractionalCarry != state.CumulativeScaled%EmissionFixedPointScale {
		return fmt.Errorf("fractional emission carry is inconsistent")
	}
	if state.RewardPoolUzyra > state.CumulativeScaled/EmissionFixedPointScale {
		return fmt.Errorf("reward pool exceeds emitted whole uzyra")
	}
	return nil
}

// AccruePoUWEmission applies one committed block's emission and returns the
// whole uzyra newly minted into the reward pool. Fractional uzyra is carried.
func AccruePoUWEmission(state PoUWEmissionState) (PoUWEmissionState, uint64, error) {
	if err := state.Validate(); err != nil {
		return state, 0, err
	}
	capScaled := EmissionSupplyCapUzyra * EmissionFixedPointScale
	if state.CumulativeScaled == capScaled {
		return state, 0, nil
	}

	step := state.CumulativeScaled / (EmissionStepThresholdUzyra * EmissionFixedPointScale)
	if step > emissionMaxStep {
		step = emissionMaxStep
	}
	rateScaled := emissionBaseRateUzyra * EmissionFixedPointScale
	for i := uint64(0); i < step; i++ {
		rateScaled = rateScaled * 3 / 4
	}
	remaining := capScaled - state.CumulativeScaled
	if rateScaled > remaining {
		rateScaled = remaining
	}

	oldWhole := state.CumulativeScaled / EmissionFixedPointScale
	state.CumulativeScaled += rateScaled
	newWhole := state.CumulativeScaled / EmissionFixedPointScale
	newlyMinted := newWhole - oldWhole
	state.FractionalCarry = state.CumulativeScaled % EmissionFixedPointScale
	if newlyMinted > EmissionSupplyCapUzyra-state.RewardPoolUzyra {
		return state, 0, fmt.Errorf("emission would exceed available supply")
	}
	state.RewardPoolUzyra += newlyMinted
	return state, newlyMinted, nil
}
