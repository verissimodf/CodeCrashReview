package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
)

// UserRepository handles user database operations
type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (username, email, password_hash, rating, rating_deviation, rating_volatility, peak_rating, preferred_agents, preferred_languages)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	preferredAgentsJSON, _ := json.Marshal(user.PreferredAgents)
	preferredLanguagesJSON, _ := json.Marshal(user.PreferredLanguages)

	return r.pool.QueryRow(ctx, query,
		user.Username, user.Email, user.PasswordHash,
		user.Rating, user.RatingDeviation, user.RatingVolatility, user.PeakRating,
		preferredAgentsJSON, preferredLanguagesJSON,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `SELECT id, username, email, password_hash, avatar_url, rating, rating_deviation, rating_volatility, peak_rating, total_matches, wins, losses, draws, preferred_agents, preferred_languages, created_at, updated_at FROM users WHERE id = $1`
	user := &User{}
	var preferredAgents, preferredLanguages []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.AvatarURL,
		&user.Rating, &user.RatingDeviation, &user.RatingVolatility, &user.PeakRating,
		&user.TotalMatches, &user.Wins, &user.Losses, &user.Draws,
		&preferredAgents, &preferredLanguages,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(preferredAgents, &user.PreferredAgents)
	json.Unmarshal(preferredLanguages, &user.PreferredLanguages)
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, username, email, password_hash, avatar_url, rating, rating_deviation, rating_volatility, peak_rating, total_matches, wins, losses, draws, preferred_agents, preferred_languages, created_at, updated_at FROM users WHERE email = $1`
	user := &User{}
	var preferredAgents, preferredLanguages []byte

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.AvatarURL,
		&user.Rating, &user.RatingDeviation, &user.RatingVolatility, &user.PeakRating,
		&user.TotalMatches, &user.Wins, &user.Losses, &user.Draws,
		&preferredAgents, &preferredLanguages,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(preferredAgents, &user.PreferredAgents)
	json.Unmarshal(preferredLanguages, &user.PreferredLanguages)
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users SET
			username = $2, email = $3, avatar_url = $4, rating = $5,
			rating_deviation = $6, rating_volatility = $7, peak_rating = $8,
			total_matches = $9, wins = $10, losses = $11, draws = $12,
			preferred_agents = $13, preferred_languages = $14, updated_at = NOW()
		WHERE id = $1
	`
	preferredAgentsJSON, _ := json.Marshal(user.PreferredAgents)
	preferredLanguagesJSON, _ := json.Marshal(user.PreferredLanguages)

	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Username, user.Email, user.AvatarURL,
		user.Rating, user.RatingDeviation, user.RatingVolatility, user.PeakRating,
		user.TotalMatches, user.Wins, user.Losses, user.Draws,
		preferredAgentsJSON, preferredLanguagesJSON,
	)
	return err
}

func (r *UserRepository) UpdateRating(ctx context.Context, userID uuid.UUID, rating int, deviation, volatility float64) error {
	query := `
		UPDATE users SET
			rating = $2, rating_deviation = $3, rating_volatility = $4,
			peak_rating = GREATEST(peak_rating, $2),
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, userID, rating, deviation, volatility)
	return err
}

func (r *UserRepository) IncrementMatchStats(ctx context.Context, userID uuid.UUID, win, loss, draw bool) error {
	query := `
		UPDATE users SET
			total_matches = total_matches + 1,
			wins = wins + $2,
			losses = losses + $3,
			draws = draws + $4,
			updated_at = NOW()
		WHERE id = $1
	`
	winInt, lossInt, drawInt := 0, 0, 0
	if win { winInt = 1 }
	if loss { lossInt = 1 }
	if draw { drawInt = 1 }

	_, err := r.pool.Exec(ctx, query, userID, winInt, lossInt, drawInt)
	return err
}

// MatchRepository handles match database operations
type MatchRepository struct {
	pool *pgxpool.Pool
}

func NewMatchRepository(pool *pgxpool.Pool) *MatchRepository {
	return &MatchRepository{pool: pool}
}

func (r *MatchRepository) Create(ctx context.Context, match *Match) error {
	query := `
		INSERT INTO matches (mode, status, challenge_id, current_turn, max_turns, time_limit_seconds, time_remaining_seconds, scoring)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	scoringJSON, _ := json.Marshal(match.Scoring)

	return r.pool.QueryRow(ctx, query,
		match.Mode, match.Status, match.ChallengeID,
		match.CurrentTurn, match.MaxTurns, match.TimeLimitSeconds, match.TimeRemainingSeconds,
		scoringJSON,
	).Scan(&match.ID, &match.CreatedAt, &match.UpdatedAt)
}

