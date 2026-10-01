package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythprotocol/x/mythprotocol/types"
)

func (k Keeper) settleTaskReward(ctx context.Context, lease *types.TaskLease, criteria []types.AcceptanceCriterion) error {
	if lease == nil || lease.TaskCategory == "" || lease.RewardSettled || lease.Status != TaskLeaseStatusApproved {
		return nil
	}
	totalReward, err := types.CalculateTaskReward(lease.TaskCategory, lease.CriteriaScoreNumerator, lease.CriteriaScoreDenominator)
	if err != nil {
		return err
	}
	alignedJudges, err := types.QuorumAlignedJudges(criteria, lease.Votes, lease.CanonicalCriteriaResultsJson)
	if err != nil {
		return err
	}
	lease.RewardAmountUzyra = totalReward
	lease.JudgeRewardAddresses = alignedJudges
	split := types.SplitTaskReward(totalReward)
	perJudge := uint64(0)
	if len(alignedJudges) > 0 {
		perJudge = split.JudgePool / uint64(len(alignedJudges))
	}
	judgeAmount := perJudge * uint64(len(alignedJudges))
	amountToPay := split.Miner + judgeAmount

	runtimeState, err := k.GetRuntimeState(ctx)
	if err != nil {
		return err
	}
	if runtimeState.Emission.RewardPoolUzyra < amountToPay {
		return nil
	}
	if amountToPay > 0 && k.bankKeeper == nil {
		return fmt.Errorf("bank keeper is required for PoUW settlement")
	}
	if split.Miner > 0 {
		minerAddress, err := k.addressCodec.StringToBytes(lease.MinerAddress)
		if err != nil {
			return types.ErrInvalidAddress
		}
		coins := sdk.NewCoins(sdk.NewCoin(types.ZYRADenom, math.NewIntFromUint64(split.Miner)))
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.PoUWModuleName, minerAddress, coins); err != nil {
			return err
		}
		if err := k.MarkFirstReward(ctx, lease.MinerAddress); err != nil {
			return err
		}
	}
	if perJudge > 0 {
		coins := sdk.NewCoins(sdk.NewCoin(types.ZYRADenom, math.NewIntFromUint64(perJudge)))
		for _, judgeAddress := range alignedJudges {
			judge, err := k.addressCodec.StringToBytes(judgeAddress)
			if err != nil {
				return types.ErrInvalidAddress
			}
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.PoUWModuleName, judge, coins); err != nil {
				return err
			}
			if err := k.MarkFirstReward(ctx, judgeAddress); err != nil {
				return err
			}
		}
	}
	runtimeState.Emission.RewardPoolUzyra -= amountToPay
	if err := k.SetRuntimeState(ctx, runtimeState); err != nil {
		return err
	}
	lease.RewardSettled = true
	return nil
}

func (k Keeper) SettlePendingTaskRewards(ctx context.Context) error {
	type pendingReward struct {
		taskID string
		lease  types.TaskLease
	}
	var pending []pendingReward
	if err := k.TaskLeases.Walk(ctx, nil, func(taskID string, lease types.TaskLease) (bool, error) {
		if lease.TaskCategory != "" && !lease.RewardSettled && lease.Status == TaskLeaseStatusApproved {
			pending = append(pending, pendingReward{taskID: taskID, lease: lease})
		}
		return false, nil
	}); err != nil {
		return err
	}
	for _, item := range pending {
		criteria, err := types.ParseAcceptanceCriteria(item.lease.CriteriaJson)
		if err != nil {
			return err
		}
		lease := item.lease
		if err := k.settleTaskReward(ctx, &lease, criteria); err != nil {
			return err
		}
		if lease.RewardSettled {
			if err := k.TaskLeases.Set(ctx, item.taskID, lease); err != nil {
				return err
			}
		}
	}
	return nil
}
