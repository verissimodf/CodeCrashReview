package reviews

import (
	"encoding/json"
	"testing"
)

func scorecard(value float64) map[string]any {
	return map[string]any{"correctness": value, "security": value, "maintainability": value, "reliability": value, "efficiency": value, "evidence": value, "total": 1}
}
func TestScoreMeanAndWinnerConsistency(t *testing.T) {
	for _, tc := range []struct {
		winner string
		a, b   float64
		valid  bool
	}{{"A", 90, 80, true}, {"B", 90, 80, false}, {"TIE", 90, 90, true}, {"A", 90, 90, false}, {"A", 101, 80, false}} {
		raw, _ := json.Marshal(Verdict{Winner: tc.winner, Rationale: "Evidence checked", Scores: map[string]any{"A": scorecard(tc.a), "B": scorecard(tc.b)}})
		v, err := parseVerdict(string(raw))
		if (err == nil) != tc.valid {
			t.Errorf("%+v: %v", tc, err)
		}
		if err == nil && v.Scores["A"].(map[string]any)["total"] != tc.a {
			t.Fatal("untrusted total used instead of computed mean")
		}
	}
}
