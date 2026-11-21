-- +goose Up
CREATE TABLE review_publications (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 review_id uuid NOT NULL REFERENCES repository_reviews(id),
 user_id uuid NOT NULL REFERENCES users(id),
 branch text NOT NULL,
 request jsonb NOT NULL,
 status text NOT NULL DEFAULT 'running',
 commit_sha text NOT NULL DEFAULT '',
 receipt jsonb NOT NULL DEFAULT '{}'::jsonb,
 error_message text,
 lease_expires_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(review_id,branch)
);

-- +goose Down
SELECT 1;
