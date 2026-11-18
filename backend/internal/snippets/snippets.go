package snippets

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
	snippetRepo   *database.SnippetRepository
	collectionRepo *database.SnippetCollectionRepository
	pool          *pgxpool.Pool
	logger        zerolog.Logger
}

func NewService(snippetRepo *database.SnippetRepository, collectionRepo *database.SnippetCollectionRepository, pool *pgxpool.Pool) *Service {
	return &Service{
		snippetRepo:    snippetRepo,
		collectionRepo: collectionRepo,
		pool:           pool,
		logger:         zerolog.DefaultContextLogger,
	}
}

func (s *Service) RegisterRoutes(router *gin.RouterGroup) {
	snippets := router.Group("/snippets")
	{
		snippets.GET("", s.ListSnippets)
		snippets.GET("/:id", s.GetSnippet)
		snippets.POST("/import", s.ImportSnippets)
		snippets.POST("", s.CreateSnippet)
		snippets.PATCH("/:id", s.UpdateSnippet)
		snippets.DELETE("/:id", s.DeleteSnippet)
	}
}

func (s *Service) ListSnippets(c *gin.Context) {
	language := c.Query("language")
	tag := c.Query("tag")
	complexity := c.Query("complexity")
	search := c.Query("search")

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
		"language":   language,
		"tag":        tag,
		"complexity": complexity,
		"search":     search,
	}

	ctx := c.Request.Context()
	snippets, err := s.snippetRepo.List(ctx, filters, limit, offset)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to list snippets")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list snippets"})
		return
	}

	count, err := s.snippetRepo.Count(ctx, filters)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to count snippets")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count snippets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    snippets,
		"meta": gin.H{
			"page":         0, // simplified
			"per_page":     limit,
			"total":        count,
			"total_pages":  (count + limit - 1) / limit,
		},
	})
}

func (s *Service) GetSnippet(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid snippet ID"})
		return
	}

	ctx := c.Request.Context()
	snippet, err := s.snippetRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Snippet not found"})
			return
		}
		s.logger.Error().Err(err).Msg("Failed to get snippet")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get snippet"})
		return
	}

	// Increment usage count
	go func() {
		s.snippetRepo.IncrementUsage(context.Background(), id)
	}()

	c.JSON(http.StatusOK, gin.H{"success": true, "data": snippet})
}

func (s *Service) CreateSnippet(c *gin.Context) {
	var req CreateSnippetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	snippet := &database.Snippet{
		Title:       req.Title,
		Description: req.Description,
		Language:    req.Language,
		Code:        req.Code,
		Tags:        req.Tags,
		Complexity:  req.Complexity,
		Author:      req.Author,
		AuthorID:    &userID,
		SourceURL:   req.SourceURL,
		IsOfficial:  false,
	}

	if err := s.snippetRepo.Create(context.Background(), snippet); err != nil {
		s.logger.Error().Err(err).Msg("Failed to create snippet")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create snippet"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": snippet})
}

func (s *Service) ImportSnippets(c *gin.Context) {
	var req ImportSnippetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// This would trigger the snippet import job
	// For now, return a placeholder response
	s.logger.Info().Str("source", req.Source).Msg("Snippet import requested")

	c.JSON(http.StatusAccepted, gin.H{
		"success": true,
		"message": "Import job started",
		"job_id":  uuid.New().String(),
	})
}

func (s *Service) UpdateSnippet(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid snippet ID"})
		return
	}

	var req UpdateSnippetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	ctx := c.Request.Context()
	snippet, err := s.snippetRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Snippet not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get snippet"})
		return
	}

	// Check ownership
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	if snippet.AuthorID != nil && *snippet.AuthorID != userID.(uuid.UUID) && !snippet.IsOfficial {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this snippet"})
		return
	}

	// Update fields
	if req.Title != nil {
		snippet.Title = *req.Title
	}
	if req.Description != nil {
		snippet.Description = *req.Description
	}
	if req.Language != nil {
		snippet.Language = *req.Language
	}
	if req.Code != nil {
		snippet.Code = *req.Code
	}
	if req.Tags != nil {
		snippet.Tags = *req.Tags
	}
	if req.Complexity != nil {
		snippet.Complexity = *req.Complexity
	}
	if req.SourceURL != nil {
		snippet.SourceURL = req.SourceURL
	}
	if req.Rating != nil {
		snippet.Rating = *req.Rating
	}
	if req.IsOfficial != nil {
		snippet.IsOfficial = *req.IsOfficial
	}

	if err := s.snippetRepo.Update(c.Request.Context(), snippet); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update snippet")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update snippet"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": snippet})
}

func (s *Service) DeleteSnippet(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid snippet ID"})
		return
	}

	ctx := c.Request.Context()
	snippet, err := s.snippetRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Snippet not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get snippet"})
		return
	}

	// Check ownership
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	if snippet.AuthorID != nil && *snippet.AuthorID != userID.(uuid.UUID) && !snippet.IsOfficial {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this snippet"})
		return
	}

	if err := s.snippetRepo.Delete(ctx, id); err != nil {
		s.logger.Error().Err(err).Msg("Failed to delete snippet")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete snippet"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Snippet deleted"})
}

// Request types

type CreateSnippetRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description" binding:"required"`
	Language    string   `json:"language" binding:"required,oneof=cpp python java javascript typescript rust shell html gherkin"`
	Code        string   `json:"code" binding:"required"`
	Tags        []string `json:"tags"`
	Complexity  string   `json:"complexity" binding:"required,oneof=basic intermediate advanced"`
	Author      string   `json:"author" binding:"required"`
	SourceURL   *string  `json:"source_url"`
}

type UpdateSnippetRequest struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Language    *string  `json:"language"`
	Code        *string  `json:"code"`
	Tags        *[]string `json:"tags"`
	Complexity  *string  `json:"complexity"`
	SourceURL   *string  `json:"source_url"`
	Rating      *float64 `json:"rating"`
	IsOfficial  *bool    `json:"is_official"`
}

type ImportSnippetsRequest struct {
	Source string `json:"source" binding:"required,oneof=awesome-algorithms cp-algorithms programming-idioms leetcode-patterns"`
}