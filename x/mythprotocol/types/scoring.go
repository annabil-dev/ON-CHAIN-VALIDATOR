package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

const (
	MaxAcceptanceCriteria        = 100
	MaxAcceptanceCriterionWeight = uint64(1_000_000)
	MaxAcceptanceWeightTotal     = uint64(1_000_000)
	MaxCriterionEvidenceBytes    = 1000
	MaxCriterionResultsJSON      = 256 * 1024
	AcceptancePassingPercent     = uint64(75)
	CriteriaJudgeCommitteeSize   = 4
	CriteriaJudgeMajority        = 3
)

var acceptanceCriterionID = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

type AcceptanceCriterion struct {
	ID          string          `json:"id"`
	Description string          `json:"description"`
	Weight      uint64          `json:"weight"`
	HardGate    bool            `json:"hard_gate"`
	Check       json.RawMessage `json:"check"`
}

type CriterionResult struct {
	Passed   bool   `json:"passed"`
	Evidence string `json:"evidence"`
}

type CanonicalCriteriaScore struct {
	ResultsJSON     string
	PassedWeight    uint64
	TotalWeight     uint64
	HardGatesPassed bool
	Passed          bool
	Resolved        bool
}

func decodeStrictJSON(raw string, target interface{}) error {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values are not allowed")
		}
		return err
	}
	return nil
}

func validateCriterionCheck(raw json.RawMessage) error {
	var check map[string]json.RawMessage
	if len(raw) == 0 || len(raw) > 16*1024 || json.Unmarshal(raw, &check) != nil || len(check) == 0 {
		return fmt.Errorf("check must be a bounded JSON object")
	}
	var checkType string
	if err := json.Unmarshal(check["type"], &checkType); err != nil {
		return fmt.Errorf("check type is required")
	}
	allowed := map[string]map[string]bool{
		"application_runs": {"type": true},
		"stdout_contains":  {"type": true, "text": true},
		"browser_contains": {"type": true, "text": true},
		"browser_fetch":    {"type": true, "path": true},
		"http":             {"type": true, "path": true, "status": true, "content_type": true, "contains": true, "json_keys": true, "json_equals": true},
	}
	fields, exists := allowed[checkType]
	if !exists {
		return fmt.Errorf("unsupported executable check type %q", checkType)
	}
	for field := range check {
		if !fields[field] {
			return fmt.Errorf("unsupported field %q for check type %q", field, checkType)
		}
	}
	if checkType == "application_runs" {
		return nil
	}
	if checkType == "stdout_contains" || checkType == "browser_contains" {
		var text string
		if err := json.Unmarshal(check["text"], &text); err != nil || strings.TrimSpace(text) == "" || len(text) > 1000 {
			return fmt.Errorf("%s check requires text between 1 and 1000 bytes", checkType)
		}
		return nil
	}
	var path string
	if err := json.Unmarshal(check["path"], &path); err != nil || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.Contains(path, "#") {
		return fmt.Errorf("%s check requires a local path starting with /", checkType)
	}
	if checkType == "browser_fetch" {
		return nil
	}
	if value, exists := check["status"]; exists {
		var status int
		if err := json.Unmarshal(value, &status); err != nil || status < 200 || status > 599 {
			return fmt.Errorf("http check status must be between 200 and 599")
		}
	}
	if value, exists := check["content_type"]; exists {
		var contentType string
		if err := json.Unmarshal(value, &contentType); err != nil {
			return fmt.Errorf("http check content_type must be text")
		}
	}
	for _, field := range []string{"contains", "json_keys"} {
		if value, exists := check[field]; exists {
			var items []string
			if err := json.Unmarshal(value, &items); err != nil {
				return fmt.Errorf("http check %s must be a string list", field)
			}
		}
	}
	if value, exists := check["json_equals"]; exists {
		var expected map[string]interface{}
		if err := json.Unmarshal(value, &expected); err != nil || expected == nil {
			return fmt.Errorf("http check json_equals must be an object")
		}
	}
	return nil
}

