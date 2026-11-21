package reviews

import (
	"context"
	"encoding/json"

	"codecrashreview/backend/pkg/freellm"
	"github.com/google/uuid"
)

const (
	defaultObjective = "Review this repository for correctness, security, maintainability, reliability, availability, efficiency and effectiveness. Propose a minimal, testable patch for the most important findings."
	maxObjectiveLen  = 4000
	maxPatchBytes    = 256 * 1024
)

type LLM interface {
	ChatCompletion(context.Context, *freellm.ChatRequest) (*freellm.ChatResponse, error)
}

type CreateRequest struct {
	ValidationProfile   string   `json:"validation_profile"`
	ValidationDirectory string   `json:"validation_directory"`
	SourceType          string   `json:"source_type"`
	Source              string   `json:"source"`
	Branch              string   `json:"branch"`
	FocusPaths          []string `json:"focus_paths"`
	Objective           string   `json:"objective"`
	ReviewFocus         string   `json:"review_focus"`
	AgentA              string   `json:"agent_a"`
	AgentB              string   `json:"agent_b"`
	EvaluatorMode       string   `json:"evaluator_mode"`
}

type HumanVerdictRequest struct {
	Winner    string         `json:"winner"`
	Rationale string         `json:"rationale"`
	Scores    map[string]any `json:"scores"`
}

type claimedReview struct {
	ValidationProfile   string
	ValidationDirectory string
	ID                  uuid.UUID
	UserID              uuid.UUID
	SourceType          string
	Source              string
	Branch              string
	FocusPaths          []string
	Objective           string
	ReviewFocus         string
	AgentA              string
	AgentB              string
	EvaluatorMode       string
}

type Finding struct {
	Severity       string `json:"severity"`
	Category       string `json:"category"`
	File           string `json:"file"`
	Line           int    `json:"line"`
	Title          string `json:"title"`
	Evidence       string `json:"evidence"`
	Recommendation string `json:"recommendation"`
}

type Proposal struct {
	Summary    string    `json:"summary"`
	Findings   []Finding `json:"findings"`
	Edits      []Edit    `json:"edits,omitempty"`
	Patch      string    `json:"patch,omitempty"`
	Delivery   Delivery  `json:"delivery"`
	Confidence float64   `json:"confidence"`
}

type Delivery struct {
	Topology              string   `json:"topology"`
	Design                string   `json:"design"`
	DeploymentPlan        []string `json:"deployment_plan"`
	ApprovalCriteria      []string `json:"approval_criteria"`
	FinalObjective        string   `json:"final_objective"`
	Methodology           []string `json:"methodology"`
	EffectivenessEvidence []string `json:"effectiveness_evidence"`
	ExecutedSteps         []string `json:"executed_steps"`
}

type Edit struct {
	File string `json:"file"`
	Old  string `json:"old"`
	New  string `json:"new"`
}

type proposalResult struct {
	Slot            string
	Persona         string
	Model           string
	Proposal        Proposal
	PatchValid      bool
	ValidationError string
	Tokens          int
	LatencyMS       int64
	Verification    Verification
	Err             error
}

type Verification struct {
	PreviousPatches    []string          `json:"previous_patches,omitempty"`
	PreviousProofs     []FunctionalProof `json:"previous_proofs,omitempty"`
	Proof              *FunctionalProof  `json:"proof,omitempty"`
	PatchApplicable    bool              `json:"patch_applicable"`
	ExactEditsVerified bool              `json:"exact_edits_verified"`
	FunctionalTestsRun bool              `json:"functional_tests_run"`
	QualityState       string            `json:"quality_state"`
	SystemEvidence     []string          `json:"system_evidence"`
	UnsupportedClaims  []string          `json:"unsupported_claims"`
}

type FunctionalProof struct {
	Baseline       *TestExecution `json:"baseline,omitempty"`
	BaselineSHA256 string         `json:"baseline_sha256,omitempty"`
	TestsPreserved bool           `json:"tests_preserved,omitempty"`
	TestCount      int            `json:"test_count,omitempty"`
	Profile        string         `json:"profile"`
	Passed         bool           `json:"passed"`
	ExitCode       int            `json:"exit_code"`
	SourceSHA256   string         `json:"source_sha256"`
	TestsSHA256    string         `json:"tests_sha256"`
	Image          string         `json:"image"`
	Output         string         `json:"output"`
}

type TestExecution struct {
	Passed    bool   `json:"passed"`
	ExitCode  int    `json:"exit_code"`
	TestCount int    `json:"test_count"`
	Output    string `json:"output"`
	Image     string `json:"image"`
}

type Verdict struct {
	Winner    string         `json:"winner"`
	Rationale string         `json:"rationale"`
	Scores    map[string]any `json:"scores"`
}

type Snapshot struct {
	IncludedFiles       []string
	ValidationProfile   string
	ValidationDirectory string
	Dir                 string
	SHA                 string
	Context             string
	FileCount           int
	ByteCount           int64
}

func rawJSON(value any) json.RawMessage {
	b, _ := json.Marshal(value)
	return b
}
