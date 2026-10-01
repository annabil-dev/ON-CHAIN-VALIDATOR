package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythprotocol/x/mythprotocol/types"
)

const (
	TaskLeaseStatusOpen      = "OPEN"
	TaskLeaseStatusLeased    = "LEASED"
	TaskLeaseStatusReleased  = "RELEASED"
	TaskLeaseStatusSubmitted = "SUBMITTED"
	TaskLeaseStatusApproved  = "APPROVED"
	TaskLeaseStatusRejected  = "REJECTED"
	TaskJudgeQuorum          = types.CriteriaJudgeMajority
)

func validSHA256(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validTaskID(value string) bool {
	if len(value) == 0 || len(value) > 128 {
		return false
	}
	for _, ch := range value {
		if !(ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '-' || ch == '_') {
			return false
		}
	}
	return true
}

func (k Keeper) canonicalAddress(value string) (string, error) {
	address, err := k.addressCodec.StringToBytes(value)
	if err != nil {
		return "", errorsmod.Wrap(types.ErrInvalidAddress, "invalid signer address")
	}
	canonical, err := k.addressCodec.BytesToString(address)
	if err != nil || canonical != value {
		return "", errorsmod.Wrap(types.ErrInvalidAddress, "non-canonical signer address")
	}
	return canonical, nil
}

func taskAttemptID(chainID, taskID, miner, acceptanceHash, nonce string, height int64) string {
	preimage := fmt.Sprintf("MYTHCHAIN-TASK-LEASE-v1\x00%s\x00%s\x00%s\x00%s\x00%s\x00%d", chainID, taskID, miner, acceptanceHash, nonce, height)
	digest := sha256.Sum256([]byte(preimage))
	return hex.EncodeToString(digest[:])
}

func (k Keeper) RegisterTask(ctx context.Context, msg *types.MsgRegisterTask) (*types.MsgRegisterTaskResponse, error) {
	if msg == nil || !validTaskID(msg.TaskId) || !validSHA256(msg.AcceptanceHash) {
		return nil, errorsmod.Wrap(types.ErrInvalidTaskID, "task ID or acceptance hash is invalid")
	}
	client, err := k.canonicalAddress(msg.Creator)
	if err != nil {
		return nil, err
	}
	exists, err := k.TaskLeases.Has(ctx, msg.TaskId)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to inspect task registry")
	}
	if exists {
		return nil, errorsmod.Wrap(types.ErrTaskAlreadyLeased, "task ID has already been registered")
	}
	criteriaJSON := ""
	if msg.CriteriaJson != "" {
		criteria, err := types.ParseAcceptanceCriteria(msg.CriteriaJson)
		if err != nil {
			return nil, errorsmod.Wrap(types.ErrInvalidAcceptanceCriteria, err.Error())
		}
		canonical, err := json.Marshal(criteria)
		if err != nil {
			return nil, errorsmod.Wrap(types.ErrInvalidAcceptanceCriteria, "could not canonicalize weighted acceptance criteria")
		}
		criteriaJSON = string(canonical)
	}
	if msg.TaskCategory != "" {
		if criteriaJSON == "" {
			return nil, errorsmod.Wrap(types.ErrInvalidAcceptanceCriteria, "rewarded tasks require weighted acceptance criteria")
		}
		if _, err := types.BaseReward(msg.TaskCategory); err != nil {
			return nil, errorsmod.Wrap(types.ErrInvalidAcceptanceCriteria, err.Error())
		}
	}
	task := types.TaskLease{
		TaskId: msg.TaskId, AcceptanceHash: msg.AcceptanceHash, ClientAddress: client,
		Status: TaskLeaseStatusOpen, Votes: []types.TaskVote{}, CriteriaJson: criteriaJSON, TaskCategory: msg.TaskCategory,
	}
	if err := k.TaskLeases.Set(ctx, msg.TaskId, task); err != nil {
		return nil, errorsmod.Wrap(err, "failed to register task")
	}
	return &types.MsgRegisterTaskResponse{Task: &task}, nil
}

