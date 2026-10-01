package keeper_test

import (
	"bytes"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"mythprotocol/x/mythprotocol/keeper"
	"mythprotocol/x/mythprotocol/types"
)

func leaseAddress(t *testing.T, f *fixture, seed byte) string {
	t.Helper()
	address := sdk.AccAddress(bytes.Repeat([]byte{seed}, 20))
	result, err := f.addressCodec.BytesToString(address)
	require.NoError(t, err)
	return result
}

func setBlockHeight(f *fixture, height int64) {
	f.ctx = sdk.UnwrapSDKContext(f.ctx).WithBlockHeight(height)
}

func claimMessage(creator, taskID, nonce string, blocks uint64) *types.MsgClaimTask {
	return &types.MsgClaimTask{
		Creator: creator, TaskId: taskID,
		AcceptanceHash: fmt.Sprintf("%064x", 123), LeaseBlocks: blocks, Nonce: nonce,
	}
}

func registerTask(t *testing.T, f *fixture, taskID string, client string) {
	t.Helper()
	_, err := f.keeper.RegisterTask(f.ctx, &types.MsgRegisterTask{
		Creator: client, TaskId: taskID, AcceptanceHash: fmt.Sprintf("%064x", 123),
	})
	require.NoError(t, err)
}

func registerWeightedTask(t *testing.T, f *fixture, taskID, client string, criteriaJSON string) *types.TaskLease {
	t.Helper()
	response, err := f.keeper.RegisterTask(f.ctx, &types.MsgRegisterTask{
		Creator: client, TaskId: taskID, AcceptanceHash: fmt.Sprintf("%064x", 123), CriteriaJson: criteriaJSON,
	})
	require.NoError(t, err)
	return response.Task
}

func TestClaimTaskConflictExpiresAndReclaims(t *testing.T) {
	f := initFixture(t)
	minerA, minerB := leaseAddress(t, f, 1), leaseAddress(t, f, 2)
	registerTask(t, f, "task-1", leaseAddress(t, f, 90))
	setBlockHeight(f, 100)

	first, err := f.keeper.ClaimTask(f.ctx, claimMessage(minerA, "task-1", fmt.Sprintf("%064x", 1), 2))
	require.NoError(t, err)
	require.Equal(t, int64(100), first.Lease.StartHeight)
	require.Equal(t, int64(102), first.Lease.ExpiresAtHeight)
	require.Len(t, first.Lease.AttemptId, 64)

	_, err = f.keeper.ClaimTask(f.ctx, claimMessage(minerB, "task-1", fmt.Sprintf("%064x", 2), 2))
	require.ErrorContains(t, err, "active lease")

	setBlockHeight(f, 101)
	_, err = f.keeper.ClaimTask(f.ctx, claimMessage(minerB, "task-1", fmt.Sprintf("%064x", 2), 2))
	require.ErrorContains(t, err, "active lease")

	setBlockHeight(f, 102)
	second, err := f.keeper.ClaimTask(f.ctx, claimMessage(minerB, "task-1", fmt.Sprintf("%064x", 2), 2))
	require.NoError(t, err)
	require.Equal(t, minerB, second.Lease.MinerAddress)
	require.NotEqual(t, first.Lease.AttemptId, second.Lease.AttemptId)
}

func TestReleaseOnlyByLeaseOwnerAndAttempt(t *testing.T) {
	f := initFixture(t)
	minerA, minerB := leaseAddress(t, f, 3), leaseAddress(t, f, 4)
	registerTask(t, f, "task-release", leaseAddress(t, f, 91))
	setBlockHeight(f, 25)
	lease, err := f.keeper.ClaimTask(f.ctx, claimMessage(minerA, "task-release", fmt.Sprintf("%064x", 3), 10))
	require.NoError(t, err)

	_, err = f.keeper.ReleaseTask(f.ctx, &types.MsgReleaseTask{Creator: minerB, TaskId: "task-release", AttemptId: lease.Lease.AttemptId})
	require.ErrorContains(t, err, "does not own")
	_, err = f.keeper.ReleaseTask(f.ctx, &types.MsgReleaseTask{Creator: minerA, TaskId: "task-release", AttemptId: fmt.Sprintf("%064x", 99)})
	require.ErrorContains(t, err, "does not match")

	_, err = f.keeper.ReleaseTask(f.ctx, &types.MsgReleaseTask{Creator: minerA, TaskId: "task-release", AttemptId: lease.Lease.AttemptId})
	require.NoError(t, err)
	stored, err := f.keeper.TaskLeases.Get(f.ctx, "task-release")
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusReleased, stored.Status)
	require.Equal(t, int64(25), stored.ExpiresAtHeight)

	next, err := f.keeper.ClaimTask(f.ctx, claimMessage(minerB, "task-release", fmt.Sprintf("%064x", 4), 10))
	require.NoError(t, err)
	require.Equal(t, minerB, next.Lease.MinerAddress)
}

