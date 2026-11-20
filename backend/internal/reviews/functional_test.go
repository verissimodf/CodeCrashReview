package reviews

import (
	"context"
	"testing"
)

func TestApplicablePatchDoesNotImplyFunctionalApproval(t *testing.T) {
	result := proposalResult{PatchValid: true}
	if proposalApproved(result) {
		t.Fatal("unexecuted patch approved")
	}
	result.Verification.Proof = &FunctionalProof{Passed: false, ExitCode: 1}
	if proposalApproved(result) {
		t.Fatal("failed compilation approved")
	}
	result.Verification.Proof = &FunctionalProof{Passed: true, ExitCode: 0}
	if !proposalApproved(result) {
		t.Fatal("passing proof rejected")
	}
	result.PatchValid = false
	if proposalApproved(result) {
		t.Fatal("inapplicable patch approved")
	}
}

func TestJudgeCannotInventFunctionalApproval(t *testing.T) {
	service := &Service{}
	_, err := service.evaluate(context.Background(), claimedReview{}, []proposalResult{{PatchValid: true}, {PatchValid: true}})
	if err == nil {
		t.Fatal("judge accepted patches without independent functional proof")
	}
}
