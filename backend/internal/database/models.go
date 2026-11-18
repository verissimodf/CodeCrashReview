package database

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// User represents a user in the system
type User struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	Username          string     `json:"username" db:"username"`
	Email             string     `json:"email" db:"email"`
	PasswordHash      string     `json:"-" db:"password_hash"`
	AvatarURL         *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	Rating            int        `json:"rating" db:"rating"`
	RatingDeviation   float64    `json:"rating_deviation" db:"rating_deviation"`
	RatingVolatility  float64    `json:"rating_volatility" db:"rating_volatility"`
	PeakRating        int        `json:"peak_rating" db:"peak_rating"`
	TotalMatches      int        `json:"total_matches" db:"total_matches"`
	Wins              int        `json:"wins" db:"wins"`
	Losses            int        `json:"losses" db:"losses"`
	Draws             int        `json:"draws" db:"draws"`
	PreferredAgents   []string   `json:"preferred_agents" db:"preferred_agents"`
	PreferredLanguages []string  `json:"preferred_languages" db:"preferred_languages"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// Match represents a competitive match
type Match struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	Mode               string     `json:"mode" db:"mode"`
	Status             string     `json:"status" db:"status"`
	ChallengeID        uuid.UUID  `json:"challenge_id" db:"challenge_id"`
	CurrentTurn        int        `json:"current_turn" db:"current_turn"`
	MaxTurns           int        `json:"max_turns" db:"max_turns"`
	TimeLimitSeconds   int        `json:"time_limit_seconds" db:"time_limit_seconds"`
	TimeRemainingSeconds int      `json:"time_remaining_seconds" db:"time_remaining_seconds"`
	StartedAt          *time.Time `json:"started_at,omitempty" db:"started_at"`
	EndedAt            *time.Time `json:"ended_at,omitempty" db:"ended_at"`
	WinnerID           *uuid.UUID `json:"winner_id,omitempty" db:"winner_id"`
	WinnerPersona      *string    `json:"winner_persona,omitempty" db:"winner_persona"`
	Scoring            json.RawMessage `json:"scoring" db:"scoring"`
	ReplayURL          *string    `json:"replay_url,omitempty" db:"replay_url"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// MatchParticipant represents a participant in a match
type MatchParticipant struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	MatchID        uuid.UUID  `json:"match_id" db:"match_id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	AgentPersona   string     `json:"agent_persona" db:"agent_persona"`
	RatingBefore   int        `json:"rating_before" db:"rating_before"`
	RatingAfter    *int       `json:"rating_after,omitempty" db:"rating_after"`
	RatingChange   *int       `json:"rating_change,omitempty" db:"rating_change"`
	Score          float64    `json:"score" db:"score"`
	Status         string     `json:"status" db:"status"`
	Connected      bool       `json:"connected" db:"connected"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// Challenge represents a coding challenge
type Challenge struct {
	ID                   uuid.UUID      `json:"id" db:"id"`
	Title                string         `json:"title" db:"title"`
	Description          string         `json:"description" db:"description"`
	RepositoryURL        string         `json:"repository_url" db:"repository_url"`
	RepositoryBranch     string         `json:"repository_branch" db:"repository_branch"`
	TargetPaths          pgtype.TextArray `json:"target_paths" db:"target_paths"`
	ExcludePaths         pgtype.TextArray `json:"exclude_paths" db:"exclude_paths"`
	Severity             string         `json:"severity" db:"severity"`
	Category             string         `json:"category" db:"category"`
	TimeLimitSeconds     int            `json:"time_limit_seconds" db:"time_limit_seconds"`
	MaxTurns             int            `json:"max_turns" db:"max_turns"`
	AcceptanceCriteria   json.RawMessage `json:"acceptance_criteria" db:"acceptance_criteria"`
	ScoringWeights       json.RawMessage `json:"scoring_weights" db:"scoring_weights"`
	KnownIssues          json.RawMessage `json:"known_issues" db:"known_issues"`
	Tags                 pgtype.TextArray `json:"tags" db:"tags"`
	TemplateID           *uuid.UUID     `json:"template_id,omitempty" db:"template_id"`
	IsPublic             bool           `json:"is_public" db:"is_public"`
	CreatedBy            uuid.UUID      `json:"created_by" db:"created_by"`
	PlayCount            int            `json:"play_count" db:"play_count"`
	AverageRating        float64        `json:"average_rating" db:"average_rating"`
	CreatedAt            time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at" db:"updated_at"`
}

