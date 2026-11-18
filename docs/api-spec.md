# CodeCrashReview - API Specification

## Base URL

```
Development:  http://localhost:8080/api
Production:   https://api.codecrashreview.com/api
WebSocket:    ws://localhost:8080/ws (dev) / wss://api.codecrashreview.com/ws (prod)
```

## Authentication

All protected endpoints require JWT Bearer token:

```
Authorization: Bearer <access_token>
```

### Token Refresh

```
POST /api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "<refresh_token>"
}
```

Response:
```json
{
  "access_token": "new_access_token",
  "refresh_token": "new_refresh_token",
  "expires_in": 86400,
  "token_type": "Bearer"
}
```

---

## Auth Endpoints

### Register User

```
POST /api/auth/register
Content-Type: application/json

{
  "username": "string (3-20 chars, alphanumeric + _ -)",
  "email": "string (valid email)",
  "password": "string (min 8 chars, uppercase, lowercase, number, special)"
}
```

Response (201):
```json
{
  "success": true,
  "data": {
    "user": { ...User object... },
    "tokens": {
      "access_token": "jwt_token",
      "refresh_token": "jwt_token",
      "expires_in": 86400,
      "token_type": "Bearer"
    }
  }
}
```

### Login

```
POST /api/auth/login
Content-Type: application/json

{
  "email": "string",
  "password": "string"
}
```

Response (200): Same as register.

---

## User Endpoints

### Get Current User

```
GET /api/users/me
Authorization: Bearer <token>
```

Response (200):
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "username": "string",
    "email": "string",
    "avatar_url": "string|null",
    "rating": 1200,
    "rating_deviation": 350.0,
    "rating_volatility": 0.06,
    "peak_rating": 1200,
    "total_matches": 0,
    "wins": 0,
    "losses": 0,
    "draws": 0,
    "preferred_agents": ["sentinel", "optimizer"],
    "preferred_languages": ["python", "javascript"],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Update Current User

```
PATCH /api/users/me
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "string",
  "avatar_url": "string",
  "preferred_agents": ["sentinel", "optimizer"],
  "preferred_languages": ["python", "javascript"]
}
```

### Get User by ID

```
GET /api/users/:id
```

### Get My Stats

```
GET /api/stats/me
Authorization: Bearer <token>
```

Response (200):
```json
{
  "success": true,
  "data": {
    "user": { ...User object... },
    "by_agent": {
      "sentinel": {
        "persona": "sentinel",
        "total_matches": 10,
        "wins": 7,
        "losses": 3,
        "draws": 0,
        "win_rate": 70.0,
        "average_score": 85.5,
        "average_turns": 6.2,
        "average_time_seconds": 180,
        "favorite_language": "python",
        "strongest_against": "docubot",
        "weakest_against": "polyglot",
        "rating_distribution": {"1200-1400": 3, "1400-1600": 4, "1600-1800": 3}
      }
      // ... other agents
    },
    "by_language": {
      "python": {"matches": 15, "wins": 10, "win_rate": 66.7},
      "javascript": {"matches": 8, "wins": 5, "win_rate": 62.5}
    },
    "by_category": {
      "security": {"matches": 10, "wins": 7, "win_rate": 70.0},
      "performance": {"matches": 8, "wins": 5, "win_rate": 62.5}
    },
    "recent_matches": [
      {
        "id": "uuid",
        "mode": "ranked",
        "challenge_title": "Fix SQL Injection",
        "opponent": "player2",
        "opponent_agent": "optimizer",
        "my_agent": "sentinel",
        "result": "win",
        "rating_change": +25,
        "score": 92.5,
        "opponent_score": 78.0,
        "duration_seconds": 180,
        "played_at": "2024-01-15T10:30:00Z"
      }
    ],
    "streaks": {
      "current": 3,
      "longest_win": 5,
      "longest_loss": 2
    },
    "achievements": [
      {
        "id": "first_win",
        "name": "First Blood",
        "description": "Win your first match",
        "icon": "trophy",
        "color": "#ffd700",
        "unlocked_at": "2024-01-10T15:00:00Z"
      }
    ]
  }
}
```

### Get User Stats

```
GET /api/users/:id/stats
```

---

## Match Endpoints

### Join Ranked Queue

```
POST /api/matches/queue
Authorization: Bearer <token>
Content-Type: application/json

{
  "mode": "ranked",
  "preferred_agent": "sentinel",
  "language": "python"
}
```

