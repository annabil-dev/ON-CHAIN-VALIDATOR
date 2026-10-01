package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	"cosmossdk.io/collections"

	"mythprotocol/x/mythprotocol/types"
)

const GaslessRateLimitSeconds = int64(60)

func (k Keeper) GetRuntimeState(ctx context.Context) (types.RuntimeState, error) {
	raw, err := k.RuntimeState.Get(ctx)
	if errors.Is(err, collections.ErrNotFound) {
		raw = ""
	} else if err != nil {
		return types.RuntimeState{}, err
	}
	state, err := types.DecodeRuntimeState(raw)
	if err != nil {
		return types.RuntimeState{}, err
	}
	state.GaslessAccounts = make(map[string]types.GaslessAccountState)
	addresses, err := k.getGaslessAccountIndex(ctx)
	if err != nil {
		return types.RuntimeState{}, err
	}
	for _, address := range addresses {
		encoded, err := k.GaslessAccounts.Get(ctx, address)
		if err != nil {
			return types.RuntimeState{}, err
		}
		var account types.GaslessAccountState
		if err := json.Unmarshal([]byte(encoded), &account); err != nil {
			return types.RuntimeState{}, err
		}
		state.GaslessAccounts[address] = account
	}
	return state, nil
}

func (k Keeper) SetRuntimeState(ctx context.Context, state types.RuntimeState) error {
	accounts := state.GaslessAccounts
	state.GaslessAccounts = nil
	raw, err := types.EncodeRuntimeState(state)
	if err != nil {
		return err
	}
	if err := k.RuntimeState.Set(ctx, raw); err != nil {
		return err
	}
	existing, err := k.getGaslessAccountIndex(ctx)
	if err != nil {
		return err
	}
	for _, address := range existing {
		if err := k.GaslessAccounts.Remove(ctx, address); err != nil {
			return err
		}
	}
	existing = existing[:0]
	for address, account := range accounts {
		if address == "" || account.LastGaslessUnix < 0 {
			return types.ErrInvalidAddress
		}
		bz, err := json.Marshal(account)
		if err != nil {
			return err
		}
		if err := k.GaslessAccounts.Set(ctx, address, string(bz)); err != nil {
			return err
		}
		existing = append(existing, address)
	}
	return k.setGaslessAccountIndex(ctx, existing)
}

func (k Keeper) getGaslessAccountIndex(ctx context.Context) ([]string, error) {
	raw, err := k.GaslessAccountIndex.Get(ctx)
	if errors.Is(err, collections.ErrNotFound) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	var addresses []string
	if err := json.Unmarshal([]byte(raw), &addresses); err != nil {
		return nil, err
	}
	sort.Strings(addresses)
	return addresses, nil
}

func (k Keeper) setGaslessAccountIndex(ctx context.Context, addresses []string) error {
	addresses = append([]string(nil), addresses...)
	sort.Strings(addresses)
	unique := addresses[:0]
	for _, address := range addresses {
		if len(unique) == 0 || unique[len(unique)-1] != address {
			unique = append(unique, address)
		}
	}
	bz, err := json.Marshal(unique)
	if err != nil {
		return err
	}
	return k.GaslessAccountIndex.Set(ctx, string(bz))
}

func (k Keeper) GetGaslessAccount(ctx context.Context, address string) (types.GaslessAccountState, bool, error) {
	encoded, err := k.GaslessAccounts.Get(ctx, address)
	if errors.Is(err, collections.ErrNotFound) {
		return types.GaslessAccountState{}, false, nil
	}
	if err != nil {
		return types.GaslessAccountState{}, false, err
	}
	var account types.GaslessAccountState
	if err := json.Unmarshal([]byte(encoded), &account); err != nil {
		return types.GaslessAccountState{}, false, err
	}
	return account, true, nil
}

func (k Keeper) SetGaslessAccount(ctx context.Context, address string, account types.GaslessAccountState) error {
	if address == "" || account.LastGaslessUnix < 0 {
		return types.ErrInvalidAddress
	}
	_, found, err := k.GetGaslessAccount(ctx, address)
	if err != nil {
		return err
	}
	bz, err := json.Marshal(account)
	if err != nil {
		return err
	}
	if err := k.GaslessAccounts.Set(ctx, address, string(bz)); err != nil {
		return err
	}
	if !found {
		addresses, err := k.getGaslessAccountIndex(ctx)
		if err != nil {
			return err
		}
		addresses = append(addresses, address)
		return k.setGaslessAccountIndex(ctx, addresses)
	}
	return nil
}

// ConsumeGaslessAllowance enforces one free transaction per account per minute
// until the account has received its first ZYRA reward.
func (k Keeper) ConsumeGaslessAllowance(ctx context.Context, address string, nowUnix int64) (bool, error) {
	if nowUnix < 0 {
		return false, nil
	}
	account, _, err := k.GetGaslessAccount(ctx, address)
	if err != nil {
		return false, err
	}
	if account.ReceivedFirstReward {
		return false, nil
	}
	if account.HasUsedGasless && (nowUnix < account.LastGaslessUnix || nowUnix-account.LastGaslessUnix < GaslessRateLimitSeconds) {
		return false, nil
	}
	account.LastGaslessUnix = nowUnix
	account.HasUsedGasless = true
	if err := k.SetGaslessAccount(ctx, address, account); err != nil {
		return false, err
	}
	return true, nil
}

func (k Keeper) RevokeGasless(ctx context.Context, address string) error {
	account, _, err := k.GetGaslessAccount(ctx, address)
	if err != nil {
		return err
	}
	account.ReceivedFirstReward = true
	return k.SetGaslessAccount(ctx, address, account)
}

// MarkFirstReward is retained as a semantic alias for task payout call sites.
func (k Keeper) MarkFirstReward(ctx context.Context, address string) error {
	return k.RevokeGasless(ctx, address)
}

// AccrueBlockEmission advances one committed block and returns the whole uzyra
// amount that must be minted into the PoUW module account.
func (k Keeper) AccrueBlockEmission(ctx context.Context) (uint64, error) {
	params, err := k.Params.Get(ctx)
	if errors.Is(err, collections.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if !params.EnablePouwEmissions {
		return 0, nil
	}
	raw, err := k.RuntimeState.Get(ctx)
	if errors.Is(err, collections.ErrNotFound) {
		raw = ""
	} else if err != nil {
		return 0, err
	}
	state, err := types.DecodeRuntimeState(raw)
	if err != nil {
		return 0, err
	}
	emission, minted, err := types.AccruePoUWEmission(state.Emission)
	if err != nil {
		return 0, err
	}
	state.Emission = emission
	state.GaslessAccounts = nil
	encoded, err := types.EncodeRuntimeState(state)
	if err != nil {
		return 0, err
	}
	if err := k.RuntimeState.Set(ctx, encoded); err != nil {
		return 0, err
	}
	return minted, nil
}