func TestSubmitResultMustMatchUnexpiredLeaseAndIsOneShot(t *testing.T) {
	f := initFixture(t)
	minerA, minerB := leaseAddress(t, f, 5), leaseAddress(t, f, 6)
	registerTask(t, f, "task-result", leaseAddress(t, f, 92))
	setBlockHeight(f, 50)
	lease, err := f.keeper.ClaimTask(f.ctx, claimMessage(minerA, "task-result", fmt.Sprintf("%064x", 5), 4))
	require.NoError(t, err)
	result := &types.MsgSubmitTaskResult{
		Creator: minerA, TaskId: "task-result", AttemptId: lease.Lease.AttemptId,
		ResultCid: "sha256:" + fmt.Sprintf("%064x", 42), ProofHash: fmt.Sprintf("%064x", 43),
	}

	wrongOwner := *result
	wrongOwner.Creator = minerB
	_, err = f.keeper.SubmitTaskResult(f.ctx, &wrongOwner)
	require.ErrorContains(t, err, "does not own")
	wrongAttempt := *result
	wrongAttempt.AttemptId = fmt.Sprintf("%064x", 99)
	_, err = f.keeper.SubmitTaskResult(f.ctx, &wrongAttempt)
	require.ErrorContains(t, err, "does not match")

	response, err := f.keeper.SubmitTaskResult(f.ctx, result)
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusSubmitted, response.Lease.Status)
	require.Equal(t, result.ResultCid, response.Lease.ResultCid)
	_, err = f.keeper.SubmitTaskResult(f.ctx, result)
	require.ErrorContains(t, err, "already submitted")
	_, err = f.keeper.ClaimTask(f.ctx, claimMessage(minerB, "task-result", fmt.Sprintf("%064x", 6), 4))
	require.ErrorContains(t, err, "active lease")

	// An unfinalized submission does not lock the task forever: after its
	// block-height lease expires, a new attempt can supersede it.
	setBlockHeight(f, lease.Lease.ExpiresAtHeight)
	next, err := f.keeper.ClaimTask(f.ctx, claimMessage(minerB, "task-result", fmt.Sprintf("%064x", 6), 4))
	require.NoError(t, err)
	_, err = f.keeper.SubmitTaskResult(f.ctx, result)
	require.ErrorContains(t, err, "does not own")
	require.NotEqual(t, lease.Lease.AttemptId, next.Lease.AttemptId)
}

func TestClaimValidationAndTaskLeaseQuery(t *testing.T) {
	f := initFixture(t)
	miner := leaseAddress(t, f, 7)
	registerTask(t, f, "task-query", leaseAddress(t, f, 93))
	setBlockHeight(f, 1)

	_, err := f.keeper.ClaimTask(f.ctx, claimMessage(miner, "bad/task", fmt.Sprintf("%064x", 1), 1))
	require.Error(t, err)
	_, err = f.keeper.ClaimTask(f.ctx, claimMessage(miner, "task-query", fmt.Sprintf("%064x", 1), types.MaxTaskLeaseBlocks+1))
	require.ErrorContains(t, err, "lease_blocks")

	lease, found, err := f.keeper.GetTaskLease(f.ctx, "task-query")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, keeper.TaskLeaseStatusOpen, lease.Status)

	_, err = f.keeper.ClaimTask(f.ctx, claimMessage(miner, "task-query", fmt.Sprintf("%064x", 2), 3))
	require.NoError(t, err)
	lease, found, err = f.keeper.GetTaskLease(f.ctx, "task-query")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, miner, lease.MinerAddress)
}

