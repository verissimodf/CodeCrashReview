package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"codecrashreview/backend/internal/database"
)

type Service struct {
	challengeRepo *database.ChallengeRepository
	templateRepo  *database.ChallengeTemplateRepository
	pool          *pgxpool.Pool
	logger        zerolog.Logger
}

func NewService(challengeRepo *database.ChallengeRepository, templateRepo *database.ChallengeTemplateRepository, pool *pgxpool.Pool) *Service {
	return &Service{
		challengeRepo: challengeRepo,
		templateRepo:  templateRepo,
		pool:          pool,
		logger:        zerolog.DefaultContextLogger,
	}
}

func (s *Service) RegisterRoutes(router *gin.RouterGroup) {
	challenges := router.Group("/challenges")
	{
		challenges.POST("", s.CreateChallenge)
		challenges.GET("", s.ListChallenges)
		challenges.GET("/templates", s.ListTemplates)
		challenges.GET("/:id", s.GetChallenge)
		challenges.PATCH("/:id", s.UpdateChallenge)
		challenges.DELETE("/:id", s.DeleteChallenge)
	}
}

func (s *Service) CreateChallenge(c *gin.Context) {
	var req CreateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate required fields
	if req.Title == "" || req.Description == "" || req.RepositoryURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title, description, and repository_url are required"})
		return
	}

	if len(req.TargetPaths) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one target path is required"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Validate acceptance criteria weights sum to 1
	weightSum := 0.0
	for _, ac := range req.AcceptanceCriteria {
		weightSum += ac.Weight
	}
	if weightSum > 0 && (weightSum < 0.99 || weightSum > 1.01) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Acceptance criteria weights must sum to 1.0"})
		return
	}

	// Validate scoring weights sum to 1
	scoreSum := 0.0
	for _, v := range req.ScoringWeights {
		scoreSum += v
	}
	if scoreSum > 0 && (scoreSum < 0.99 || scoreSum > 1.01) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Scoring weights must sum to 1.0"})
		return
	}

	challenge := &database.Challenge{
		Title:                req.Title,
		Description:          req.Description,
		RepositoryURL:        req.RepositoryURL,
		RepositoryBranch:     req.RepositoryBranch,
		TargetPaths:          req.TargetPaths,
		ExcludePaths:         req.ExcludePaths,
		Severity:             req.Severity,
		Category:             req.Category,
		TimeLimitSeconds:     req.TimeLimitSeconds,
		MaxTurns:             req.MaxTurns,
		AcceptanceCriteria:   req.AcceptanceCriteria,
		ScoringWeights:       req.ScoringWeights,
		KnownIssues:          req.KnownIssues,
		Tags:                 req.Tags,
		TemplateID:           req.TemplateID,
		IsPublic:             req.IsPublic,
		CreatedBy:            userID.(uuid.UUID),
	}

	// Set defaults
	if challenge.TimeLimitSeconds == 0 {
		challenge.TimeLimitSeconds = 300
	}
	if challenge.MaxTurns == 0 {
		challenge.MaxTurns = 10
	}
	if challenge.RepositoryBranch == "" {
		challenge.RepositoryBranch = "main"
	}

	// Set default scoring weights if not provided
	if len(challenge.ScoringWeights) == 0 {
		challenge.ScoringWeights = map[string]float64{
			"correctness":    0.4,
			"security":       0.3,
			"performance":    0.15,
			"style":          0.1,
			"test_coverage":  0.05,
		}
	}

	ctx := c.Request.Context()
	if err := s.challengeRepo.Create(ctx, challenge); err != nil {
		s.logger.Error().Err(err).Msg("Failed to create challenge")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create challenge"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": challenge})
}

func (s *Service) ListChallenges(c *gin.Context) {
	category := c.Query("category")
	severity := c.Query("severity")
	search := c.Query("search")
	language := c.Query("language")
	templateID := c.Query("template_id")

	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	filters := map[string]interface{}{
		"category":  category,
		"severity":  severity,
		"search":    search,
		"language":  language,
		"template_id": templateID,
	}

	ctx := c.Request.Context()
	challenges, err := s.challengeRepo.List(ctx, filters, limit, offset)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to list challenges")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list challenges"})
		return
	}

	// Get total count
	count, err := s.challengeRepo.Count(ctx, filters)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to count challenges")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count challenges"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    challenges,
		"meta": gin.H{
			"page":         offset/limit + 1,
			"per_page":     limit,
			"total":        count,
			"total_pages":  (count + limit - 1) / limit,
		},
	})
}

func (s *Service) GetChallenge(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid challenge ID"})
		return
	}

	ctx := c.Request.Context()
	challenge, err := s.challengeRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Challenge not found"})
			return
		}
		s.logger.Error().Err(err).Msg("Failed to get challenge")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get challenge"})
		return
	}

	// Increment play count
	go func() {
		ctx := context.Background()
		s.challengeRepo.IncrementPlayCount(context.Background(), id)
	}()

	c.JSON(http.StatusOK, gin.H{"success": true, "data": challenge})
}