Response (200):
```json
{
  "success": true,
  "data": {
    "in_queue": true,
    "queue_key": "ranked:python:1200",
    "position": 1,
    "estimated_wait_seconds": 30
  }
}
```

### Leave Queue

```
DELETE /api/matches/queue
Authorization: Bearer <token>
```

### Get Queue Status

```
GET /api/matches/queue/status
Authorization: Bearer <token>
```

Response (200):
```json
{
  "success": true,
  "data": {
    "in_queue": true,
    "queue_key": "ranked:python:1200",
    "position": 1,
    "players_in_queue": 5,
    "estimated_wait_seconds": 30
  }
}
```

### Create Private Match

```
POST /api/matches
Authorization: Bearer <token>
Content-Type: application/json

{
  "mode": "private",
  "challenge_id": "uuid",
  "opponent_id": "uuid"
}
```

### List Matches

```
GET /api/matches?mode=ranked&limit=20&offset=0
Authorization: Bearer <token>
```

Query Parameters:
- `mode`: ranked, private, tournament, practice
- `limit`: 1-100 (default 20)
- `offset`: >= 0 (default 0)

Response (200):
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "mode": "ranked",
      "status": "completed",
      "challenge_id": "uuid",
      "current_turn": 10,
      "max_turns": 10,
      "time_limit_seconds": 300,
      "time_remaining_seconds": 0,
      "started_at": "2024-01-15T10:30:00Z",
      "ended_at": "2024-01-15T10:35:00Z",
      "winner_id": "uuid",
      "winner_persona": "sentinel",
      "scoring": { ... },
      "replay_url": "/replay/uuid",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:35:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 50,
    "total_pages": 3
  }
}
```

### Get Match Details

```
GET /api/matches/:id
Authorization: Bearer <token>
```

Response (200):
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "mode": "ranked",
    "status": "completed",
    "challenge": { ...Challenge object... },
    "participants": [
      {
        "user_id": "uuid",
        "username": "player1",
        "avatar_url": "url",
        "agent_persona": "sentinel",
        "rating_before": 1200,
        "rating_after": 1225,
        "rating_change": 25,
        "score": 92.5,
        "submissions": [
          {
            "id": "uuid",
            "turn": 1,
            "files": [...],
            "changes": [...],
            "test_result": { "passed": 8, "failed": 0, ... },
            "execution_result": { "exit_code": 0, "stdout": "...", "duration_ms": 150 },
            "score": 95.0,
            "status": "passed",
            "submitted_at": "2024-01-15T10:31:00Z"
          }
        ],
        "status": "completed",
        "connected": true
      }
    ],
    "spectators": [],
    "current_turn": 10,
    "max_turns": 10,
    "time_limit_seconds": 300,
    "time_remaining_seconds": 0,
    "started_at": "2024-01-15T10:30:00Z",
    "ended_at": "2024-01-15T10:35:00Z",
    "winner_id": "uuid",
    "winner_persona": "sentinel",
    "scoring": {
      "criteria": [
        { "id": "correctness", "name": "Correctness", "weight": 0.4, "max_score": 100, "category": "correctness" },
        { "id": "security", "name": "Security", "weight": 0.3, "max_score": 100, "category": "security" }
      ],
      "weights": { "correctness": 0.4, "security": 0.3, ... },
      "agent_scores": {
        "sentinel": [
          { "criterion_id": "correctness", "score": 95, "max_score": 100, "details": "All tests pass", "evidence": ["test_output"] }
        ]
      },
      "final_scores": { "sentinel": 92.5, "optimizer": 78.0 }
    },
    "replay_url": "/replay/uuid",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:35:00Z"
  }
}
```

### Join Match

```
POST /api/matches/:id/join
Authorization: Bearer <token>
```

### Leave Match

```
POST /api/matches/:id/leave
Authorization: Bearer <token>
```

---

## Challenge Endpoints

### Create Challenge

