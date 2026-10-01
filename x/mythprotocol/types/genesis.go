package types

import (
	"fmt"
	"strings"
)

func isLowerHex(value string, length int) bool {
	if len(value) != length {
		return false
	}
	for _, ch := range value {
		if !(ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'f') {
			return false
		}
	}
	return true
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if _, err := DecodeRuntimeState(gs.RuntimeStateJson); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(gs.TaskLeases))
	for _, lease := range gs.TaskLeases {
		if lease.TaskId == "" || lease.ClientAddress == "" || !isLowerHex(lease.AcceptanceHash, 64) {
			return fmt.Errorf("genesis task lease is missing required identity fields")
		}
		if _, exists := seen[lease.TaskId]; exists {
			return fmt.Errorf("duplicate genesis task lease for task %q", lease.TaskId)
		}
		seen[lease.TaskId] = struct{}{}
		if lease.Status != "OPEN" && lease.Status != "LEASED" && lease.Status != "RELEASED" &&
			lease.Status != "SUBMITTED" && lease.Status != "APPROVED" && lease.Status != "REJECTED" {
			return fmt.Errorf("genesis task %q has invalid lease status %q", lease.TaskId, lease.Status)
		}
		if lease.Status != "OPEN" && (lease.MinerAddress == "" || !isLowerHex(lease.AttemptId, 64)) {
			return fmt.Errorf("genesis task %q is missing its active attempt identity", lease.TaskId)
		}
		if lease.StartHeight < 0 || lease.ExpiresAtHeight < lease.StartHeight {
			return fmt.Errorf("genesis task %q has invalid lease heights", lease.TaskId)
		}
		if lease.Status == "LEASED" && lease.ExpiresAtHeight <= lease.StartHeight {
			return fmt.Errorf("genesis task %q has an empty active lease", lease.TaskId)
		}
		if (lease.Status == "SUBMITTED" || lease.Status == "APPROVED" || lease.Status == "REJECTED") &&
			(!strings.HasPrefix(lease.ResultCid, "sha256:") ||
				!isLowerHex(strings.TrimPrefix(lease.ResultCid, "sha256:"), 64) || !isLowerHex(lease.ProofHash, 64)) {
			return fmt.Errorf("genesis task %q has an invalid submitted result", lease.TaskId)
		}
		seenJudges := make(map[string]struct{}, len(lease.Votes))
		passCount, failCount := 0, 0
		var criteria []AcceptanceCriterion
		if lease.CriteriaJson != "" {
			parsedCriteria, parseErr := ParseAcceptanceCriteria(lease.CriteriaJson)
			if parseErr != nil {
				return fmt.Errorf("genesis task %q has invalid weighted criteria: %w", lease.TaskId, parseErr)
			}
			criteria = parsedCriteria
			if len(lease.Votes) > CriteriaJudgeCommitteeSize {
				return fmt.Errorf("genesis task %q has too many weighted judge votes", lease.TaskId)
			}
		} else if lease.CanonicalCriteriaResultsJson != "" || lease.CriteriaScoreNumerator != 0 || lease.CriteriaScoreDenominator != 0 {
			return fmt.Errorf("genesis task %q has weighted score state without a rubric", lease.TaskId)
		}
		for _, vote := range lease.Votes {
			if vote.JudgeAddress == "" || vote.AttemptId != lease.AttemptId || (vote.Verdict != "PASS" && vote.Verdict != "FAIL") {
				return fmt.Errorf("genesis task %q has an invalid judge vote", lease.TaskId)
			}
			if _, exists := seenJudges[vote.JudgeAddress]; exists {
				return fmt.Errorf("genesis task %q contains duplicate judge votes", lease.TaskId)
			}
			seenJudges[vote.JudgeAddress] = struct{}{}
			if vote.Verdict == "PASS" {
				passCount++
			} else {
				failCount++
			}
			if len(criteria) > 0 {
				results, err := ParseCriterionResults(criteria, vote.CriteriaResultsJson)
				if err != nil {
					return fmt.Errorf("genesis task %q has invalid criterion results: %w", lease.TaskId, err)
				}
				passed, total, gates := ScoreCriterionResults(criteria, results)
				expected := "FAIL"
				if gates && passed*100 >= total*AcceptancePassingPercent {
					expected = "PASS"
				}
				if vote.Verdict != expected {
					return fmt.Errorf("genesis task %q has a judge verdict inconsistent with its weighted score", lease.TaskId)
				}
			} else if vote.CriteriaResultsJson != "" {
				return fmt.Errorf("genesis task %q has criterion results without a rubric", lease.TaskId)
			}
		}
		if len(criteria) > 0 {
			canonical, err := AggregateCriteriaVotes(criteria, lease.Votes)
			if err != nil {
				return fmt.Errorf("genesis task %q has invalid weighted quorum: %w", lease.TaskId, err)
			}
			if lease.Status == "APPROVED" || lease.Status == "REJECTED" {
				if !canonical.Resolved || lease.CanonicalCriteriaResultsJson != canonical.ResultsJSON ||
					lease.CriteriaScoreNumerator != canonical.PassedWeight || lease.CriteriaScoreDenominator != canonical.TotalWeight {
					return fmt.Errorf("genesis task %q has inconsistent canonical weighted score", lease.TaskId)
				}
				expectedStatus := "REJECTED"
				if canonical.Passed {
					expectedStatus = "APPROVED"
				}
				if lease.Status != expectedStatus {
					return fmt.Errorf("genesis task %q has status inconsistent with canonical weighted score", lease.TaskId)
				}
			} else {
				if canonical.Resolved {
					return fmt.Errorf("genesis task %q has a weighted judge majority but is not finalized", lease.TaskId)
				}
				if lease.CanonicalCriteriaResultsJson != "" || lease.CriteriaScoreNumerator != 0 || lease.CriteriaScoreDenominator != 0 {
					return fmt.Errorf("genesis task %q stores an unfinished weighted score", lease.TaskId)
				}
			}
		} else {
			if lease.Status == "APPROVED" && passCount < CriteriaJudgeMajority {
				return fmt.Errorf("genesis task %q is approved without judge quorum", lease.TaskId)
			}
			if lease.Status == "REJECTED" && failCount < CriteriaJudgeMajority {
				return fmt.Errorf("genesis task %q is rejected without judge quorum", lease.TaskId)
			}
		}
		if lease.TaskCategory == "" {
			if lease.RewardSettled || lease.RewardAmountUzyra != 0 || len(lease.JudgeRewardAddresses) != 0 {
				return fmt.Errorf("genesis task %q has reward state without a task category", lease.TaskId)
			}
		} else {
			if _, err := BaseReward(lease.TaskCategory); err != nil || len(criteria) == 0 {
				return fmt.Errorf("genesis task %q has an invalid reward category or missing rubric", lease.TaskId)
			}
			if lease.Status == "REJECTED" {
				if lease.RewardSettled || lease.RewardAmountUzyra != 0 || len(lease.JudgeRewardAddresses) != 0 {
					return fmt.Errorf("rejected genesis task %q cannot have a payout", lease.TaskId)
				}
			} else if lease.Status != "APPROVED" {
				if lease.RewardSettled || lease.RewardAmountUzyra != 0 || len(lease.JudgeRewardAddresses) != 0 {
					return fmt.Errorf("genesis task %q has reward state before quorum", lease.TaskId)
				}
			} else {
				expectedReward, err := CalculateTaskReward(lease.TaskCategory, lease.CriteriaScoreNumerator, lease.CriteriaScoreDenominator)
				if err != nil || expectedReward != lease.RewardAmountUzyra {
					return fmt.Errorf("genesis task %q has inconsistent reward amount", lease.TaskId)
				}
				expectedJudges, err := QuorumAlignedJudges(criteria, lease.Votes, lease.CanonicalCriteriaResultsJson)
				if err != nil || len(expectedJudges) != len(lease.JudgeRewardAddresses) {
					return fmt.Errorf("genesis task %q has inconsistent eligible Judges", lease.TaskId)
				}
				for i := range expectedJudges {
					if expectedJudges[i] != lease.JudgeRewardAddresses[i] {
						return fmt.Errorf("genesis task %q has inconsistent eligible Judges", lease.TaskId)
					}
				}
			}
		}
	}
	return nil
}
