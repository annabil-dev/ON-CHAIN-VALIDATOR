package keeper

import (
	"context"

	"mythprotocol/x/mythprotocol/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
	if err := k.Params.Set(ctx, genState.Params); err != nil {
		return err
	}
	runtimeState, err := types.DecodeRuntimeState(genState.RuntimeStateJson)
	if err != nil {
		return err
	}
	if err := k.SetRuntimeState(ctx, runtimeState); err != nil {
		return err
	}
	for _, lease := range genState.TaskLeases {
		if err := k.TaskLeases.Set(ctx, lease.TaskId, lease); err != nil {
			return err
		}
	}
	return nil
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	var err error

	genesis := types.DefaultGenesis()
	genesis.Params, err = k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	genesis.TaskLeases = make([]types.TaskLease, 0)
	err = k.TaskLeases.Walk(ctx, nil, func(_ string, lease types.TaskLease) (bool, error) {
		genesis.TaskLeases = append(genesis.TaskLeases, lease)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	runtimeState, err := k.GetRuntimeState(ctx)
	if err != nil {
		return nil, err
	}
	genesis.RuntimeStateJson, err = types.EncodeRuntimeState(runtimeState)
	if err != nil {
		return nil, err
	}

	return genesis, nil
}
