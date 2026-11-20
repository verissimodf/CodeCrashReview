package reviews

import "testing"

func TestValidateCreateRequest(t *testing.T) {
	req := CreateRequest{
		SourceType: "github", Source: "https://github.com/example/project", Branch: "main",
		AgentA: "sentinel", AgentB: "guardian", EvaluatorMode: "ai",
	}
	if err := validateCreateRequest(&req); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	if req.Objective != defaultObjective {
		t.Fatal("default objective not applied")
	}
	if req.ReviewFocus != "comprehensive" {
		t.Fatal("default review focus not applied")
	}
	req.AgentB = "sentinel"
	if err := validateCreateRequest(&req); err == nil {
		t.Fatal("duplicate agents accepted")
	}
	req.AgentB = "docubot"
	req.ReviewFocus = "security"
	if err := validateCreateRequest(&req); err == nil {
		t.Fatal("incompatible specialist accepted")
	}
}