func TestTaskLeasesSurviveGenesisExportAndImport(t *testing.T) {
	f := initFixture(t)
	miner := leaseAddress(t, f, 8)
	registerTask(t, f, "task-genesis", leaseAddress(t, f, 94))
	setBlockHeight(f, 70)
	claimed, err := f.keeper.ClaimTask(f.ctx, claimMessage(miner, "task-genesis", fmt.Sprintf("%064x", 8), 12))
	require.NoError(t, err)

	genesis, err := f.keeper.ExportGenesis(f.ctx)
	require.NoError(t, err)
	require.Len(t, genesis.TaskLeases, 1)
	require.NoError(t, genesis.Validate())

	second := initFixture(t)
	require.NoError(t, second.keeper.InitGenesis(second.ctx, *genesis))
	restored, found, err := second.keeper.GetTaskLease(second.ctx, "task-genesis")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, claimed.Lease.TaskId, restored.TaskId)
	require.Equal(t, claimed.Lease.AcceptanceHash, restored.AcceptanceHash)
	require.Equal(t, claimed.Lease.MinerAddress, restored.MinerAddress)
	require.Equal(t, claimed.Lease.AttemptId, restored.AttemptId)
	require.Equal(t, claimed.Lease.ExpiresAtHeight, restored.ExpiresAtHeight)
	require.Empty(t, restored.Votes)
}

func TestTaskJudgeQuorumFinalizesAttempt(t *testing.T) {
	f := initFixture(t)
	client := leaseAddress(t, f, 11)
	miner := leaseAddress(t, f, 12)
	judgeA, judgeB, judgeC, judgeD := leaseAddress(t, f, 13), leaseAddress(t, f, 14), leaseAddress(t, f, 15), leaseAddress(t, f, 16)
	registerTask(t, f, "task-vote", client)
	setBlockHeight(f, 100)
	lease, err := f.keeper.ClaimTask(f.ctx, claimMessage(miner, "task-vote", fmt.Sprintf("%064x", 11), 20))
	require.NoError(t, err)
	_, err = f.keeper.SubmitTaskResult(f.ctx, &types.MsgSubmitTaskResult{
		Creator: miner, TaskId: "task-vote", AttemptId: lease.Lease.AttemptId,
		ResultCid: "sha256:" + fmt.Sprintf("%064x", 21), ProofHash: fmt.Sprintf("%064x", 22),
	})
	require.NoError(t, err)

	vote := func(address, verdict string) *types.MsgVoteTask {
		return &types.MsgVoteTask{Creator: address, TaskId: "task-vote", AttemptId: lease.Lease.AttemptId,
			Verdict: verdict, Reason: "checked against task criteria"}
	}
	_, err = f.keeper.VoteTask(f.ctx, vote(miner, "PASS"))
	require.ErrorContains(t, err, "cannot judge")
	_, err = f.keeper.VoteTask(f.ctx, vote(client, "PASS"))
	require.ErrorContains(t, err, "cannot judge")

	first, err := f.keeper.VoteTask(f.ctx, vote(judgeA, "PASS"))
	require.NoError(t, err)
	require.EqualValues(t, 1, first.VotesForVerdict)
	require.Equal(t, keeper.TaskLeaseStatusSubmitted, first.Lease.Status)
	_, err = f.keeper.VoteTask(f.ctx, vote(judgeA, "PASS"))
	require.ErrorContains(t, err, "already voted")
	second, err := f.keeper.VoteTask(f.ctx, vote(judgeB, "PASS"))
	require.NoError(t, err)
	require.EqualValues(t, 2, second.VotesForVerdict)
	require.Equal(t, keeper.TaskLeaseStatusSubmitted, second.Lease.Status)
	third, err := f.keeper.VoteTask(f.ctx, vote(judgeC, "PASS"))
	require.NoError(t, err)
	require.EqualValues(t, 3, third.VotesForVerdict)
	require.Equal(t, keeper.TaskLeaseStatusApproved, third.Lease.Status)
	_, err = f.keeper.VoteTask(f.ctx, vote(judgeD, "PASS"))
	require.Error(t, err)
	_, err = f.keeper.ClaimTask(f.ctx, claimMessage(judgeC, "task-vote", fmt.Sprintf("%064x", 12), 5))
	require.ErrorContains(t, err, "already been approved")
}