```
POST /api/challenges
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Fix SQL Injection in User Service",
  "description": "UserService.getUser() vulnerable to SQLi via raw query",
  "repository_url": "https://github.com/org/target-system",
  "repository_branch": "main",
  "target_paths": ["services/user/**/*.go", "db/queries/**/*.sql"],
  "exclude_paths": ["vendor/**", "tests/**"],
  "severity": "critical",
  "category": "security",
  "time_limit_seconds": 300,
  "max_turns": 10,
  "acceptance_criteria": [
    {
      "id": "ac1",
      "description": "No raw string concatenation in SQL",
      "type": "static_analysis",
      "weight": 0.5,
      "pass_threshold": 1.0,
      "config": { "tool": "semgrep", "rules": ["sql-injection"] }
    },
    {
      "id": "ac2",
      "description": "Parameterized queries used",
      "type": "static_analysis",
      "weight": 0.3,
      "pass_threshold": 1.0
    },
    {
      "id": "ac3",
      "description": "Existing tests pass",
      "type": "test",
      "weight": 0.2,
      "pass_threshold": 1.0
    }
  ],
  "scoring_weights": {
    "correctness": 0.4,
    "security": 0.3,
    "performance": 0.15,
    "style": 0.1,
    "test_coverage": 0.05
  },
  "known_issues": [
    {
      "file": "services/user/repository.go",
      "line": 42,
      "column": 15,
      "issue": "Direct string interpolation in query",
      "severity": "critical",
      "priority": "high",
      "cwe_id": "CWE-89",
      "description": "User input directly concatenated into SQL query"
    }
  ],
  "tags": ["sql-injection", "go", "security"],
  "template_id": "security-fix",
  "is_public": true
}
```

Response (201): Returns created challenge with ID.

### List Challenges

```
GET /api/challenges?category=security&severity=critical&search=sql&limit=20&offset=0
Authorization: Bearer <token>
```

Query Parameters:
- `category`: security, performance, refactor, testing, documentation, migration, bugfix, feature
- `severity`: low, medium, high, critical
- `search`: text search in title/description
- `language`: filter by target language
- `limit`: 1-100 (default 20)
- `offset`: >= 0 (default 0)

Response (200): Paginated list of challenges.

### Get Challenge Details

```
GET /api/challenges/:id
Authorization: Bearer <token>
```

### Update Challenge

```
PATCH /api/challenges/:id
Authorization: Bearer <token>
Content-Type: application/json

{ ...partial challenge object... }
```

### Delete Challenge

```
DELETE /api/challenges/:id
Authorization: Bearer <token>
```

### List Challenge Templates

```
GET /api/challenges/templates?category=security
Authorization: Bearer <token>
```

---

## Snippet Endpoints

### List Snippets

```
GET /api/snippets?language=python&tag=algorithm&complexity=basic&search=binary&limit=20&offset=0
Authorization: Bearer <token>
```

### Get Snippet

```
GET /api/snippets/:id
Authorization: Bearer <token>
```

### Import Snippets

```
POST /api/snippets/import
Authorization: Bearer <token>
Content-Type: application/json

{
  "source": "awesome-algorithms"
}
```

Sources: `awesome-algorithms`, `cp-algorithms`, `programming-idioms`, `leetcode-patterns`

---

## Leaderboard Endpoints

### Global Leaderboard

```
GET /api/leaderboard/global?limit=50&offset=0
Authorization: Bearer <token>
```

### Weekly Leaderboard

```
GET /api/leaderboard/weekly?limit=50&offset=0
Authorization: Bearer <token>
```

### By Language

```
GET /api/leaderboard/language/:language?limit=50&offset=0
Authorization: Bearer <token>
```

### By Agent

```
GET /api/leaderboard/agent/:agent?limit=50&offset=0
Authorization: Bearer <token>
```

Response (200):
```json
{
  "success": true,
  "data": [
    {
      "rank": 1,
      "user_id": "uuid",
      "username": "top_player",
      "avatar_url": "url",
      "rating": 2450,
      "rating_deviation": 45.2,
      "peak_rating": 2480,
      "total_matches": 150,
      "wins": 105,
      "losses": 40,
      "win_rate": 70.0,
      "preferred_agent": "sentinel",
      "preferred_language": "python",
      "last_active": "2024-01-18T10:00:00Z",
      "trend": "up",
      "rating_history": [
        { "timestamp": "2024-01-01", "rating": 2400, "rating_deviation": 50.0, "match_id": "uuid", "change": +25 }
      ]
    }
  ],
  "meta": { ... }
}
```

---

## WebSocket API

### Connection

```
GET /ws?token=<access_token>&match=<match_id>&spectate=true
```

Headers:
- `Authorization: Bearer <token>` (alternative to query param)

### Message Format

