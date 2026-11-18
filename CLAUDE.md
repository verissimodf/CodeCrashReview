# CodeCrashReview — 1v1 Competitive Code Review Arena

## Project Overview
Real-time competitive coding platform where two AI agents battle to solve real codebase problems. Users watch live as agents propose fixes in isolated sandboxes, with gaming-style UI, leaderboards, and replay system.

## Tech Stack

### Frontend
- **Framework**: SvelteKit (TypeScript, App Router)
- **Editor**: Monaco Editor (VS Code experience)
- **Styling**: Tailwind CSS + custom cyberpunk/neon theme
- **Real-time**: Native WebSocket client
- **Charts**: Chart.js / uPlot for performance graphs
- **State**: Svelte stores + WebSocket sync

### Backend
- **Language**: Go (Gin or Fiber)
- **WebSocket**: gorilla/ws or nhooyr/websocket + Redis Pub/Sub
- **Auth**: JWT + Argon2id (custom implementation)
- **Database**: PostgreSQL (primary) + Redis (sessions, pub/sub, cache)
- **ORM**: sqlc + Goose migrations (or GORM if preferred)
- **Container Runtime**: Docker Engine API (containerd)

### AI Agents (FreeLLM-Local)
- **Inference**: Ollama / vLLM / TGI (OpenAI-compatible API)
- **Models**: codellama, deepseek-coder, qwen2.5-coder, starcoder2
- **Personas**: System prompts + tool access + model selection + temperature
- **Tools per agent**: semgrep, bandit, eslint, go vet, cargo clippy, pytest/jest, custom analyzers

### Sandbox Execution
- **Isolation**: Docker containers per language + seccomp/apparmor profiles
- **Resource Limits**: CPU (--cpus), Memory (--memory), PIDs (--pids-limit), no-new-privileges
- **Network**: None (--network none)
- **Languages**: C++, Python, Java, JavaScript, TypeScript, Rust, Shell, HTML, Gherkin
- **Pool**: Pre-warmed containers per language for low latency

### Matchmaking
- **Algorithm**: Elo/Glicko-2 rating system
- **Queue**: ±200 rating range, language preference, agent affinity
- **Modes**: Ranked 1v1, Private Challenges, Tournament (Swiss)

### Visual Design: Cyberpunk/Neon
- **Base**: #0d1117 (VS Code dark)
- **Accents**: #00ffff (cyan), #ff00ff (magenta), #00ff87 (green), #ffaa00 (amber)
- **Fonts**: JetBrains Mono (code), Orbitron/Space Grotesk (display), IBM Plex Sans (UI)
- **Effects**: Glow, scanlines, grid backgrounds, particle animations
- **Motion**: 150-300ms easings, stagger reveals, reduced-motion support

## Project Structure

```
codecrashreview/
├── CLAUDE.md
├── docker-compose.yml
├── .env.example
├── Makefile
├── docs/
│   ├── architecture.md
│   ├── api-spec.md
│   ├── agent-personas.md
│   ├── challenge-format.md
│   └── sandbox-spec.md
├── frontend/
│   ├── package.json
│   ├── svelte.config.js
│   ├── tailwind.config.js
│   ├── vite.config.ts
│   ├── src/
│   │   ├── app.html
│   │   ├── app.css
│   │   ├── lib/
│   │   │   ├── components/
│   │   │   │   ├── ui/
│   │   │   │   ├── arena/
│   │   │   │   ├── editor/
│   │   │   │   ├── leaderboard/
│   │   │   │   ├── challenges/
│   │   │   │   ├── snippets/
│   │   │   │   └── profile/
│   │   │   ├── stores/
│   │   │   ├── hooks/
│   │   │   ├── utils/
│   │   │   ├── types/
│   │   │   └── styles/
│   │   ├── routes/
│   │   │   ├── +layout.svelte
│   │   │   ├── +page.svelte
│   │   │   ├── arena/
│   │   │   │   ├── +page.svelte
│   │   │   │   └── [matchId]/
│   │   │   │       └── +page.svelte
│   │   │   ├── challenges/
│   │   │   │   ├── +page.svelte
│   │   │   │   ├── create/
│   │   │   │   └── [id]/
│   │   │   ├── leaderboard/
│   │   │   ├── snippets/
│   │   │   ├── profile/
│   │   │   └── auth/
│   │   └── app.d.ts
│   └── static/
├── backend/
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   ├── api/
│   │   ├── ws/
│   │   ├── worker/
│   │   └── migrate/
│   ├── internal/
│   │   ├── auth/
│   │   ├── matchmaking/
│   │   ├── arena/
│   │   ├── sandbox/
│   │   ├── agents/
│   │   ├── challenges/
│   │   ├── snippets/
│   │   ├── leaderboard/
│   │   ├── database/
│   │   ├── websocket/
│   │   └── config/
│   ├── pkg/
│   │   ├── docker/
│   │   ├── freellm/
│   │   └── ratelimit/
│   ├── migrations/
│   └── sqlc.yaml
├── sandbox/
│   ├── cpp/
│   ├── python/
│   ├── java/
│   ├── javascript/
│   ├── typescript/
│   ├── rust/
│   ├── shell/
│   ├── html/
│   └── gherkin/
├── shared/
│   ├── types/
│   └── constants/
└── scripts/
    ├── dev.sh
    ├── build.sh
    ├── test.sh
    └── sync-snippets.sh
```

## Agent Personas (Differentiated by Prompt + Tools + Model)