func (r *MatchRepository) GetByID(ctx context.Context, id uuid.UUID) (*Match, error) {
	query := `SELECT id, mode, status, challenge_id, current_turn, max_turns, time_limit_seconds, time_remaining_seconds, started_at, ended_at, winner_id, winner_persona, scoring, replay_url, created_at, updated_at FROM matches WHERE id = $1`
	match := &Match{}
	var scoring []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&match.ID, &match.Mode, &match.Status, &match.ChallengeID,
		&match.CurrentTurn, &match.MaxTurns, &match.TimeLimitSeconds, &match.TimeRemainingSeconds,
		&match.StartedAt, &match.EndedAt, &match.WinnerID, &match.WinnerPersona,
		&scoring, &match.ReplayURL, &match.CreatedAt, &match.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(scoring, &match.Scoring)
	return match, nil
}

func (r *MatchRepository) Update(ctx context.Context, match *Match) error {
	query := `
		UPDATE matches SET
			mode = $2, status = $3, challenge_id = $4, current_turn = $5,
			max_turns = $6, time_limit_seconds = $7, time_remaining_seconds = $8,
			started_at = $9, ended_at = $10, winner_id = $11, winner_persona = $12,
			scoring = $13, replay_url = $14, updated_at = NOW()
		WHERE id = $1
	`
	scoringJSON, _ := json.Marshal(match.Scoring)

	_, err := r.pool.Exec(ctx, query,
		match.ID, match.Mode, match.Status, match.ChallengeID,
		match.CurrentTurn, match.MaxTurns, match.TimeLimitSeconds, match.TimeRemainingSeconds,
		match.StartedAt, match.EndedAt, match.WinnerID, match.WinnerPersona,
		scoringJSON, match.ReplayURL,
	)
	return err
}

func (r *MatchRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Match, error) {
	query := `
		SELECT m.id, m.mode, m.status, m.challenge_id, m.current_turn, m.max_turns,
			m.time_limit_seconds, m.time_remaining_seconds, m.started_at, m.ended_at,
			m.winner_id, m.winner_persona, m.scoring, m.replay_url, m.created_at, m.updated_at
		FROM matches m
		JOIN match_participants mp ON mp.match_id = m.id
		WHERE mp.user_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []*Match
	for rows.Next() {
		match := &Match{}
		var scoring []byte
		err := rows.Scan(
			&match.ID, &match.Mode, &match.Status, &match.ChallengeID,
			&match.CurrentTurn, &match.MaxTurns, &match.TimeLimitSeconds, &match.TimeRemainingSeconds,
			&match.StartedAt, &match.EndedAt, &match.WinnerID, &match.WinnerPersona,
			&scoring, &match.ReplayURL, &match.CreatedAt, &match.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(scoring, &match.Scoring)
		matches = append(matches, match)
	}
	return matches, nil
}

// MatchParticipantRepository handles match participant operations
type MatchParticipantRepository struct {
	pool *pgxpool.Pool
}

func NewMatchParticipantRepository(pool *pgxpool.Pool) *MatchParticipantRepository {
	return &MatchParticipantRepository{pool: pool}
}

func (r *MatchParticipantRepository) Create(ctx context.Context, participant *MatchParticipant) error {
	query := `
		INSERT INTO match_participants (match_id, user_id, agent_persona, rating_before, score, status, connected)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query,
		participant.MatchID, participant.UserID, participant.AgentPersona,
		participant.RatingBefore, participant.Score, participant.Status, participant.Connected,
	).Scan(&participant.ID, &participant.CreatedAt)
}