// ClaimTask reserves a task until an absolute block height. CometBFT transaction
// execution order is canonical: at most one unexpired lease is stored per task.
func (k Keeper) ClaimTask(ctx context.Context, msg *types.MsgClaimTask) (*types.MsgClaimTaskResponse, error) {
	if msg == nil || !validTaskID(msg.TaskId) || !validSHA256(msg.AcceptanceHash) || !validSHA256(msg.Nonce) {
		return nil, errorsmod.Wrap(types.ErrInvalidTaskID, "task ID or acceptance hash is invalid")
	}
	miner, err := k.canonicalAddress(msg.Creator)
	if err != nil {
		return nil, err
	}
	if msg.LeaseBlocks == 0 || msg.LeaseBlocks > types.MaxTaskLeaseBlocks {
		return nil, errorsmod.Wrapf(types.ErrInvalidLease, "lease_blocks must be between 1 and %d", types.MaxTaskLeaseBlocks)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height := sdkCtx.BlockHeight()
	if height < 0 || uint64(height) > uint64(math.MaxInt64)-msg.LeaseBlocks {
		return nil, errorsmod.Wrap(types.ErrInvalidLease, "lease expiry height overflows")
	}

	exists, err := k.TaskLeases.Has(ctx, msg.TaskId)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to inspect task lease")
	}
	if !exists {
		return nil, errorsmod.Wrap(types.ErrLeaseNotFound, "register task before claiming it")
	}
	previous, err := k.TaskLeases.Get(ctx, msg.TaskId)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to read task lease")
	}
	if previous.AcceptanceHash != msg.AcceptanceHash {
		return nil, errorsmod.Wrap(types.ErrInvalidLease, "acceptance hash differs from registered task")
	}
	if previous.Status == TaskLeaseStatusApproved {
		return nil, types.ErrTaskFinalized
	}
	if (previous.Status == TaskLeaseStatusLeased || previous.Status == TaskLeaseStatusSubmitted) && height < previous.ExpiresAtHeight {
		return nil, errorsmod.Wrapf(types.ErrTaskAlreadyLeased, "active lease belongs to %s until height %d", previous.MinerAddress, previous.ExpiresAtHeight)
	}

	lease := previous
	lease.MinerAddress = miner
	lease.AttemptId = taskAttemptID(sdkCtx.ChainID(), msg.TaskId, miner, msg.AcceptanceHash, msg.Nonce, height)
	lease.StartHeight = height
	lease.ExpiresAtHeight = height + int64(msg.LeaseBlocks)
	lease.Status = TaskLeaseStatusLeased
	lease.ResultCid = ""
	lease.ProofHash = ""
	lease.Votes = []types.TaskVote{}
	lease.CanonicalCriteriaResultsJson = ""
	lease.CriteriaScoreNumerator = 0
	lease.CriteriaScoreDenominator = 0
	lease.RewardSettled = false
	lease.RewardAmountUzyra = 0
	lease.JudgeRewardAddresses = []string{}
	if err := k.TaskLeases.Set(ctx, msg.TaskId, lease); err != nil {
		return nil, errorsmod.Wrap(err, "failed to save task lease")
	}
	return &types.MsgClaimTaskResponse{Lease: &lease}, nil
}

func (k Keeper) ReleaseTask(ctx context.Context, msg *types.MsgReleaseTask) (*types.MsgReleaseTaskResponse, error) {
	if msg == nil || !validTaskID(msg.TaskId) || !validSHA256(msg.AttemptId) {
		return nil, errorsmod.Wrap(types.ErrInvalidLease, "task ID or attempt ID is invalid")
	}
	miner, err := k.canonicalAddress(msg.Creator)
	if err != nil {
		return nil, err
	}
	lease, err := k.TaskLeases.Get(ctx, msg.TaskId)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrLeaseNotFound, msg.TaskId)
	}
	if lease.MinerAddress != miner {
		return nil, types.ErrLeaseOwnerMismatch
	}
	if lease.AttemptId != msg.AttemptId {
		return nil, types.ErrAttemptMismatch
	}
	if lease.Status != TaskLeaseStatusLeased {
		return nil, errorsmod.Wrap(types.ErrInvalidLease, "only an unused active lease can be released")
	}
	lease.Status = TaskLeaseStatusReleased
	lease.ExpiresAtHeight = sdk.UnwrapSDKContext(ctx).BlockHeight()
	if err := k.TaskLeases.Set(ctx, msg.TaskId, lease); err != nil {
		return nil, errorsmod.Wrap(err, "failed to release task lease")
	}
	return &types.MsgReleaseTaskResponse{}, nil
}

