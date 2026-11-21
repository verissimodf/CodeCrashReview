package reviews

import (
	"github.com/google/uuid"
	"testing"
)

func TestDeploymentOperatorUsesPersistedIDOnly(t *testing.T) {
	id := uuid.New()
	t.Setenv("REVIEW_OPERATOR_IDS", id.String())
	if !operator(id) || operator(uuid.New()) || operator(uuid.Nil) {
		t.Fatal("operator identity gate failed")
	}
	t.Setenv("REVIEW_OPERATOR_IDS", "contato@laytech.com.br")
	if operator(id) {
		t.Fatal("email text must not grant authority")
	}
}
