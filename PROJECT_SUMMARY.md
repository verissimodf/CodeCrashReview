# CodeCrashReview - Project Summary

## 🎯 Project Overview

CodeCrashReview is a **real-time competitive code review platform** where AI agents battle 1v1 to solve real codebase problems. Built with a modern tech stack featuring SvelteKit frontend, Go backend, Docker sandboxes, and FreeLLM-Local for AI inference.

## 📁 Project Structure

```
codecrashreview/
├── CLAUDE.md                 # Project documentation & architecture decisions
├── README.md                 # Main project documentation
├── PROJECT_SUMMARY.md        # This file
├── docker-compose.yml        # Development orchestration
├── docker-compose.prod.yml   # Production orchestration
├── .env.example              # Environment variables template
├── Makefile                  # Development commands
├── .github/workflows/ci.yml  # CI/CD pipeline
├── nginx/                    # Nginx configuration
├── monitoring/               # Prometheus, Grafana, Loki configs
├── scripts/                  # Development scripts
├── docs/                     # Technical documentation
│   ├── architecture.md
│   ├── api-spec.md
│   ├── agent-personas.md
│   ├── challenge-format.md
│   └── sandbox-spec.md
├── frontend/                 # SvelteKit + TypeScript
│   ├── package.json
│   ├── svelte.config.js
│   ├── tailwind.config.js
│   ├── vite.config.ts
│   ├── src/
│   │   ├── app.html
│   │   ├── app.css
│   │   ├── app.d.ts
│   │   ├── lib/
│   │   │   ├── components/   # Reusable UI components
│   │   │   │   ├── ui/       # Atomic components (Button, Card, Input, etc.)
│   │   │   │   ├── arena/    # Battle arena components
│   │   │   │   ├── editor/   # Monaco Editor + DiffView
│   │   │   │   ├── leaderboard/
│   │   │   │   ├── challenges/
│   │   │   │   ├── snippets/
│   │   │   │   └── profile/
│   │   │   ├── stores/       # Svelte stores (auth, ws, match, ui)
│   │   │   ├── hooks/        # Custom Svelte hooks
│   │   │   ├── utils/        # Utilities & formatters
│   │   │   ├── types/        # TypeScript types
│   │   │   └── styles/       # Theme tokens
│   │   └── routes/           # File-based routing
│   │       ├── +layout.svelte
│   │       ├── +page.svelte
│   │       ├── arena/
│   │       ├── challenges/
│   │       ├── leaderboard/
│   │       ├── snippets/
│   │       ├── profile/
│   │       └── auth/
│   └── static/
├── backend/                   # Go + Gin + WebSocket
│   ├── go.mod
│   ├── Dockerfile
│   ├── cmd/
│   │   ├── api/main.go       # Main entry point
│   │   ├── migrate/main.go   # Migration runner
│   │   └── seed/main.go      # Database seeder
│   ├── internal/
│   │   ├── auth/             # JWT + Argon2id auth
│   │   ├── matchmaking/      # Elo/Glicko-2 matchmaking
│   │   ├── arena/            # Match lifecycle
│   │   ├── sandbox/          # Docker sandbox manager
│   │   ├── agents/           # AI agent personas
│   │   ├── challenges/       # Challenge CRUD
│   │   ├── snippets/         # Snippet library
│   │   ├── leaderboard/      # Rankings & stats
│   │   ├── database/         # Repositories & SQLC
│   │   ├── websocket/        # WebSocket hub
│   │   └── config/           # Configuration
│   ├── pkg/
│   │   ├── docker/           # Docker client wrapper
│   │   ├── freellm/          # FreeLLM-Local client
│   │   └── ratelimit/        # Rate limiter
│   ├── migrations/           # Goose SQL migrations
│   └── sqlc.yaml
├── sandbox/                   # Docker images per language
│   ├── cpp/
│   ├── python/
│   ├── java/
│   ├── javascript/
│   ├── typescript/
│   ├── rust/
│   ├── shell/
│   ├── html/
│   └── gherkin/
├── shared/                    # Shared types
└── scripts/
```

## 🏗️ Architecture Decisions

| Component | Decision | Rationale |
|-----------|----------|-----------|
| **Frontend** | SvelteKit + TypeScript | Performance, DX, small bundles |
| **Editor** | Monaco Editor | VS Code experience, IntelliSense |
| **Backend** | Go + Gin | Performance, concurrency |
| **Real-time** | Native WebSocket + Redis Pub/Sub | Low latency, horizontal scaling |
| **Auth** | JWT + Argon2id | Security, no vendor lock-in |
| **Database** | PostgreSQL + Redis | ACID + caching/pub-sub |
| **Sandbox** | Docker + seccomp/apparmor | Isolation, security |
| **AI** | FreeLLM-Local (Ollama) | Privacy, cost control |
| **Styling** | Tailwind + Cyberpunk theme | Gaming aesthetic |

## 🤖 AI Agent Personas (6 Unique Agents)

| Agent | Specialty | Model | Temp | Tools |
|-------|-----------|-------|------|-------|
| **Sentinel** | Security & Vulnerabilities | deepseek-coder:33b | 0.2 | semgrep, bandit, gosec, trivy |
| **Optimizer** | Performance & Complexity | codellama:34b | 0.3 | benchmarks, profilers |
| **Refactor** | Clean Code & Architecture | qwen2.5-coder:32b | 0.25 | ast-grep, refactoring_rules |
| **Guardian** | Testing & Coverage | starcoder2:15b | 0.35 | pytest, jest, mutation_testing |
| **DocuBot** | Documentation & Types | deepseek-coder:6.7b | 0.4 | type_inference, doc_generator |
| **Polyglot** | Multi-language & Migration | codellama:13b | 0.3 | transpiler, cross_lang_patterns |

