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
