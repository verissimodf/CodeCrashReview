-- +goose Up
ALTER TABLE review_proposals
    ADD COLUMN verification JSONB NOT NULL DEFAULT '{}'::jsonb;

-- +goose Down
ALTER TABLE review_proposals DROP COLUMN verification;
