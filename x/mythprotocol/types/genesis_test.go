package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"mythprotocol/x/mythprotocol/types"
)

func TestGenesisState_Validate(t *testing.T) {
	lease := types.TaskLease{
		TaskId: "task-1", AcceptanceHash: strings.Repeat("a", 64), ClientAddress: "myth1client", MinerAddress: "myth1miner",
		AttemptId: strings.Repeat("b", 64), StartHeight: 10, ExpiresAtHeight: 20, Status: "LEASED",
	}
	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc:     "valid genesis state",
			genState: &types.GenesisState{},
			valid:    true,
		},
		{
			desc:     "valid task lease state",
			genState: &types.GenesisState{Params: types.DefaultParams(), TaskLeases: []types.TaskLease{lease}},
			valid:    true,
		},
		{
			desc:     "duplicate lease task id",
			genState: &types.GenesisState{Params: types.DefaultParams(), TaskLeases: []types.TaskLease{lease, lease}},
			valid:    false,
		},
		{
			desc: "submitted lease must include a content hash and proof hash",
			genState: &types.GenesisState{Params: types.DefaultParams(), TaskLeases: []types.TaskLease{{
				TaskId: "task-2", AcceptanceHash: strings.Repeat("a", 64), ClientAddress: "myth1client", MinerAddress: "myth1miner",
				AttemptId: strings.Repeat("b", 64), StartHeight: 10, ExpiresAtHeight: 20, Status: "SUBMITTED",
			}}},
			valid: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
