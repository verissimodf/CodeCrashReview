-- +goose Up
-- CodeCrashReview Database Schema
-- Migration: 000001_initial_schema.sql
-- Description: Initial database schema for users, matches, challenges, snippets

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    avatar_url TEXT,
    rating INTEGER NOT NULL DEFAULT 1200,
    rating_deviation DOUBLE PRECISION NOT NULL DEFAULT 350.0,
    rating_volatility DOUBLE PRECISION NOT NULL DEFAULT 0.06,
    peak_rating INTEGER NOT NULL DEFAULT 1200,
    total_matches INTEGER NOT NULL DEFAULT 0,
    wins INTEGER NOT NULL DEFAULT 0,
    losses INTEGER NOT NULL DEFAULT 0,
    draws INTEGER NOT NULL DEFAULT 0,
    preferred_agents JSONB NOT NULL DEFAULT '[]',
    preferred_languages JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_rating ON users(rating DESC);
CREATE INDEX idx_users_rating_deviation ON users(rating_deviation);

-- Matches table
CREATE TABLE matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mode VARCHAR(20) NOT NULL CHECK (mode IN ('ranked', 'private', 'tournament', 'practice')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'matching', 'preparing', 'active', 'paused', 'completed', 'cancelled', 'error')),
    challenge_id UUID NOT NULL,
    current_turn INTEGER NOT NULL DEFAULT 0,
    max_turns INTEGER NOT NULL DEFAULT 10,
    time_limit_seconds INTEGER NOT NULL DEFAULT 300,
    time_remaining_seconds INTEGER NOT NULL DEFAULT 300,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    winner_id UUID REFERENCES users(id),
    winner_persona VARCHAR(20),
    scoring JSONB NOT NULL DEFAULT '{}',
    replay_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for matches
CREATE INDEX idx_matches_status ON matches(status);
CREATE INDEX idx_matches_mode ON matches(mode);
CREATE INDEX idx_matches_challenge_id ON matches(challenge_id);
CREATE INDEX idx_matches_created_at ON matches(created_at DESC);
CREATE INDEX idx_matches_winner_id ON matches(winner_id);

-- Match participants table
CREATE TABLE match_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    agent_persona VARCHAR(20) NOT NULL CHECK (agent_persona IN ('sentinel', 'optimizer', 'refactor', 'guardian', 'docubot', 'polyglot')),
    rating_before INTEGER NOT NULL,
    rating_after INTEGER,
    rating_change INTEGER,
    score DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    status VARCHAR(20) NOT NULL DEFAULT 'waiting' CHECK (status IN ('waiting', 'thinking', 'editing', 'testing', 'submitting', 'completed', 'left')),
    connected BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(match_id, user_id)
);

-- Indexes for match_participants
CREATE INDEX idx_match_participants_match_id ON match_participants(match_id);
CREATE INDEX idx_match_participants_user_id ON match_participants(user_id);
CREATE INDEX idx_match_participants_agent_persona ON match_participants(agent_persona);

-- Challenge templates table
CREATE TABLE challenge_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    category VARCHAR(20) NOT NULL CHECK (category IN ('security', 'performance', 'refactor', 'testing', 'documentation', 'migration', 'bugfix', 'feature')),
    default_severity VARCHAR(20) NOT NULL CHECK (default_severity IN ('low', 'medium', 'high', 'critical')),
    default_time_limit_seconds INTEGER NOT NULL DEFAULT 300,
    default_max_turns INTEGER NOT NULL DEFAULT 10,
    acceptance_criteria_template JSONB NOT NULL DEFAULT '[]',
    scoring_weights_template JSONB NOT NULL DEFAULT '{}',
    known_issues_template JSONB NOT NULL DEFAULT '[]',
    example_repository TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Challenges table
CREATE TABLE challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    repository_url TEXT NOT NULL,
    repository_branch VARCHAR(100) NOT NULL DEFAULT 'main',
    target_paths TEXT[] NOT NULL DEFAULT '{}',
    exclude_paths TEXT[] NOT NULL DEFAULT '{}',
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    category VARCHAR(20) NOT NULL CHECK (category IN ('security', 'performance', 'refactor', 'testing', 'documentation', 'migration', 'bugfix', 'feature')),
    time_limit_seconds INTEGER NOT NULL DEFAULT 300,
    max_turns INTEGER NOT NULL DEFAULT 10,
    acceptance_criteria JSONB NOT NULL DEFAULT '[]',
    scoring_weights JSONB NOT NULL DEFAULT '{}',
    known_issues JSONB NOT NULL DEFAULT '[]',
    tags TEXT[] NOT NULL DEFAULT '{}',
    template_id UUID REFERENCES challenge_templates(id),
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    created_by UUID NOT NULL REFERENCES users(id),
    play_count INTEGER NOT NULL DEFAULT 0,
    average_rating DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for challenges
