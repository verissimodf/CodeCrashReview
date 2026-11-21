-- +goose Up
CREATE TABLE review_deployments (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 review_id UUID NOT NULL REFERENCES repository_reviews(id),
 user_id UUID NOT NULL REFERENCES users(id),
 target TEXT NOT NULL,
 status TEXT NOT NULL,
 receipt JSONB NOT NULL DEFAULT '{}',
 error_message TEXT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 completed_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX review_deployments_target_running ON review_deployments(target) WHERE status='running';
-- +goose Down
-- Audit receipts retained intentionally.
