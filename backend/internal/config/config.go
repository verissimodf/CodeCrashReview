package config

import (
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	// Database
	DatabaseURL string
	RedisURL    string

	// Auth
	JWTSecret     string
	JWTExpiry     time.Duration
	Argon2Time    uint32
	Argon2Memory  uint32
	Argon2Threads uint8
	Argon2KeyLen  uint32

	// FreeLLM-Local
	FreeLLMBaseURL string
	FreeLLMAPIKey  string
	DefaultModel   string

	// Docker/Sandbox
	DockerHost            string
	SandboxNetwork        string
	SandboxCPULimit       float64
	SandboxMemoryLimit    string
	SandboxPIDsLimit      int64
	SandboxSeccompProfile string

	// Matchmaking
	EloKFactor           float64
	GlickoTau            float64
	MatchTimeoutSeconds  int
	QueueCheckIntervalMs int
	MinRating            int
	MaxRating            int
	DefaultRating        int
	MatchmakingRange     int

	// Server
	GinMode           string
	ServerHost        string
	ServerPort        int
	RateLimitRequests int
	RateLimitWindow   int

	// Logging
	LogLevel  string
	LogFormat string

	// Features
	EnableSpectatorMode      bool
	EnableReplaySystem       bool
	EnableChallengeTemplates bool
	EnableSnippetImport      bool
	EnableAgentAnalytics     bool
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/codecrashreview")
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("database_url", "postgres://codecrash:codecrash@localhost:5432/codecrash?sslmode=disable")
	viper.SetDefault("redis_url", "redis://localhost:6379")
	viper.SetDefault("jwt_secret", "dev-secret-change-in-production")
	viper.SetDefault("jwt_expiry", "24h")
	viper.SetDefault("argon2_time", 3)
	viper.SetDefault("argon2_memory", 65536)
	viper.SetDefault("argon2_threads", 4)
	viper.SetDefault("argon2_keylen", 32)
	viper.SetDefault("freellm_base_url", "http://localhost:11434/v1")
	viper.SetDefault("freellm_api_key", "ollama")
	viper.SetDefault("default_model", "codellama:34b")
	viper.SetDefault("docker_host", "unix:///var/run/docker.sock")
	viper.SetDefault("sandbox_network", "codecrash-sandbox")
	viper.SetDefault("sandbox_cpu_limit", 2.0)
	viper.SetDefault("sandbox_memory_limit", "2g")
	viper.SetDefault("sandbox_pids_limit", 100)
	viper.SetDefault("elo_k_factor", 32.0)
	viper.SetDefault("glicko_tau", 0.5)
	viper.SetDefault("match_timeout_seconds", 300)
	viper.SetDefault("queue_check_interval_ms", 5000)
	viper.SetDefault("min_rating", 100)
	viper.SetDefault("max_rating", 3000)
	viper.SetDefault("default_rating", 1200)
	viper.SetDefault("matchmaking_range", 200)
	viper.SetDefault("gin_mode", "debug")
	viper.SetDefault("server_host", "0.0.0.0")
	viper.SetDefault("server_port", 8080)
	viper.SetDefault("rate_limit_requests", 100)
	viper.SetDefault("rate_limit_window", 60)
	viper.SetDefault("log_level", "debug")
	viper.SetDefault("log_format", "json")
	viper.SetDefault("enable_spectator_mode", true)
	viper.SetDefault("enable_replay_system", true)
	viper.SetDefault("enable_challenge_templates", true)
	viper.SetDefault("enable_snippet_import", true)
	viper.SetDefault("enable_agent_analytics", true)

	_ = viper.ReadInConfig()

	cfg := &Config{
		DatabaseURL:              getEnv("DATABASE_URL", viper.GetString("database_url")),
		RedisURL:                 getEnv("REDIS_URL", viper.GetString("redis_url")),
		JWTSecret:                getEnv("JWT_SECRET", viper.GetString("jwt_secret")),
		JWTExpiry:                parseDuration(getEnv("JWT_EXPIRY", viper.GetString("jwt_expiry"))),
		Argon2Time:               uint32(getEnvInt("ARGON2_TIME", viper.GetInt("argon2_time"))),
		Argon2Memory:             uint32(getEnvInt("ARGON2_MEMORY", viper.GetInt("argon2_memory"))),
		Argon2Threads:            uint8(getEnvInt("ARGON2_THREADS", viper.GetInt("argon2_threads"))),
		Argon2KeyLen:             uint32(getEnvInt("ARGON2_KEYLEN", viper.GetInt("argon2_keylen"))),
		FreeLLMBaseURL:           getEnv("FREELLM_BASE_URL", viper.GetString("freellm_base_url")),
		FreeLLMAPIKey:            getEnv("FREELLM_API_KEY", viper.GetString("freellm_api_key")),
		DefaultModel:             getEnv("DEFAULT_MODEL", viper.GetString("default_model")),
		DockerHost:               getEnv("DOCKER_HOST", viper.GetString("docker_host")),
		SandboxNetwork:           getEnv("SANDBOX_NETWORK", viper.GetString("sandbox_network")),
		SandboxCPULimit:          getEnvFloat("SANDBOX_CPU_LIMIT", viper.GetFloat64("sandbox_cpu_limit")),
		SandboxMemoryLimit:       getEnv("SANDBOX_MEMORY_LIMIT", viper.GetString("sandbox_memory_limit")),
		SandboxPIDsLimit:         int64(getEnvInt("SANDBOX_PIDS_LIMIT", viper.GetInt("sandbox_pids_limit"))),
		SandboxSeccompProfile:    getEnv("SANDBOX_SECCOMP_PROFILE", viper.GetString("sandbox_seccomp_profile")),
		EloKFactor:               getEnvFloat("ELO_K_FACTOR", viper.GetFloat64("elo_k_factor")),
		GlickoTau:                getEnvFloat("GLICKO_TAU", viper.GetFloat64("glicko_tau")),
		MatchTimeoutSeconds:      getEnvInt("MATCH_TIMEOUT_SECONDS", viper.GetInt("match_timeout_seconds")),
		QueueCheckIntervalMs:     getEnvInt("QUEUE_CHECK_INTERVAL_MS", viper.GetInt("queue_check_interval_ms")),
		MinRating:                getEnvInt("MIN_RATING", viper.GetInt("min_rating")),
		MaxRating:                getEnvInt("MAX_RATING", viper.GetInt("max_rating")),
		DefaultRating:            getEnvInt("DEFAULT_RATING", viper.GetInt("default_rating")),
		MatchmakingRange:         getEnvInt("MATCHMAKING_RANGE", viper.GetInt("matchmaking_range")),
		GinMode:                  getEnv("GIN_MODE", viper.GetString("gin_mode")),
		ServerHost:               getEnv("SERVER_HOST", viper.GetString("server_host")),
		ServerPort:               getEnvInt("SERVER_PORT", viper.GetInt("server_port")),
		RateLimitRequests:        getEnvInt("RATE_LIMIT_REQUESTS", viper.GetInt("rate_limit_requests")),
		RateLimitWindow:          getEnvInt("RATE_LIMIT_WINDOW", viper.GetInt("rate_limit_window")),
		LogLevel:                 getEnv("LOG_LEVEL", viper.GetString("log_level")),
		LogFormat:                getEnv("LOG_FORMAT", viper.GetString("log_format")),
		EnableSpectatorMode:      getEnvBool("ENABLE_SPECTATOR_MODE", viper.GetBool("enable_spectator_mode")),
		EnableReplaySystem:       getEnvBool("ENABLE_REPLAY_SYSTEM", viper.GetBool("enable_replay_system")),
		EnableChallengeTemplates: getEnvBool("ENABLE_CHALLENGE_TEMPLATES", viper.GetBool("enable_challenge_templates")),
		EnableSnippetImport:      getEnvBool("ENABLE_SNIPPET_IMPORT", viper.GetBool("enable_snippet_import")),
		EnableAgentAnalytics:     getEnvBool("ENABLE_AGENT_ANALYTICS", viper.GetBool("enable_agent_analytics")),
	}

	return cfg
}

func (c *Config) ServerAddress() string {
	return c.ServerHost + ":" + strconv.Itoa(c.ServerPort)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}
