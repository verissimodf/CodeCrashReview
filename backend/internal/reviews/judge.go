package reviews

import (
	"codecrashreview/backend/pkg/freellm"
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

func excerpt(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + " [truncated]"
}

// Full evidence stays in PostgreSQL. A bounded summary avoids overflowing local
// model context with test logs, repeated repair attempts and generated plans.
func judgeSummary(p proposalResult) map[string]any {
	claims := []string{}
	for i, claim := range p.Verification.UnsupportedClaims {
		if i == 2 {
			break
		}
		claims = append(claims, excerpt(claim, 160))
	}
	findings := []string{}
	for i, finding := range p.Proposal.Findings {
		if i == 2 {
			break
		}
		findings = append(findings, excerpt(finding.Title+": "+finding.Evidence, 200))
	}
	proof := map[string]any{"passed": false, "profile": "unavailable"}
	if p.Verification.Proof != nil {
		proof = map[string]any{"passed": p.Verification.Proof.Passed, "profile": p.Verification.Proof.Profile, "exit_code": p.Verification.Proof.ExitCode}
	}
	return map[string]any{"persona": p.Persona, "summary": excerpt(p.Proposal.Summary, 250), "findings": findings,
		"design": excerpt(p.Proposal.Delivery.Design, 250), "topology": excerpt(p.Proposal.Delivery.Topology, 150),
		"patch_excerpt": excerpt(p.Proposal.Patch, 1000), "patch_applicable": p.PatchValid,
		"verification": map[string]any{"system_evidence": proof, "unsupported_claims": claims}}
}

func (s *Service) judgeWithEvidence(ctx context.Context, job claimedReview, proposals []proposalResult) (Verdict, error) {
	payload, _ := json.Marshal(map[string]any{"objective": excerpt(job.Objective, 500), "review_focus": job.ReviewFocus, "proposal_a": judgeSummary(proposals[0]), "proposal_b": judgeSummary(proposals[1])})
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		instruction := judgeSystemPrompt
		if lastErr != nil {
			instruction += "\nYour previous response was rejected: " + lastErr.Error() + ". Use the literal winner A, B or TIE and numeric scorecards for both slots. Return the corrected JSON object only."
		}
		response, err := s.llm.ChatCompletion(ctx, &freellm.ChatRequest{Model: s.model, Temperature: 0.1, MaxTokens: 900, ResponseFormat: map[string]any{"type": "json_object"}, Messages: []freellm.ChatMessage{{Role: "system", Content: instruction}, {Role: "user", Content: string(payload)}}})
		if err != nil {
			return Verdict{}, err
		}
		raw := ""
		var verdict Verdict
		if len(response.Choices) == 0 {
			err = errors.New("judge returned no choices")
		} else {
			raw = response.Choices[0].Message.Content
			verdict, err = parseVerdict(raw)
		}
		if err == nil && proposalApproved(proposals[0]) != proposalApproved(proposals[1]) {
			winner := "A"
			if proposalApproved(proposals[1]) {
				winner = "B"
			}
			if verdict.Winner != winner {
				err = errors.New("judge must select the proposal with passing independent proof")
			}
		}
		if s.db != nil {
			record := map[string]any{"attempt": attempt, "model": s.model, "response": excerpt(raw, 8000), "error": sanitizeError(err)}
			if _, saveErr := s.db.Exec(ctx, `UPDATE repository_reviews SET evaluation_attempts=evaluation_attempts || $2::jsonb WHERE id=$1`, job.ID, rawJSON([]any{record})); saveErr != nil {
				return Verdict{}, fmt.Errorf("persist judge evidence: %w", saveErr)
			}
		}
		if err == nil {
			return verdict, nil
		}
		lastErr = err
	}
	return Verdict{}, lastErr
}
