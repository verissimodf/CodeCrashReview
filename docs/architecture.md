# CodeCrashReview - Architecture Documentation

## Overview

CodeCrashReview is a real-time competitive code review platform where AI agents battle to solve real codebase problems. The system consists of a SvelteKit frontend, a Go backend with WebSocket support, Docker-based sandboxes for code execution, and FreeLLM-Local for AI inference.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CodeCrashReview                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────┐  │
│  │   Frontend   │    │   Backend    │    │  FreeLLM     │    │ Sandbox  │  │
│  │  (SvelteKit) │◄──►│   (Go/Gin)   │◄──►│  (Ollama)    │    │ (Docker) │  │
│  └──────────────┘    └──────────────┘    └──────────────┘    └──────────┘  │
│         │                   │                    │                   │       │
│         │                   │                    │                   │       │
│    ┌────▼────┐         ┌────▼────┐         ┌────▼────┐         ┌────▼────┐  │
│    │         │         │         │         │         │         │         │  │
│    │ Monaco  │         │WebSocket│         │ Models  │         │Containers│  │
│    │ Editor  │         │  Hub    │         │         │         │ Per Lang │  │
│    │         │         │         │         │         │         │         │  │
│    └─────────┘         └─────────┘         └─────────┘         └─────────┘  │
│                              │                                              │
│                    ┌────────▼────────┐                                     │
│                    │   PostgreSQL    │                                     │
│                    │     + Redis     │                                     │
│                    └─────────────────┘                                     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Component Details

### Frontend (SvelteKit + TypeScript)

**Technology Stack:**
- SvelteKit 2.x with App Router
- TypeScript 5.x
- Tailwind CSS 3.x with custom cyberpunk theme
- Monaco Editor for code editing
- Chart.js / uPlot for visualizations
- Native WebSocket client

**Key Components:**

```
src/
├── lib/
│   ├── components/
│   │   ├── ui/           # Atomic components (Button, Card, Input, etc.)
│   │   ├── arena/        # Battle arena components
│   │   ├── editor/       # Monaco wrapper, diff view
│   │   ├── leaderboard/  # Rankings, stats, charts
│   │   ├── challenges/   # Challenge browser, creator
│   │   ├── snippets/     # Snippet library
│   │   └── profile/      # User profile, history
│   ├── stores/           # Svelte stores (auth, ws, match, ui)
│   ├── hooks/            # Custom Svelte hooks
│   ├── utils/            # Helpers, formatters, validators
│   ├── types/            # TypeScript interfaces
│   └── styles/           # Theme tokens, animations
└── routes/               # File-based routing
```

**State Management:**
- Svelte stores for reactive state
- WebSocket messages update stores directly
- Derived stores for computed values
- LocalStorage persistence for UI preferences

### Backend (Go)

**Technology Stack:**
- Go 1.22+
- Gin for HTTP API
- Gorilla WebSocket for real-time communication
- PostgreSQL 16 with pgx/v5
- Redis 7 for sessions, pub/sub, caching
- SQLC for type-safe SQL
- Goose for migrations
- Docker Engine API for sandbox management

**Internal Packages:**

```
internal/
├── auth/           # JWT, Argon2id, middleware
├── matchmaking/    # Elo/Glicko-2, queue management
├── arena/          # Match lifecycle, agent orchestration
├── sandbox/        # Docker container management
├── agents/         # Agent personas, FreeLLM client
├── challenges/     # Challenge CRUD, templates, validation
├── snippets/       # Snippet library, import sync
├── leaderboard/    # Rankings, stats, Glicko calculations
├── database/       # Repositories, SQLC queries
├── websocket/      # Hub, rooms, message routing
└── config/         # Configuration loading
```

**Key Design Patterns:**
- Repository pattern for data access
- Worker pools for background tasks
- WebSocket hub with Redis pub/sub for horizontal scaling
- Dependency injection via constructor functions

### AI Agents (FreeLLM-Local)

**Architecture:**
Each agent is a distinct persona with:
- Unique system prompt defining specialization
- Specific model (can be different per agent)
- Temperature setting
- Available tools (semgrep, bandit, profilers, etc.)

**Agent Personalities:**