func TestConflictingJudgeVotesRequireQuorumAndRejectionUnlocksRetry(t *testing.T) {
	f := initFixture(t)
	client := leaseAddress(t, f, 21)
	miner := leaseAddress(t, f, 22)
	judgeA, judgeB, judgeC, judgeD, minerB := leaseAddress(t, f, 23), leaseAddress(t, f, 24), leaseAddress(t, f, 25), leaseAddress(t, f, 26), leaseAddress(t, f, 27)
	registerTask(t, f, "task-conflict", client)
	setBlockHeight(f, 200)
	lease, err := f.keeper.ClaimTask(f.ctx, claimMessage(miner, "task-conflict", fmt.Sprintf("%064x", 21), 30))
	require.NoError(t, err)
	_, err = f.keeper.SubmitTaskResult(f.ctx, &types.MsgSubmitTaskResult{
		Creator: miner, TaskId: "task-conflict", AttemptId: lease.Lease.AttemptId,
		ResultCid: "sha256:" + fmt.Sprintf("%064x", 31), ProofHash: fmt.Sprintf("%064x", 32),
	})
	require.NoError(t, err)
	vote := func(address, verdict string) *types.MsgVoteTask {
		return &types.MsgVoteTask{Creator: address, TaskId: "task-conflict", AttemptId: lease.Lease.AttemptId, Verdict: verdict}
	}
	_, err = f.keeper.VoteTask(f.ctx, vote(judgeA, "PASS"))
	require.NoError(t, err)
	_, err = f.keeper.VoteTask(f.ctx, vote(judgeB, "FAIL"))
	require.NoError(t, err)
	current, err := f.keeper.TaskLeases.Get(f.ctx, "task-conflict")
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusSubmitted, current.Status)

	setBlockHeight(f, 201)
	third, err := f.keeper.VoteTask(f.ctx, vote(judgeC, "FAIL"))
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusSubmitted, third.Lease.Status)
	final, err := f.keeper.VoteTask(f.ctx, vote(judgeD, "FAIL"))
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusRejected, final.Lease.Status)
	require.Equal(t, int64(201), final.Lease.ExpiresAtHeight)

	retry, err := f.keeper.ClaimTask(f.ctx, claimMessage(minerB, "task-conflict", fmt.Sprintf("%064x", 22), 30))
	require.NoError(t, err)
	require.NotEqual(t, lease.Lease.AttemptId, retry.Lease.AttemptId)
}

func TestWeightedJudgeVotesAggregatePerCriterionAndStoreCanonicalScore(t *testing.T) {
	f := initFixture(t)
	client := leaseAddress(t, f, 31)
	miner := leaseAddress(t, f, 32)
	judgeA, judgeB, judgeC, judgeD := leaseAddress(t, f, 33), leaseAddress(t, f, 34), leaseAddress(t, f, 35), leaseAddress(t, f, 36)
	criteriaJSON := `[{"id":"core","description":"Core behavior","weight":40,"hard_gate":false,"check":{"type":"application_runs"}},{"id":"secondary","description":"Secondary behavior","weight":30,"hard_gate":false,"check":{"type":"application_runs"}},{"id":"launches","description":"Application starts","weight":30,"hard_gate":true,"check":{"type":"application_runs"}}]`
	registered := registerWeightedTask(t, f, "task-weighted", client, criteriaJSON)
	require.NotEmpty(t, registered.CriteriaJson)
	setBlockHeight(f, 300)
	lease, err := f.keeper.ClaimTask(f.ctx, claimMessage(miner, "task-weighted", fmt.Sprintf("%064x", 31), 30))
	require.NoError(t, err)
	_, err = f.keeper.SubmitTaskResult(f.ctx, &types.MsgSubmitTaskResult{
		Creator: miner, TaskId: "task-weighted", AttemptId: lease.Lease.AttemptId,
		ResultCid: "sha256:" + fmt.Sprintf("%064x", 41), ProofHash: fmt.Sprintf("%064x", 42),
	})
	require.NoError(t, err)

	vote := func(address, verdict, results string) *types.MsgVoteTask {
		return &types.MsgVoteTask{Creator: address, TaskId: "task-weighted", AttemptId: lease.Lease.AttemptId,
			Verdict: verdict, CriteriaResultsJson: results}
	}
	judgeAResults := `{"core":{"passed":true,"evidence":"core passed"},"secondary":{"passed":true,"evidence":"secondary passed"},"launches":{"passed":true,"evidence":"app started"}}`
	judgeBResults := `{"core":{"passed":true,"evidence":"core passed"},"secondary":{"passed":false,"evidence":"secondary missing"},"launches":{"passed":true,"evidence":"app started"}}`
	judgeCResults := `{"core":{"passed":false,"evidence":"core failed"},"secondary":{"passed":false,"evidence":"secondary missing"},"launches":{"passed":true,"evidence":"app started"}}`
	judgeDResults := `{"core":{"passed":true,"evidence":"core passed"},"secondary":{"passed":false,"evidence":"secondary missing"},"launches":{"passed":true,"evidence":"app started"}}`

	first, err := f.keeper.VoteTask(f.ctx, vote(judgeA, "PASS", judgeAResults))
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusSubmitted, first.Lease.Status)
	second, err := f.keeper.VoteTask(f.ctx, vote(judgeB, "FAIL", judgeBResults))
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusSubmitted, second.Lease.Status)
	require.Zero(t, second.Lease.CriteriaScoreDenominator)

	third, err := f.keeper.VoteTask(f.ctx, vote(judgeC, "FAIL", judgeCResults))
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusSubmitted, third.Lease.Status)
	final, err := f.keeper.VoteTask(f.ctx, vote(judgeD, "FAIL", judgeDResults))
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusRejected, final.Lease.Status)
	require.EqualValues(t, 70, final.Lease.CriteriaScoreNumerator)
	require.EqualValues(t, 100, final.Lease.CriteriaScoreDenominator)
	require.Contains(t, final.Lease.CanonicalCriteriaResultsJson, `"secondary":{"passed":false`)
	require.Equal(t, int64(300), final.Lease.ExpiresAtHeight)
	genesis, err := f.keeper.ExportGenesis(f.ctx)
	require.NoError(t, err)
	require.NoError(t, genesis.Validate())
}

