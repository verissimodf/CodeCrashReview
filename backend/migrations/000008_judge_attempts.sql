-- +goose Up
ALTER TABLE repository_reviews ADD COLUMN evaluation_attempts jsonb NOT NULL DEFAULT '[]'::jsonb;

-- +goose Down
-- Keep the audit history on application rollback.
SELECT 1;