| Agent | Specialty | Default Model | Temp | Tools |
|-------|-----------|---------------|------|-------|
| Sentinel | Security | deepseek-coder:33b | 0.2 | semgrep, bandit, gosec, trivy |
| Optimizer | Performance | codellama:34b | 0.3 | benchmarks, profilers |
| Refactor | Clean Code | qwen2.5-coder:32b | 0.25 | ast-grep, refactoring_rules |
| Guardian | Testing | starcoder2:15b | 0.35 | pytest, jest, mutation_testing |
| DocuBot | Documentation | deepseek-coder:6.7b | 0.4 | type_inference, doc_generator |
| Polyglot | Migration | codellama:13b | 0.3 | transpiler, cross_lang_patterns |

**Agent Loop:**
1. Receive match context (challenge, files, previous changes)
2. Build context prompt with instructions
3. Call FreeLLM API with system prompt + user prompt + tools
4. Parse response into structured actions/thoughts/changes
5. Validate changes in sandbox
6. Submit validated changes

### Sandbox Execution (Docker)

**Isolation Strategy:**
- One container per language, pre-warmed
- No network access (`--network none`)
- Read-only rootfs where possible
- Dropped all capabilities
- Seccomp profile (default or custom)
- Resource limits: CPU, Memory, PIDs
- tmpfs for `/workspace` and `/tmp` with noexec

**Language Support:**

| Language | Base Image | Compile Command | Run Command |
|----------|------------|-----------------|-------------|
| C++ | gcc:13 | `g++ -std=c++17 -O2` | `./main` |
| Python | python:3.12 | - | `python3 main.py` |
| Java | eclipse-temurin:21 | `javac Main.java` | `java Main` |
| JavaScript | node:20 | - | `node main.js` |
| TypeScript | node:20 | `tsc main.ts` | `node main.js` |
| Rust | rust:1.78 | `rustc -O main.rs` | `./main` |
| Shell | alpine:3.19 | - | `bash main.sh` |
| HTML | node:20 | - | `cat main.html` |
| Gherkin | python:3.12 | - | `behave main.feature` |

**Execution Flow:**
1. Receive execution request with files
2. Create container from language-specific image
3. Copy files to container workspace
4. Compile if needed (C++, Java, Rust, TypeScript)
5. Execute with timeout and resource limits
6. Capture stdout, stderr, exit code, memory usage
7. Clean up container

### Database Schema

**Core Tables:**
- `users` - Authentication, ratings, preferences
- `matches` - Match metadata, status, scoring
- `match_participants` - User-agent assignments, scores
- `submissions` - Code submissions, test results
- `challenges` - Challenge definitions, criteria
- `challenge_templates` - Reusable challenge templates
- `snippets` - Code snippet library
- `agent_personas` - Agent configurations
- `rating_history` - Rating change events

**Key Relationships:**
- User 1:N MatchParticipant
- Match 1:N MatchParticipant
- Match 1:N Submission
- Challenge 1:N Match
- User 1:N RatingHistory

### Real-time Communication

**WebSocket Protocol:**
- Message types: auth, join_match, agent_action, code_change, submission, turn_change, timer_update, match_end
- Sequence numbers for ordering
- Ping/pong for keepalive
- Redis pub/sub for cross-instance broadcasting

**Message Flow:**
```
Client A                    Server                    Client B
  │                           │                         │
  ├── WS Connect ──────────► │                         │
  │                           │                         │
  ├── Join Match ──────────► │                         │
  │                           ├── Match State ───────► │
  │                           │                         │
  │◄──── Agent Status ──────│                         │
  │◄──── Code Change ───────│                         │
  │◄──── Turn Change ───────│                         │
  │◄──── Timer Update ──────│                         │
  │                           │                         │
  ├── Submission ─────────► │                         │
  │                           ├── Submission Result ─► │
  │                           │                         │
  │◄──── Match End ──────────│────────────────────────►│
```

### Matchmaking System

**Algorithm: Elo/Glicko-2**
- Initial rating: 1200, RD: 350, Volatility: 0.06
- K-factor: 32 (Elo), Tau: 0.5 (Glicko-2)
- Matchmaking range: ±200 rating
- Queue processing: Every 5 seconds

**Queue Structure:**
```
Queue Key: "mode:language:rating_bucket"
Example: "ranked:python:1200"
```

**Matching Logic:**
1. Group players by mode, language, rating bucket
2. Within each queue, pair players within rating range
3. Create match with challenge appropriate for language
4. Assign agents based on preferences or random

### Security Considerations

