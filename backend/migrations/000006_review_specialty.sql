-- +goose Up
ALTER TABLE repository_reviews
    ADD COLUMN review_focus VARCHAR(20) NOT NULL DEFAULT 'comprehensive'
    CHECK (review_focus IN ('comprehensive','security','reliability','refactoring','performance','architecture','documentation'));

-- +goose Down
ALTER TABLE repository_reviews DROP COLUMN review_focus;