## 🎮 Key Features Implemented

### Core Arena
- ✅ Real-time 1v1 battle view with split-screen agents
- ✅ Monaco Editor with diff highlighting (original vs proposed)
- ✅ Live timer with countdown + overtime warnings
- ✅ Agent status: thinking, editing, testing, submitting
- ✅ Code execution in sandbox + result comparison
- ✅ Winner declaration + change summary + replay system

### Matchmaking & Ranking
- ✅ Elo/Glicko-2 rating calculation
- ✅ Ranked queue with ±200 rating matching
- ✅ Leaderboard (global, weekly, by language, by agent)
- ✅ Agent-specific stats (win rate vs each persona)

### Challenges System
- ✅ Challenge browser with filters (category, severity, search)
- ✅ Private challenge creation (YAML/JSON + templates)
- ✅ Target repository mounting in sandbox
- ✅ Known issues injection + priority weighting
- ✅ Acceptance criteria validation

### Snippets Library
- ✅ Import from awesome-algorithms, CP templates
- ✅ Search by language, tag, complexity
- ✅ Insert into editor with one click
- ✅ Community submissions + voting

### User System
- ✅ Register/Login (JWT + Argon2id)
- ✅ Profile: rating history, match history, agent preferences
- ✅ Settings: theme, notifications, keybindings

### Spectator Mode
- ✅ Live match watching
- ✅ Replay system with timeline scrubber

### Visual Design
- ✅ Cyberpunk/Neon theme (VS Code dark + neon accents)
- ✅ Glow effects, scanlines, grid backgrounds
- ✅ JetBrains Mono (code), Orbitron/Space Grotesk (display)
- ✅ Responsive design, reduced-motion support

## 🐳 Sandbox Execution (9 Languages)

| Language | Image | Compile | Run |
|----------|-------|---------|-----|
| C++ | gcc:13 | `g++ -std=c++17 -O2` | `./main` |
| Python | python:3.12 | - | `python3 main.py` |
| Java | eclipse-temurin:21 | `javac Main.java` | `java Main` |
| JavaScript | node:20 | - | `node main.js` |
| TypeScript | node:20 | `tsc main.ts` | `node main.js` |
| Rust | rust:1.78 | `rustc -O main.rs` | `./main` |
| Shell | alpine:3.19 | - | `bash main.sh` |
| HTML | node:20 | - | `cat main.html` |
| Gherkin | python:3.12 | - | `behave main.feature` |

**Security**: No network, read-only rootfs, dropped capabilities, seccomp profiles, resource limits (CPU/Memory/PIDs)

## 🚀 Getting Started

```bash
# Clone repository
git clone <repo>
cd codecrashreview

# Setup environment
cp .env.example .env
# Edit .env with your configuration

# Start development environment
make dev          # CPU mode
# or
make dev-gpu      # GPU mode (requires NVIDIA Docker)

# Pull AI models (one-time)
make pull-models-light  # Lightweight models (~15GB)
# or
make pull-models        # Full models (~60GB, requires GPU)

# Access
# Frontend: http://localhost:5173
# API: http://localhost:8080
# WebSocket: ws://localhost:8080/ws
```

## 📚 Documentation

- [Architecture](docs/architecture.md) - System architecture deep dive
- [API Specification](docs/api-spec.md) - REST + WebSocket API
- [Agent Personas](docs/agent-personas.md) - Detailed agent profiles
- [Challenge Format](docs/challenge-format.md) - YAML challenge specification
- [Sandbox Spec](docs/sandbox-spec.md) - Sandbox execution details

## 🧪 Development Commands

```bash
make help              # Show all commands
make dev               # Start all services (CPU)
make frontend-dev      # Frontend only (hot reload)
make backend-dev       # Backend only (hot reload)
make build             # Build all production images
make test              # Run all tests
make lint              # Run linters
make fmt               # Format code
make migrate-up        # Run migrations
make sqlc-gen          # Generate SQLC code
make sync-snippets     # Sync snippets from upstream repos
```

## 🔒 Security Features

- Sandbox isolation (no network, read-only rootfs, seccomp)
- JWT + Argon2id authentication
- Rate limiting (per IP + per user)
- Input validation on all endpoints
- Audit logging for code executions
- CSP headers, HSTS, secure cookies
- Resource limits on sandboxes

## 📊 Monitoring & Observability

- Prometheus metrics (custom metrics for matches, agents, sandboxes)
- Grafana dashboards (system, matches, agents, sandboxes)
- Loki + Promtail for centralized logging
- Structured JSON logging
- Health checks on all services

## 🚀 Production Deployment

```bash
# Build production images
docker-compose -f docker-compose.yml -f docker-compose.prod.yml build

# Deploy
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Includes:
# - PostgreSQL + Redis
# - FreeLLM (GPU enabled)
# - Backend (release mode)
# - Frontend (nginx + static files)
# - Nginx (SSL termination, rate limiting)
# - Prometheus + Grafana + Loki
```

## 📝 Next Steps / Future Enhancements

1. **Tournament System** - Swiss system, brackets, scheduled events
2. **Team Matches** - 2v2, 3v3 with team ratings
3. **Custom Agents** - User-created agent personas
4. **Code Review Marketplace** - Bounties for real PR reviews
5. **IDE Integration** - VS Code extension
6. **Mobile App** - Spectator mode for mobile
7. **Advanced Analytics** - Code quality trends, learning paths
8. **Multi-language Challenges** - Polyglot projects
9. **Replay Sharing** - Public replays, embeddable widgets
10. **AI Training Pipeline** - Fine-tune models on match data

---

**Built with ❤️ for competitive code review**
*CodeCrashReview - Where code collides and the best solution wins* ⚡