func (s *Service) UpdateChallenge(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid challenge ID"})
		return
	}

	var req UpdateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	ctx := c.Request.Context()
	existing, err := s.challengeRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Challenge not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get challenge"})
		return
	}

	// Check ownership
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	if existing.CreatedBy != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this challenge"})
		return
	}

	// Update fields
	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.RepositoryURL != nil {
		existing.RepositoryURL = *req.RepositoryURL
	}
	if req.RepositoryBranch != nil {
		existing.RepositoryBranch = *req.RepositoryBranch
	}
	if req.TargetPaths != nil {
		existing.TargetPaths = *req.TargetPaths
	}
	if req.ExcludePaths != nil {
		existing.ExcludePaths = *req.ExcludePaths
	}
	if req.Severity != nil {
		existing.Severity = *req.Severity
	}
	if req.Category != nil {
		existing.Category = *req.Category
	}
	if req.TimeLimitSeconds != nil {
		existing.TimeLimitSeconds = *req.TimeLimitSeconds
	}
	if req.MaxTurns != nil {
		existing.MaxTurns = *req.MaxTurns
	}
	if req.AcceptanceCriteria != nil {
		existing.AcceptanceCriteria = *req.AcceptanceCriteria
	}
	if req.ScoringWeights != nil {
		existing.ScoringWeights = *req.ScoringWeights
	}
	if req.KnownIssues != nil {
		existing.KnownIssues = *req.KnownIssues
	}
	if req.Tags != nil {
		existing.Tags = *req.Tags
	}
	if req.TemplateID != nil {
		existing.TemplateID = req.TemplateID
	}
	if req.IsPublic != nil {
		existing.IsPublic = *req.IsPublic
	}

	if err := s.challengeRepo.Update(ctx, existing); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update challenge")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update challenge"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": existing})
}

func (s *Service) DeleteChallenge(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid challenge ID"})
		return
	}

	ctx := c.Request.Context()
	existing, err := s.challengeRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Challenge not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get challenge"})
		return
	}

	// Check ownership
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	if existing.CreatedBy != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this challenge"})
		return
	}

	if err := s.challengeRepo.Delete(ctx, id); err != nil {
		s.logger.Error().Err(err).Msg("Failed to delete challenge")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete challenge"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Challenge deleted"})
}

func (s *Service) ListTemplates(c *gin.Context) {
	category := c.Query("category")

	ctx := c.Request.Context()
	templates, err := s.templateRepo.List(ctx, category)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to list templates")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": templates})
}

// Request/Response types

type CreateChallengeRequest struct {
	Title                string                 `json:"title" binding:"required"`
	Description          string                 `json:"description" binding:"required"`
	RepositoryURL        string                 `json:"repository_url" binding:"required,url"`
	RepositoryBranch     string                 `json:"repository_branch"`
	TargetPaths          []string               `json:"target_paths" binding:"required,min=1"`
	ExcludePaths         []string               `json:"exclude_paths"`
	Severity             string                 `json:"severity" binding:"required,oneof=low medium high critical"`
	Category             string                 `json:"category" binding:"required,oneof=security performance refactor testing documentation migration bugfix feature"`
	TimeLimitSeconds     int                    `json:"time_limit_seconds"`
	MaxTurns             int                    `json:"max_turns"`
	AcceptanceCriteria   []AcceptanceCriterion  `json:"acceptance_criteria"`
	ScoringWeights       map[string]float64     `json:"scoring_weights"`
	KnownIssues          []KnownIssue           `json:"known_issues"`
	Tags                 []string               `json:"tags"`
	TemplateID           *uuid.UUID             `json:"template_id"`
	IsPublic             bool                   `json:"is_public"`
}

type UpdateChallengeRequest struct {
	Title                *string                `json:"title"`
	Description          *string                `json:"description"`
	RepositoryURL        *string                `json:"repository_url"`
	RepositoryBranch     *string                `json:"repository_branch"`
	TargetPaths          *[]string              `json:"target_paths"`
	ExcludePaths         *[]string              `json:"exclude_paths"`
	Severity             *string                `json:"severity"`
	Category             *string                `json:"category"`
	TimeLimitSeconds     *int                   `json:"time_limit_seconds"`
	MaxTurns             *int                   `json:"max_turns"`
	AcceptanceCriteria   *[]AcceptanceCriterion `json:"acceptance_criteria"`
	ScoringWeights       *map[string]float64    `json:"scoring_weights"`
	KnownIssues          *[]KnownIssue          `json:"known_issues"`
	Tags                 *[]string              `json:"tags"`
	TemplateID           *uuid.UUID             `json:"template_id"`
	IsPublic             *bool                  `json:"is_public"`
}

type AcceptanceCriterion struct {
	ID              string                 `json:"id"`
	Description     string                 `json:"description" binding:"required"`
	Type            string                 `json:"type" binding:"required,oneof=test static_analysis metric manual"`
	Weight          float64                `json:"weight" binding:"required,min=0,max=1"`
	PassThreshold   float64                `json:"pass_threshold" binding:"required,min=0,max=1"`
	Config          map[string]interface{} `json:"config"`
}

type KnownIssue struct {
	File        string  `json:"file" binding:"required"`
	Line        int     `json:"line" binding:"required,min=0"`
	Column      *int    `json:"column"`
	Issue       string  `json:"issue" binding:"required"`
	Severity    string  `json:"severity" binding:"required,oneof=low medium high critical"`
	Priority    string  `json:"priority" binding:"required,oneof=low medium high critical"`
	CWEID       string  `json:"cwe_id"`
	Description string  `json:"description"`
}