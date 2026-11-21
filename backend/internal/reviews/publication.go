package reviews

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type publicationRequest struct {
	Token   string `json:"token,omitempty"`
	Slot    string `json:"slot"`
	Branch  string `json:"branch"`
	Tag     string `json:"tag"`
	Release bool   `json:"release"`
	Title   string `json:"title"`
}

type winnerArtifact struct {
	SourceType string   `json:"source_type"`
	Source     string   `json:"source_location"`
	Branch     string   `json:"source_branch"`
	SHA        string   `json:"snapshot_sha"`
	Status     string   `json:"status"`
	Verdict    *Verdict `json:"verdict"`
	Proposals  []struct {
		Slot         string       `json:"slot"`
		Patch        string       `json:"patch"`
		PatchValid   bool         `json:"patch_valid"`
		Verification Verification `json:"verification"`
	} `json:"proposals"`
}

func (s *Service) prepareWinner(ctx context.Context, user, id uuid.UUID, slot string) (winnerArtifact, string, map[string]publishedFile, func(), error) {
	cleanup := func() {}
	var artifact winnerArtifact
	raw, err := s.Get(ctx, user, id)
	if err != nil {
		return artifact, "", nil, cleanup, err
	}
	if err = json.Unmarshal(raw, &artifact); err != nil {
		return artifact, "", nil, cleanup, err
	}
	if artifact.Status != "completed" || artifact.Verdict == nil {
		return artifact, "", nil, cleanup, errors.New("review is not completed")
	}
	if artifact.Verdict.Winner != "TIE" && artifact.Verdict.Winner != slot {
		return artifact, "", nil, cleanup, errors.New("only the selected winner may be delivered")
	}
	var patch string
	var proof *FunctionalProof
	for _, p := range artifact.Proposals {
		if p.Slot == slot && p.PatchValid {
			patch = p.Patch
			proof = p.Verification.Proof
		}
	}
	if proof == nil || !proof.Passed || patch == "" {
		return artifact, "", nil, cleanup, errUnverifiedWinner
	}
	snapshot := filepath.Join(s.sources.workspaceRoot, id.String(), "source")
	head, err := runGit(ctx, snapshot, "rev-parse", "HEAD")
	if err != nil || strings.TrimSpace(head) != artifact.SHA {
		return artifact, "", nil, cleanup, errors.New("review snapshot is no longer available or changed")
	}
	workspace, err := os.MkdirTemp(s.sources.workspaceRoot, "delivery-")
	if err != nil {
		return artifact, "", nil, cleanup, err
	}
	cleanup = func() { os.RemoveAll(workspace) }
	dir := filepath.Join(workspace, "source")
	if _, err = runGit(ctx, workspace, "-c", "protocol.file.allow=always", "clone", "--quiet", "--no-hardlinks", "--", snapshot, dir); err != nil {
		return artifact, "", nil, cleanup, err
	}
	patchFile := filepath.Join(workspace, "winner.patch")
	if err = os.WriteFile(patchFile, []byte(strings.TrimSpace(patch)+"\n"), 0600); err != nil {
		return artifact, "", nil, cleanup, err
	}
	if _, err = runGit(ctx, dir, "apply", "--recount", "--whitespace=error-all", "--", patchFile); err != nil {
		return artifact, "", nil, cleanup, errors.New("winning patch no longer applies")
	}
	files, err := projectFiles(ctx, dir)
	if err != nil {
		return artifact, "", nil, cleanup, err
	}
	if proof.Profile == "arena-grade-v1" {
		digest := sha256.Sum256(files["backend/internal/arena/grade.go"])
		if hex.EncodeToString(digest[:]) != proof.SourceSHA256 {
			return artifact, "", nil, cleanup, errors.New("winning source does not match functional proof")
		}
	} else if manifestHash(files) != proof.SourceSHA256 {
		return artifact, "", nil, cleanup, errors.New("winning project does not match functional proof")
	}
	changed, err := runGit(ctx, dir, "diff", "--name-only", "-z")
	if err != nil {
		return artifact, "", nil, cleanup, err
	}
	publication := map[string]publishedFile{}
	for _, name := range strings.Split(changed, "\x00") {
		if name == "" {
			continue
		}
		content, ok := files[name]
		if !ok {
			return artifact, "", nil, cleanup, errors.New("patch changes a file outside the verified project")
		}
		if proof.Profile == "arena-grade-v1" && name != "backend/internal/arena/grade.go" {
			return artifact, "", nil, cleanup, errors.New("patch exceeds the grading proof scope")
		}
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			return artifact, "", nil, cleanup, err
		}
		mode := "100644"
		if info.Mode()&0111 != 0 {
			mode = "100755"
		}
		publication[name] = publishedFile{Content: content, Mode: mode}
	}
	if len(publication) == 0 || len(publication) > 5 {
		return artifact, "", nil, cleanup, errors.New("invalid delivery change count")
	}
	return artifact, dir, publication, cleanup, nil
}

