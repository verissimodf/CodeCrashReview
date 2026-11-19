-- name: GetGlobalLeaderboard :many
SELECT u.id as user_id, u.username, u.avatar_url, u.rating, u.rating_deviation, u.peak_rating,
       u.total_matches, u.wins, u.losses,
       CASE WHEN u.total_matches > 0 THEN (u.wins::float / u.total_matches) * 100 ELSE 0 END as win_rate,
       u.preferred_agents, u.preferred_languages, u.updated_at as last_active
FROM users u
WHERE u.total_matches > 0
ORDER BY u.rating DESC, u.rating_deviation ASC
LIMIT $1 OFFSET $2;

-- name: GetWeeklyLeaderboard :many
SELECT u.id as user_id, u.username, u.avatar_url, u.rating, u.rating_deviation, u.peak_rating,
       u.total_matches, u.wins, u.losses,
       CASE WHEN u.total_matches > 0 THEN (u.wins::float / u.total_matches) * 100 ELSE 0 END as win_rate,
       u.preferred_agents, u.preferred_languages, u.updated_at as last_active
FROM users u
WHERE u.total_matches > 0
  AND u.updated_at > NOW() - INTERVAL '7 days'
ORDER BY u.rating DESC, u.rating_deviation ASC
LIMIT $1 OFFSET $2;

-- name: GetLeaderboardByLanguage :many
SELECT u.id as user_id, u.username, u.avatar_url, u.rating, u.rating_deviation, u.peak_rating,
       u.total_matches, u.wins, u.losses,
       CASE WHEN u.total_matches > 0 THEN (u.wins::float / u.total_matches) * 100 ELSE 0 END as win_rate,
       u.preferred_agents, u.preferred_languages, u.updated_at as last_active
FROM users u
JOIN match_participants mp ON mp.user_id = u.id
JOIN matches m ON m.id = mp.match_id
JOIN challenges c ON c.id = m.challenge_id
WHERE u.total_matches > 0
  AND c.category = $1  -- Using category as proxy for language, could add language column
ORDER BY u.rating DESC, u.rating_deviation ASC
LIMIT $2 OFFSET $3;

-- name: GetLeaderboardByAgent :many
SELECT u.id as user_id, u.username, u.avatar_url, u.rating, u.rating_deviation, u.peak_rating,
       u.total_matches, u.wins, u.losses,
       CASE WHEN u.total_matches > 0 THEN (u.wins::float / u.total_matches) * 100 ELSE 0 END as win_rate,
       u.preferred_agents, u.preferred_languages, u.updated_at as last_active
FROM users u
JOIN match_participants mp ON mp.user_id = u.id
WHERE u.total_matches > 0
  AND mp.agent_persona = $1
ORDER BY u.rating DESC, u.rating_deviation ASC
LIMIT $2 OFFSET $3;

-- name: GetUserRatingHistory :many
SELECT rh.id, rh.user_id, rh.match_id, rh.rating_before, rh.rating_after, rh.rating_change,
       rh.deviation_before, rh.deviation_after, rh.created_at,
       m.mode, m.challenge_id
FROM rating_history rh
JOIN matches m ON m.id = rh.match_id
WHERE rh.user_id = $1
ORDER BY rh.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetAgentStats :many
SELECT mp.agent_persona,
       COUNT(*) as total_matches,
       SUM(CASE WHEN m.winner_persona = mp.agent_persona THEN 1 ELSE 0 END) as wins,
       SUM(CASE WHEN m.winner_persona IS NOT NULL AND m.winner_persona != mp.agent_persona THEN 1 ELSE 0 END) as losses,
       SUM(CASE WHEN m.winner_id IS NULL THEN 1 ELSE 0 END) as draws,
       CASE WHEN COUNT(*) > 0 THEN (SUM(CASE WHEN m.winner_persona = mp.agent_persona THEN 1 ELSE 0 END)::float / COUNT(*)) * 100 ELSE 0 END as win_rate,
       AVG(mp.score) as average_score,
       AVG(m.current_turn) as average_turns,
       AVG(EXTRACT(EPOCH FROM (m.ended_at - m.started_at))) as average_time_seconds
FROM match_participants mp
JOIN matches m ON m.id = mp.match_id
WHERE m.status = 'completed'
GROUP BY mp.agent_persona;

-- name: GetAgentMatchupStats :many
SELECT mp1.agent_persona as agent_a,
       mp2.agent_persona as agent_b,
       COUNT(*) as total_matches,
       SUM(CASE WHEN m.winner_persona = mp1.agent_persona THEN 1 ELSE 0 END) as agent_a_wins,
       SUM(CASE WHEN m.winner_persona = mp2.agent_persona THEN 1 ELSE 0 END) as agent_b_wins
FROM match_participants mp1
JOIN match_participants mp2 ON mp2.match_id = mp1.match_id AND mp2.user_id != mp1.user_id
JOIN matches m ON m.id = mp1.match_id
WHERE m.status = 'completed'
  AND mp1.agent_persona < mp2.agent_persona
GROUP BY mp1.agent_persona, mp2.agent_persona;