func (k Keeper) SubmitTaskResult(ctx context.Context, msg *types.MsgSubmitTaskResult) (*types.MsgSubmitTaskResultResponse, error) {
	if msg == nil || !validTaskID(msg.TaskId) || !validSHA256(msg.AttemptId) || !validSHA256(msg.ProofHash) {
		return nil, errorsmod.Wrap(types.ErrInvalidLease, "invalid task ID, attempt ID, or proof hash")
	}
	if !strings.HasPrefix(msg.ResultCid, "sha256:") || !validSHA256(strings.TrimPrefix(msg.ResultCid, "sha256:")) {
		return nil, errorsmod.Wrap(types.ErrInvalidLease, "result CID must be sha256:<64 lowercase hex characters>")
	}
	miner, err := k.canonicalAddress(msg.Creator)
	if err != nil {
		return nil, err
	}
	lease, err := k.TaskLeases.Get(ctx, msg.TaskId)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrLeaseNotFound, msg.TaskId)
	}
	if lease.MinerAddress != miner {
		return nil, types.ErrLeaseOwnerMismatch
	}
	if lease.AttemptId != msg.AttemptId {
		return nil, types.ErrAttemptMismatch
	}
	if lease.Status == TaskLeaseStatusSubmitted {
		return nil, types.ErrResultAlreadySubmitted
	}
	if lease.Status != TaskLeaseStatusLeased {
		return nil, errorsmod.Wrap(types.ErrInvalidLease, "task lease is not active")
	}
	if sdk.UnwrapSDKContext(ctx).BlockHeight() >= lease.ExpiresAtHeight {
		return nil, types.ErrLeaseExpired
	}
	lease.ResultCid = msg.ResultCid
	lease.ProofHash = msg.ProofHash
	lease.Status = TaskLeaseStatusSubmitted
	if err := k.TaskLeases.Set(ctx, msg.TaskId, lease); err != nil {
		return nil, errorsmod.Wrap(err, "failed to store submitted task result")
	}
	return &types.MsgSubmitTaskResultResponse{Lease: &lease}, nil
}

