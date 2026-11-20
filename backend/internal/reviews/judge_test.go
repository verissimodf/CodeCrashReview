package reviews

import (
	"codecrashreview/backend/pkg/freellm"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

type judgeSequence struct {
	calls       int
	lastPayload string
}

func (f *judgeSequence) ChatCompletion(_ context.Context, req *freellm.ChatRequest) (*freellm.ChatResponse, error) {
	f.calls++
	f.lastPayload = req.Messages[1].Content
	winner := "B"
	if f.calls == 1 {
		winner = "proposal B"
	}
	score := map[string]any{"correctness": 90, "security": 90, "maintainability": 90, "reliability": 90, "efficiency": 90, "evidence": 90, "total": 90}
	raw, _ := json.Marshal(Verdict{Winner: winner, Rationale: "Independent acceptance proof passed.", Scores: map[string]any{"A": score, "B": score}})
	return &freellm.ChatResponse{Choices: []freellm.Choice{{Message: freellm.ChatMessage{Content: string(raw)}}}}, nil
}

func TestJudgeRetriesMalformedWinnerAndBoundsContext(t *testing.T) {
	fake := &judgeSequence{}
	service := &Service{llm: fake, model: "fixture"}
	p := proposalResult{PatchValid: true, Proposal: Proposal{Summary: strings.Repeat("summary", 1000), Patch: strings.Repeat("patch", 1000)}, Verification: Verification{Proof: &FunctionalProof{Passed: true, Output: strings.Repeat("logs", 10000)}}}
	verdict, err := service.evaluate(context.Background(), claimedReview{Objective: strings.Repeat("objective", 1000)}, []proposalResult{p, p})
	if err != nil || verdict.Winner != "B" || fake.calls != 2 {
		t.Fatalf("verdict=%+v calls=%d err=%v", verdict, fake.calls, err)
	}
	if len(fake.lastPayload) > 7000 || strings.Contains(fake.lastPayload, "logslogs") {
		t.Fatal("unbounded logs entered judge context")
	}
}
