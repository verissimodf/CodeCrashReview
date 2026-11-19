-- name: CreateChallenge :one
INSERT INTO challenges (title, description, repository_url, repository_branch, target_paths, exclude_paths, severity, category, time_limit_seconds, max_turns, acceptance_criteria, scoring_weights, known_issues, tags, template_id, is_public, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING id, created_at, updated_at;

-- name: GetChallengeByID :one
SELECT id, title, description, repository_url, repository_branch, target_paths, exclude_paths, severity, category,
       time_limit_seconds, max_turns, acceptance_criteria, scoring_weights, known_issues, tags,
       template_id, is_public, created_by, play_count, average_rating, created_at, updated_at
FROM challenges WHERE id = $1;

-- name: ListChallenges :many
SELECT id, title, description, repository_url, repository_branch, target_paths, exclude_paths, severity, category,
       time_limit_seconds, max_turns, acceptance_criteria, scoring_weights, known_issues, tags,
       template_id, is_public, created_by, play_count, average_rating, created_at, updated_at
FROM challenges
WHERE is_public = TRUE
  AND ($1::text IS NULL OR category = $1)
  AND ($2::text IS NULL OR severity = $2)
  AND ($3::text IS NULL OR title ILIKE '%' || $3 || '%' OR description ILIKE '%' || $3 || '%')
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountChallenges :one
SELECT COUNT(*) FROM challenges
WHERE is_public = TRUE
  AND ($1::text IS NULL OR category = $1)
  AND ($2::text IS NULL OR severity = $2)
  AND ($3::text IS NULL OR title ILIKE '%' || $3 || '%' OR description ILIKE '%' || $3 || '%');

-- name: UpdateChallenge :exec
UPDATE challenges SET
    title = $2, description = $3, repository_url = $4, repository_branch = $5,
    target_paths = $6, exclude_paths = $7, severity = $8, category = $9,
    time_limit_seconds = $10, max_turns = $11, acceptance_criteria = $12,
    scoring_weights = $13, known_issues = $14, tags = $15, template_id = $16,
    is_public = $17, updated_at = NOW()
WHERE id = $1;

-- name: DeleteChallenge :exec
DELETE FROM challenges WHERE id = $1;

-- name: IncrementChallengePlayCount :exec
UPDATE challenges SET play_count = play_count + 1 WHERE id = $1;

-- name: CreateChallengeTemplate :one
INSERT INTO challenge_templates (name, description, category, default_severity, default_time_limit_seconds, default_max_turns, acceptance_criteria_template, scoring_weights_template, known_issues_template, example_repository)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, created_at, updated_at;

-- name: GetChallengeTemplateByID :one
SELECT id, name, description, category, default_severity, default_time_limit_seconds, default_max_turns,
       acceptance_criteria_template, scoring_weights_template, known_issues_template, example_repository,
       created_at, updated_at
FROM challenge_templates WHERE id = $1;

-- name: ListChallengeTemplates :many
SELECT id, name, description, category, default_severity, default_time_limit_seconds, default_max_turns,
       acceptance_criteria_template, scoring_weights_template, known_issues_template, example_repository,
       created_at, updated_at
FROM challenge_templates
WHERE ($1::text IS NULL OR category = $1)
ORDER BY name;