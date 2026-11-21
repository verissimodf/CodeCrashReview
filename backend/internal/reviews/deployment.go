package reviews

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type deploymentTarget struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Repository string `json:"repository"`
	Profile    string `json:"profile"`
	Port       int    `json:"port"`
}

func operator(user uuid.UUID) bool {
	for _, value := range strings.Split(os.Getenv("REVIEW_OPERATOR_IDS"), ",") {
		id, err := uuid.Parse(strings.TrimSpace(value))
		if err == nil && id != uuid.Nil && id == user {
			return true
		}
	}
	return false
}
func deploymentTargets() ([]deploymentTarget, error) {
	file := os.Getenv("DEPLOY_TARGETS_FILE")
	if file == "" {
		return []deploymentTarget{}, nil
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var targets []deploymentTarget
	err = json.Unmarshal(raw, &targets)
	return targets, err
}
func (s *Service) targetsHTTP(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Sign in first."})
		return
	}
	if !operator(user) {
		c.JSON(200, gin.H{"authorized": false, "targets": []deploymentTarget{}})
		return
	}
	targets, err := deploymentTargets()
	if err != nil {
		c.JSON(503, gin.H{"error": "Deployment targets unavailable."})
		return
	}
	c.JSON(200, gin.H{"authorized": true, "targets": targets})
}
func (s *Service) deployHTTP(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok || !operator(user) {
		c.JSON(403, gin.H{"error": "Only the configured deployment operator may deploy."})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid review."})
		return
	}
	var req struct {
		Target string `json:"target"`
		Slot   string `json:"slot"`
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2048)
	if c.ShouldBindJSON(&req) != nil || (req.Slot != "A" && req.Slot != "B") {
		c.JSON(400, gin.H{"error": "Select a target and proposal."})
		return
	}
	targets, err := deploymentTargets()
	if err != nil {
		c.JSON(503, gin.H{"error": "Deployment targets unavailable."})
		return
	}
	var target *deploymentTarget
	for i := range targets {
		if targets[i].ID == req.Target {
			target = &targets[i]
			break
		}
	}
	if target == nil {
		c.JSON(400, gin.H{"error": "Unknown deployment target."})
		return
	}
	ctx := c.Request.Context()
	artifact, dir, _, cleanup, err := s.prepareWinner(ctx, user, id, req.Slot)
	defer cleanup()
	if err != nil {
		c.JSON(409, gin.H{"error": sanitizeError(err)})
		return
	}
	origin := artifact.Source
	if artifact.SourceType == "local" {
		local, pathErr := s.sources.validateLocalPath(origin)
		if pathErr != nil {
			c.JSON(409, gin.H{"error": "Source no longer allowed."})
			return
		}
		origin, err = runGit(ctx, local, "remote", "get-url", "origin")
		if err != nil {
			c.JSON(409, gin.H{"error": "Configure the source GitHub origin."})
			return
		}
	}
	repo, err := githubRepository(strings.TrimSpace(origin))
	if err != nil || repo != target.Repository {
		c.JSON(409, gin.H{"error": "This target does not accept the source repository."})
		return
	}
	var proof *FunctionalProof
	for _, p := range artifact.Proposals {
		if p.Slot == req.Slot {
			proof = p.Verification.Proof
		}
	}
	if proof == nil || proof.Profile != target.Profile {
		c.JSON(409, gin.H{"error": "Target runtime does not match the tested profile."})
		return
	}
	files, err := projectFiles(ctx, dir)
	if err != nil {
		c.JSON(409, gin.H{"error": sanitizeError(err)})
		return
	}
	if manifestHash(files) != proof.SourceSHA256 {
		c.JSON(409, gin.H{"error": "Deployment artifact differs from the tested artifact."})
		return
	}
	receiptID := uuid.New()
	_, err = s.db.Exec(ctx, `INSERT INTO review_deployments(id,review_id,user_id,target,status) VALUES($1,$2,$3,$4,'running')`, receiptID, id, user, target.ID)
	if err != nil {
		c.JSON(409, gin.H{"error": "Target has an active deployment; check its receipt before retrying."})
		return
	}
	// Delivery must finish or restore the previous service even if the browser disconnects.
	execution, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	payload := rawJSON(map[string]any{"target": target.ID, "review_id": id, "files": files, "image": proof.Image, "source_sha256": proof.SourceSHA256})
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", os.Getenv("REVIEW_VALIDATOR_SOCKET"))
	}}
	defer transport.CloseIdleConnections()
	request, _ := http.NewRequestWithContext(execution, "POST", "http://validator/deploy", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Transport: transport}).Do(request)
	var receipt json.RawMessage
	if err == nil {
		defer response.Body.Close()
		if response.StatusCode != 200 {
			err = errors.New("deployment failed; inspect the target and broker receipt before retrying")
		} else {
			receipt, err = io.ReadAll(io.LimitReader(response.Body, 65536))
			if err == nil && !json.Valid(receipt) {
				err = errors.New("invalid deployment receipt")
			}
		}
	}
	if err == nil {
		var confirmed struct {
			Source string `json:"source_sha256"`
			Target string `json:"target"`
			Passed bool   `json:"functional_probe_passed"`
		}
		if json.Unmarshal(receipt, &confirmed) != nil || confirmed.Source != proof.SourceSHA256 || confirmed.Target != target.ID || !confirmed.Passed {
			err = errors.New("deployment receipt does not match the validated artifact")
		}
	}
	saveCtx, saveCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer saveCancel()
	if err != nil {
		s.db.Exec(saveCtx, `UPDATE review_deployments SET status='failed',error_message=$2,completed_at=NOW() WHERE id=$1`, receiptID, sanitizeError(err))
		c.JSON(502, gin.H{"error": "Deployment was not confirmed. Inspect the recorded target state before retrying.", "id": receiptID})
		return
	}
	_, err = s.db.Exec(saveCtx, `UPDATE review_deployments SET status='completed',receipt=$2,completed_at=NOW() WHERE id=$1`, receiptID, receipt)
	if err != nil {
		c.JSON(500, gin.H{"error": "Service responded, but the deployment receipt could not be saved.", "id": receiptID})
		return
	}
	c.Data(200, "application/json", receipt)
}

func (s *Service) deploymentsHTTP(c *gin.Context) {
	user, ok := currentUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if !ok || err != nil {
		c.JSON(400, gin.H{"error": "Invalid review."})
		return
	}
	var data json.RawMessage
	err = s.db.QueryRow(c.Request.Context(), `SELECT COALESCE(jsonb_agg(jsonb_build_object('id',d.id,'target',d.target,'status',d.status,'receipt',d.receipt,'error_message',d.error_message,'created_at',d.created_at) ORDER BY d.created_at DESC),'[]'::jsonb) FROM review_deployments d JOIN repository_reviews r ON r.id=d.review_id WHERE r.id=$1 AND r.user_id=$2`, id, user).Scan(&data)
	if err != nil {
		c.JSON(500, gin.H{"error": "Could not load deployment receipts."})
		return
	}
	c.Data(200, "application/json", data)
}
