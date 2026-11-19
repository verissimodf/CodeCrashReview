package auth

import (
	"codecrashreview/backend/internal/database"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"
)

func (s *Service) RegisterPublic(api *gin.RouterGroup) {
	api.POST("/register", func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required,min=2,max=50"`
			Email    string `json:"email" binding:"required,email,max=255"`
			Password string `json:"password" binding:"required,min=8,max=256"`
		}
		if c.ShouldBindJSON(&req) != nil {
			c.JSON(400, gin.H{"error": "Enter a username, valid email and password of at least eight characters"})
			return
		}
		u, t, err := s.Register(c.Request.Context(), req.Username, strings.ToLower(req.Email), req.Password)
		authResponse(c, 201, u, t, err)
	})
	api.POST("/login", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,max=256"`
		}
		if c.ShouldBindJSON(&req) != nil {
			c.JSON(400, gin.H{"error": "Enter an email and password"})
			return
		}
		u, t, err := s.Login(c.Request.Context(), strings.ToLower(req.Email), req.Password)
		authResponse(c, 200, u, t, err)
	})
	api.POST("/refresh", func(c *gin.Context) {
		var req struct {
			Token string `json:"refresh_token" binding:"required"`
		}
		if c.ShouldBindJSON(&req) != nil {
			c.JSON(400, gin.H{"error": "Refresh token required"})
			return
		}
		u, t, err := s.RefreshToken(c.Request.Context(), req.Token)
		authResponse(c, 200, u, t, err)
	})
}
func authResponse(c *gin.Context, status int, u *database.User, t *Tokens, err error) {
	if err != nil {
		respondError(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(status, gin.H{"user": u, "tokens": t})
}
func respondError(c *gin.Context, err error) {
	var dbErr *pgconn.PgError
	switch {
	case errors.Is(err, ErrUserExists), errors.As(err, &dbErr) && dbErr.Code == "23505":
		c.JSON(409, gin.H{"error": "Username or email already in use"})
	case errors.Is(err, ErrInvalidCredentials), errors.Is(err, ErrTokenInvalid):
		c.JSON(401, gin.H{"error": "Invalid credentials"})
	case errors.Is(err, pgx.ErrNoRows), errors.Is(err, ErrUserNotFound):
		c.JSON(404, gin.H{"error": "User not found"})
	default:
		c.JSON(500, gin.H{"error": "Account operation failed"})
	}
}
func (s *Service) RegisterProtected(api *gin.RouterGroup) {
	profile := func(c *gin.Context) {
		id := c.MustGet("user_id").(uuid.UUID)
		if raw := c.Param("id"); raw != "" {
			var err error
			id, err = uuid.Parse(raw)
			if err != nil {
				c.JSON(400, gin.H{"error": "Invalid user ID"})
				return
			}
		}
		if strings.HasSuffix(c.FullPath(), "/stats") || c.FullPath() == "/api/stats/me" {
			stats, err := s.GetMyStats(c.Request.Context(), id)
			if err != nil {
				respondError(c, err)
				return
			}
			c.JSON(200, stats)
			return
		}
		u, err := s.GetUser(c.Request.Context(), id)
		if err != nil {
			respondError(c, err)
			return
		}
		if id != c.MustGet("user_id").(uuid.UUID) {
			u.Email = ""
		}
		c.JSON(200, u)
	}
	api.GET("/users/me", profile)
	api.GET("/users/:id", profile)
	api.GET("/stats/me", profile)
	api.GET("/users/:id/stats", profile)
	api.PATCH("/users/me", func(c *gin.Context) {
		var req struct {
			Username           *string  `json:"username" binding:"omitempty,min=2,max=50"`
			AvatarURL          *string  `json:"avatar_url"`
			PreferredAgents    []string `json:"preferred_agents"`
			PreferredLanguages []string `json:"preferred_languages"`
		}
		if c.ShouldBindJSON(&req) != nil {
			c.JSON(400, gin.H{"error": "Invalid profile fields"})
			return
		}
		updates := map[string]any{}
		if req.Username != nil {
			updates["username"] = *req.Username
		}
		if req.AvatarURL != nil {
			updates["avatar_url"] = *req.AvatarURL
		}
		if req.PreferredAgents != nil {
			updates["preferred_agents"] = req.PreferredAgents
		}
		if req.PreferredLanguages != nil {
			updates["preferred_languages"] = req.PreferredLanguages
		}
		u, err := s.UpdateCurrentUser(c.Request.Context(), c.MustGet("user_id").(uuid.UUID), updates)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(200, u)
	})
}
