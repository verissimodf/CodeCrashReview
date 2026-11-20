package arena

import (
	"math"
	"testing"
)

func TestGradeAcceptance(t *testing.T) {
	for _, tc := range []struct {
		name, code string
		cases      [][3]int64
		score      int
	}{
		{"signs_and_zero", "a+b", [][3]int64{{2, 3, 5}, {-2, -3, -5}, {-2, 2, 0}}, 100},
		{"partial", "a", [][3]int64{{2, 0, 2}, {2, 0, 3}, {2, 0, 4}}, 33},
		{"wrong", "a-b", [][3]int64{{2, 3, 5}}, 0},
		{"large_intermediate", "a*b-a*b+a", [][3]int64{{math.MaxInt64, math.MaxInt64, math.MaxInt64}, {math.MinInt64, math.MaxInt64, math.MinInt64}}, 100},
		{"absolute", "abs(a)", [][3]int64{{-3, 0, 3}, {0, 0, 0}}, 100},
		{"minimum", "min(a,b)", [][3]int64{{-3, 4, -3}}, 100},
		{"maximum", "max(a,b)", [][3]int64{{-3, 4, 4}}, 100},
		{"function", "def solve(a, b):\n    return a+b", [][3]int64{{2, 3, 5}}, 100},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := grade(tc.code, tc.cases)
			if err != nil || got.Score != tc.score {
				t.Fatalf("score=%d err=%v; want %d", got.Score, err, tc.score)
			}
		})
	}
}

func TestGradeRejectsInvalidInput(t *testing.T) {
	for _, code := range []string{"", "a/b", "exec(a)", "abs(a,b)", "min(a)", "a; b", "9999999999", "unknown", "a**b"} {
		if _, err := grade(code, [][3]int64{{1, 2, 3}}); err == nil {
			t.Errorf("accepted %q", code)
		}
	}
	if _, err := grade("a+b", nil); err == nil {
		t.Fatal("accepted missing acceptance cases")
	}
}