func (r *MatchParticipantRepository) GetByMatch(ctx context.Context, matchID uuid.UUID) ([]*MatchParticipant, error) {
	query := `SELECT id, match_id, user_id, agent_persona, rating_before, rating_after, rating_change, score, status, connected, created_at FROM match_participants WHERE match_id = $1`
	rows, err := r.pool.Query(ctx, query, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []*MatchParticipant
	for rows.Next() {
		p := &MatchParticipant{}
		err := rows.Scan(
			&p.ID, &p.MatchID, &p.UserID, &p.AgentPersona,
			&p.RatingBefore, &p.RatingAfter, &p.RatingChange,
			&p.Score, &p.Status, &p.Connected, &p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}
	return participants, nil
}

func (r *MatchParticipantRepository) Update(ctx context.Context, participant *MatchParticipant) error {
	query := `
		UPDATE match_participants SET
			rating_after = $2, rating_change = $3, score = $4, status = $5, connected = $6
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		participant.ID, participant.RatingAfter, participant.RatingChange,
		participant.Score, participant.Status, participant.Connected,
	)
	return err
}

// ChallengeRepository handles challenge operations
type ChallengeRepository struct {
	pool *pgxpool.Pool
}

func NewChallengeRepository(pool *pgxpool.Pool) *ChallengeRepository {
	return &ChallengeRepository{pool: pool}
}

func (r *ChallengeRepository) Create(ctx context.Context, challenge *Challenge) error {
	query := `
		INSERT INTO challenges (title, description, repository_url, repository_branch, target_paths, exclude_paths, severity, category, time_limit_seconds, max_turns, acceptance_criteria, scoring_weights, known_issues, tags, template_id, is_public, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, updated_at
	`
	acceptanceCriteriaJSON, _ := json.Marshal(challenge.AcceptanceCriteria)
	scoringWeightsJSON, _ := json.Marshal(challenge.ScoringWeights)
	knownIssuesJSON, _ := json.Marshal(challenge.KnownIssues)

	var targetPaths, excludePaths, tags pgtype.TextArray
	targetPaths.Set(challenge.TargetPaths)
	excludePaths.Set(challenge.ExcludePaths)
	tags.Set(challenge.Tags)

	return r.pool.QueryRow(ctx, query,
		challenge.Title, challenge.Description, challenge.RepositoryURL, challenge.RepositoryBranch,
		targetPaths, excludePaths, challenge.Severity, challenge.Category,
		challenge.TimeLimitSeconds, challenge.MaxTurns,
		acceptanceCriteriaJSON, scoringWeightsJSON, knownIssuesJSON,
		tags, challenge.TemplateID, challenge.IsPublic, challenge.CreatedBy,
	).Scan(&challenge.ID, &challenge.CreatedAt, &challenge.UpdatedAt)
}

func (r *ChallengeRepository) GetByID(ctx context.Context, id uuid.UUID) (*Challenge, error) {
	query := `SELECT id, title, description, repository_url, repository_branch, target_paths, exclude_paths, severity, category, time_limit_seconds, max_turns, acceptance_criteria, scoring_weights, known_issues, tags, template_id, is_public, created_by, play_count, average_rating, created_at, updated_at FROM challenges WHERE id = $1`
	challenge := &Challenge{}
	var acceptanceCriteria, scoringWeights, knownIssues []byte
	var targetPaths, excludePaths, tags pgtype.TextArray

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&challenge.ID, &challenge.Title, &challenge.Description, &challenge.RepositoryURL, &challenge.RepositoryBranch,
		&targetPaths, &excludePaths, &challenge.Severity, &challenge.Category,
		&challenge.TimeLimitSeconds, &challenge.MaxTurns,
		&acceptanceCriteria, &scoringWeights, &knownIssues,
		&tags, &challenge.TemplateID, &challenge.IsPublic, &challenge.CreatedBy,
		&challenge.PlayCount, &challenge.AverageRating, &challenge.CreatedAt, &challenge.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	challenge.TargetPaths = targetPaths.Elements
	challenge.ExcludePaths = excludePaths.Elements
	challenge.Tags = tags.Elements
	json.Unmarshal(acceptanceCriteria, &challenge.AcceptanceCriteria)
	json.Unmarshal(scoringWeights, &challenge.ScoringWeights)
	json.Unmarshal(knownIssues, &challenge.KnownIssues)
	return challenge, nil
}

func (r *ChallengeRepository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*Challenge, error) {
	query := `SELECT id, title, description, repository_url, repository_branch, target_paths, exclude_paths, severity, category, time_limit_seconds, max_turns, acceptance_criteria, scoring_weights, known_issues, tags, template_id, is_public, created_by, play_count, average_rating, created_at, updated_at FROM challenges WHERE is_public = true`
	args := []interface{}{}
	argCount := 1

	if category, ok := filters["category"].(string); ok && category != "" {
		query += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, category)
		argCount++
	}
	if severity, ok := filters["severity"].(string); ok && severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", argCount)
		args = append(args, severity)
		argCount++
	}
	if search, ok := filters["search"].(string); ok && search != "" {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+search+"%")
		argCount++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var challenges []*Challenge
	for rows.Next() {
		c := &Challenge{}
		var acceptanceCriteria, scoringWeights, knownIssues []byte
		var targetPaths, excludePaths, tags pgtype.TextArray

		err := rows.Scan(
			&c.ID, &c.Title, &c.Description, &c.RepositoryURL, &c.RepositoryBranch,
			&targetPaths, &excludePaths, &c.Severity, &c.Category,
			&c.TimeLimitSeconds, &c.MaxTurns,
			&acceptanceCriteria, &scoringWeights, &knownIssues,
			&tags, &c.TemplateID, &c.IsPublic, &c.CreatedBy,
			&c.PlayCount, &c.AverageRating, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		c.TargetPaths = targetPaths.Elements
		c.ExcludePaths = excludePaths.Elements
		c.Tags = tags.Elements
		json.Unmarshal(acceptanceCriteria, &c.AcceptanceCriteria)
		json.Unmarshal(scoringWeights, &c.ScoringWeights)
		json.Unmarshal(knownIssues, &c.KnownIssues)
		challenges = append(challenges, c)
	}
	return challenges, nil
}

// SnippetRepository handles snippet operations
type SnippetRepository struct {
	pool *pgxpool.Pool
}

func NewSnippetRepository(pool *pgxpool.Pool) *SnippetRepository {
	return &SnippetRepository{pool: pool}
}

func (r *SnippetRepository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*Snippet, error) {
	query := `SELECT id, title, description, language, code, tags, complexity, author, author_id, source_url, rating, usage_count, is_official, created_at, updated_at FROM snippets WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	if language, ok := filters["language"].(string); ok && language != "" {
		query += fmt.Sprintf(" AND language = $%d", argCount)
		args = append(args, language)
		argCount++
	}
	if tag, ok := filters["tag"].(string); ok && tag != "" {
		query += fmt.Sprintf(" AND $%d = ANY(tags)", argCount)
		args = append(args, tag)
		argCount++
	}
	if complexity, ok := filters["complexity"].(string); ok && complexity != "" {
		query += fmt.Sprintf(" AND complexity = $%d", argCount)
		args = append(args, complexity)
		argCount++
	}
	if search, ok := filters["search"].(string); ok && search != "" {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d OR code ILIKE $%d)", argCount, argCount, argCount)
		args = append(args, "%"+search+"%")
		argCount++
	}

	query += fmt.Sprintf(" ORDER BY rating DESC, usage_count DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snippets []*Snippet
	for rows.Next() {
		s := &Snippet{}
		var tags pgtype.TextArray
		err := rows.Scan(
			&s.ID, &s.Title, &s.Description, &s.Language, &s.Code,
			&tags, &s.Complexity, &s.Author, &s.AuthorID, &s.SourceURL,
			&s.Rating, &s.UsageCount, &s.IsOfficial, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		s.Tags = tags.Elements
		snippets = append(snippets, s)
	}
	return snippets, nil
}