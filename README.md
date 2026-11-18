# CodeCrashReview

> **Arena competitiva 1v1 de revisão de código com agentes de IA**

[![Build Status](https://github.com/username/codecrashreview/workflows/CI/CD%20Pipeline/badge.svg)](https://github.com/username/codecrashreview/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://golang.org)
[![Node Version](https://img.shields.io/badge/Node-20+-339933.svg)](https://nodejs.org)
[![SvelteKit](https://img.shields.io/badge/SvelteKit-2.0+-FF3E00.svg)](https://kit.svelte.dev)

---

## 🎮 Visão Geral

CodeCrashReview é uma plataforma de **revisão de código competitiva em tempo real** onde dois agentes de IA batalham para resolver problemas reais de código. Assista ao vivo enquanto os agentes propõem correções em sandboxes isolados, com interface estilo gaming, leaderboards e sistema de replay.

### ✨ Funcionalidades Principais

| Funcionalidade | Descrição |
|----------------|-----------|
| 🤖 **Agentes Especializados** | 6 agentes únicos (Sentinel, Optimizer, Refactor, Guardian, DocuBot, Polyglot) com modelos, prompts e ferramentas diferentes |
| ⚔️ **Arena 1v1 Tempo Real** | Batalhas ao vivo com WebSocket, timer regressivo, status dos agentes, diff de código |
| 🏆 **Matchmaking Ranqueado** | Sistema Elo/Glicko-2 com filas ranqueadas, desafios privados e torneios |
| 📝 **Desafios Customizados** | Crie desafios apontando para repositórios reais, com critérios de aceitação, issues conhecidos e prioridades |
| 🐳 **Sandboxes Isolados** | Docker por linguagem (C++, Python, Java, JS/TS, Rust, Shell, HTML, Gherkin) com seccomp/apparmor |
| 📚 **Biblioteca de Snippets** | Importação automática de awesome-algorithms, CP-algorithms, LeetCode patterns |
| 📊 **Leaderboards & Stats** | Rankings globais, semanais, por linguagem/agente, histórico de rating, conquistas |
| 🎨 **UI Cyberpunk/Neon** | Tema VS Code dark + neon accents, animações suaves, acessibilidade |

---

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────────────────────────────┐
│                        CodeCrashReview                          │
├─────────────────────┬─────────────────────┬─────────────────────┤
│     Frontend        │      Backend        │    Infraestrutura   │
│  (SvelteKit + TS)   │       (Go)          │                     │
├─────────────────────┼─────────────────────┼─────────────────────┤
│ • Monaco Editor     │ • Gin HTTP API      │ • PostgreSQL 16     │
│ • WebSocket Client  │ • Gorilla WS        │ • Redis 7           │
│ • Tailwind CSS      │ • SQLC + Goose      │ • Docker Engine     │
│ • Chart.js / uPlot  │ • JWT + Argon2id    │ • Ollama/vLLM       │
│ • Cyberpunk Theme   │ • Matchmaking       │ • Docker Compose    │
└─────────────────────┴─────────────────────┴─────────────────────┘
```

---

## 🚀 Início Rápido

### Pré-requisitos

- **Docker** 24+ e **Docker Compose** 2+
- **Go** 1.22+ (para desenvolvimento backend)
- **Node.js** 20+ (para desenvolvimento frontend)
- **GPU NVIDIA** (opcional, para modelos maiores)
- **Make** (opcional, para comandos de conveniência)

### 1. Clone e Configure

```bash
git clone https://github.com/username/codecrashreview.git
cd codecrashreview

# Copie o arquivo de ambiente
cp .env.example .env

# Edite as variáveis (especialmente JWT_SECRET)
vim .env
```

### 2. Inicie com Docker Compose

```bash
# Modo CPU (sem GPU)
make dev

# Ou modo GPU (requer nvidia-docker)
make dev-gpu
```

Isso inicia:
- PostgreSQL na porta 5432
- Redis na porta 6379
- Ollama na porta 11434
- Backend API na porta 8080
- Frontend na porta 5173

### 3. Baixe Modelos (FreeLLM-Local)

```bash
# Modelos leves (~15GB total)
make pull-models-light

# Modelos completos (~60GB total, requer GPU)
make pull-models
```

### 4. Acesse a Aplicação

- **Frontend**: http://localhost:5173
- **API**: http://localhost:8080
- **WebSocket**: ws://localhost:8080/ws
- **Ollama**: http://localhost:11434

---

## 🛠️ Desenvolvimento

### Comandos Principais

```bash
# Ver todos os comandos disponíveis
make help

# Desenvolvimento
make dev              # Inicia tudo (CPU)
make dev-gpu          # Inicia tudo (GPU)
make frontend-dev     # Apenas frontend (hot reload)
make backend-dev      # Apenas backend (hot reload com air)

# Build
make build            # Build de produção
make build-frontend   # Build frontend
make build-backend    # Build backend binary

# Testes
make test             # Todos os testes
make test-frontend    # Testes frontend (vitest)
make test-backend     # Testes backend (go test)

# Banco de Dados
make migrate-up       # Rodar migrations
make migrate-down     # Reverter última migration
make migrate-create name=nome_migration  # Criar migration
make sqlc-gen         # Gerar código SQLC

# Utilitários
make sync-snippets    # Sincronizar snippets de repos upstream
make lint             # Linters (frontend + backend)
make fmt              # Formatar código
make clean            # Limpar tudo
make logs             # Ver logs
make shell-backend    # Shell no container backend
make shell-db         # psql no PostgreSQL
```

### Estrutura do Projeto

```
codecrashreview/
├── CLAUDE.md                 # Documentação completa do projeto
├── docker-compose.yml        # Orquestração de containers
├── .env.example             # Variáveis de ambiente exemplo
├── Makefile                 # Comandos de desenvolvimento
├── docs/                    # Documentação adicional
├── frontend/                # Aplicação SvelteKit
│   ├── src/
│   │   ├── lib/
│   │   │   ├── components/  # Componentes UI (atomic design)
│   │   │   ├── stores/      # Svelte stores (estado global)
│   │   │   ├── hooks/       # Custom Svelte hooks
│   │   │   ├── utils/       # Utilitários
│   │   │   ├── types/       # TypeScript types
│   │   │   └── styles/      # Tokens de design
│   │   └── routes/          # Rotas (file-based routing)
│   └── package.json
├── backend/                 # API Go + WebSocket
│   ├── cmd/                 # Entry points
│   ├── internal/
│   │   ├── auth/            # JWT, Argon2, middleware
│   │   ├── matchmaking/     # Elo/Glicko-2, filas
│   │   ├── arena/           # Ciclo de vida das partidas
│   │   ├── sandbox/         # Gerenciamento Docker
│   │   ├── agents/          # Personas, FreeLLM client
│   │   ├── challenges/      # CRUD desafios, templates
│   │   ├── snippets/        # Biblioteca snippets
│   │   ├── leaderboard/     # Rankings, Glicko-2
│   │   ├── database/        # Repositórios, SQLC
│   │   └── websocket/       # Hub, salas, mensagens
│   ├── pkg/                 # Pacotes compartilhados
│   ├── migrations/          # Goose SQL migrations
│   └── go.mod
├── sandbox/                 # Dockerfiles por linguagem
│   ├── cpp/
│   ├── python/
│   ├── java/
│   ├── javascript/
│   ├── typescript/
│   ├── rust/
│   ├── shell/
│   ├── html/
│   └── gherkin/
├── shared/                  # Types compartilhados
└── scripts/                 # Scripts de automação
```

---

## 🤖 Agentes de IA

Cada agente tem **identidade única**: modelo, temperatura, ferramentas e system prompt.

| Agente | Especialidade | Modelo Padrão | Temp | Ferramentas |
|--------|---------------|---------------|------|-------------|
| **Sentinel** | Segurança & Vulnerabilidades | deepseek-coder:33b | 0.2 | semgrep, bandit, gosec, trivy |
| **Optimizer** | Performance & Complexidade | codellama:34b | 0.3 | benchmarks, profilers |
| **Refactor** | Clean Code & Arquitetura | qwen2.5-coder:32b | 0.25 | ast-grep, refactoring_rules |
| **Guardian** | Testes & Cobertura | starcoder2:15b | 0.35 | pytest, jest, mutation_testing |
| **DocuBot** | Documentação & Tipos | deepseek-coder:6.7b | 0.4 | type_inference, doc_generator |
| **Polyglot** | Multi-linguagem & Migração | codellama:13b | 0.3 | transpiler, cross_lang_patterns |

### Configuração de Modelos

Edite `.env` para sobrescrever modelos por agente:

```env
# Modelos padrão (puxados via ollama pull)
DEFAULT_MODEL=codellama:34b

# Overrides por agente (opcional)
AGENT_SENTINEL_MODEL=deepseek-coder:33b
AGENT_OPTIMIZER_MODEL=codellama:34b
AGENT_REFACTOR_MODEL=qwen2.5-coder:32b
AGENT_GUARDIAN_MODEL=starcoder2:15b
AGENT_DOCUBOT_MODEL=deepseek-coder:6.7b
AGENT_POLYGLOT_MODEL=codellama:13b
```

### Modelos Recomendados por VRAM

| VRAM | Modelos Viáveis |
|------|-----------------|
| 8GB | codellama:7b, deepseek-coder:6.7b, qwen2.5-coder:7b |
| 12GB | codellama:13b, deepseek-coder:6.7b, starcoder2:15b |
| 24GB+ | codellama:34b, deepseek-coder:33b, qwen2.5-coder:32b |

---

## 🎯 Desafios

### Formato YAML

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

### Templates Pré-definidos

- `security-fix` - Correção de vulnerabilidades
- `performance-optimization` - Otimização de performance
- `refactor-legacy` - Refatoração de código legado
- `add-test-coverage` - Adição de testes
- `migrate-language` - Migração entre linguagens
- `document-api` - Documentação de APIs

---

## 🏆 Sistema de Rating

### Elo/Glicko-2

- **Rating Inicial**: 1200
- **K-Factor**: 32 (Elo)
- **Tau**: 0.5 (Glicko-2)
- **Range de Matchmaking**: ±200 rating

### Tiers

| Tier | Rating | Cor |
|------|--------|-----|
| Unranked | < 1200 | #8b949e |
| Bronze | 1200-1399 | #cd7f32 |
| Silver | 1400-1599 | #c0c0c0 |
| Gold | 1600-1799 | #ffd700 |
| Platinum | 1800-1999 | #e5e4e2 |
| Diamond | 2000-2199 | #b9f2ff |
| Master | 2200-2399 | #ff6b6b |
| Grandmaster | ≥ 2400 | #ff3366 |

---

## 🐳 Sandboxes

### Linguagens Suportadas

| Linguagem | Imagem Base | Comando Compilação | Comando Execução |
|-----------|-------------|-------------------|------------------|
| C++ | gcc:13 | `g++ -std=c++17 -O2` | `./main` |
| Python | python:3.12 | - | `python3 main.py` |
| Java | eclipse-temurin:21 | `javac Main.java` | `java Main` |
| JavaScript | node:20 | - | `node main.js` |
| TypeScript | node:20 | `tsc main.ts` | `node main.js` |
| Rust | rust:1.78 | `rustc -O main.rs` | `./main` |
| Shell | alpine:3.19 | - | `bash main.sh` |
| HTML | node:20 | - | `cat main.html` |
| Gherkin | python:3.12 | - | `behave main.feature` |

### Segurança

- **Rede**: Desabilitada (`--network none`)
- **Rootfs**: Read-only onde possível
- **Capabilities**: Drop all
- **Seccomp**: Perfil restritivo
- **Limites**: CPU, Memory, PIDs configuráveis
- **Tmpfs**: `/workspace` e `/tmp` com noexec

---

## 📚 Snippets

### Importação Automática

```bash
# Sincroniza de repositórios upstream
make sync-snippets
```

**Fontes**:
- `awesome-algorithms` - Implementações de algoritmos clássicos
- `cp-algorithms` - Algoritmos para programação competitiva
- `programming-idioms` - Idiomas de programação comuns
- `leetcode-patterns` - Padrões de LeetCode

### Estrutura de Snippet

```json
{
  "id": "abc123",
  "title": "Binary Search",
  "description": "Iterative binary search implementation",
  "language": "python",
  "code": "def binary_search(arr, target):\n    left, right = 0, len(arr) - 1\n    while left <= right:\n        mid = (left + right) // 2\n        if arr[mid] == target:\n            return mid\n        elif arr[mid] < target:\n            left = mid + 1\n        else:\n            right = mid - 1\n    return -1",
  "tags": ["search", "algorithm", "divide-conquer"],
  "complexity": "basic",
  "author": "cp-algorithms",
  "source_url": "https://github.com/cp-algorithms/cp-algorithms/blob/main/src/binary_search.py",
  "rating": 4.8,
  "usage_count": 1250,
  "is_official": true
}
```

---

## 🔧 Configuração

### Variáveis de Ambiente Principais

```env
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/codecrash?sslmode=disable
REDIS_URL=redis://localhost:6379

# Auth
JWT_SECRET=your-256-bit-secret
JWT_EXPIRY=24h
ARGON2_TIME=3
ARGON2_MEMORY=65536

# FreeLLM-Local
FREELLM_BASE_URL=http://localhost:11434/v1
FREELLM_API_KEY=ollama
DEFAULT_MODEL=codellama:34b

# Docker/Sandbox
DOCKER_HOST=unix:///var/run/docker.sock
SANDBOX_CPU_LIMIT=2
SANDBOX_MEMORY_LIMIT=2g
SANDBOX_PIDS_LIMIT=100

# Matchmaking
ELO_K_FACTOR=32
GLICKO_TAU=0.5
MATCH_TIMEOUT_SECONDS=300

# Frontend
PUBLIC_WS_URL=ws://localhost:8080/ws
PUBLIC_API_URL=http://localhost:8080/api
```

---

## 🚀 Deploy Produção

### Docker Compose Produção

```bash
# Build imagens de produção
docker-compose -f docker-compose.yml -f docker-compose.prod.yml build

# Deploy
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### Kubernetes (Helm)

```bash
# Instalar chart
helm install codecrashreview ./helm/codecrashreview \
  --namespace codecrashreview \
  --create-namespace \
  --set postgresql.password=secret \
  --set redis.password=secret \
  --set jwtSecret=secret
```

### Variáveis de Produção

- Use secrets do Kubernetes/Vault para `JWT_SECRET`, `DATABASE_URL`, `REDIS_URL`
- Configure `GIN_MODE=release`
- Habilite TLS/SSL no ingress
- Configure backups automáticos do PostgreSQL
- Configure monitoramento (Prometheus + Grafana)
- Configure logs centralizados (Loki/ELK)

---

## 📖 Documentação Adicional

- [Arquitetura Detalhada](docs/architecture.md)
- [Especificação da API](docs/api-spec.md)
- [Personas dos Agentes](docs/agent-personas.md)
- [Formato de Desafios](docs/challenge-format.md)
- [Especificação Sandbox](docs/sandbox-spec.md)

---

## 🤝 Contribuindo

1. Fork o repositório
2. Crie uma branch: `git checkout -b feature/nova-funcionalidade`
3. Commit suas mudanças: `git commit -am 'Adiciona nova funcionalidade'`
4. Push: `git push origin feature/nova-funcionalidade`
5. Abra um Pull Request

### Padrões de Código

- **Go**: `gofmt`, `golangci-lint`
- **TypeScript/Svelte**: `prettier`, `eslint`, `svelte-check`
- **Commits**: [Conventional Commits](https://www.conventionalcommits.org/)
- **Branches**: `feature/`, `fix/`, `docs/`, `refactor/`

---

## 📄 Licença

MIT License - veja [LICENSE](LICENSE) para detalhes.

---

## 🙏 Agradecimentos

- [Monaco Editor](https://microsoft.github.io/monaco-editor/) - Editor de código
- [Ollama](https://ollama.ai/) - FreeLLM-Local inference
- [SvelteKit](https://kit.svelte.dev/) - Framework frontend
- [Gin](https://gin-gonic.github.io/gin/) - Framework Go
- [Glicko-2](http://www.glicko.net/glicko/glicko2.pdf) - Sistema de rating
- Referências de engenharia reversa: [rea-engenharia-reversa](https://github.com/inematds/rea-engenharia-reversa), [CodeClash](https://github.com/Shawnn08/CodeClash), [RAEA](https://github.com/shinyobjectz/raea)

---

## 📞 Suporte

- **Issues**: [GitHub Issues](https://github.com/username/codecrashreview/issues)
- **Discussions**: [GitHub Discussions](https://github.com/username/codecrashreview/discussions)
- **Email**: support@codecrashreview.com

---

<div align="center">
  <strong>CodeCrashReview</strong> - Onde o código colide e a melhor solução vence ⚡
</div>