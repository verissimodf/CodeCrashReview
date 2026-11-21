package reviews

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var validPersonas = map[string]bool{
	"sentinel": true, "optimizer": true, "refactor": true,
	"guardian": true, "docubot": true, "polyglot": true,
}

var errUnverifiedWinner = errors.New("selected proposal has not passed independent functional validation")

var eligiblePersonas = map[string]map[string]bool{
	"comprehensive": {"sentinel": true, "optimizer": true, "refactor": true, "guardian": true, "polyglot": true},
	"security":      {"sentinel": true, "guardian": true, "polyglot": true},
	"reliability":   {"guardian": true, "sentinel": true, "refactor": true},
	"refactoring":   {"refactor": true, "guardian": true, "polyglot": true},
	"performance":   {"optimizer": true, "refactor": true, "guardian": true},
	"architecture":  {"refactor": true, "docubot": true, "guardian": true, "polyglot": true},
	"documentation": {"docubot": true, "refactor": true, "polyglot": true},
}

func (s *Service) Register(protected *gin.RouterGroup) {
	g := protected.Group("/reviews")
	g.POST("", s.createHTTP)
	g.GET("", s.listHTTP)
	g.GET("/deployment-targets", s.targetsHTTP)
	g.GET("/:id", s.getHTTP)
	g.POST("/:id/verdict", s.verdictHTTP)
	g.POST("/:id/publish", s.publishHTTP)
	g.POST("/:id/deploy", s.deployHTTP)
	g.GET("/:id/deployments", s.deploymentsHTTP)
	g.GET("/:id/publications", s.publicationsHTTP)
}

func currentUser(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("user_id")
	user, ok := value.(uuid.UUID)
	return user, exists && ok && user != uuid.Nil
}

func (s *Service) createHTTP(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sign in to start a review."})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32*1024)
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review request."})
		return
	}
	if err := validateCreateRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := s.Create(c.Request.Context(), user, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not queue the review."})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"id": id, "status": "queued"})
}

func (s *Service) listHTTP(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sign in to view reviews."})
		return
	}
	data, err := s.List(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not list reviews."})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

func (s *Service) getHTTP(c *gin.Context) {
	user, ok := currentUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if !ok || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID."})
		return
	}
	data, err := s.Get(c.Request.Context(), user, id)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found."})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load the review."})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

func (s *Service) verdictHTTP(c *gin.Context) {
	user, ok := currentUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if !ok || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID."})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16*1024)
	var req HumanVerdictRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verdict."})
		return
	}
	req.Winner = strings.ToUpper(strings.TrimSpace(req.Winner))
	if req.Winner != "A" && req.Winner != "B" && req.Winner != "TIE" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Winner must be A, B or TIE."})
		return
	}
	if len(strings.TrimSpace(req.Rationale)) < 10 || len(req.Rationale) > 4000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provide a rationale between 10 and 4000 characters."})
		return
	}
	checked, scoreErr := parseVerdict(string(rawJSON(Verdict{Winner: req.Winner, Rationale: req.Rationale, Scores: req.Scores})))
	if scoreErr != nil {
		c.JSON(400, gin.H{"error": scoreErr.Error()})
		return
	}
	req.Scores = checked.Scores
	if err = s.HumanVerdict(c.Request.Context(), user, id, req); errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusConflict, gin.H{"error": "This review is not awaiting your verdict."})
		return
	}
	if err != nil {
		if errors.Is(err, errUnverifiedWinner) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not persist the verdict."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "status": "completed"})
}

func validateCreateRequest(req *CreateRequest) error {
	if req.ValidationProfile == "" {
		req.ValidationProfile = "arena-grade-v1"
	}
	if !map[string]bool{"arena-grade-v1": true, "go-test-v1": true, "python-unittest-v1": true, "node-test-v1": true}[req.ValidationProfile] {
		return errors.New("unsupported validation profile")
	}
	if req.ValidationDirectory == "" {
		req.ValidationDirectory = "."
	}
	if strings.Contains(req.ValidationDirectory, "\\") || filepath.IsAbs(req.ValidationDirectory) || filepath.Clean(req.ValidationDirectory) != req.ValidationDirectory || req.ValidationDirectory == ".." || strings.HasPrefix(req.ValidationDirectory, "../") {
		return errors.New("invalid validation directory")
	}
	req.SourceType = strings.ToLower(strings.TrimSpace(req.SourceType))
	req.Source = strings.TrimSpace(req.Source)
	req.Branch = strings.TrimSpace(req.Branch)
	req.AgentA = strings.ToLower(strings.TrimSpace(req.AgentA))
	req.AgentB = strings.ToLower(strings.TrimSpace(req.AgentB))
	req.EvaluatorMode = strings.ToLower(strings.TrimSpace(req.EvaluatorMode))
	req.ReviewFocus = strings.ToLower(strings.TrimSpace(req.ReviewFocus))
	req.Objective = strings.TrimSpace(req.Objective)
	if req.Branch == "" {
		req.Branch = "main"
	}
	if req.Objective == "" {
		req.Objective = defaultObjective
	}
	if req.ReviewFocus == "" {
		req.ReviewFocus = "comprehensive"
	}
	if req.SourceType != "github" && req.SourceType != "local" {
		return errors.New("source_type must be github or local")
	}
	if req.Source == "" || len(req.Source) > 2000 {
		return errors.New("provide a repository source")
	}
	if req.SourceType == "github" {
		if _, err := validateGitHubURL(req.Source); err != nil {
			return err
		}
	}
	if !branchPattern.MatchString(req.Branch) || strings.Contains(req.Branch, "..") {
		return errors.New("invalid branch name")
	}
	if !validPersonas[req.AgentA] || !validPersonas[req.AgentB] || req.AgentA == req.AgentB {
		return errors.New("select two different valid agents")
	}
	eligible, ok := eligiblePersonas[req.ReviewFocus]
	if !ok {
		return errors.New("select a valid review focus")
	}
	if !eligible[req.AgentA] || !eligible[req.AgentB] {
		return errors.New("both agents must be specialists compatible with the review focus")
	}
	if req.EvaluatorMode != "ai" && req.EvaluatorMode != "human" {
		return errors.New("evaluator_mode must be ai or human")
	}
	if len(req.Objective) > maxObjectiveLen {
		return errors.New("objective exceeds 4000 characters")
	}
	if len(req.FocusPaths) > 20 {
		return errors.New("at most 20 focus paths are allowed")
	}
	for i, value := range req.FocusPaths {
		value = filepath.ToSlash(filepath.Clean(strings.TrimSpace(value)))
		if value == "" || value == "." || value == ".." || strings.HasPrefix(value, "../") || filepath.IsAbs(value) {
			return errors.New("focus paths must be relative repository paths")
		}
		req.FocusPaths[i] = value
	}
	return nil
}