func TestWeightedJudgeTieWaitsForThirdAndVerdictMustMatchScore(t *testing.T) {
	f := initFixture(t)
	client := leaseAddress(t, f, 41)
	miner := leaseAddress(t, f, 42)
	judgeA, judgeB, judgeC, judgeD := leaseAddress(t, f, 43), leaseAddress(t, f, 44), leaseAddress(t, f, 45), leaseAddress(t, f, 46)
	criteriaJSON := `[{"id":"required","description":"Required behavior","weight":100,"hard_gate":true,"check":{"type":"application_runs"}}]`
	registerWeightedTask(t, f, "task-weighted-tie", client, criteriaJSON)
	setBlockHeight(f, 400)
	lease, err := f.keeper.ClaimTask(f.ctx, claimMessage(miner, "task-weighted-tie", fmt.Sprintf("%064x", 41), 30))
	require.NoError(t, err)
	_, err = f.keeper.SubmitTaskResult(f.ctx, &types.MsgSubmitTaskResult{
		Creator: miner, TaskId: "task-weighted-tie", AttemptId: lease.Lease.AttemptId,
		ResultCid: "sha256:" + fmt.Sprintf("%064x", 51), ProofHash: fmt.Sprintf("%064x", 52),
	})
	require.NoError(t, err)

	vote := func(address, verdict, result string) *types.MsgVoteTask {
		return &types.MsgVoteTask{Creator: address, TaskId: "task-weighted-tie", AttemptId: lease.Lease.AttemptId,
			Verdict: verdict, CriteriaResultsJson: result}
	}
	passed := `{"required":{"passed":true,"evidence":"passed"}}`
	failed := `{"required":{"passed":false,"evidence":"failed"}}`
	_, err = f.keeper.VoteTask(f.ctx, vote(judgeA, "FAIL", passed))
	require.ErrorContains(t, err, "does not match")
	_, err = f.keeper.VoteTask(f.ctx, vote(judgeA, "PASS", passed))
	require.NoError(t, err)
	second, err := f.keeper.VoteTask(f.ctx, vote(judgeB, "FAIL", failed))
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusSubmitted, second.Lease.Status)
	third, err := f.keeper.VoteTask(f.ctx, vote(judgeC, "FAIL", failed))
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusSubmitted, third.Lease.Status)
	fourth, err := f.keeper.VoteTask(f.ctx, vote(judgeD, "FAIL", failed))
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusRejected, fourth.Lease.Status)
	require.EqualValues(t, 0, fourth.Lease.CriteriaScoreNumerator)
}

