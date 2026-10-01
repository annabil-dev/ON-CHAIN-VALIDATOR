package mythprotocol

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"mythprotocol/x/mythprotocol/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: types.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the module",
				},
				{
					RpcMethod: "TaskLease",
					Use:       "task-lease [task-id]",
					Short:     "Queries the current on-chain task lease",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              types.Msg_serviceDesc.ServiceName,
			EnhanceCustomCommand: true, // only required if you want to use the custom command
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "RegisterTask",
					Use:       "register-task [task-id] [acceptance-hash] [criteria-json]",
					Short:     "Registers a task, acceptance hash, and optional weighted criteria",
				},
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
				{
					RpcMethod: "ClaimTask",
					Use:       "claim-task [task-id] [acceptance-hash] [lease-blocks] [nonce]",
					Short:     "Claims a task until its block-height lease expires",
				},
				{
					RpcMethod: "ReleaseTask",
					Use:       "release-task [task-id] [attempt-id]",
					Short:     "Releases an unused task lease",
				},
				{
					RpcMethod: "SubmitTaskResult",
					Use:       "submit-task-result [task-id] [attempt-id] [result-cid] [proof-hash]",
					Short:     "Submits a result for the active task lease",
				},
				{
					RpcMethod: "VoteTask",
					Use:       "vote-task [task-id] [attempt-id] [PASS|FAIL] [reason] [criteria-results-json]",
					Short:     "Records a judge vote and optional per-criterion results for the active attempt",
				},
			},
		},
	}
}