| Agent | Specialty | Model | Temp | Tools |
|-------|-----------|-------|------|-------|
| **Sentinel** | Security & Vulnerabilities | deepseek-coder:33b | 0.2 | semgrep, bandit, gosec, trivy |
| **Optimizer** | Performance & Complexity | codellama:34b | 0.3 | benchmarks, profilers, complexity |
| **Refactor** | Clean Code & Architecture | qwen2.5-coder:32b | 0.25 | ast-grep, refactoring rules |
| **Guardian** | Testing & Coverage | starcoder2:15b | 0.35 | pytest, jest, coverage, mutation |
| **DocuBot** | Documentation & Types | deepseek-coder:6.7b | 0.4 | type inference, doc generators |
| **Polyglot** | Multi-language & Migration | codellama:13b | 0.3 | cross-lang patterns, transpilers |

## Challenge Format (YAML)

```yaml
id: "chal-001"
title: "Fix SQL Injection in User Service"
description: "UserService.getUser() vulnerable to SQLi via raw query"
repository: "https://github.com/org/target-system"
target_paths:
  - "services/user/**/*.go"
  - "db/queries/**/*.sql"
severity: "critical"
category: "security"
time_limit_seconds: 300
max_turns: 10
acceptance_criteria:
  - "No raw string concatenation in SQL"
  - "Parameterized queries used"
  - "Existing tests pass"
  - "New test for injection attempt"
scoring_weights:
  correctness: 0.4
  security: 0.3
  performance: 0.15
  style: 0.1
  test_coverage: 0.05
known_issues:
  - file: "services/user/repository.go"
    line: 42
    issue: "Direct string interpolation in query"
    priority: "high"
tags: ["sql-injection", "go", "security"]
template: "security-fix"
```

## Key Features Checklist

### Core Arena (MVP)
- [ ] Real-time 1v1 battle view with split-screen agents
- [ ] Monaco Editor with diff highlighting (original vs proposed)
- [ ] Live timer with countdown + overtime
- [ ] Agent status: thinking, editing, testing, submitting
- [ ] Code execution in sandbox + result comparison
- [ ] Winner declaration + change summary + replay

### Matchmaking & Ranking
- [ ] Elo/Glicko-2 rating calculation
- [ ] Ranked queue with ±200 matching
- [ ] Leaderboard (global, weekly, by language)
- [ ] Agent-specific stats (win rate vs each persona)

### Challenges
- [ ] Challenge browser with filters
- [ ] Private challenge creation (YAML/JSON + template)
- [ ] Target repository mounting in sandbox
- [ ] Known issues injection + priority weighting
- [ ] Acceptance criteria validation

### Snippets Library
- [ ] Import from awesome-algorithms, CP templates
- [ ] Search by language, tag, complexity
- [ ] Insert into editor with one click
- [ ] Community submissions + voting

### User System
- [ ] Register/Login (JWT + Argon2id)
- [ ] Profile: rating history, match history, agent preferences
- [ ] Settings: theme, notifications, keybindings

### Spectator Mode
- [ ] Live match watching
- [ ] Replay system with timeline scrubber
- [ ] Commentary overlay (optional)

## Development Commands

```bash
# Start all services (dev)
make dev

# Frontend only
make frontend-dev

# Backend only
make backend-dev

# Build all
make build

# Run tests
make test

# Sync snippets from upstream repos
make sync-snippets

# Database migrations
make migrate-up
make migrate-down

# Generate sqlc code
make sqlc-gen
```

## Environment Variables

```env
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/codecrash?sslmode=disable
REDIS_URL=redis://localhost:6379

# Auth
JWT_SECRET=your-256-bit-secret
JWT_EXPIRY=24h
ARGON2_TIME=3
ARGON2_MEMORY=65536
ARGON2_THREADS=4
ARGON2_KEYLEN=32

# FreeLLM-Local
FREELLM_BASE_URL=http://localhost:11434/v1
FREELLM_API_KEY=ollama
DEFAULT_MODEL=codellama:34b

# Docker
DOCKER_HOST=unix:///var/run/docker.sock
SANDBOX_NETWORK=codecrash-sandbox
SANDBOX_CPU_LIMIT=2
SANDBOX_MEMORY_LIMIT=2g
SANDBOX_PIDS_LIMIT=100

# Matchmaking
ELO_K_FACTOR=32
GLICKO_TAU=0.5
MATCH_TIMEOUT_SECONDS=300
QUEUE_CHECK_INTERVAL_MS=5000

# Frontend
PUBLIC_WS_URL=ws://localhost:8080/ws
PUBLIC_API_URL=http://localhost:8080/api
```

## Security Considerations

- Sandbox containers: no network, read-only rootfs where possible, drop all capabilities
- Rate limiting on API + WebSocket (per IP + per user)
- Input validation on all challenge/agent submissions
- Audit logging for all code executions
- Secrets never in logs or client bundle
- CSP headers, HSTS, secure cookies

## Performance Targets

- WebSocket latency: <50ms local, <150ms LAN
- Sandbox cold start: <3s (pre-warmed: <500ms)
- Code execution (typical): <5s
- Frontend bundle: <500KB gzipped (excl. Monaco)
- 100 concurrent matches on 4-core/16GB server

## References

- Reverse engineering: https://github.com/inematds/rea-engenharia-reversa
- CodeClash: https://github.com/Shawnn08/CodeClash
- RAEA: https://github.com/shinyobjectz/raea
- Glicko-2: http://www.glicko.net/glicko/glicko2.pdf
- Monaco Editor: https://microsoft.github.io/monaco-editor/