func TestRegisterTaskRejectsInvalidWeightedRubric(t *testing.T) {
	f := initFixture(t)
	client := leaseAddress(t, f, 51)
	_, err := f.keeper.RegisterTask(f.ctx, &types.MsgRegisterTask{
		Creator: client, TaskId: "task-bad-rubric", AcceptanceHash: fmt.Sprintf("%064x", 123),
		CriteriaJson: `[{"id":"bad id","description":"bad","weight":1,"hard_gate":false}]`,
	})
	require.ErrorContains(t, err, "invalid weighted acceptance criteria")
}

func TestWeightedThreeMatchingJudgeResultsApproveCanonicalTask(t *testing.T) {
	f := initFixture(t)
	client := leaseAddress(t, f, 61)
	miner := leaseAddress(t, f, 62)
	judgeA, judgeB, judgeC, judgeD := leaseAddress(t, f, 63), leaseAddress(t, f, 64), leaseAddress(t, f, 65), leaseAddress(t, f, 66)
	criteriaJSON := `[{"id":"required","description":"Required behavior","weight":100,"hard_gate":true,"check":{"type":"application_runs"}}]`
	registerWeightedTask(t, f, "task-weighted-pass", client, criteriaJSON)
	setBlockHeight(f, 500)
	lease, err := f.keeper.ClaimTask(f.ctx, claimMessage(miner, "task-weighted-pass", fmt.Sprintf("%064x", 61), 30))
	require.NoError(t, err)
	_, err = f.keeper.SubmitTaskResult(f.ctx, &types.MsgSubmitTaskResult{
		Creator: miner, TaskId: "task-weighted-pass", AttemptId: lease.Lease.AttemptId,
		ResultCid: "sha256:" + fmt.Sprintf("%064x", 61), ProofHash: fmt.Sprintf("%064x", 62),
	})
	require.NoError(t, err)
	results := `{"required":{"passed":true,"evidence":"application started"}}`
	vote := func(address string) *types.MsgVoteTask {
		return &types.MsgVoteTask{Creator: address, TaskId: "task-weighted-pass", AttemptId: lease.Lease.AttemptId,
			Verdict: "PASS", CriteriaResultsJson: results}
	}

	_, err = f.keeper.VoteTask(f.ctx, vote(judgeA))
	require.NoError(t, err)
	second, err := f.keeper.VoteTask(f.ctx, vote(judgeB))
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusSubmitted, second.Lease.Status)
	third, err := f.keeper.VoteTask(f.ctx, vote(judgeC))
	require.NoError(t, err)
	final := third
	require.Equal(t, keeper.TaskLeaseStatusApproved, final.Lease.Status)
	require.EqualValues(t, 100, final.Lease.CriteriaScoreNumerator)
	require.EqualValues(t, 100, final.Lease.CriteriaScoreDenominator)
	require.Contains(t, final.Lease.CanonicalCriteriaResultsJson, `"required":{"passed":true`)
	_, err = f.keeper.VoteTask(f.ctx, vote(judgeD))
	require.Error(t, err)
}

