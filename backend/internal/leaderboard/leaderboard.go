package leaderboard

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"codecrashreview/backend/internal/database"
)

type Service struct {
	matchRepo       *database.MatchRepository
	userRepo        *database.UserRepository
	redis           *redis.Client
	logger          zerolog.Logger
	leaderboardTTL  time.Duration
}

func NewService(matchRepo *database.MatchRepository, userRepo *database.UserRepository, redis *redis.Client) *Service {
	return &Service{
		matchRepo:      matchRepo,
		userRepo:       userRepo,
		redis:          redis,
		logger:         zerolog.DefaultContextLogger,
		leaderboardTTL: 5 * time.Minute,
	}
}

func (s *Service) RegisterRoutes(router *gin.RouterGroup) {
	lb := router.Group("/leaderboard")
	{
		lb.GET("/global", s.GetGlobalLeaderboard)
		lb.GET("/weekly", s.GetWeeklyLeaderboard)
		lb.GET("/language/:language", s.GetLeaderboardByLanguage)
		lb.GET("/agent/:agent", s.GetLeaderboardByAgent)
	}
}

func (s *Service) GetGlobalLeaderboard(c *gin.Context) {
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			l = strconv.Itoa(v)
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			o = strconv.Itoa(v)
		}
	}

	// Try to get from cache
	cacheKey := fmt.Sprintf("leaderboard:global:%s:%s", c.Query("limit"), c.Query("offset"))
	cached, err := s.redis.Get(c.Request.Context(), cacheKey).Result()
	if err == nil {
		c.Data(http.StatusOK, "application/json", []byte(cached))
		return
	}

	ctx := c.Request.Context()
	entries, err := s.getGlobalLeaderboard(ctx, limit, offset)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get global leaderboard")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get leaderboard"})
		return
	}

	response := gin.H{
		"success": true,
		"data":    entries,
		"meta": gin.H{
			"page":         1,
			"per_page":     len(entries),
			"total":        0, // Would need separate count query
			"total_pages":  1,
		},
	}

	// Cache the response
	data, _ := json.Marshal(response)
	s.redis.Set(c.Request.Context(), cacheKey, data, s.leaderboardTTL)

	c.JSON(http.StatusOK, response)
}

func (s *Service) GetWeeklyLeaderboard(c *gin.Context) {
	limit := 50
	offset := 0
	// ... similar to global but filtered by week
	c.JSON(http.StatusOK, gin.H{"success": true, "data": []interface{}{}, "meta": gin.H{}})
}

func (s *Service) GetLeaderboardByLanguage(c *gin.Context) {
	language := c.Param("language")
	c.JSON(http.StatusOK, gin.H{"success": true, "data": []interface{}{}, "meta": gin.H{}})
}

func (s *Service) GetLeaderboardByAgent(c *gin.Context) {
	agent := c.Param("agent")
	c.JSON(http.StatusOK, gin.H{"success": true, "data": []interface{}{}, "meta": gin.H{}})
}

func (s *Service) getGlobalLeaderboard(ctx context.Context, limit, offset int) ([]database.LeaderboardEntry, error) {
	query := `
		SELECT u.id as user_id, u.username, u.avatar_url, u.rating, u.rating_deviation, u.peak_rating,
		       u.total_matches, u.wins, u.losses,
		       CASE WHEN u.total_matches > 0 THEN (u.wins::float / u.total_matches) * 100 ELSE 0 END as win_rate,
		       u.preferred_agents, u.preferred_languages, u.updated_at as last_active
		FROM users u
		WHERE u.total_matches > 0
		ORDER BY u.rating DESC, u.rating_deviation ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.matchRepo.GetPool().Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []database.LeaderboardEntry
	rank := 1
	for rows.Next() {
		var entry database.LeaderboardEntry
		var preferredAgents, preferredLanguages []byte

		err := rows.Scan(
			&entry.UserID, &entry.Username, &entry.AvatarURL,
			&entry.Rating, &entry.RatingDeviation, &entry.PeakRating,
			&entry.TotalMatches, &entry.Wins, &entry.Losses, &entry.WinRate,
			&preferredAgents, &preferredLanguages, &entry.LastActive,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON arrays
		// ... parsing logic

		entry.Rank = rank
		rank++
		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *Service) UpdateLeaderboardCache(ctx context.Context) error {
	// Update global leaderboard cache
	entries, err := s.getGlobalLeaderboard(ctx, 100, 0)
	if err != nil {
		return err
	}

	data, _ := json.Marshal(gin.H{"success": true, "data": entries})
	return s.redis.Set(ctx, "leaderboard:global:100:0", data, s.leaderboardTTL).Err()
}

// Periodic cache updater
func (s *Service) StartCacheUpdater(ctx context.Context) {
	ticker := time.NewTicker(s.leaderboardTTL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.UpdateLeaderboardCache(ctx)
		}
	}
}