func (s *Service) Create(ctx context.Context, user uuid.UUID, req CreateRequest) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.db.QueryRow(ctx, `INSERT INTO repository_reviews(user_id,source_type,source_location,source_branch,focus_paths,objective,review_focus,agent_a,agent_b,evaluator_mode,validation_profile,validation_directory)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id`, user, req.SourceType, req.Source, req.Branch, rawJSON(req.FocusPaths), req.Objective, req.ReviewFocus, req.AgentA, req.AgentB, req.EvaluatorMode, req.ValidationProfile, req.ValidationDirectory).Scan(&id)
	if err == nil {
		s.notify()
	}
	return id, err
}

func (s *Service) List(ctx context.Context, user uuid.UUID) (json.RawMessage, error) {
	var data json.RawMessage
	err := s.db.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(item) ORDER BY created_at DESC),'[]'::jsonb)
		FROM (SELECT id,source_type,source_location,source_branch,objective,review_focus,agent_a,agent_b,evaluator_mode,status,snapshot_sha,snapshot_files,snapshot_bytes,error_message,created_at,started_at,completed_at
              FROM repository_reviews WHERE user_id=$1 ORDER BY created_at DESC LIMIT 100) item`, user).Scan(&data)
	return data, err
}

func (s *Service) Get(ctx context.Context, user, id uuid.UUID) (json.RawMessage, error) {
	var data json.RawMessage
	err := s.db.QueryRow(ctx, `SELECT to_jsonb(r) || jsonb_build_object(
          'proposals', COALESCE((SELECT jsonb_agg(jsonb_build_object(
              'id',p.id,'slot',p.slot,'persona',p.persona,'model',p.model,'status',p.status,
			  'summary',p.summary,'findings',p.findings,'patch',p.patch,'patch_valid',p.patch_valid,'delivery',p.details,'verification',p.verification,
              'validation_error',p.validation_error,'confidence',p.confidence,'tokens_used',p.tokens_used,
              'latency_ms',p.latency_ms,'error_message',p.error_message,'started_at',p.started_at,'completed_at',p.completed_at
          ) ORDER BY p.slot) FROM review_proposals p WHERE p.review_id=r.id),'[]'::jsonb),
          'verdict', (SELECT jsonb_build_object('evaluator_type',v.evaluator_type,'winner',v.winner,'rationale',v.rationale,'scores',v.scores,'model',v.model,'created_at',v.created_at) FROM review_verdicts v WHERE v.review_id=r.id)
        ) FROM repository_reviews r WHERE r.id=$1 AND r.user_id=$2`, id, user).Scan(&data)
	return data, err
}

func (s *Service) HumanVerdict(ctx context.Context, user, id uuid.UUID, req HumanVerdictRequest) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	var exists bool
	err = tx.QueryRow(ctx, `SELECT TRUE FROM repository_reviews WHERE id=$1 AND user_id=$2 AND status='awaiting_human' FOR UPDATE`, id, user).Scan(&exists)
	if err != nil {
		return err
	}
	verdict := Verdict{Winner: req.Winner, Rationale: strings.TrimSpace(req.Rationale), Scores: req.Scores}
	slots := []string{req.Winner}
	if req.Winner == "TIE" {
		slots = []string{"A", "B"}
	}
	for _, slot := range slots {
		var approved bool
		err = tx.QueryRow(ctx, `SELECT patch_valid AND COALESCE((verification->'proof'->>'passed')::boolean,FALSE) FROM review_proposals WHERE review_id=$1 AND slot=$2`, id, slot).Scan(&approved)
		if err != nil {
			return err
		}
		if !approved {
			return errUnverifiedWinner
		}
	}
	if verdict.Scores == nil {
		verdict.Scores = map[string]any{}
	}
	if _, err = tx.Exec(ctx, `INSERT INTO review_verdicts(review_id,evaluator_type,evaluator_user_id,winner,rationale,scores)
		VALUES($1,'human',$2,$3,$4,$5)`, id, user, verdict.Winner, verdict.Rationale, rawJSON(verdict.Scores)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE repository_reviews SET status='completed',completed_at=NOW(),lease_expires_at=NULL,error_message=NULL WHERE id=$1`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