func TestWeightedQuorumSettlesSixtyFortyAndMarksOnlyAlignedJudges(t *testing.T) {
	f := initFixture(t)
	client, miner := leaseAddress(t, f, 71), leaseAddress(t, f, 72)
	judgeA, judgeB, judgeC, judgeD := leaseAddress(t, f, 73), leaseAddress(t, f, 74), leaseAddress(t, f, 75), leaseAddress(t, f, 76)
	criteriaJSON := `[{"id":"first","description":"First criterion","weight":80,"hard_gate":true,"check":{"type":"application_runs"}},{"id":"second","description":"Second criterion","weight":20,"hard_gate":false,"check":{"type":"application_runs"}}]`
	registered, err := f.keeper.RegisterTask(f.ctx, &types.MsgRegisterTask{
		Creator: client, TaskId: "task-reward", AcceptanceHash: fmt.Sprintf("%064x", 123),
		CriteriaJson: criteriaJSON, TaskCategory: types.TaskCategoryMedium,
	})
	require.NoError(t, err)

	state := types.RuntimeState{Emission: types.PoUWEmissionState{
		CumulativeScaled: 1_000_000 * types.EmissionFixedPointScale,
		RewardPoolUzyra:  300_000,
	}}
	require.NoError(t, f.keeper.SetRuntimeState(f.ctx, state))
	params := types.DefaultParams()
	params.EnablePouwEmissions = true
	require.NoError(t, f.keeper.Params.Set(f.ctx, params))
	f.bank.add(types.PoUWModuleName, types.ZYRADenom, 300_000)
	setBlockHeight(f, 700)
	lease, err := f.keeper.ClaimTask(f.ctx, claimMessage(miner, "task-reward", fmt.Sprintf("%064x", 72), 30))
	require.NoError(t, err)
	_, err = f.keeper.SubmitTaskResult(f.ctx, &types.MsgSubmitTaskResult{
		Creator: miner, TaskId: "task-reward", AttemptId: lease.Lease.AttemptId,
		ResultCid: "sha256:" + fmt.Sprintf("%064x", 73), ProofHash: fmt.Sprintf("%064x", 74),
	})
	require.NoError(t, err)
	resultsA := `{"first":{"passed":true,"evidence":"first passed"},"second":{"passed":true,"evidence":"second passed"}}`
	resultsB := `{"first":{"passed":true,"evidence":"first passed"},"second":{"passed":false,"evidence":"second failed"}}`
	vote := func(judge, verdict, results string) *types.MsgVoteTask {
		return &types.MsgVoteTask{Creator: judge, TaskId: "task-reward", AttemptId: lease.Lease.AttemptId,
			Verdict: verdict, CriteriaResultsJson: results}
	}
	first, err := f.keeper.VoteTask(f.ctx, vote(judgeA, "PASS", resultsA))
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusSubmitted, first.Lease.Status)
	second, err := f.keeper.VoteTask(f.ctx, vote(judgeB, "PASS", resultsB))
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusSubmitted, second.Lease.Status)
	third, err := f.keeper.VoteTask(f.ctx, vote(judgeC, "PASS", resultsB))
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusSubmitted, third.Lease.Status)
	final, err := f.keeper.VoteTask(f.ctx, vote(judgeD, "PASS", resultsB))
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusApproved, final.Lease.Status)
	require.False(t, final.Lease.RewardSettled)
	require.EqualValues(t, 437_500, final.Lease.RewardAmountUzyra)
	require.ElementsMatch(t, []string{judgeB, judgeC, judgeD}, final.Lease.JudgeRewardAddresses)
	minted, err := f.keeper.AccrueBlockEmission(f.ctx)
	require.NoError(t, err)
	require.EqualValues(t, 200_000, minted)
	require.NoError(t, f.bank.MintCoins(f.ctx, types.PoUWModuleName, sdk.NewCoins(sdk.NewCoin(types.ZYRADenom, sdkmath.NewIntFromUint64(minted)))))
	require.NoError(t, f.keeper.SettlePendingTaskRewards(f.ctx))
	settled, found, err := f.keeper.GetTaskLease(f.ctx, registered.Task.TaskId)
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, settled.RewardSettled)
	balance := func(address string) uint64 {
		addressBytes, err := f.addressCodec.StringToBytes(address)
		require.NoError(t, err)
		return f.bank.balances[string(addressBytes)][types.ZYRADenom]
	}
	require.EqualValues(t, 262_500, balance(miner))
	require.EqualValues(t, 58_333, balance(judgeB))
	require.EqualValues(t, 58_333, balance(judgeC))
	require.EqualValues(t, 58_333, balance(judgeD))
	require.Zero(t, balance(judgeA))
	remaining, err := f.keeper.GetRuntimeState(f.ctx)
	require.NoError(t, err)
	require.EqualValues(t, 62_501, remaining.Emission.RewardPoolUzyra)
	genesis, err := f.keeper.ExportGenesis(f.ctx)
	require.NoError(t, err)
	require.NoError(t, genesis.Validate())

	_, err = f.keeper.ClaimTask(f.ctx, claimMessage(miner, "task-reward", fmt.Sprintf("%064x", 77), 30))
	require.ErrorContains(t, err, "already been approved")
	_, found, err = f.keeper.GetTaskLease(f.ctx, registered.Task.TaskId)
	require.NoError(t, err)
	require.True(t, found)
}