func (k Keeper) VoteTask(ctx context.Context, msg *types.MsgVoteTask) (*types.MsgVoteTaskResponse, error) {
	if msg == nil || !validTaskID(msg.TaskId) || !validSHA256(msg.AttemptId) ||
		(msg.Verdict != "PASS" && msg.Verdict != "FAIL") || len(msg.Reason) > 2000 {
		return nil, types.ErrInvalidVerdict
	}
	judge, err := k.canonicalAddress(msg.Creator)
	if err != nil {
		return nil, err
	}
	lease, err := k.TaskLeases.Get(ctx, msg.TaskId)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrLeaseNotFound, msg.TaskId)
	}
	if lease.Status != TaskLeaseStatusSubmitted || lease.MinerAddress == "" {
		return nil, errorsmod.Wrap(types.ErrInvalidLease, "task has no submitted result awaiting judge votes")
	}
	if lease.AttemptId != msg.AttemptId {
		return nil, types.ErrAttemptMismatch
	}
	if sdk.UnwrapSDKContext(ctx).BlockHeight() >= lease.ExpiresAtHeight {
		return nil, types.ErrLeaseExpired
	}
	if judge == lease.MinerAddress || judge == lease.ClientAddress {
		return nil, errorsmod.Wrap(types.ErrInvalidSigner, "miner and client cannot judge this task")
	}
	for _, vote := range lease.Votes {
		if vote.JudgeAddress == judge {
			return nil, types.ErrDuplicateVote
		}
	}
	if len(lease.Votes) >= types.CriteriaJudgeCommitteeSize {
		return nil, errorsmod.Wrap(types.ErrInvalidVerdict, "task already has four judge votes")
	}
	weighted := lease.CriteriaJson != ""
	var criteria []types.AcceptanceCriterion
	if weighted {
		criteria, err = types.ParseAcceptanceCriteria(lease.CriteriaJson)
		if err != nil {
			return nil, errorsmod.Wrap(types.ErrInvalidAcceptanceCriteria, "stored acceptance rubric is invalid: "+err.Error())
		}
		results, resultErr := types.ParseCriterionResults(criteria, msg.CriteriaResultsJson)
		if resultErr != nil {
			return nil, errorsmod.Wrap(types.ErrInvalidCriterionResults, resultErr.Error())
		}
		passedWeight, totalWeight, hardGatesPassed := types.ScoreCriterionResults(criteria, results)
		judgePasses := hardGatesPassed && passedWeight*100 >= totalWeight*types.AcceptancePassingPercent
		expectedVerdict := "FAIL"
		if judgePasses {
			expectedVerdict = "PASS"
		}
		if msg.Verdict != expectedVerdict {
			return nil, errorsmod.Wrap(types.ErrInvalidVerdict, "verdict does not match this judge's weighted criteria score")
		}
	} else if msg.CriteriaResultsJson != "" {
		return nil, errorsmod.Wrap(types.ErrInvalidCriterionResults, "task has no weighted rubric; criterion results are not accepted")
	}
	lease.Votes = append(lease.Votes, types.TaskVote{
		JudgeAddress: judge, AttemptId: msg.AttemptId, Verdict: msg.Verdict,
		Reason: msg.Reason, Height: sdk.UnwrapSDKContext(ctx).BlockHeight(),
		CriteriaResultsJson: msg.CriteriaResultsJson,
	})
	passCount, failCount := uint32(0), uint32(0)
	for _, vote := range lease.Votes {
		if vote.Verdict == "PASS" {
			passCount++
		} else if vote.Verdict == "FAIL" {
			failCount++
		}
	}
	if weighted {
		canonical, aggregateErr := types.AggregateCriteriaVotes(criteria, lease.Votes)
		if aggregateErr != nil {
			return nil, errorsmod.Wrap(types.ErrInvalidCriterionResults, "could not aggregate weighted judge votes: "+aggregateErr.Error())
		}
		if canonical.Resolved {
			lease.CanonicalCriteriaResultsJson = canonical.ResultsJSON
			lease.CriteriaScoreNumerator = canonical.PassedWeight
			lease.CriteriaScoreDenominator = canonical.TotalWeight
			if canonical.Passed {
				lease.Status = TaskLeaseStatusApproved
			} else {
				lease.Status = TaskLeaseStatusRejected
				lease.ExpiresAtHeight = sdk.UnwrapSDKContext(ctx).BlockHeight()
			}
			if lease.Status == TaskLeaseStatusApproved && lease.TaskCategory != "" {
				if err := k.settleTaskReward(ctx, &lease, criteria); err != nil {
					return nil, errorsmod.Wrap(types.ErrInvalidVerdict, "failed to settle task reward: "+err.Error())
				}
			}
		}
	} else if passCount >= TaskJudgeQuorum {
		lease.Status = TaskLeaseStatusApproved
	} else if failCount >= TaskJudgeQuorum {
		lease.Status = TaskLeaseStatusRejected
		lease.ExpiresAtHeight = sdk.UnwrapSDKContext(ctx).BlockHeight()
	}
	if err := k.TaskLeases.Set(ctx, msg.TaskId, lease); err != nil {
		return nil, errorsmod.Wrap(err, "failed to record judge vote")
	}
	count := passCount
	if msg.Verdict == "FAIL" {
		count = failCount
	}
	return &types.MsgVoteTaskResponse{Lease: &lease, VotesForVerdict: count}, nil
}

func (k Keeper) GetTaskLease(ctx context.Context, taskID string) (*types.TaskLease, bool, error) {
	if !validTaskID(taskID) {
		return nil, false, types.ErrInvalidTaskID
	}
	found, err := k.TaskLeases.Has(ctx, taskID)
	if err != nil || !found {
		return nil, found, err
	}
	lease, err := k.TaskLeases.Get(ctx, taskID)
	if err != nil {
		return nil, false, err
	}
	return &lease, true, nil
}
