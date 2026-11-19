package matchmaking

import (
	"context"
	"encoding/json"
	"fmt"

	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"codecrashreview/backend/internal/database"
)

type Service struct {
	matchRepo       *database.MatchRepository
	participantRepo *database.MatchParticipantRepository
	userRepo        *database.UserRepository
	challengeRepo   *database.ChallengeRepository
	redis           *redis.Client
	logger          zerolog.Logger
	eloKFactor      float64
	glickoTau       float64

	// Queue management
	queueMu    sync.Mutex
	queues     map[string][]*QueueEntry // key: "mode:language:rating_bucket"
	processing bool
}

type QueueEntry struct {
	UserID         uuid.UUID
	Username       string
	Mode           string
	PreferredAgent string
	Language       string
	Rating         int
	JoinedAt       time.Time
}

type MatchResult struct {
	MatchID      uuid.UUID
	ParticipantA *MatchParticipantInfo
	ParticipantB *MatchParticipantInfo
	ChallengeID  uuid.UUID
}

type MatchParticipantInfo struct {
	UserID       uuid.UUID
	Username     string
	AgentPersona string
	Rating       int
}

type CodeFile struct {
	Path     string
	Language string
	Content  string
}

type CodeChange struct {
	ID              string
	FilePath        string
	Language        string
	OriginalContent string
	NewContent      string
	Diff            []DiffHunk
	Description     string
	Turn            int
	Timestamp       string
	Accepted        *bool
}

type DiffHunk struct {
	OldStart int
	OldLines int
	NewStart int
	NewLines int
	Lines    []DiffLine
}

type DiffLine struct {
	Type       string
	Content    string
	LineNumber int
}

func NewService(
	matchRepo *database.MatchRepository,
	participantRepo *database.MatchParticipantRepository,
	userRepo *database.UserRepository,
	challengeRepo *database.ChallengeRepository,
	redis *redis.Client,
	eloKFactor float64,
	glickoTau float64,
) *Service {
	return &Service{
		matchRepo:       matchRepo,
		participantRepo: participantRepo,
		userRepo:        userRepo,
		challengeRepo:   challengeRepo,
		redis:           redis,
		logger:          *zerolog.Ctx(context.Background()),
		eloKFactor:      eloKFactor,
		glickoTau:       glickoTau,
		queues:          make(map[string][]*QueueEntry),
	}
}

func (s *Service) StartWorker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processQueues(ctx)
		}
	}
}

func (s *Service) processQueues(ctx context.Context) {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()
	// Process each queue
	for queueKey, queue := range s.queues {
		if len(queue) < 2 {
			continue
		}

		// Try to match players
		matched := s.tryMatch(queue)
		if matched != nil {
			var challengeID uuid.UUID
			err := s.matchRepo.GetPool().QueryRow(ctx, `SELECT id FROM challenges WHERE is_public ORDER BY play_count,created_at LIMIT 1`).Scan(&challengeID)
			if err != nil {
				continue
			}
			people := make([]*MatchParticipantInfo, 0, len(matched))
			for _, entry := range matched {
				people = append(people, &MatchParticipantInfo{UserID: entry.UserID, AgentPersona: entry.PreferredAgent, Rating: entry.Rating})
			}
			if _, err := s.CreateMatch(ctx, challengeID, matched[0].Mode, people); err != nil {
				s.logger.Error().Err(err).Msg("Matchmaking failed")
				continue
			}

			s.removeMatchedFromQueue(queueKey, matched)
		}
	}
}

func (s *Service) tryMatch(queue []*QueueEntry) []*QueueEntry {
	// Simple matching: pair first two players within rating range
	for i := 0; i < len(queue)-1; i++ {
		for j := i + 1; j < len(queue); j++ {
			if s.canMatch(queue[i], queue[j]) {
				return []*QueueEntry{queue[i], queue[j]}
			}
		}
	}
	return nil
}

func (s *Service) canMatch(a, b *QueueEntry) bool {
	ratingDiff := int(math.Abs(float64(a.Rating - b.Rating)))
	return ratingDiff <= 200 // Matchmaking range
}

func (s *Service) removeMatchedFromQueue(queueKey string, matched []*QueueEntry) {
	queue := s.queues[queueKey]
	matchedIDs := make(map[uuid.UUID]bool)
	for _, m := range matched {
		matchedIDs[m.UserID] = true
	}

	newQueue := make([]*QueueEntry, 0, len(queue))
	for _, entry := range queue {
		if !matchedIDs[entry.UserID] {
			newQueue = append(newQueue, entry)
		}
	}
	s.queues[queueKey] = newQueue
}

func (s *Service) JoinQueue(ctx context.Context, userID uuid.UUID, mode, preferredAgent, language string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	entry := &QueueEntry{
		UserID:         userID,
		Username:       user.Username,
		Mode:           mode,
		PreferredAgent: preferredAgent,
		Language:       language,
		Rating:         user.Rating,
		JoinedAt:       time.Now(),
	}

	queueKey := fmt.Sprintf("%s:%s", mode, language)
	s.queueMu.Lock()
	for _, entries := range s.queues {
		for _, e := range entries {
			if e.UserID == userID {
				s.queueMu.Unlock()
				return fmt.Errorf("already queued")
			}
		}
	}
	s.queues[queueKey] = append(s.queues[queueKey], entry)
	s.queueMu.Unlock()

	s.logger.Info().Str("user_id", userID.String()).Str("queue", queueKey).Msg("Player joined queue")
	return nil
}