func TestRejectedAttemptPaysZeroAndTaskPaysOnlyOnceAfterApproval(t *testing.T) {
	f := initFixture(t)
	client, miner := leaseAddress(t, f, 81), leaseAddress(t, f, 82)
	judges := []string{leaseAddress(t, f, 83), leaseAddress(t, f, 84), leaseAddress(t, f, 85)}
	criteriaJSON := `[{"id":"required","description":"Must start","weight":100,"hard_gate":true,"check":{"type":"application_runs"}}]`
	_, err := f.keeper.RegisterTask(f.ctx, &types.MsgRegisterTask{
		Creator: client, TaskId: "task-retry-payout", AcceptanceHash: fmt.Sprintf("%064x", 123),
		CriteriaJson: criteriaJSON, TaskCategory: types.TaskCategoryLight,
	})
	require.NoError(t, err)
	require.NoError(t, f.keeper.SetRuntimeState(f.ctx, types.RuntimeState{Emission: types.PoUWEmissionState{
		CumulativeScaled: 1_000_000 * types.EmissionFixedPointScale,
		RewardPoolUzyra:  500_000,
	}}))
	f.bank.add(types.PoUWModuleName, types.ZYRADenom, 500_000)
	setBlockHeight(f, 800)

	failedLease, err := f.keeper.ClaimTask(f.ctx, claimMessage(miner, "task-retry-payout", fmt.Sprintf("%064x", 81), 20))
	require.NoError(t, err)
	submit := func(attemptID string, nonce int) {
		_, submitErr := f.keeper.SubmitTaskResult(f.ctx, &types.MsgSubmitTaskResult{
			Creator: miner, TaskId: "task-retry-payout", AttemptId: attemptID,
			ResultCid: "sha256:" + fmt.Sprintf("%064x", nonce+90), ProofHash: fmt.Sprintf("%064x", nonce+100),
		})
		require.NoError(t, submitErr)
	}
	submit(failedLease.Lease.AttemptId, 1)
	failedResults := `{"required":{"passed":false,"evidence":"application did not start"}}`
	failedVerdict := func(judge string) *types.MsgVoteTask {
		return &types.MsgVoteTask{Creator: judge, TaskId: "task-retry-payout", AttemptId: failedLease.Lease.AttemptId,
			Verdict: "FAIL", CriteriaResultsJson: failedResults}
	}
	for _, judge := range judges[:2] {
		_, err = f.keeper.VoteTask(f.ctx, failedVerdict(judge))
		require.NoError(t, err)
	}
	failed, err := f.keeper.VoteTask(f.ctx, failedVerdict(judges[2]))
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusRejected, failed.Lease.Status)
	require.False(t, failed.Lease.RewardSettled)
	require.Zero(t, failed.Lease.RewardAmountUzyra)
	require.EqualValues(t, 500_000, f.bank.balances[types.PoUWModuleName][types.ZYRADenom])

	setBlockHeight(f, 801)
	passedLease, err := f.keeper.ClaimTask(f.ctx, claimMessage(miner, "task-retry-payout", fmt.Sprintf("%064x", 82), 20))
	require.NoError(t, err)
	submit(passedLease.Lease.AttemptId, 2)
	passedResults := `{"required":{"passed":true,"evidence":"application started"}}`
	passedVote := func(judge string) *types.MsgVoteTask {
		return &types.MsgVoteTask{Creator: judge, TaskId: "task-retry-payout", AttemptId: passedLease.Lease.AttemptId,
			Verdict: "PASS", CriteriaResultsJson: passedResults}
	}
	for _, judge := range judges[:2] {
		_, err = f.keeper.VoteTask(f.ctx, passedVote(judge))
		require.NoError(t, err)
	}
	approved, err := f.keeper.VoteTask(f.ctx, passedVote(judges[2]))
	require.NoError(t, err)
	require.Equal(t, keeper.TaskLeaseStatusApproved, approved.Lease.Status)
	require.True(t, approved.Lease.RewardSettled)
	require.EqualValues(t, 200_000, approved.Lease.RewardAmountUzyra)
	require.EqualValues(t, 300_002, f.bank.balances[types.PoUWModuleName][types.ZYRADenom])
	_, err = f.keeper.ClaimTask(f.ctx, claimMessage(miner, "task-retry-payout", fmt.Sprintf("%064x", 83), 20))
	require.ErrorContains(t, err, "already been approved")
	genesis, err := f.keeper.ExportGenesis(f.ctx)
	require.NoError(t, err)
	require.NoError(t, genesis.Validate())
}

func judgeCAddress(f *fixture, t *testing.T) string {
	t.Helper()
	return leaseAddress(t, f, 65)
}