func ParseAcceptanceCriteria(raw string) ([]AcceptanceCriterion, error) {
	if len(raw) == 0 || len(raw) > MaxCriterionResultsJSON {
		return nil, fmt.Errorf("criteria JSON must be between 1 byte and %d bytes", MaxCriterionResultsJSON)
	}
	var criteria []AcceptanceCriterion
	if err := decodeStrictJSON(raw, &criteria); err != nil {
		return nil, fmt.Errorf("invalid criteria JSON: %w", err)
	}
	if len(criteria) == 0 || len(criteria) > MaxAcceptanceCriteria {
		return nil, fmt.Errorf("criteria must contain between 1 and %d entries", MaxAcceptanceCriteria)
	}
	seen := make(map[string]struct{}, len(criteria))
	var total uint64
	for index, criterion := range criteria {
		if !acceptanceCriterionID.MatchString(criterion.ID) {
			return nil, fmt.Errorf("criteria[%d] has an invalid id", index)
		}
		if _, exists := seen[criterion.ID]; exists {
			return nil, fmt.Errorf("duplicate criterion id %q", criterion.ID)
		}
		seen[criterion.ID] = struct{}{}
		if strings.TrimSpace(criterion.Description) == "" || len(criterion.Description) > 500 {
			return nil, fmt.Errorf("criteria[%d] description must contain 1 to 500 bytes", index)
		}
		if criterion.Weight == 0 || criterion.Weight > MaxAcceptanceCriterionWeight {
			return nil, fmt.Errorf("criteria[%d] weight must be between 1 and %d", index, MaxAcceptanceCriterionWeight)
		}
		if err := validateCriterionCheck(criterion.Check); err != nil {
			return nil, fmt.Errorf("criteria[%d] has invalid executable check: %w", index, err)
		}
		total += criterion.Weight
		if total > MaxAcceptanceWeightTotal {
			return nil, fmt.Errorf("total criterion weight exceeds %d", MaxAcceptanceWeightTotal)
		}
		criteria[index].Description = strings.TrimSpace(criterion.Description)
	}
	return criteria, nil
}

func ParseCriterionResults(criteria []AcceptanceCriterion, raw string) (map[string]CriterionResult, error) {
	if len(raw) == 0 || len(raw) > MaxCriterionResultsJSON {
		return nil, fmt.Errorf("criterion results JSON must be between 1 byte and %d bytes", MaxCriterionResultsJSON)
	}
	var results map[string]CriterionResult
	if err := decodeStrictJSON(raw, &results); err != nil {
		return nil, fmt.Errorf("invalid criterion results JSON: %w", err)
	}
	if len(results) != len(criteria) {
		return nil, fmt.Errorf("criterion results must contain every declared criterion exactly once")
	}
	for _, criterion := range criteria {
		result, exists := results[criterion.ID]
		if !exists {
			return nil, fmt.Errorf("missing result for criterion %q", criterion.ID)
		}
		if strings.TrimSpace(result.Evidence) == "" || len(result.Evidence) > MaxCriterionEvidenceBytes {
			return nil, fmt.Errorf("criterion %q evidence must contain 1 to %d bytes", criterion.ID, MaxCriterionEvidenceBytes)
		}
		result.Evidence = strings.TrimSpace(result.Evidence)
		results[criterion.ID] = result
	}
	return results, nil
}

func ScoreCriterionResults(criteria []AcceptanceCriterion, results map[string]CriterionResult) (passedWeight, totalWeight uint64, hardGatesPassed bool) {
	hardGatesPassed = true
	for _, criterion := range criteria {
		totalWeight += criterion.Weight
		result := results[criterion.ID]
		if result.Passed {
			passedWeight += criterion.Weight
		} else if criterion.HardGate {
			hardGatesPassed = false
		}
	}
	return passedWeight, totalWeight, hardGatesPassed
}

