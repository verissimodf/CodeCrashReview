# CodeCrashReview Makefile
# Usage: make <target>

.PHONY: help dev build test clean migrate-up migrate-down sqlc-gen sync-snippets frontend-dev backend-dev sandbox-build sandbox-pull

# Default target
help:
	@echo "CodeCrashReview - Competitive Code Review Arena"
	@echo ""
	@echo "Development:"
	@echo "  make dev              Start all services (docker-compose)"
	@echo "  make frontend-dev     Start frontend only (SvelteKit dev server)"
	@echo "  make backend-dev      Start backend only (Go with air hot reload)"
	@echo "  make sandbox-build    Build all sandbox Docker images"
	@echo "  make sandbox-pull     Pull base images for sandboxes"
	@echo ""
	@echo "Building:"
	@echo "  make build            Build all production images"
	@echo "  make build-frontend   Build frontend production bundle"
	@echo "  make build-backend    Build backend binary"
	@echo ""
	@echo "Testing:"
	@echo "  make test             Run all tests"
	@echo "  make test-frontend    Run frontend tests (vitest)"
	@echo "  make test-backend     Run backend tests (go test)"
	@echo "  make test-integration Run integration tests"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up       Run database migrations"
	@echo "  make migrate-down     Rollback last migration"
	@echo "  make migrate-create   Create new migration (usage: make migrate-create name=add_users_table)"
	@echo "  make sqlc-gen         Generate SQLC code from queries"
	@echo "  make db-reset         Reset database (drop + migrate)"
	@echo ""
	@echo "Utilities:"
	@echo "  make sync-snippets    Sync snippets from upstream repos"
	@echo "  make lint             Run linters (frontend + backend)"
	@echo "  make fmt              Format code (prettier + gofmt)"
	@echo "  make clean            Clean build artifacts"
	@echo "  make logs             Show docker-compose logs"
	@echo "  make shell-backend    Shell into backend container"
	@echo "  make shell-frontend   Shell into frontend container"
	@echo "  make shell-db         PostgreSQL shell"
	@echo "  make shell-redis      Redis CLI"

# ===========================================
# Development
# ===========================================
dev:
	docker-compose --profile cpu up --build

dev-gpu:
	docker-compose --profile gpu up --build

frontend-dev:
	cd frontend && npm run dev

backend-dev:
	cd backend && air -c .air.toml

# ===========================================
# Building
# ===========================================
build: build-frontend build-backend sandbox-build

build-frontend:
	cd frontend && npm run build

build-backend:
	cd backend && CGO_ENABLED=0 GOOS=linux go build -o bin/server ./cmd/api

sandbox-build:
	docker-compose build sandbox-cpp sandbox-python sandbox-java sandbox-javascript sandbox-typescript sandbox-rust sandbox-shell sandbox-html sandbox-gherkin

sandbox-pull:
	docker pull golang:1.22-alpine
	docker pull python:3.12-alpine
	docker pull openjdk:21-slim
	docker pull node:20-alpine
	docker pull rust:1.78-alpine
	docker pull alpine:latest

# ===========================================
# Testing
# ===========================================
test: test-frontend test-backend

test-frontend:
	cd frontend && npm run test

test-backend:
	cd backend && go test -v -race -coverprofile=coverage.out ./...

test-integration:
	cd backend && go test -v -tags=integration ./...

# ===========================================
# Database
# ===========================================
migrate-up:
	cd backend && goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	cd backend && goose -dir migrations postgres "$(DATABASE_URL)" down

migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=migration_name"; exit 1; fi
	cd backend && goose -dir migrations create $(name) sql

sqlc-gen:
	cd backend && sqlc generate

db-reset:
	cd backend && goose -dir migrations postgres "$(DATABASE_URL)" reset
	cd backend && goose -dir migrations postgres "$(DATABASE_URL)" up

# ===========================================
# Utilities
# ===========================================
sync-snippets:
	@echo "Syncing snippets from upstream repositories..."
	./scripts/sync-snippets.sh

lint: lint-frontend lint-backend

lint-frontend:
	cd frontend && npm run lint

lint-backend:
	cd backend && golangci-lint run ./...

fmt: fmt-frontend fmt-backend

fmt-frontend:
	cd frontend && npm run fmt

fmt-backend:
	cd backend && gofmt -w .

clean:
	docker-compose down -v --remove-orphans
	rm -rf frontend/build frontend/.svelte-kit frontend/node_modules
	rm -rf backend/bin backend/coverage.out
	docker system prune -f

logs:
	docker-compose logs -f

logs-backend:
	docker-compose logs -f backend

logs-frontend:
	docker-compose logs -f frontend

shell-backend:
	docker-compose exec backend sh

shell-frontend:
	docker-compose exec frontend sh

shell-db:
	docker-compose exec postgres psql -U codecrash -d codecrash

shell-redis:
	docker-compose exec redis redis-cli

# ===========================================
# FreeLLM Model Management
# ===========================================
pull-models:
	docker-compose exec freellm ollama pull codellama:34b
	docker-compose exec freellm ollama pull deepseek-coder:33b
	docker-compose exec freellm ollama pull qwen2.5-coder:32b
	docker-compose exec freellm ollama pull starcoder2:15b
	docker-compose exec freellm ollama pull codellama:13b
	docker-compose exec freellm ollama pull deepseek-coder:6.7b

pull-models-light:
	docker-compose exec freellm ollama pull codellama:13b
	docker-compose exec freellm ollama pull deepseek-coder:6.7b
	docker-compose exec freellm ollama pull qwen2.5-coder:7b

list-models:
	docker-compose exec freellm ollama list

# ===========================================
# Production
# ===========================================
prod-build:
	docker compose -f compose.host.yml build

prod-up:
	docker compose -f compose.host.yml up -d --wait

prod-down:
	docker compose -f compose.host.yml down

# ===========================================
# CI Helpers
# ===========================================
ci-test: test lint

ci-build: build

ci-deploy:
	@echo "Deploy to production (configure your CI/CD)"
