-- +goose Up
-- Reserve the built-in challenge identifiers for ranked expression matches.
CREATE TABLE arena_challenges (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    starter_code TEXT NOT NULL,
    cases JSONB NOT NULL CHECK (jsonb_typeof(cases) = 'array' AND jsonb_array_length(cases) > 0)
);
INSERT INTO arena_challenges VALUES
('ec000000-0000-4000-8000-000000000001','Distance','Return the nonnegative distance between integers a and b.',E'def solve(a, b):\n    return 0','[[1,4,3],[5,1,4],[-3,-8,5],[0,0,0],[-7,4,11],[42,42,0]]'),
('ec000000-0000-4000-8000-000000000002','Rectangle perimeter','Return the perimeter of a rectangle with nonnegative side lengths a and b.',E'def solve(a, b):\n    return 0','[[1,4,10],[5,1,12],[3,8,22],[0,0,0],[7,4,22],[42,42,168]]'),
('ec000000-0000-4000-8000-000000000003','Largest square','Return the larger of the squares of a and b. Either integer may be negative.',E'def solve(a, b):\n    return 0','[[1,4,16],[5,1,25],[-3,-8,64],[0,0,0],[-7,4,49],[42,42,1764]]');
INSERT INTO users(id,username,email,password_hash) VALUES
('ec000000-0000-4000-8000-0000000000ff','Arena AI','arena-ai@codecrash.invalid','disabled');
CREATE TABLE arena_requests (
    user_id UUID NOT NULL REFERENCES users(id),
    request_key VARCHAR(100) NOT NULL,
    fingerprint TEXT NOT NULL,
    match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    PRIMARY KEY(user_id, request_key)
);
-- +goose Down
DROP TABLE arena_requests;
DELETE FROM matches WHERE mode='practice' AND challenge_id IN (SELECT id FROM arena_challenges);
DELETE FROM users WHERE id='ec000000-0000-4000-8000-0000000000ff';
DROP TABLE arena_challenges;
