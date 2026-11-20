package reviews

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// This profile proves only the published arena grading contract. Other source
// files require their own trusted acceptance profile, never model-authored tests.
func functionalProof(ctx context.Context, snapshot Snapshot, proposal Proposal) (*FunctionalProof, error) {
	const target = "backend/internal/arena/grade.go"
	socket := os.Getenv("REVIEW_VALIDATOR_SOCKET")
	if socket == "" {
		return nil, nil
	}
	if len(proposal.Edits) == 0 {
		return nil, nil
	}
	for _, edit := range proposal.Edits {
		if edit.File != target {
			return nil, nil
		}
	}
	data, err := os.ReadFile(filepath.Join(snapshot.Dir, target))
	if err != nil {
		return nil, err
	}
	source := string(data)
	for _, edit := range proposal.Edits {
		if strings.Count(source, edit.Old) != 1 {
			return nil, fmt.Errorf("functional proof cannot reproduce exact edit")
		}
		source = strings.Replace(source, edit.Old, edit.New, 1)
	}
	if err := validateGradeProfile(source); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]string{"profile": "arena-grade-v1", "source": source})
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socket)
	}}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 200 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://validator/validate", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("functional validator unavailable: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("functional validator returned %d", response.StatusCode)
	}
	var proof FunctionalProof
	if err := json.NewDecoder(io.LimitReader(response.Body, 32768)).Decode(&proof); err != nil {
		return nil, err
	}
	digest := sha256.Sum256([]byte(source))
	if proof.Profile != "arena-grade-v1" || proof.SourceSHA256 != hex.EncodeToString(digest[:]) || proof.TestsSHA256 == "" || proof.Image == "" || proof.Passed != (proof.ExitCode == 0) {
		return nil, fmt.Errorf("functional proof identity mismatch")
	}
	return &proof, nil
}

func validateGradeProfile(source string) error {
	if strings.Contains(source, "//go:") {
		return fmt.Errorf("compiler directives are outside the grading profile")
	}
	file, err := parser.ParseFile(token.NewFileSet(), "grade.go", source, 0)
	if err != nil {
		return err
	}
	if file.Name.Name != "arena" {
		return fmt.Errorf("unexpected grading package")
	}
	allowed := map[string]bool{"errors": true, "fmt": true, "go/ast": true, "go/parser": true, "go/token": true, "math/big": true, "regexp": true, "strings": true}
	for _, spec := range file.Imports {
		name, err := strconv.Unquote(spec.Path.Value)
		if err != nil || !allowed[name] {
			return fmt.Errorf("import outside trusted grading profile: %s", spec.Path.Value)
		}
	}
	for _, declaration := range file.Decls {
		if fn, ok := declaration.(*ast.FuncDecl); ok && (fn.Name.Name == "init" || fn.Recv != nil) {
			return fmt.Errorf("initializers and methods are outside the grading profile")
		}
	}
	return nil
}

func proposalApproved(result proposalResult) bool {
	return result.PatchValid && result.Verification.Proof != nil && result.Verification.Proof.Passed
}