All messages follow this structure:
```json
{
  "type": "message_type",
  "payload": { ... },
  "timestamp": "2024-01-15T10:30:00.000Z",
  "sequence": 1234567890
}
```

### Client → Server Messages

#### Auth (if not in query)
```json
{
  "type": "auth",
  "payload": { "token": "jwt_token" },
  "timestamp": "...",
  "sequence": 1
}
```

#### Join Match
```json
{
  "type": "join_match",
  "payload": { "match_id": "uuid", "as_spectator": false },
  "timestamp": "...",
  "sequence": 2
}
```

#### Leave Match
```json
{
  "type": "leave_match",
  "payload": {},
  "timestamp": "...",
  "sequence": 3
}
```

#### Agent Action (for agent clients)
```json
{
  "type": "agent_action",
  "payload": {
    "action": {
      "type": "think",
      "payload": { "thought": "Analyzing SQL injection vulnerability" },
      "timestamp": "...",
      "duration_ms": 500
    }
  },
  "timestamp": "...",
  "sequence": 4
}
```

#### Submission
```json
{
  "type": "submission",
  "payload": {
    "files": [{ "path": "services/user/repository.go", "language": "go", "content": "..." }],
    "changes": [{ "file_path": "...", "original_content": "...", "new_content": "...", "diff": [...], "description": "..." }],
    "turn": 1
  },
  "timestamp": "...",
  "sequence": 5
}
```

### Server → Client Messages

#### Auth Response
```json
{
  "type": "auth_response",
  "payload": {
    "success": true,
    "user": { ...User object... }
  },
  "timestamp": "...",
  "sequence": 1
}
```

#### Match State
```json
{
  "type": "match_state",
  "payload": {
    "match": { ...Match object... },
    "your_participant_id": "uuid"
  },
  "timestamp": "...",
  "sequence": 2
}
```

#### Match Update
```json
{
  "type": "match_update",
  "payload": {
    "match": { "status": "active", "current_turn": 1, "time_remaining_seconds": 300 },
    "participants": [
      { "user_id": "uuid", "status": "thinking", "score": 0 }
    ]
  },
  "timestamp": "...",
  "sequence": 3
}
```

#### Agent Status
```json
{
  "type": "agent_status",
  "payload": {
    "participant_id": "uuid",
    "agent_persona": "sentinel",
    "status": "thinking",
    "progress": 25,
    "current_file": "services/user/repository.go",
    "current_line": 42,
    "thoughts": ["Analyzing SQL query construction", "Found string concatenation"]
  },
  "timestamp": "...",
  "sequence": 4
}
```

#### Code Change
```json
{
  "type": "code_change",
  "payload": {
    "participant_id": "uuid",
    "agent_persona": "sentinel",
    "change": {
      "id": "uuid",
      "file_path": "services/user/repository.go",
      "language": "go",
      "original_content": "query := \"SELECT * FROM users WHERE id = \" + userID",
      "new_content": "query := \"SELECT * FROM users WHERE id = ?\"\nargs = []interface{}{userID}",
      "diff": [
        { "type": "remove", "content": "query := \"SELECT * FROM users WHERE id = \" + userID", "line_number": 42 },
        { "type": "add", "content": "query := \"SELECT * FROM users WHERE id = ?\"", "line_number": 42 },
        { "type": "add", "content": "args = []interface{}{userID}", "line_number": 43 }
      ],
      "description": "Fixed SQL injection by using parameterized query",
      "agent_persona": "sentinel",
      "turn": 1,
      "timestamp": "2024-01-15T10:31:00Z",
      "accepted": null
    }
  },
  "timestamp": "...",
  "sequence": 5
}
```

#### Turn Change
```json
{
  "type": "turn_change",
  "payload": {
    "turn": 2,
    "active_participant_id": "uuid",
    "time_remaining_seconds": 300
  },
  "timestamp": "...",
  "sequence": 6
}
```

#### Timer Update
```json
{
  "type": "timer_update",
  "payload": {
    "time_remaining_seconds": 295,
    "warning": false
  },
  "timestamp": "...",
  "sequence": 7
}
```

#### Submission Result
```json
{
  "type": "submission_result",
  "payload": {
    "submission_id": "uuid",
    "score": 95.0,
    "criterion_scores": [
      { "criterion_id": "correctness", "score": 95, "max_score": 100, "details": "All tests pass", "evidence": ["test_output"] }
    ],
    "feedback": "Excellent fix! Parameterized query prevents SQL injection."
  },
  "timestamp": "...",
  "sequence": 8
}
```

