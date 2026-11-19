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