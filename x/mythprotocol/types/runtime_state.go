package types

import (
	"encoding/json"
	"fmt"
)

type GaslessAccountState struct {
	LastGaslessUnix int64 `json:"last_gasless_unix"`
	HasUsedGasless  bool  `json:"has_used_gasless"`
	// ReceivedFirstReward remains as the state field name; it means gasless is permanently revoked after a positive balance in an accepted fee denom.
	ReceivedFirstReward bool `json:"received_first_reward"`
}

type RuntimeState struct {
	Emission        PoUWEmissionState              `json:"emission"`
	GaslessAccounts map[string]GaslessAccountState `json:"gasless_accounts,omitempty"`
}

func DecodeRuntimeState(raw string) (RuntimeState, error) {
	state := RuntimeState{GaslessAccounts: map[string]GaslessAccountState{}}
	if raw == "" {
		return state, nil
	}
	if err := decodeStrictJSON(raw, &state); err != nil {
		return RuntimeState{}, fmt.Errorf("invalid runtime state JSON: %w", err)
	}
	if state.GaslessAccounts == nil {
		state.GaslessAccounts = map[string]GaslessAccountState{}
	}
	if err := state.Emission.Validate(); err != nil {
		return RuntimeState{}, err
	}
	for address, account := range state.GaslessAccounts {
		if address == "" || account.LastGaslessUnix < 0 {
			return RuntimeState{}, fmt.Errorf("invalid gasless account state")
		}
	}
	return state, nil
}

func EncodeRuntimeState(state RuntimeState) (string, error) {
	if err := state.Emission.Validate(); err != nil {
		return "", err
	}
	if state.GaslessAccounts == nil {
		state.GaslessAccounts = map[string]GaslessAccountState{}
	}
	bz, err := json.Marshal(state)
	if err != nil {
		return "", fmt.Errorf("failed to encode runtime state: %w", err)
	}
	return string(bz), nil
}
