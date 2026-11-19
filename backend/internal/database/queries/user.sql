-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, rating, rating_deviation, rating_volatility, peak_rating, preferred_agents, preferred_languages)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, username, email, password_hash, avatar_url, rating, rating_deviation, rating_volatility, peak_rating,
       total_matches, wins, losses, draws, preferred_agents, preferred_languages, created_at, updated_at
FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, username, email, password_hash, avatar_url, rating, rating_deviation, rating_volatility, peak_rating,
       total_matches, wins, losses, draws, preferred_agents, preferred_languages, created_at, updated_at
FROM users WHERE email = $1;

-- name: GetUserByUsername :one
SELECT id, username, email, password_hash, avatar_url, rating, rating_deviation, rating_volatility, peak_rating,
       total_matches, wins, losses, draws, preferred_agents, preferred_languages, created_at, updated_at
FROM users WHERE username = $1;

-- name: UpdateUser :exec
UPDATE users SET
    username = $2, email = $3, avatar_url = $4, rating = $5,
    rating_deviation = $6, rating_volatility = $7, peak_rating = $8,
    total_matches = $9, wins = $10, losses = $11, draws = $12,
    preferred_agents = $13, preferred_languages = $14, updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserRating :exec
UPDATE users SET
    rating = $2, rating_deviation = $3, rating_volatility = $4,
    peak_rating = GREATEST(peak_rating, $2), updated_at = NOW()
WHERE id = $1;

-- name: IncrementMatchStats :exec
UPDATE users SET
    total_matches = total_matches + 1,
    wins = wins + $2,
    losses = losses + $3,
    draws = draws + $4,
    updated_at = NOW()
WHERE id = $1;

-- name: ListUsersByRating :many
SELECT id, username, email, avatar_url, rating, rating_deviation, peak_rating,
       total_matches, wins, losses, draws, last_active
FROM users
ORDER BY rating DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;