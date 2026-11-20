package reviews

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"codecrashreview/backend/pkg/freellm"
)

type barrierLLM struct {
	started chan struct{}
	release chan struct{}
	body    string
}

func (f *barrierLLM) ChatCompletion(ctx context.Context, _ *freellm.ChatRequest) (*freellm.ChatResponse, error) {
	f.started <- struct{}{}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-f.release:
	}
	return &freellm.ChatResponse{Choices: []freellm.Choice{{Message: freellm.ChatMessage{Role: "assistant", Content: f.body}}}, Usage: freellm.Usage{TotalTokens: 10}}, nil
}

func TestRunAgentsStartsBothIndependently(t *testing.T) {
	repo := t.TempDir()
	if _, err := runGit(context.Background(), repo, "init"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "file.txt"), []byte("old\n"), 0600); err != nil {
		t.Fatal(err)
	}
	patch := "diff --git a/file.txt b/file.txt\n--- a/file.txt\n+++ b/file.txt\n@@ -1 +1 @@\n-old\n+new"
	body, _ := json.Marshal(Proposal{Summary: "Fix the observed defect.", Findings: []Finding{{Severity: "high", Category: "correctness", File: "file.txt", Line: 1, Title: "Wrong value", Evidence: "The fixture contains old.", Recommendation: "Replace it."}}, Patch: patch, Delivery: testDelivery(), Confidence: .8})
	fake := &barrierLLM{started: make(chan struct{}, 2), release: make(chan struct{}), body: string(body)}
	service := newService(nil, fake, "test-model", nil, time.Minute)
	done := make(chan []proposalResult, 1)
	go func() {
		done <- service.runAgents(context.Background(), claimedReview{AgentA: "sentinel", AgentB: "refactor", Objective: "review", ReviewFocus: "comprehensive"}, Snapshot{Dir: repo, SHA: "abc", Context: "file"})
	}()
	for i := 0; i < 2; i++ {
		select {
		case <-fake.started:
		case <-time.After(time.Second):
			t.Fatal("agents did not start concurrently")
		}
	}
	close(fake.release)
	select {
	case results := <-done:
		if len(results) != 2 || results[0].Err != nil || results[1].Err != nil {
			t.Fatalf("unexpected results: %#v", results)
		}
		if !results[0].PatchValid || !results[1].PatchValid {
			t.Fatal("agent patches were not validated")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("agents did not finish")
	}
}

func testDelivery() Delivery {
	return Delivery{
		Topology: "One file in the fixture.", Design: "Replace the incorrect value.",
		DeploymentPlan: []string{"Apply the patch."}, ApprovalCriteria: []string{"The patch applies cleanly."},
		FinalObjective: "The fixture contains the corrected value.", Methodology: []string{"Inspected the supplied file."},
		EffectivenessEvidence: []string{"Exact replacement is validated by git apply."}, ExecutedSteps: []string{"Compared the old and new values."},
	}
}

func TestParseVerdictRejectsInventedWinner(t *testing.T) {
	if _, err := parseVerdict(`{"winner":"C","rationale":"bad","scores":{"A":1}}`); err == nil {
		t.Fatal("invalid winner accepted")
	}
}

func TestEditsProduceApplicablePatch(t *testing.T) {
	repo := t.TempDir()
	if _, err := runGit(context.Background(), repo, "init"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "file.txt"), []byte("before\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := runGit(context.Background(), repo, "add", "file.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := runGit(context.Background(), repo, "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-m", "fixture"); err != nil {
		t.Fatal(err)
	}
	patch, err := editsToPatch(context.Background(), repo, []Edit{{File: "file.txt", Old: "before\n", New: "after\n"}})
	if err != nil {
		t.Fatal(err)
	}
	valid, reason := validatePatch(context.Background(), repo, patch)
	if !valid {
		t.Fatalf("generated patch is invalid: %s\n%s", reason, patch)
	}
}

func TestEditsRejectPathTraversal(t *testing.T) {
	if _, err := editsToPatch(context.Background(), t.TempDir(), []Edit{{File: "../outside", Old: "a", New: "b"}}); err == nil {
		t.Fatal("path traversal edit accepted")
	}
}

func TestVerificationSeparatesUnsupportedModelClaims(t *testing.T) {
	proposal := Proposal{Edits: []Edit{{File: "file.txt", Old: "a", New: "b"}}, Delivery: testDelivery()}
	proposal.Delivery.EffectivenessEvidence = []string{"All tests passed."}
	verification := verifyProposal(proposal, true, "")
	if !verification.PatchApplicable || verification.FunctionalTestsRun {
		t.Fatalf("unexpected verification flags: %#v", verification)
	}
	if len(verification.UnsupportedClaims) != 1 {
		t.Fatalf("unsupported claim was not separated: %#v", verification)
	}
}
