-- name: CreateSnippet :one
INSERT INTO snippets (title, description, language, code, tags, complexity, author, author_id, source_url, is_official)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, created_at, updated_at;

-- name: GetSnippetByID :one
SELECT id, title, description, language, code, tags, complexity, author, author_id, source_url, rating, usage_count, is_official, created_at, updated_at
FROM snippets WHERE id = $1;

-- name: ListSnippets :many
SELECT id, title, description, language, code, tags, complexity, author, author_id, source_url, rating, usage_count, is_official, created_at, updated_at
FROM snippets
WHERE ($1::text IS NULL OR language = $1)
  AND ($2::text IS NULL OR $2 = ANY(tags))
  AND ($3::text IS NULL OR complexity = $3)
  AND ($4::text IS NULL OR title ILIKE '%' || $4 || '%' OR description ILIKE '%' || $4 || '%')
ORDER BY rating DESC, usage_count DESC
LIMIT $5 OFFSET $6;

-- name: CountSnippets :one
SELECT COUNT(*) FROM snippets
WHERE ($1::text IS NULL OR language = $1)
  AND ($2::text IS NULL OR $2 = ANY(tags))
  AND ($3::text IS NULL OR complexity = $3)
  AND ($4::text IS NULL OR title ILIKE '%' || $4 || '%' OR description ILIKE '%' || $4 || '%');

-- name: UpdateSnippet :exec
UPDATE snippets SET
    title = $2, description = $3, language = $4, code = $5,
    tags = $6, complexity = $7, author = $8, author_id = $9,
    source_url = $10, rating = $11, usage_count = $12, is_official = $13, updated_at = NOW()
WHERE id = $1;

-- name: DeleteSnippet :exec
DELETE FROM snippets WHERE id = $1;

-- name: IncrementSnippetUsage :exec
UPDATE snippets SET usage_count = usage_count + 1 WHERE id = $1;

-- name: UpdateSnippetRating :exec
UPDATE snippets SET rating = $2 WHERE id = $1;

-- name: SearchSnippets :many
SELECT id, title, description, language, code, tags, complexity, author, author_id, source_url, rating, usage_count, is_official, created_at, updated_at
FROM snippets
WHERE code ILIKE '%' || $1 || '%'
   OR title ILIKE '%' || $1 || '%'
   OR description ILIKE '%' || $1 || '%'
ORDER BY rating DESC, usage_count DESC
LIMIT $2 OFFSET $3;