#!/usr/bin/env bash
set -Eeuo pipefail

usage() { echo "Uso: $0 [--dir DIRETORIO] [--freellm-url URL] [--freellm-key CHAVE]"; }
ROOT="${CCR_VALIDATION_DIR:-$PWD/CodeCrashReview-validation}"
FREELLM_URL="${FREELLM_BASE_URL:-http://host.docker.internal:8082/v1}"
FREELLM_KEY="${FREELLM_API_KEY:-local-validation}"
while (($#)); do case "$1" in
  --dir) ROOT="$2"; shift 2;; --freellm-url) FREELLM_URL="$2"; shift 2;; --freellm-key) FREELLM_KEY="$2"; shift 2;; -h|--help) usage; exit 0;; *) echo "Opção inválida: $1" >&2; usage >&2; exit 2;; esac; done
command -v docker >/dev/null || { echo "Docker Engine/Docker Desktop não encontrado" >&2; exit 1; }
docker compose version >/dev/null || { echo "Docker Compose v2 não encontrado" >&2; exit 1; }
mkdir -p "$ROOT"; ROOT="$(cd "$ROOT" && pwd)"
if [[ ! -d "$ROOT/.git" ]]; then git clone --branch Develop --depth 1 https://github.com/verissimodf/CodeCrashReview.git "$ROOT"; fi
cd "$ROOT"; umask 077
if [[ ! -f .env.validation ]]; then
  secret() { openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n'; }
  cat > .env.validation <<EOF
POSTGRES_USER=ccr_validation
POSTGRES_DB=ccr_validation
POSTGRES_PASSWORD=$(secret)
REDIS_PASSWORD=$(secret)
JWT_SECRET=$(secret)
FREELLM_BASE_URL=$FREELLM_URL
FREELLM_API_KEY=$FREELLM_KEY
DEFAULT_MODEL=claude-3-freellm-no-thinking/ollama/qwen2.5:7b
VALIDATOR_IMAGE=golang@sha256:ecb67f987aa1a57794c91d19ade7ae0209f2f0f5496aaef9e6ae85ddf92d02ab
VALIDATOR_PYTHON_IMAGE=python@sha256:9ef1a599f863231c3d91e941389bfdc4a2d6070557243d54df1a4e57015b960e
VALIDATOR_NODE_IMAGE=node@sha256:1314fb30ec37592d5845e5a160d522c0e680972aa534549940b8d73bde88481a
EOF
fi
docker compose --env-file .env.validation -f compose.validation.yml up -d --build
for _ in {1..60}; do curl -fsS "http://127.0.0.1:${CCR_FRONTEND_PORT:-18080}/" >/dev/null && break; sleep 2; done
curl -fsS "http://127.0.0.1:${CCR_FRONTEND_PORT:-18080}/" >/dev/null
echo "CodeCrashReview disponível em http://127.0.0.1:${CCR_FRONTEND_PORT:-18080}"
echo "Desligar: docker compose --env-file .env.validation -f compose.validation.yml down"
