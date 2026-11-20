package reviews

import "strings"

func verifyProposal(proposal Proposal, patchValid bool, validationError string) Verification {
	verification := Verification{
		PatchApplicable: patchValid, ExactEditsVerified: patchValid && len(proposal.Edits) > 0,
		FunctionalTestsRun: false, QualityState: "requires_functional_validation",
		SystemEvidence: []string{"No functional test command was executed by the review service."},
	}
	if patchValid {
		verification.QualityState = "patch_applicability_verified"
		if verification.ExactEditsVerified {
			verification.SystemEvidence = append(verification.SystemEvidence, "The patch was generated from an exact, unique replacement in an isolated clone.")
		}
		verification.SystemEvidence = append(verification.SystemEvidence, "git apply --check --recount --whitespace=error-all accepted the patch against the immutable snapshot.")
	} else if strings.TrimSpace(validationError) != "" {
		verification.SystemEvidence = append(verification.SystemEvidence, "Patch applicability failed: "+validationError)
	}

	claims := append([]string{}, proposal.Delivery.Methodology...)
	claims = append(claims, proposal.Delivery.EffectivenessEvidence...)
	claims = append(claims, proposal.Delivery.ExecutedSteps...)
	for _, value := range claims {
		lower := strings.ToLower(value)
		if strings.Contains(lower, "test") || strings.Contains(lower, "static analysis") || strings.Contains(lower, "análise estática") || strings.Contains(lower, "benchmark") || strings.Contains(lower, "compil") || strings.Contains(lower, "passes") || strings.Contains(lower, "passou") {
			verification.UnsupportedClaims = append(verification.UnsupportedClaims, value)
		}
	}
	return verification
}
