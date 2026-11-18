package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"codecrashreview/backend/internal/auth"
	"codecrashreview/backend/internal/challenges"
	"codecrashreview/backend/internal/config"
	"codecrashreview/backend/internal/database"
	"codecrashreview/backend/internal/leaderboard"
	"codecrashreview/backend/internal/matchmaking"
	"codecrashreview/backend/internal/snippets"
	"codecrashreview/backend/internal/websocket"
	"codecrashreview/backend/pkg/ratelimit"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &logger

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		logger.Fatal().Err(err).Msg("Failed to run migrations")
	}

	// Initialize Redis
	redisClient := database.NewRedisClient(cfg.RedisURL)
	defer redisClient.Close()

	// Initialize repositories
	userRepo := database.NewUserRepository(db)
	matchRepo := database.NewMatchRepository(db)
	participantRepo := database.NewMatchParticipantRepository(db)
	challengeRepo := database.NewChallengeRepository(db)
	templateRepo := database.NewChallengeTemplateRepository(db)
	snippetRepo := database.NewSnippetRepository(db)
	collectionRepo := database.NewSnippetCollectionRepository(db)

	// Initialize services
	authService := auth.NewService(userRepo, cfg.JWTSecret, cfg.JWTExpiry)
	matchmakingService := matchmaking.NewService(matchRepo, participantRepo, userRepo, challengeRepo, redisClient, cfg.EloKFactor, cfg.GlickoTau)
	challengeService := challenges.NewService(challengeRepo, templateRepo, db)
	snippetService := snippets.NewService(snippetRepo, nil, db)
	leaderboardService := leaderboard.NewService(matchRepo, nil, redisClient)
	wsHub := websocket.NewHub(redisClient, logger)

	// Set dependencies
	wsHub.SetAuthService(authService)
	wsHub.SetMatchRepository(matchRepo)

	// Start WebSocket hub
	go wsHub.Run()

	// Start matchmaking worker
	go matchmakingService.StartWorker(context.Background())

	// Start leaderboard cache updater
	go leaderboardService.StartCacheUpdater(context.Background())

	// Setup Gin router
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(ratelimit.RateLimiter(redisClient, cfg.RateLimitRequests, cfg.RateLimitWindow))
	router.Use(corsMiddleware())
	router.Use(requestLogger(&logger))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
	})

	// API routes
	api := router.Group("/api")
	{
		// Auth routes (public)
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", authService.Register)
			authRoutes.POST("/login", authService.Login)
			authRoutes.POST("/refresh", authService.RefreshToken)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(authService.AuthMiddleware())
		{
			// User routes
			protected.GET("/users/me", authService.GetCurrentUser)
			protected.PATCH("/users/me", authService.UpdateCurrentUser)
			protected.GET("/users/:id", authService.GetUser)
			protected.GET("/stats/me", authService.GetMyStats)
			protected.GET("/users/:id/stats", authService.GetUserStats)

			// Match routes
			protected.POST("/matches", matchmakingService.CreateMatch)
			protected.GET("/matches", matchmakingService.ListMatches)
			protected.GET("/matches/:id", matchmakingService.GetMatch)
			protected.POST("/matches/:id/join", matchmakingService.JoinMatch)
			protected.POST("/matches/:id/leave", matchmakingService.LeaveMatch)
			protected.POST("/matches/queue", matchmakingService.JoinQueue)
			protected.DELETE("/matches/queue", matchmakingService.LeaveQueue)
			protected.GET("/matches/queue/status", matchmakingService.GetQueueStatus)

			// Challenge routes
			challengeService.RegisterRoutes(protected)

			// Snippet routes
			snippetService.RegisterRoutes(protected)

			// Leaderboard routes
			leaderboardService.RegisterRoutes(protected)
		}
	}

	// WebSocket endpoint
	router.GET("/ws", wsHub.HandleWebSocket)

	// Start server
	srv := &http.Server{
		Addr:    cfg.ServerAddress(),
		Handler: router,
	}

	go func() {
		logger.Info().Str("addr", cfg.ServerAddress()).Msg("Starting HTTP server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Server exited")
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func requestLogger(logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info().
			Str("method", method).
			Str("path", path).
			Str("ip", clientIP).
			Int("status", statusCode).
			Dur("latency", latency).
			Msg("HTTP request")
	}
}