func (s *Service) LeaveQueue(ctx context.Context, userID uuid.UUID) error {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()

	for queueKey, queue := range s.queues {
		newQueue := make([]*QueueEntry, 0, len(queue))
		for _, entry := range queue {
			if entry.UserID != userID {
				newQueue = append(newQueue, entry)
			}
		}
		s.queues[queueKey] = newQueue
	}

	// Remove from Redis (would need to scan, simplified here)
	s.logger.Info().Str("user_id", userID.String()).Msg("Player left queue")
	return nil
}

func (s *Service) GetQueueStatus(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()

	// Find user in queues
	for queueKey, queue := range s.queues {
		for i, entry := range queue {
			if entry.UserID == userID {
				return map[string]interface{}{
					"in_queue":           true,
					"queue_key":          queueKey,
					"position":           i + 1,
					"players_in_queue":   len(queue),
					"estimated_wait_sec": len(queue) * 30, // rough estimate
				}, nil
			}
		}
	}

	return map[string]interface{}{
		"in_queue": false,
	}, nil
}

func (s *Service) CreateMatch(ctx context.Context, challengeID uuid.UUID, mode string, participants []*MatchParticipantInfo) (*database.Match, error) {
	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	match := &database.Match{
		Mode:                 mode,
		Status:               "preparing",
		ChallengeID:          challengeID,
		CurrentTurn:          0,
		MaxTurns:             challenge.MaxTurns,
		TimeLimitSeconds:     challenge.TimeLimitSeconds,
		TimeRemainingSeconds: challenge.TimeLimitSeconds,
		Scoring:              json.RawMessage(`{}`),
	}

	tx, err := s.matchRepo.GetPool().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	match.Status = "active"
	err = tx.QueryRow(ctx, `INSERT INTO matches(mode,status,challenge_id,max_turns,time_limit_seconds,time_remaining_seconds,started_at) VALUES($1,'active',$2,$3,$4,$4,now()) RETURNING id,created_at,updated_at,started_at`, mode, challengeID, challenge.MaxTurns, challenge.TimeLimitSeconds).Scan(&match.ID, &match.CreatedAt, &match.UpdatedAt, &match.StartedAt)
	if err != nil {
		return nil, err
	}
	for _, p := range participants {
		_, err = tx.Exec(ctx, `INSERT INTO match_participants(match_id,user_id,agent_persona,rating_before) VALUES($1,$2,$3,$4)`, match.ID, p.UserID, p.AgentPersona, p.Rating)
		if err != nil {
			return nil, err
		}
	}
	_, err = tx.Exec(ctx, `UPDATE challenges SET play_count=play_count+1 WHERE id=$1`, challengeID)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	s.logger.Info().Str("match_id", match.ID.String()).Msg("Match created")
	return match, nil
}

func (s *Service) JoinMatch(ctx context.Context, matchID, userID uuid.UUID) error {
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return err
	}

	if match.Status != "preparing" && match.Status != "active" {
		return fmt.Errorf("match not joinable")
	}

	participants, err := s.participantRepo.GetByMatch(ctx, matchID)
	if err != nil {
		return err
	}

	// Find participant slot for this user
	var participant *database.MatchParticipant
	for _, p := range participants {
		if p.UserID == userID {
			participant = p
			break
		}
	}

	if participant == nil {
		return fmt.Errorf("user not in match")
	}

	participant.Connected = true
	participant.Status = "waiting"
	return s.participantRepo.Update(ctx, participant)
}

func (s *Service) LeaveMatch(ctx context.Context, matchID, userID uuid.UUID) error {
	participants, err := s.participantRepo.GetByMatch(ctx, matchID)
	if err != nil {
		return err
	}

	for _, p := range participants {
		if p.UserID == userID {
			p.Connected = false
			p.Status = "left"
			return s.participantRepo.Update(ctx, p)
		}
	}

	return fmt.Errorf("user not in match")
}

func (s *Service) ListMatches(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*database.Match, error) {
	return s.matchRepo.ListByUser(ctx, userID, limit, offset)
}

func (s *Service) GetMatch(ctx context.Context, matchID uuid.UUID) (*database.Match, error) {
	return s.matchRepo.GetByID(ctx, matchID)
}

// Elo rating calculation
func (s *Service) CalculateEloChange(ratingA, ratingB int, scoreA float64) int {
	expectedA := 1.0 / (1.0 + math.Pow(10, float64(ratingB-ratingA)/400))
	change := s.eloKFactor * (scoreA - expectedA)
	return int(math.Round(change))
}

// Glicko-2 rating calculation (simplified)
func (s *Service) CalculateGlicko2Change(rating, deviation, volatility float64, results []GlickoResult) (newRating, newDeviation, newVolatility float64) {
	// This is a simplified version - full implementation is complex
	// See: http://www.glicko.net/glicko/glicko2.pdf

	// For now, use Elo as approximation
	// Real implementation would:
	// 1. Convert to Glicko-2 scale (mu = (rating - 1500) / 173.7178, phi = deviation / 173.7178)
	// 2. Calculate new volatility using Illinois algorithm
	// 3. Update mu and phi based on results
	// 4. Convert back to rating scale

	return rating, deviation, volatility
}

type GlickoResult struct {
	OpponentRating    float64
	OpponentDeviation float64
	Score             float64 // 1.0 win, 0.5 draw, 0.0 loss
}