// AggregateCriteriaVotes computes the canonical 3-of-4 result for every rubric item.
// A rubric item without three matching votes remains unresolved while more votes fit.
func AggregateCriteriaVotes(criteria []AcceptanceCriterion, votes []TaskVote) (CanonicalCriteriaScore, error) {
	if len(votes) > CriteriaJudgeCommitteeSize {
		return CanonicalCriteriaScore{}, fmt.Errorf("weighted tasks accept at most %d judge votes", CriteriaJudgeCommitteeSize)
	}
	type judgeResult struct {
		address string
		result  CriterionResult
	}
	canonical := make(map[string]CriterionResult, len(criteria))
	for _, criterion := range criteria {
		var passed, failed []judgeResult
		for _, vote := range votes {
			results, err := ParseCriterionResults(criteria, vote.CriteriaResultsJson)
			if err != nil {
				return CanonicalCriteriaScore{}, err
			}
			entry := judgeResult{address: vote.JudgeAddress, result: results[criterion.ID]}
			if entry.result.Passed {
				passed = append(passed, entry)
			} else {
				failed = append(failed, entry)
			}
		}
		var selected []judgeResult
		passedValue := false
		if len(passed) >= CriteriaJudgeMajority {
			selected = passed
			passedValue = true
		} else if len(failed) >= CriteriaJudgeMajority {
			selected = failed
		} else {
			return CanonicalCriteriaScore{Resolved: false}, nil
		}
		sort.Slice(selected, func(i, j int) bool { return selected[i].address < selected[j].address })
		evidence := make([]string, 0, len(selected))
		for _, judge := range selected {
			evidence = append(evidence, judge.address+": "+judge.result.Evidence)
		}
		canonical[criterion.ID] = CriterionResult{Passed: passedValue, Evidence: strings.Join(evidence, " | ")}
	}

	passedWeight, totalWeight, hardGatesPassed := ScoreCriterionResults(criteria, canonical)
	canonicalJSON, err := json.Marshal(canonical)
	if err != nil {
		return CanonicalCriteriaScore{}, fmt.Errorf("failed to encode canonical criteria results: %w", err)
	}
	passed := hardGatesPassed && passedWeight*100 >= totalWeight*AcceptancePassingPercent
	return CanonicalCriteriaScore{
		ResultsJSON:     string(bytes.TrimSpace(canonicalJSON)),
		PassedWeight:    passedWeight,
		TotalWeight:     totalWeight,
		HardGatesPassed: hardGatesPassed,
		Passed:          passed,
		Resolved:        true,
	}, nil
}

// QuorumAlignedJudges returns Judges whose PASS/FAIL result matches the canonical
// result for every criterion. Evidence text is intentionally excluded from the
// comparison because each Judge supplies independent evidence.
func QuorumAlignedJudges(criteria []AcceptanceCriterion, votes []TaskVote, canonicalJSON string) ([]string, error) {
	var canonical map[string]CriterionResult
	if err := decodeStrictJSON(canonicalJSON, &canonical); err != nil {
		return nil, fmt.Errorf("invalid canonical criteria results: %w", err)
	}
	if len(canonical) != len(criteria) {
		return nil, fmt.Errorf("canonical criteria results do not match rubric")
	}
	for _, criterion := range criteria {
		if _, ok := canonical[criterion.ID]; !ok {
			return nil, fmt.Errorf("canonical results are missing criterion %q", criterion.ID)
		}
	}

	aligned := make([]string, 0, len(votes))
	for _, vote := range votes {
		results, err := ParseCriterionResults(criteria, vote.CriteriaResultsJson)
		if err != nil {
			return nil, fmt.Errorf("invalid results for Judge %q: %w", vote.JudgeAddress, err)
		}
		matches := true
		for _, criterion := range criteria {
			if results[criterion.ID].Passed != canonical[criterion.ID].Passed {
				matches = false
				break
			}
		}
		if matches {
			aligned = append(aligned, vote.JudgeAddress)
		}
	}
	sort.Strings(aligned)
	return aligned, nil
}
