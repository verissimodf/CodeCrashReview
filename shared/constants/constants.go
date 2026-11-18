package constants

// CodeCrashReview - Shared Constants
// Constants used across frontend, backend, and sandbox

// Version
const (
	Version     = "0.1.0"
	APIVersion  = "v1"
	BuildDate   = "2024-01-01"
	GitCommit   = "dev"
)

// API
const (
	APIBasePath     = "/api"
	WSPath          = "/ws"
	HealthPath      = "/health"
	MetricsPath     = "/metrics"
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Auth
const (
	AccessTokenCookieName  = "ccr_access_token"
	RefreshTokenCookieName = "ccr_refresh_token"
	TokenHeaderName        = "Authorization"
	TokenPrefix            = "Bearer "
	CSRFHeaderName         = "X-CSRF-Token"
)

// Timeouts (seconds)
const (
	DefaultRequestTimeout     = 30
	WebSocketPingInterval     = 30
	WebSocketPongTimeout      = 10
	SandboxDefaultTimeout     = 30
	SandboxMaxTimeout         = 300
	MatchDefaultTimeout       = 300
	MatchMaxTimeout           = 1800
	QueueCheckInterval        = 5
	AgentActionTimeout        = 60
	CodeExecutionTimeout      = 30
)

// Matchmaking
const (
	DefaultRating           = 1200
	DefaultRatingDeviation  = 350.0
	DefaultRatingVolatility = 0.06
	MinRating               = 100
	MaxRating               = 3000
	KFactor                 = 32.0
	GlickoTau               = 0.5
	MatchmakingRange        = 200
	MaxQueueWaitTime        = 300
	MinPlayersForMatch      = 2
	MaxPlayersPerMatch      = 2
)

// Match
const (
	MaxTurnsDefault      = 10
	MaxTurnsMax          = 50
	TimeLimitDefault     = 300
	TimeLimitMin         = 60
	TimeLimitMax         = 3600
	WarningTimeThreshold = 30
	DangerTimeThreshold  = 10
	OvertimeEnabled      = true
	OvertimeDuration     = 60
)

// Agent
const (
	MaxAgentsPerMatch     = 2
	AgentThinkingMinTime  = 1
	AgentThinkingMaxTime  = 30
	AgentEditingMaxTime   = 60
	AgentTestingMaxTime   = 30
	AgentSubmittingMaxTime = 10
	MaxThoughtsPerTurn    = 10
	MaxChangesPerTurn     = 5
)

// Sandbox
const (
	SandboxCPULimitDefault     = 2.0
	SandboxMemoryLimitDefault  = "2g"
	SandboxPIDsLimitDefault    = 100
	SandboxNetworkDisabled     = true
	SandboxReadOnlyRootfs      = true
	SandboxSeccompProfile      = "default"
	MaxConcurrentSandboxes     = 50
	SandboxPoolSizePerLanguage = 3
	ContainerNamePrefix        = "codecrash-sandbox"
)

// Languages
const (
	LanguageCpp        = "cpp"
	LanguagePython     = "python"
	LanguageJava       = "java"
	LanguageJavaScript = "javascript"
	LanguageTypeScript = "typescript"
	LanguageRust       = "rust"
	LanguageShell      = "shell"
	LanguageHTML       = "html"
	LanguageGherkin    = "gherkin"
)

var SupportedLanguages = []string{
	LanguageCpp,
	LanguagePython,
	LanguageJava,
	LanguageJavaScript,
	LanguageTypeScript,
	LanguageRust,
	LanguageShell,
	LanguageHTML,
	LanguageGherkin,
}

var LanguageExtensions = map[string]string{
	LanguageCpp:        ".cpp",
	LanguagePython:     ".py",
	LanguageJava:       ".java",
	LanguageJavaScript: ".js",
	LanguageTypeScript: ".ts",
	LanguageRust:       ".rs",
	LanguageShell:      ".sh",
	LanguageHTML:       ".html",
	LanguageGherkin:    ".feature",
}

var LanguageMonacoIDs = map[string]string{
	LanguageCpp:        "cpp",
	LanguagePython:     "python",
	LanguageJava:       "java",
	LanguageJavaScript: "javascript",
	LanguageTypeScript: "typescript",
	LanguageRust:       "rust",
	LanguageShell:      "shell",
	LanguageHTML:       "html",
	LanguageGherkin:    "gherkin",
}

var LanguageDockerImages = map[string]string{
	LanguageCpp:        "codecrash-sandbox-cpp:latest",
	LanguagePython:     "codecrash-sandbox-python:latest",
	LanguageJava:       "codecrash-sandbox-java:latest",
	LanguageJavaScript: "codecrash-sandbox-javascript:latest",
	LanguageTypeScript: "codecrash-sandbox-typescript:latest",
	LanguageRust:       "codecrash-sandbox-rust:latest",
	LanguageShell:      "codecrash-sandbox-shell:latest",
	LanguageHTML:       "codecrash-sandbox-html:latest",
	LanguageGherkin:    "codecrash-sandbox-gherkin:latest",
}

// Agent Personas
const (
	AgentSentinel  = "sentinel"
	AgentOptimizer = "optimizer"
	AgentRefactor  = "refactor"
	AgentGuardian  = "guardian"
	AgentDocuBot   = "docubot"
	AgentPolyglot  = "polyglot"
)

var AgentPersonas = []string{
	AgentSentinel,
	AgentOptimizer,
	AgentRefactor,
	AgentGuardian,
	AgentDocuBot,
	AgentPolyglot,
}

var AgentColors = map[string]string{
	AgentSentinel:  "#ff3366",
	AgentOptimizer: "#00ff87",
	AgentRefactor:  "#0099ff",
	AgentGuardian:  "#ffaa00",
	AgentDocuBot:   "#ff00ff",
	AgentPolyglot:  "#00ffff",
}

var AgentGlowColors = map[string]string{
	AgentSentinel:  "#ff336640",
	AgentOptimizer: "#00ff8740",
	AgentRefactor:  "#0099ff40",
	AgentGuardian:  "#ffaa0040",
	AgentDocuBot:   "#ff00ff40",
	AgentPolyglot:  "#00ffff40",
}

var AgentIcons = map[string]string{
	AgentSentinel:  "shield-alert",
	AgentOptimizer: "zap",
	AgentRefactor:  "code-2",
	AgentGuardian:  "check-circle-2",
	AgentDocuBot:   "book-open",
	AgentPolyglot:  "globe",
}

var AgentDefaultModels = map[string]string{
	AgentSentinel:  "deepseek-coder:33b",
	AgentOptimizer: "codellama:34b",
	AgentRefactor:  "qwen2.5-coder:32b",
	AgentGuardian:  "starcoder2:15b",
	AgentDocuBot:   "deepseek-coder:6.7b",
	AgentPolyglot:  "codellama:13b",
}

var AgentDefaultTemperatures = map[string]float64{
	AgentSentinel:  0.2,
	AgentOptimizer: 0.3,
	AgentRefactor:  0.25,
	AgentGuardian:  0.35,
	AgentDocuBot:   0.4,
	AgentPolyglot:  0.3,
}

// Challenge
const (
	ChallengeSeverityLow      = "low"
	ChallengeSeverityMedium   = "medium"
	ChallengeSeverityHigh     = "high"
	ChallengeSeverityCritical = "critical"
)

var ChallengeSeverities = []string{
	ChallengeSeverityLow,
	ChallengeSeverityMedium,
	ChallengeSeverityHigh,
	ChallengeSeverityCritical,
}

const (
	ChallengeCategorySecurity       = "security"
	ChallengeCategoryPerformance    = "performance"
	ChallengeCategoryRefactor       = "refactor"
	ChallengeCategoryTesting        = "testing"
	ChallengeCategoryDocumentation  = "documentation"
	ChallengeCategoryMigration      = "migration"
	ChallengeCategoryBugFix         = "bugfix"
	ChallengeCategoryFeature        = "feature"
)

var ChallengeCategories = []string{
	ChallengeCategorySecurity,
	ChallengeCategoryPerformance,
	ChallengeCategoryRefactor,
	ChallengeCategoryTesting,
	ChallengeCategoryDocumentation,
	ChallengeCategoryMigration,
	ChallengeCategoryBugFix,
	ChallengeCategoryFeature,
}

const (
	ChallengePriorityLow      = "low"
	ChallengePriorityMedium   = "medium"
	ChallengePriorityHigh     = "high"
	ChallengePriorityCritical = "critical"
)

var ChallengePriorities = []string{
	ChallengePriorityLow,
	ChallengePriorityMedium,
	ChallengePriorityHigh,
	ChallengePriorityCritical,
}

// Match Modes
const (
	MatchModeRanked     = "ranked"
	MatchModePrivate    = "private"
	MatchModeTournament = "tournament"
	MatchModePractice   = "practice"
)

var MatchModes = []string{
	MatchModeRanked,
	MatchModePrivate,
	MatchModeTournament,
	MatchModePractice,
}

// Match Status
const (
	MatchStatusPending     = "pending"
	MatchStatusMatching    = "matching"
	MatchStatusPreparing   = "preparing"
	MatchStatusActive      = "active"
	MatchStatusPaused      = "paused"
	MatchStatusCompleted   = "completed"
	MatchStatusCancelled   = "cancelled"
	MatchStatusError       = "error"
)

var MatchStatuses = []string{
	MatchStatusPending,
	MatchStatusMatching,
	MatchStatusPreparing,
	MatchStatusActive,
	MatchStatusPaused,
	MatchStatusCompleted,
	MatchStatusCancelled,
	MatchStatusError,
}

// Agent Status
const (
	AgentStatusIdle       = "idle"
	AgentStatusThinking   = "thinking"
	AgentStatusEditing    = "editing"
	AgentStatusTesting    = "testing"
	AgentStatusSubmitting = "submitting"
	AgentStatusWaiting    = "waiting"
	AgentStatusError      = "error"
	AgentStatusCompleted  = "completed"
)

var AgentStatuses = []string{
	AgentStatusIdle,
	AgentStatusThinking,
	AgentStatusEditing,
	AgentStatusTesting,
	AgentStatusSubmitting,
	AgentStatusWaiting,
	AgentStatusError,
	AgentStatusCompleted,
}

// Submission Status
const (
	SubmissionStatusPending   = "pending"
	SubmissionStatusRunning   = "running"
	SubmissionStatusPassed    = "passed"
	SubmissionStatusFailed    = "failed"
	SubmissionStatusTimeout   = "timeout"
	SubmissionStatusError     = "error"
)

var SubmissionStatuses = []string{
	SubmissionStatusPending,
	SubmissionStatusRunning,
	SubmissionStatusPassed,
	SubmissionStatusFailed,
	SubmissionStatusTimeout,
	SubmissionStatusError,
}

// Snippet
const (
	SnippetComplexityBasic       = "basic"
	SnippetComplexityIntermediate = "intermediate"
	SnippetComplexityAdvanced    = "advanced"
)

var SnippetComplexities = []string{
	SnippetComplexityBasic,
	SnippetComplexityIntermediate,
	SnippetComplexityAdvanced,
}

// Rating Tiers
const (
	RatingTierUnranked     = "unranked"
	RatingTierBronze       = "bronze"
	RatingTierSilver       = "silver"
	RatingTierGold         = "gold"
	RatingTierPlatinum     = "platinum"
	RatingTierDiamond      = "diamond"
	RatingTierMaster       = "master"
	RatingTierGrandmaster  = "grandmaster"
)

var RatingTiers = []string{
	RatingTierUnranked,
	RatingTierBronze,
	RatingTierSilver,
	RatingTierGold,
	RatingTierPlatinum,
	RatingTierDiamond,
	RatingTierMaster,
	RatingTierGrandmaster,
}

var RatingTierThresholds = map[string]int{
	RatingTierUnranked:    0,
	RatingTierBronze:      1200,
	RatingTierSilver:      1400,
	RatingTierGold:        1600,
	RatingTierPlatinum:    1800,
	RatingTierDiamond:     2000,
	RatingTierMaster:      2200,
	RatingTierGrandmaster: 2400,
}

var RatingTierColors = map[string]string{
	RatingTierUnranked:    "#8b949e",
	RatingTierBronze:      "#cd7f32",
	RatingTierSilver:      "#c0c0c0",
	RatingTierGold:        "#ffd700",
	RatingTierPlatinum:    "#e5e4e2",
	RatingTierDiamond:     "#b9f2ff",
	RatingTierMaster:      "#ff6b6b",
	RatingTierGrandmaster: "#ff3366",
}

// WebSocket Message Types
const (
	WSMsgAuth              = "auth"
	WSMsgAuthResponse      = "auth_response"
	WSMsgJoinMatch         = "join_match"
	WSMsgLeaveMatch        = "leave_match"
	WSMsgMatchState        = "match_state"
	WSMsgMatchUpdate       = "match_update"
	WSMsgAgentAction       = "agent_action"
	WSMsgAgentStatus       = "agent_status"
	WSMsgCodeChange        = "code_change"
	WSMsgSubmission        = "submission"
	WSMsgSubmissionResult  = "submission_result"
	WSMsgTestResult        = "test_result"
	WSMsgExecutionResult   = "execution_result"
	WSMsgTurnChange        = "turn_change"
	WSMsgTimerUpdate       = "timer_update"
	WSMsgMatchEnd          = "match_end"
	WSMsgLeaderboardUpdate = "leaderboard_update"
	WSMsgQueueStatus       = "queue_status"
	WSMsgError             = "error"
	WSMsgPing              = "ping"
	WSMsgPong              = "pong"
)

var WSMessageTypes = []string{
	WSMsgAuth,
	WSMsgAuthResponse,
	WSMsgJoinMatch,
	WSMsgLeaveMatch,
	WSMsgMatchState,
	WSMsgMatchUpdate,
	WSMsgAgentAction,
	WSMsgAgentStatus,
	WSMsgCodeChange,
	WSMsgSubmission,
	WSMsgSubmissionResult,
	WSMsgTestResult,
	WSMsgExecutionResult,
	WSMsgTurnChange,
	WSMsgTimerUpdate,
	WSMsgMatchEnd,
	WSMsgLeaderboardUpdate,
	WSMsgQueueStatus,
	WSMsgError,
	WSMsgPing,
	WSMsgPong,
}

// Limits
const (
	MaxUsernameLength     = 50
	MinUsernameLength     = 3
	MaxEmailLength        = 255
	MaxPasswordLength     = 128
	MinPasswordLength     = 8
	MaxTitleLength        = 255
	MaxDescriptionLength  = 10000
	MaxCodeLength         = 100000
	MaxTagsPerItem        = 20
	MaxTagLength          = 50
	MaxFilePathLength     = 500
	MaxFilesPerSubmission = 50
	MaxThoughtsLength     = 500
	MaxFeedbackLength     = 2000
)

// Pagination
const (
	DefaultLimit  = 20
	MaxLimit      = 100
	DefaultOffset = 0
)

// Cache
const (
	CacheTTLShort  = 60      // 1 minute
	CacheTTLMedium = 300     // 5 minutes
	CacheTTLLong   = 3600    // 1 hour
	CacheTTLDaily  = 86400   // 1 day
)

// Rate Limiting
const (
	DefaultRateLimitRequests = 100
	DefaultRateLimitWindow   = 60 // seconds
	AuthRateLimitRequests    = 10
	AuthRateLimitWindow      = 60 // seconds
	APIRateLimitRequests     = 1000
	APIRateLimitWindow       = 60 // seconds
)

// Monitoring
const (
	MetricsPort          = 9090
	HealthCheckInterval  = 30
	MetricsRetentionDays = 30
	LogRetentionDays     = 7
)

// File Paths
const (
	ConfigFilePath       = "./config.yaml"
	MigrationsPath       = "./migrations"
	SnippetsPath         = "./snippets"
	LogsPath             = "./logs"
	TempPath             = "/tmp/codecrash"
	SandboxWorkingDir    = "/workspace"
)

// Regex Patterns
const (
	UsernameRegex = `^[a-zA-Z0-9_-]{3,20}$`
	EmailRegex    = `^[^\s@]+@[^\s@]+\.[^\s@]+$`
	URLRegex      = `^https?://[^\s/$.?#].[^\s]*$`
	GitHubURLRegex = `^https://github\.com/[\w-]+/[\w.-]+`
)