CREATE INDEX idx_challenges_category ON challenges(category);
CREATE INDEX idx_challenges_severity ON challenges(severity);
CREATE INDEX idx_challenges_is_public ON challenges(is_public);
CREATE INDEX idx_challenges_created_by ON challenges(created_by);
CREATE INDEX idx_challenges_tags ON challenges USING GIN(tags);
CREATE INDEX idx_challenges_created_at ON challenges(created_at DESC);

-- Snippets table
CREATE TABLE snippets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    language VARCHAR(20) NOT NULL CHECK (language IN ('cpp', 'python', 'java', 'javascript', 'typescript', 'rust', 'shell', 'html', 'gherkin')),
    code TEXT NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}',
    complexity VARCHAR(20) NOT NULL DEFAULT 'intermediate' CHECK (complexity IN ('basic', 'intermediate', 'advanced')),
    author VARCHAR(100) NOT NULL,
    author_id UUID REFERENCES users(id),
    source_url TEXT,
    rating DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    usage_count INTEGER NOT NULL DEFAULT 0,
    is_official BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for snippets
CREATE INDEX idx_snippets_language ON snippets(language);
CREATE INDEX idx_snippets_tags ON snippets USING GIN(tags);
CREATE INDEX idx_snippets_complexity ON snippets(complexity);
CREATE INDEX idx_snippets_rating ON snippets(rating DESC);
CREATE INDEX idx_snippets_is_official ON snippets(is_official);
CREATE INDEX idx_snippets_author_id ON snippets(author_id);

-- Snippet collections table
CREATE TABLE snippet_collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    snippet_ids UUID[] NOT NULL DEFAULT '{}',
    tags TEXT[] NOT NULL DEFAULT '{}',
    is_curated BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Submissions table
CREATE TABLE submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    participant_id UUID NOT NULL REFERENCES match_participants(id) ON DELETE CASCADE,
    agent_persona VARCHAR(20) NOT NULL CHECK (agent_persona IN ('sentinel', 'optimizer', 'refactor', 'guardian', 'docubot', 'polyglot')),
    turn INTEGER NOT NULL,
    files JSONB NOT NULL DEFAULT '[]',
    changes JSONB NOT NULL DEFAULT '[]',
    test_result JSONB,
    execution_result JSONB,
    score DOUBLE PRECISION,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'passed', 'failed', 'timeout', 'error')),
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    evaluated_at TIMESTAMPTZ,
    feedback TEXT
);

-- Indexes for submissions
CREATE INDEX idx_submissions_match_id ON submissions(match_id);
CREATE INDEX idx_submissions_participant_id ON submissions(participant_id);
CREATE INDEX idx_submissions_turn ON submissions(turn);