// ChallengeTemplate represents a challenge template
type ChallengeTemplate struct {
	ID                          uuid.UUID      `json:"id" db:"id"`
	Name                        string         `json:"name" db:"name"`
	Description                 string         `json:"description" db:"description"`
	Category                    string         `json:"category" db:"category"`
	DefaultSeverity             string         `json:"default_severity" db:"default_severity"`
	DefaultTimeLimitSeconds     int            `json:"default_time_limit_seconds" db:"default_time_limit_seconds"`
	DefaultMaxTurns             int            `json:"default_max_turns" db:"default_max_turns"`
	AcceptanceCriteriaTemplate  json.RawMessage `json:"acceptance_criteria_template" db:"acceptance_criteria_template"`
	ScoringWeightsTemplate      json.RawMessage `json:"scoring_weights_template" db:"scoring_weights_template"`
	KnownIssuesTemplate         json.RawMessage `json:"known_issues_template" db:"known_issues_template"`
	ExampleRepository           *string        `json:"example_repository,omitempty" db:"example_repository"`
	CreatedAt                   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt                   time.Time      `json:"updated_at" db:"updated_at"`
}

// Snippet represents a code snippet
type Snippet struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	Title       string         `json:"title" db:"title"`
	Description string         `json:"description" db:"description"`
	Language    string         `json:"language" db:"language"`
	Code        string         `json:"code" db:"code"`
	Tags        pgtype.TextArray `json:"tags" db:"tags"`
	Complexity  string         `json:"complexity" db:"complexity"`
	Author      string         `json:"author" db:"author"`
	AuthorID    *uuid.UUID     `json:"author_id,omitempty" db:"author_id"`
	SourceURL   *string        `json:"source_url,omitempty" db:"source_url"`
	Rating      float64        `json:"rating" db:"rating"`
	UsageCount  int            `json:"usage_count" db:"usage_count"`
	IsOfficial  bool           `json:"is_official" db:"is_official"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// SnippetCollection represents a collection of snippets
type SnippetCollection struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`
	Description string         `json:"description" db:"description"`
	SnippetIDs  pgtype.TextArray `json:"snippet_ids" db:"snippet_ids"`
	Tags        pgtype.TextArray `json:"tags" db:"tags"`
	IsCurated   bool           `json:"is_curated" db:"is_curated"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
}

// Submission represents a code submission in a match
type Submission struct {
	ID               uuid.UUID      `json:"id" db:"id"`
	MatchID          uuid.UUID      `json:"match_id" db:"match_id"`
	ParticipantID    uuid.UUID      `json:"participant_id" db:"participant_id"`
	AgentPersona     string         `json:"agent_persona" db:"agent_persona"`
	Turn             int            `json:"turn" db:"turn"`
	Files            json.RawMessage `json:"files" db:"files"`
	Changes          json.RawMessage `json:"changes" db:"changes"`
	TestResult       json.RawMessage `json:"test_result,omitempty" db:"test_result"`
	ExecutionResult  json.RawMessage `json:"execution_result,omitempty" db:"execution_result"`
	Score            *float64       `json:"score,omitempty" db:"score"`
	Status           string         `json:"status" db:"status"`
	SubmittedAt      time.Time      `json:"submitted_at" db:"submitted_at"`
	EvaluatedAt      *time.Time     `json:"evaluated_at,omitempty" db:"evaluated_at"`
	Feedback         *string        `json:"feedback,omitempty" db:"feedback"`
}

// AgentPersonaConfig represents an agent persona configuration
type AgentPersonaConfig struct {
	ID          string         `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`
	Description string         `json:"description" db:"description"`
	Specialty   string         `json:"specialty" db:"specialty"`
	Model       string         `json:"model" db:"model"`
	Temperature float64        `json:"temperature" db:"temperature"`
	Tools       pgtype.TextArray `json:"tools" db:"tools"`
	SystemPrompt string        `json:"system_prompt" db:"system_prompt"`
	Color       string         `json:"color" db:"color"`
	GlowColor   string         `json:"glow_color" db:"glow_color"`
	Icon        string         `json:"icon" db:"icon"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// RatingHistory represents a rating change event
type RatingHistory struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	MatchID         uuid.UUID `json:"match_id" db:"match_id"`
	RatingBefore    int       `json:"rating_before" db:"rating_before"`
	RatingAfter     int       `json:"rating_after" db:"rating_after"`
	RatingChange    int       `json:"rating_change" db:"rating_change"`
	DeviationBefore float64   `json:"deviation_before" db:"deviation_before"`
	DeviationAfter  float64   `json:"deviation_after" db:"deviation_after"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	Username        string    `json:"username" db:"username"`
	AvatarURL       *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	Rating          int       `json:"rating" db:"rating"`
	RatingDeviation float64   `json:"rating_deviation" db:"rating_deviation"`
	PeakRating      int       `json:"peak_rating" db:"peak_rating"`
	TotalMatches    int       `json:"total_matches" db:"total_matches"`
	Wins            int       `json:"wins" db:"wins"`
	Losses          int       `json:"losses" db:"losses"`
	WinRate         float64   `json:"win_rate" db:"win_rate"`
	PreferredAgent  *string   `json:"preferred_agent,omitempty" db:"preferred_agent"`
	PreferredLanguage *string `json:"preferred_language,omitempty" db:"preferred_language"`
	LastActive      time.Time `json:"last_active" db:"last_active"`
}