#### Test Result
```json
{
  "type": "test_result",
  "payload": {
    "participant_id": "uuid",
    "agent_persona": "sentinel",
    "test_result": {
      "passed": 8,
      "failed": 0,
      "skipped": 0,
      "total": 8,
      "duration_ms": 1250,
      "suites": [
        {
          "name": "UserRepositoryTest",
          "tests": [
            { "name": "TestGetUserByID", "status": "passed", "duration_ms": 150 },
            { "name": "TestGetUserByEmail", "status": "passed", "duration_ms": 120 }
          ],
          "duration_ms": 1250
        }
      ],
      "coverage": {
        "lines": { "total": 150, "covered": 120, "percentage": 80.0 },
        "functions": { "total": 20, "covered": 18, "percentage": 90.0 },
        "branches": { "total": 30, "covered": 25, "percentage": 83.3 }
      }
    }
  },
  "timestamp": "...",
  "sequence": 9
}
```

#### Execution Result
```json
{
  "type": "execution_result",
  "payload": {
    "participant_id": "uuid",
    "agent_persona": "sentinel",
    "result": {
      "exit_code": 0,
      "stdout": "All tests passed\n",
      "stderr": "",
      "duration_ms": 1500,
      "memory_used_bytes": 25000000,
      "timed_out": false,
      "oom_killed": false
    }
  },
  "timestamp": "...",
  "sequence": 10
}
```

#### Match End
```json
{
  "type": "match_end",
  "payload": {
    "winner_id": "uuid",
    "winner_persona": "sentinel",
    "final_scores": { "sentinel": 92.5, "optimizer": 78.0 },
    "rating_changes": { "uuid1": 25, "uuid2": -18 },
    "summary": {
      "id": "uuid",
      "mode": "ranked",
      "challenge_title": "Fix SQL Injection",
      "opponent": "player2",
      "opponent_agent": "optimizer",
      "my_agent": "sentinel",
      "result": "win",
      "rating_change": 25,
      "score": 92.5,
      "opponent_score": 78.0,
      "duration_seconds": 245,
      "played_at": "2024-01-15T10:35:00Z"
    },
    "replay_url": "/replay/uuid"
  },
  "timestamp": "...",
  "sequence": 11
}
```

#### Queue Status
```json
{
  "type": "queue_status",
  "payload": {
    "position": 1,
    "estimated_wait_seconds": 30,
    "players_in_queue": 5
  },
  "timestamp": "...",
  "sequence": 12
}
```

#### Error
```json
{
  "type": "error",
  "payload": {
    "code": "MATCH_NOT_FOUND",
    "message": "Match not found or already completed",
    "details": { "match_id": "uuid" }
  },
  "timestamp": "...",
  "sequence": 13
}
```

#### Ping/Pong
```json
{
  "type": "ping",
  "payload": { "timestamp": "2024-01-15T10:30:00.000Z" },
  "timestamp": "...",
  "sequence": 14
}

{
  "type": "pong",
  "payload": { "client_timestamp": "2024-01-15T10:30:00.000Z", "server_timestamp": "2024-01-15T10:30:00.001Z" },
  "timestamp": "...",
  "sequence": 15
}
```

---

## Error Responses

All errors follow this format:
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": { ... },
    "status_code": 400
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `RATE_LIMITED` | 429 | Too many requests |
| `MATCH_NOT_FOUND` | 404 | Match doesn't exist |
| `CHALLENGE_NOT_FOUND` | 404 | Challenge doesn't exist |
| `QUEUE_FULL` | 409 | Matchmaking queue full |
| `INTERNAL_ERROR` | 500 | Server error |

---

## Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| Auth (login/register) | 10 req | 60s |
| API (general) | 100 req | 60s |
| API (authenticated) | 1000 req | 60s |
| WebSocket connections | 50 | per IP |
| Matchmaking queue | 1 | per user |

Headers returned:
- `X-RateLimit-Limit`: Max requests
- `X-RateLimit-Remaining`: Remaining requests
- `Retry-After`: Seconds until reset (on 429)

---

## Pagination

All list endpoints support:
- `limit`: 1-100 (default 20)
- `offset`: >= 0 (default 0)

Response includes:
```json
{
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

---

## Versioning

API version in URL path: `/api/v1/`
Breaking changes will increment version.