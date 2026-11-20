-- +goose Up
CREATE TABLE repository_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_type VARCHAR(10) NOT NULL CHECK (source_type IN ('github', 'local')),
    source_location TEXT NOT NULL,
    source_branch VARCHAR(200) NOT NULL DEFAULT 'main',
    focus_paths JSONB NOT NULL DEFAULT '[]',
    objective TEXT NOT NULL,
    agent_a VARCHAR(20) NOT NULL REFERENCES agent_personas(id),
    agent_b VARCHAR(20) NOT NULL REFERENCES agent_personas(id),
    evaluator_mode VARCHAR(10) NOT NULL CHECK (evaluator_mode IN ('ai', 'human')),
    status VARCHAR(24) NOT NULL DEFAULT 'queued' CHECK (status IN (
        'queued', 'acquiring', 'analyzing', 'evaluating',
        'awaiting_human', 'completed', 'failed'
    )),
    snapshot_sha VARCHAR(64),
    snapshot_files INTEGER NOT NULL DEFAULT 0,
    snapshot_bytes BIGINT NOT NULL DEFAULT 0,
    attempts INTEGER NOT NULL DEFAULT 0,
    lease_expires_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (agent_a <> agent_b)
);

CREATE INDEX idx_repository_reviews_user_created
    ON repository_reviews(user_id, created_at DESC);
CREATE INDEX idx_repository_reviews_worker
    ON repository_reviews(status, lease_expires_at, created_at);

CREATE TABLE review_proposals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL REFERENCES repository_reviews(id) ON DELETE CASCADE,
    slot CHAR(1) NOT NULL CHECK (slot IN ('A', 'B')),
    persona VARCHAR(20) NOT NULL REFERENCES agent_personas(id),
    model TEXT NOT NULL,
    status VARCHAR(16) NOT NULL CHECK (status IN ('running', 'completed', 'failed')),
    summary TEXT,
    findings JSONB NOT NULL DEFAULT '[]',
    patch TEXT NOT NULL DEFAULT '',
    patch_valid BOOLEAN NOT NULL DEFAULT FALSE,
    validation_error TEXT,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (confidence >= 0 AND confidence <= 1),
    tokens_used INTEGER NOT NULL DEFAULT 0,
    latency_ms BIGINT NOT NULL DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    UNIQUE(review_id, slot)
);

CREATE TABLE review_verdicts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL UNIQUE REFERENCES repository_reviews(id) ON DELETE CASCADE,
    evaluator_type VARCHAR(10) NOT NULL CHECK (evaluator_type IN ('ai', 'human')),
    evaluator_user_id UUID REFERENCES users(id),
    winner VARCHAR(4) NOT NULL CHECK (winner IN ('A', 'B', 'TIE')),
    rationale TEXT NOT NULL,
    scores JSONB NOT NULL DEFAULT '{}',
    model TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_repository_reviews_updated_at
    BEFORE UPDATE ON repository_reviews
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER update_repository_reviews_updated_at ON repository_reviews;
DROP TABLE review_verdicts;
DROP TABLE review_proposals;
DROP TABLE repository_reviews;