-- Agent personas table
CREATE TABLE agent_personas (
    id VARCHAR(20) PRIMARY KEY CHECK (id IN ('sentinel', 'optimizer', 'refactor', 'guardian', 'docubot', 'polyglot')),
    name VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    specialty VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    temperature DOUBLE PRECISION NOT NULL DEFAULT 0.3,
    tools TEXT[] NOT NULL DEFAULT '{}',
    system_prompt TEXT NOT NULL,
    color VARCHAR(7) NOT NULL,
    glow_color VARCHAR(9) NOT NULL,
    icon VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Rating history table
CREATE TABLE rating_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    rating_before INTEGER NOT NULL,
    rating_after INTEGER NOT NULL,
    rating_change INTEGER NOT NULL,
    deviation_before DOUBLE PRECISION NOT NULL,
    deviation_after DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for rating_history
CREATE INDEX idx_rating_history_user_id ON rating_history(user_id);
CREATE INDEX idx_rating_history_match_id ON rating_history(match_id);
CREATE INDEX idx_rating_history_created_at ON rating_history(created_at DESC);

-- Updated at trigger function
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Apply updated_at triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_matches_updated_at BEFORE UPDATE ON matches FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_challenges_updated_at BEFORE UPDATE ON challenges FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_challenge_templates_updated_at BEFORE UPDATE ON challenge_templates FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_snippets_updated_at BEFORE UPDATE ON snippets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_agent_personas_updated_at BEFORE UPDATE ON agent_personas FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default agent personas
INSERT INTO agent_personas (id, name, description, specialty, model, temperature, tools, system_prompt, color, glow_color, icon) VALUES
('sentinel', 'Sentinel', 'Security & Vulnerabilities Specialist', 'security', 'deepseek-coder:33b', 0.2,
 ARRAY['semgrep', 'bandit', 'gosec', 'trivy', 'sast'],
 'You are Sentinel, an elite security-focused code agent. Your mission is to identify and fix security vulnerabilities in code. CORE PRINCIPLES: 1. Security FIRST - Every change must improve security posture 2. Zero tolerance for vulnerabilities - SQL injection, XSS, path traversal, etc. 3. Defense in depth - Apply multiple layers of protection 4. Minimal attack surface - Remove unnecessary code, permissions, exposures 5. Secure by default - Fail closed, validate everything',
 '#ff3366', '#ff336640', 'shield-alert'),

('optimizer', 'Optimizer', 'Performance & Complexity Specialist', 'performance', 'codellama:34b', 0.3,
 ARRAY['benchmarks', 'profilers', 'complexity_analyzer', 'memory_profiler', 'cpu_profiler'],
 'You are Optimizer, a performance-obsessed code agent. Your mission is to make code faster, more efficient, and scalable. CORE PRINCIPLES: 1. Measure before optimizing - Profile first, optimize second 2. Algorithmic complexity matters - O(n) > O(n²) always 3. Memory efficiency - Minimize allocations, reuse objects 4. Cache-friendly - Locality of reference, avoid cache misses 5. Concurrency done right - Lock-free when possible, minimal contention',
 '#00ff87', '#00ff8740', 'zap'),

('refactor', 'Refactor', 'Clean Code & Architecture Specialist', 'refactor', 'qwen2.5-coder:32b', 0.25,
 ARRAY['ast_grep', 'refactoring_rules', 'design_patterns', 'solid_principles', 'coupling_analyzer'],
 'You are Refactor, a clean code and architecture specialist. Your mission is to transform messy code into elegant, maintainable software. CORE PRINCIPLES: 1. SOLID principles - Single responsibility, Open/closed, Liskov substitution, Interface segregation, Dependency inversion 2. DRY - Don''t repeat yourself, but don''t over-abstract 3. KISS - Keep it simple, stupid 4. YAGNI - You ain''t gonna need it 5. Clean architecture - Separation of concerns, dependency rule',
 '#0099ff', '#0099ff40', 'code-2'),

('guardian', 'Guardian', 'Testing & Coverage Specialist', 'testing', 'starcoder2:15b', 0.35,
 ARRAY['pytest', 'jest', 'coverage', 'mutation_testing', 'property_testing', 'contract_testing'],
 'You are Guardian, a testing and quality assurance specialist. Your mission is to ensure code correctness through comprehensive testing. CORE PRINCIPLES: 1. Test pyramid - Unit > Integration > E2E 2. Test behavior, not implementation 3. Edge cases first - Boundary conditions, error states 4. Deterministic tests - No flakiness 5. Fast feedback - Tests must run quickly 6. High coverage with meaningful assertions 7. Mutation testing - Verify test quality',
 '#ffaa00', '#ffaa0040', 'check-circle-2'),

('docubot', 'DocuBot', 'Documentation & Types Specialist', 'documentation', 'deepseek-coder:6.7b', 0.4,
 ARRAY['type_inference', 'doc_generator', 'api_doc_generator', 'readme_generator', 'changelog_generator'],
 'You are DocuBot, a documentation and type specialist. Your mission is to make code self-documenting and comprehensible. CORE PRINCIPLES: 1. Code as documentation - Types, names, structure tell the story 2. README first - Project purpose, setup, usage 3. API docs - Every public interface documented 4. Examples over explanations - Show, don''t just tell 5. Keep docs in sync - Documentation debt is technical debt 6. Type safety - Strong types prevent bugs',
 '#ff00ff', '#ff00ff40', 'book-open'),

('polyglot', 'Polyglot', 'Multi-language & Migration Specialist', 'migration', 'codellama:13b', 0.3,
 ARRAY['transpiler', 'cross_lang_patterns', 'ffi_generator', 'api_adapter', 'data_mapper'],
 'You are Polyglot, a multi-language and migration specialist. Your mission is to bridge languages and modernize codebases. CORE PRINCIPLES: 1. Idiomatic code - Write like a native in each language 2. Preserve semantics - Behavior must remain identical 3. Incremental migration - Strangler fig pattern 4. Interop layers - Clean boundaries between languages 5. Data consistency - Serialization, validation across boundaries 6. Performance awareness - FFI overhead, memory management',
 '#00ffff', '#00ffff40', 'globe')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    specialty = EXCLUDED.specialty,
    model = EXCLUDED.model,
    temperature = EXCLUDED.temperature,
    tools = EXCLUDED.tools,
    system_prompt = EXCLUDED.system_prompt,
    color = EXCLUDED.color,
    glow_color = EXCLUDED.glow_color,
    icon = EXCLUDED.icon,
    updated_at = NOW();