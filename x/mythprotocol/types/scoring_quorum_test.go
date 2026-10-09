package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"mythprotocol/x/mythprotocol/types"
)

func TestQuorumAlignedJudgesMustMatchEveryCriterion(t *testing.T) {
	criteria := []types.AcceptanceCriterion{{ID: "first"}, {ID: "second"}}
	canonical := `{"first":{"passed":true,"evidence":"canonical evidence"},"second":{"passed":false,"evidence":"canonical evidence"}}`
	votes := []types.TaskVote{
		{JudgeAddress: "myth1a", CriteriaResultsJson: `{"first":{"passed":true,"evidence":"a"},"second":{"passed":false,"evidence":"different evidence"}}`},
		{JudgeAddress: "myth1b", CriteriaResultsJson: `{"first":{"passed":true,"evidence":"a"},"second":{"passed":true,"evidence":"b"}}`},
		{JudgeAddress: "myth1c", CriteriaResultsJson: `{"first":{"passed":true,"evidence":"a"},"second":{"passed":false,"evidence":"b"}}`},
	}

	aligned, err := types.QuorumAlignedJudges(criteria, votes, canonical)
	require.NoError(t, err)
	require.Equal(t, []string{"myth1a", "myth1c"}, aligned)
}

func TestBindEvidenceToChecksRequiresObservedOutput(t *testing.T) {
	criteria := []types.AcceptanceCriterion{{
		ID: "out", Weight: 100,
		Check: []byte(`{"type":"stdout_contains","text":"BUILD OK"}`),
	}}

	// Passed result quoting the observed output is accepted.
	err := types.BindEvidenceToChecks(criteria, map[string]types.CriterionResult{
		"out": {Passed: true, Evidence: "log tail: BUILD OK in 12s"},
	})
	require.NoError(t, err)

	// Passed result with fabricated evidence is rejected.
	err = types.BindEvidenceToChecks(criteria, map[string]types.CriterionResult{
		"out": {Passed: true, Evidence: "ok"},
	})
	require.ErrorContains(t, err, "does not contain")

	// Failed results explain the miss instead of quoting the hit.
	err = types.BindEvidenceToChecks(criteria, map[string]types.CriterionResult{
		"out": {Passed: false, Evidence: "build failed, no output marker"},
	})
	require.NoError(t, err)
}

func TestHardGateRequiresExecutableCheck(t *testing.T) {
	_, err := types.ParseAcceptanceCriteria(`[{"id":"gate","description":"gate","weight":50,"hard_gate":true,"check":{"type":"bogus"}}]`)
	require.Error(t, err)
}
