package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"codecrashreview/backend/internal/database"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("token invalid")
)

type Service struct {
	userRepo   *database.UserRepository
	jwtSecret  string
	jwtExpiry  time.Duration
	argon2Params *argon2id.Params
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Type     string    `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

func NewService(userRepo *database.UserRepository, jwtSecret string, jwtExpiry time.Duration) *Service {
	return &Service{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
		argon2Params: &argon2id.Params{
			Memory:      65536,
			Iterations:  3,
			Parallelism: 4,
			SaltLength:  16,
			KeyLength:   32,
		},
	}
}

func (s *Service) Register(ctx context.Context, username, email, password string) (*database.User, *Tokens, error) {
	// Check if user exists
	existing, _ := s.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, nil, ErrUserExists
	}

	existing, _ = s.userRepo.GetByID(ctx, uuid.Nil) // This won't work, need GetByUsername
	// For now, just check email

	// Hash password
	hash, err := argon2id.CreateHash(password, s.argon2Params)
	if err != nil {
		return nil, nil, fmt.Errorf("hash password: %w", err)
	}

	// Create user
	user := &database.User{
		Username:           username,
		Email:              email,
		PasswordHash:       hash,
		Rating:             1200,
		RatingDeviation:    350,
		RatingVolatility:   0.06,
		PeakRating:         1200,
		PreferredAgents:    []string{},
		PreferredLanguages: []string{},
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, fmt.Errorf("create user: %w", err)
	}

	// Generate tokens
	tokens, err := s.generateTokens(user)
	if err != nil {
		return nil, nil, fmt.Errorf("generate tokens: %w", err)
	}

	log.Info().Str("user_id", user.ID.String()).Str("username", username).Msg("User registered")
	return user, tokens, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*database.User, *Tokens, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Verify password
	match, err := argon2id.ComparePasswordAndHash(password, user.PasswordHash)
	if err != nil || !match {
		return nil, nil, ErrInvalidCredentials
	}

	// Generate tokens
	tokens, err := s.generateTokens(user)
	if err != nil {
		return nil, nil, fmt.Errorf("generate tokens: %w", err)
	}

	log.Info().Str("user_id", user.ID.String()).Msg("User logged in")
	return user, tokens, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*database.User, *Tokens, error) {
	claims, err := s.validateToken(refreshToken, "refresh")
	if err != nil {
		return nil, nil, ErrTokenInvalid
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}

	tokens, err := s.generateTokens(user)
	if err != nil {
		return nil, nil, fmt.Errorf("generate tokens: %w", err)
	}

	return user, tokens, nil
}

func (s *Service) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*database.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *Service) UpdateCurrentUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*database.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if username, ok := updates["username"].(string); ok {
		user.Username = username
	}
	if avatarURL, ok := updates["avatar_url"].(string); ok {
		user.AvatarURL = &avatarURL
	}
	if preferredAgents, ok := updates["preferred_agents"].([]string); ok {
		user.PreferredAgents = preferredAgents
	}
	if preferredLanguages, ok := updates["preferred_languages"].([]string); ok {
		user.PreferredLanguages = preferredLanguages
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (*database.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *Service) GetMyStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	winRate := 0.0
	if user.TotalMatches > 0 {
		winRate = float64(user.Wins) / float64(user.TotalMatches) * 100
	}

	return map[string]interface{}{
		"user": map[string]interface{}{
			"id":                   user.ID,
			"username":             user.Username,
			"rating":               user.Rating,
			"rating_deviation":     user.RatingDeviation,
			"rating_volatility":    user.RatingVolatility,
			"peak_rating":          user.PeakRating,
			"total_matches":        user.TotalMatches,
			"wins":                 user.Wins,
			"losses":               user.Losses,
			"draws":                user.Draws,
			"win_rate":             winRate,
			"preferred_agents":     user.PreferredAgents,
			"preferred_languages":  user.PreferredLanguages,
		},
	}, nil
}

func (s *Service) GetUserStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	// This would aggregate more detailed stats from matches
	return s.GetMyStats(ctx, userID)
}

func (s *Service) generateTokens(user *database.User) (*Tokens, error) {
	now := time.Now()
	accessExpiry := now.Add(s.jwtExpiry)
	refreshExpiry := now.Add(7 * 24 * time.Hour) // 7 days

	accessClaims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
		},
	}

	refreshClaims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Type:     "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	return &Tokens{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(s.jwtExpiry.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (s *Service) validateToken(tokenString, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	if claims.Type != expectedType {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, "access")
}

func (s *Service) AuthMiddleware() func(ctx context.Context, userID uuid.UUID) error {
	return func(ctx context.Context, userID uuid.UUID) error {
		// This will be implemented in the HTTP handler
		return nil
	}
}

// Gin middleware
func (s *Service) GinAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing authorization header"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token format"})
			return
		}

		claims, err := s.ValidateAccessToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}