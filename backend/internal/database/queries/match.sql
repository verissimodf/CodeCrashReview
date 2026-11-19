-- name: CreateMatch :one
INSERT INTO matches (mode, status, challenge_id, current_turn, max_turns, time_limit_seconds, time_remaining_seconds, scoring)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, created_at, updated_at;

-- name: GetMatchByID :one
SELECT id, mode, status, challenge_id, current_turn, max_turns, time_limit_seconds, time_remaining_seconds,
       started_at, ended_at, winner_id, winner_persona, scoring, replay_url, created_at, updated_at
FROM matches WHERE id = $1;

-- name: UpdateMatch :exec
UPDATE matches SET
    mode = $2, status = $3, challenge_id = $4, current_turn = $5,
    max_turns = $6, time_limit_seconds = $7, time_remaining_seconds = $8,
    started_at = $9, ended_at = $10, winner_id = $11, winner_persona = $12,
    scoring = $13, replay_url = $14, updated_at = NOW()
WHERE id = $1;

-- name: ListMatchesByUser :many
SELECT m.id, m.mode, m.status, m.challenge_id, m.current_turn, m.max_turns,
       m.time_limit_seconds, m.time_remaining_seconds, m.started_at, m.ended_at,
       m.winner_id, m.winner_persona, m.scoring, m.replay_url, m.created_at, m.updated_at
FROM matches m
JOIN match_participants mp ON mp.match_id = m.id
WHERE mp.user_id = $1
ORDER BY m.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveMatches :many
SELECT id, mode, status, challenge_id, current_turn, max_turns,
       time_limit_seconds, time_remaining_seconds, started_at, created_at
FROM matches
WHERE status IN ('preparing', 'active', 'paused')
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountActiveMatches :one
SELECT COUNT(*) FROM matches WHERE status IN ('preparing', 'active', 'paused');

-- name: DeleteMatch :exec
DELETE FROM matches WHERE id = $1;

-- name: CreateMatchParticipant :one
INSERT INTO match_participants (match_id, user_id, agent_persona, rating_before, score, status, connected)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at;

-- name: GetMatchParticipants :many
SELECT id, match_id, user_id, agent_persona, rating_before, rating_after, rating_change,
       score, status, connected, created_at
FROM match_participants WHERE match_id = $1;

-- name: UpdateMatchParticipant :exec
UPDATE match_participants SET
    rating_after = $2, rating_change = $3, score = $4, status = $5, connected = $6
WHERE id = $1;

-- name: GetParticipantByMatchAndUser :one
SELECT id, match_id, user_id, agent_persona, rating_before, rating_after, rating_change,
       score, status, connected, created_at
FROM match_participants WHERE match_id = $1 AND user_id = $2;

-- name: CreateSubmission :one
INSERT INTO submissions (match_id, participant_id, agent_persona, turn, files, changes, test_result, execution_result, score, status, feedback)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, submitted_at;

-- name: GetSubmissionsByMatch :many
SELECT id, match_id, participant_id, agent_persona, turn, files, changes, test_result, execution_result, score, status, submitted_at, evaluated_at, feedback
FROM submissions WHERE match_id = $1 ORDER BY turn, submitted_at;

-- name: GetSubmissionsByParticipant :many
SELECT id, match_id, participant_id, agent_persona, turn, files, changes, test_result, execution_result, score, status, submitted_at, evaluated_at, feedback
FROM submissions WHERE participant_id = $1 ORDER BY turn;

-- name: UpdateSubmission :exec
UPDATE submissions SET
    test_result = $2, execution_result = $3, score = $4, status = $5, evaluated_at = NOW(), feedback = $6
WHERE id = $1;

-- name: CreateRatingHistory :one
INSERT INTO rating_history (user_id, match_id, rating_before, rating_after, rating_change, deviation_before, deviation_after)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at;