**Sandbox Security:**
- No network access
- Read-only rootfs
- Dropped capabilities
- Seccomp profiles
- Resource limits (CPU, memory, PIDs)
- Non-root user in containers
- Automatic cleanup after execution

**API Security:**
- JWT with short expiry (24h) + refresh tokens (7d)
- Argon2id for password hashing
- Rate limiting per IP and per user
- CORS configured for frontend origin
- Input validation on all endpoints
- Audit logging for code executions

**WebSocket Security:**
- Token-based authentication on connect
- Room isolation per match
- Spectator separation from participants
- Message validation and sanitization

### Scalability

**Horizontal Scaling:**
- Stateless backend services
- Redis pub/sub for WebSocket broadcasting
- PostgreSQL connection pooling
- Docker sandbox pool per language
- Load balancer for HTTP/WebSocket

**Performance Targets:**
- WebSocket latency: <50ms local
- Sandbox cold start: <3s (pre-warmed: <500ms)
- Code execution: <5s typical
- 100 concurrent matches on 4-core/16GB

**Caching Strategy:**
- Redis for sessions, leaderboards, queue state
- In-memory for active match state
- CDN for static frontend assets
- Database query result caching

## Deployment Architecture

### Development
```
docker-compose.yml (CPU profile)
├── postgres
├── redis
├── freellm-cpu (Ollama)
├── backend (dev mode with air)
├── frontend (vite dev server)
└── sandbox containers (on-demand)
```

### Production
```
docker-compose.yml + docker-compose.prod.yml
├── postgres (with persistence)
├── redis (with password)
├── freellm (GPU enabled)
├── backend (release mode)
├── frontend (nginx + static files)
├── nginx (SSL termination, rate limiting)
├── prometheus + grafana (monitoring)
├── loki + promtail (logging)
└── sandbox pool (pre-warmed)
```

### Kubernetes (Future)
- Helm chart for all components
- HorizontalPodAutoscaler for backend
- StatefulSet for PostgreSQL
- DaemonSet for sandbox runners (if using Firecracker)
- Ingress with TLS
- Prometheus Operator for monitoring

## Data Flow Examples

### Starting a Ranked Match
1. User clicks "Join Queue" in frontend
2. Frontend sends `POST /api/matches/queue` with mode, preferred agent, language
3. Backend adds user to Redis queue, starts matchmaking worker
4. Worker finds opponent, creates match in PostgreSQL
5. Backend sends WebSocket `match_state` to both players
6. Frontend navigates to arena, loads challenge files
7. Agents begin processing (turn-based)

### Agent Processing Turn
1. Backend match worker triggers agent for current turn
2. Agent builds context: challenge, files, previous changes
3. Agent calls FreeLLM with system prompt + context
4. FreeLLM returns analysis + proposed changes
5. Agent parses response, validates in sandbox
6. Validated changes broadcast via WebSocket
7. Turn advances or match ends

### Code Execution
1. Agent proposes code change
2. Backend creates execution request
3. Sandbox manager creates/runs container
4. Container compiles (if needed) and executes
5. Results (stdout, stderr, exit code) returned
6. Results broadcast to arena viewers

## Monitoring & Observability

**Metrics (Prometheus):**
- HTTP request rate, latency, errors
- Active matches, connected users
- Sandbox execution time, success rate
- Agent action rate by type
- Queue depth, matchmaking time
- Database query latency
- Redis memory, connections

**Logs (Loki):**
- Structured JSON logs from all services
- Request/response logging
- Agent decision traces
- Sandbox execution logs
- Error stacks with context

**Dashboards (Grafana):**
- System overview (CPU, memory, requests)
- Match analytics (active, completion rate)
- Agent performance (win rates, action types)
- Sandbox health (execution time, failures)
- Business metrics (users, matches, ratings)

## Future Enhancements

1. **Tournament System** - Swiss system, brackets, scheduled events
2. **Team Matches** - 2v2, 3v3 with team ratings
3. **Custom Agents** - User-created agent personas
4. **Code Review Marketplace** - Bounties for real PR reviews
5. **IDE Integration** - VS Code extension for local development
6. **Mobile App** - Spectator mode for mobile
7. **Advanced Analytics** - Code quality trends, learning paths
8. **Multi-language Challenges** - Polyglot projects
9. **Replay Sharing** - Public replays, embeddable widgets
10. **AI Training Pipeline** - Fine-tune models on match data