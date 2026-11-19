package leaderboard

import (
	"codecrashreview/backend/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"strconv"
)

type Service struct{ db *pgxpool.Pool }

func NewService(db *pgxpool.Pool) *Service { return &Service{db: db} }
func (s *Service) RegisterRoutes(router *gin.RouterGroup) {
	lb := router.Group("/leaderboard")
	lb.GET("/global", s.list)
	lb.GET("/weekly", s.list)
	lb.GET("/language/:language", s.list)
	lb.GET("/agent/:agent", s.list)
}
func (s *Service) list(c *gin.Context) {
	limit, e := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if e != nil || err != nil || limit < 1 || limit > 100 || offset < 0 {
		c.JSON(400, gin.H{"error": "Invalid pagination"})
		return
	}
	filter := " FROM users u WHERE total_matches>0"
	args := []any{}
	if c.FullPath() == "/api/leaderboard/weekly" {
		filter += " AND EXISTS (SELECT 1 FROM match_participants p JOIN matches m ON m.id=p.match_id WHERE p.user_id=u.id AND m.ended_at>=now()-interval '7 days')"
	}
	if v := c.Param("agent"); v != "" {
		filter += " AND EXISTS (SELECT 1 FROM match_participants p WHERE p.user_id=u.id AND p.agent_persona=$1)"
		args = append(args, v)
	}
	if v := c.Param("language"); v != "" {
		filter += " AND u.preferred_languages ? $1"
		args = append(args, v)
	}
	ctx := c.Request.Context()
	var total int
	err = s.db.QueryRow(ctx, "SELECT count(*)"+filter, args...).Scan(&total)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to count leaderboard"})
		return
	}
	q := "SELECT id,username,avatar_url,rating,rating_deviation,peak_rating,total_matches,wins,losses,100.0*wins/total_matches,preferred_agents->>0,preferred_languages->>0,updated_at" + filter + " ORDER BY rating DESC,rating_deviation,id LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	rows, err := s.db.Query(ctx, q, append(args, limit, offset)...)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to load leaderboard"})
		return
	}
	defer rows.Close()
	entries := []database.LeaderboardEntry{}
	for rows.Next() {
		var e database.LeaderboardEntry
		err = rows.Scan(&e.UserID, &e.Username, &e.AvatarURL, &e.Rating, &e.RatingDeviation, &e.PeakRating, &e.TotalMatches, &e.Wins, &e.Losses, &e.WinRate, &e.PreferredAgent, &e.PreferredLanguage, &e.LastActive)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to read leaderboard"})
			return
		}
		e.Rank = offset + len(entries) + 1
		entries = append(entries, e)
	}
	if rows.Err() != nil {
		c.JSON(500, gin.H{"error": "Failed to read leaderboard"})
		return
	}
	c.JSON(200, gin.H{"success": true, "data": entries, "meta": gin.H{"page": offset/limit + 1, "per_page": limit, "total": total, "total_pages": (total + limit - 1) / limit}})
}
