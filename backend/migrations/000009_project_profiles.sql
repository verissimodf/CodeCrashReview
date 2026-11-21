-- +goose Up
ALTER TABLE repository_reviews ADD COLUMN validation_profile text NOT NULL DEFAULT 'arena-grade-v1';
ALTER TABLE repository_reviews ADD COLUMN validation_directory text NOT NULL DEFAULT '.';

-- +goose Down
SELECT 1;