func (s *Service) publishHTTP(c *gin.Context) {
	user, ok := currentUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if !ok || err != nil {
		c.JSON(400, gin.H{"error": "Invalid review."})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16*1024)
	var req publicationRequest
	if c.ShouldBindJSON(&req) != nil || len(req.Token) < 20 || len(req.Token) > 1024 || strings.ContainsAny(req.Token, "\r\n\t ") {
		c.JSON(400, gin.H{"error": "Provide a valid transient GitHub token."})
		return
	}
	if (req.Slot != "A" && req.Slot != "B") || len(req.Branch) > 120 || (!strings.HasPrefix(req.Branch, "review/") && !strings.HasPrefix(req.Branch, "hotfix/") && !strings.HasPrefix(req.Branch, "release/")) || len(req.Tag) > 120 || (req.Release && req.Tag == "") || len(req.Title) > 200 {
		c.JSON(400, gin.H{"error": "Select a proposal, a new review/, hotfix/ or release/ branch, and a tag when creating a release."})
		return
	}
	if req.Title == "" {
		req.Title = "CodeCrashReview " + id.String()
	}
	ctx := c.Request.Context()
	artifact, dir, files, cleanup, err := s.prepareWinner(ctx, user, id, req.Slot)
	defer cleanup()
	if err != nil {
		c.JSON(409, gin.H{"error": sanitizeError(err)})
		return
	}
	for _, ref := range []string{"refs/heads/" + req.Branch, "refs/tags/" + req.Tag} {
		if ref == "refs/tags/" {
			continue
		}
		if _, err = runGit(ctx, dir, "check-ref-format", ref); err != nil {
			c.JSON(400, gin.H{"error": "Invalid publication reference."})
			return
		}
	}
	origin := artifact.Source
	if artifact.SourceType == "local" {
		local, pathErr := s.sources.validateLocalPath(origin)
		if pathErr != nil {
			c.JSON(409, gin.H{"error": "Local origin is no longer allowed."})
			return
		}
		origin, err = runGit(ctx, local, "remote", "get-url", "origin")
		if err != nil {
			c.JSON(409, gin.H{"error": "Configure the local repository's GitHub origin before publishing."})
			return
		}
	}
	repo, err := githubRepository(strings.TrimSpace(origin))
	if err != nil {
		c.JSON(409, gin.H{"error": "The source repository must have a GitHub origin."})
		return
	}
	publicReq := req
	publicReq.Token = ""
	config := rawJSON(publicReq)
	_, err = s.db.Exec(ctx, `INSERT INTO review_publications(review_id,user_id,branch,request) VALUES($1,$2,$3,$4) ON CONFLICT(review_id,branch) DO NOTHING`, id, user, req.Branch, config)
	if err != nil {
		c.JSON(500, gin.H{"error": "Could not reserve publication."})
		return
	}
	var publicationID uuid.UUID
	var commit string
	err = s.db.QueryRow(ctx, `UPDATE review_publications SET status='running',lease_expires_at=NOW()+INTERVAL '10 minutes',error_message=NULL,updated_at=NOW() WHERE review_id=$1 AND branch=$2 AND request=$3::jsonb AND status!='completed' AND (lease_expires_at IS NULL OR lease_expires_at<NOW()) RETURNING id,commit_sha`, id, req.Branch, config).Scan(&publicationID, &commit)
	if errors.Is(err, pgx.ErrNoRows) {
		var status string
		var receipt json.RawMessage
		err = s.db.QueryRow(ctx, `SELECT status,receipt FROM review_publications WHERE review_id=$1 AND branch=$2 AND request=$3::jsonb`, id, req.Branch, config).Scan(&status, &receipt)
		if err == nil && status == "completed" {
			c.Data(200, "application/json", receipt)
			return
		}
		c.JSON(409, gin.H{"error": "Publication already running or branch reserved with different options."})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "Could not claim publication."})
		return
	}
	g := newGitHub(req.Token)
	defer func() { g.token = ""; req.Token = "" }()
	receipt := map[string]any{"id": publicationID, "repository": repo, "branch": req.Branch, "tag": req.Tag, "slot": req.Slot, "snapshot_sha": artifact.SHA}
	fail := func(problem error) {
		receipt["commit_sha"] = commit
		saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, saveErr := s.db.Exec(saveCtx, `UPDATE review_publications SET status='failed',receipt=$2,error_message=$3,lease_expires_at=NULL,updated_at=NOW() WHERE id=$1`, publicationID, rawJSON(receipt), sanitizeError(problem))
		if saveErr != nil {
			c.JSON(500, gin.H{"error": "Publication interrupted; inspect the reserved branch before retrying."})
			return
		}
		c.JSON(502, gin.H{"error": sanitizeError(problem), "receipt": receipt})
	}
	if commit == "" {
		current, readErr := g.ref(ctx, repo, "heads/"+artifact.Branch)
		if readErr != nil {
			fail(readErr)
			return
		}
		if current != artifact.SHA {
			fail(errors.New("source branch advanced; run a new review before publishing"))
			return
		}
		commit, err = g.commit(ctx, repo, artifact.SHA, req.Title+"\n\nValidated review "+id.String(), files)
		if err != nil {
			fail(err)
			return
		}
		if _, err = s.db.Exec(ctx, `UPDATE review_publications SET commit_sha=$2 WHERE id=$1`, publicationID, commit); err != nil {
			fail(err)
			return
		}
	}
	if err = g.ensureRef(ctx, repo, "heads/"+req.Branch, commit); err != nil {
		fail(err)
		return
	}
	receipt["commit_sha"] = commit
	receipt["branch_url"] = "https://github.com/" + repo + "/tree/" + url.PathEscape(req.Branch)
	if req.Tag != "" {
		if err = g.ensureRef(ctx, repo, "tags/"+req.Tag, commit); err != nil {
			fail(err)
			return
		}
	}
	if req.Release {
		releaseURL, releaseErr := g.release(ctx, repo, req.Tag, commit, req.Title, fmt.Sprintf("Verified review %s. Source snapshot %s. Proposal %s. Functional proof stored in CodeCrashReview.", id, artifact.SHA, req.Slot))
		if releaseErr != nil {
			fail(releaseErr)
			return
		}
		receipt["release_url"] = releaseURL
	}
	_, err = s.db.Exec(ctx, `UPDATE review_publications SET status='completed',receipt=$2,lease_expires_at=NULL,updated_at=NOW() WHERE id=$1`, publicationID, rawJSON(receipt))
	if err != nil {
		fail(err)
		return
	}
	c.JSON(200, receipt)
}

func (s *Service) publicationsHTTP(c *gin.Context) {
	user, ok := currentUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if !ok || err != nil {
		c.JSON(400, gin.H{"error": "Invalid review."})
		return
	}
	var data json.RawMessage
	err = s.db.QueryRow(c.Request.Context(), `SELECT COALESCE(jsonb_agg(jsonb_build_object('id',p.id,'status',p.status,'receipt',p.receipt,'error_message',p.error_message,'created_at',p.created_at) ORDER BY p.created_at DESC),'[]'::jsonb) FROM review_publications p JOIN repository_reviews r ON r.id=p.review_id WHERE r.id=$1 AND r.user_id=$2`, id, user).Scan(&data)
	if err != nil {
		c.JSON(500, gin.H{"error": "Could not load publication receipts."})
		return
	}
	c.Data(200, "application/json", data)
}
