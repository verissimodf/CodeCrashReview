package matchmaking

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Service) RegisterRoutes(g *gin.RouterGroup) {
	g.GET("/matches", func(c *gin.Context) {
		u := c.MustGet("user_id").(uuid.UUID)
		r, e := s.ListMatches(c.Request.Context(), u, 20, 0)
		if e != nil {
			c.JSON(500, gin.H{"error": "Could not load matches"})
			return
		}
		c.JSON(200, gin.H{"matches": r})
	})
}
