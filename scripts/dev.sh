#!/bin/bash
# Development environment setup and management
# Usage: ./dev.sh [command]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $*"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $*"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }
log_step() { echo -e "${PURPLE}[STEP]${NC} $*"; }

# Configuration
COMPOSE_FILE="$PROJECT_ROOT/docker-compose.yml"
ENV_FILE="$PROJECT_ROOT/.env"
ENV_EXAMPLE="$PROJECT_ROOT/.env.example"

check_dependencies() {
    log_step "Checking dependencies..."
    local deps=("docker" "docker-compose" "go" "node" "npm" "make")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log_error "$dep is not installed"
            return 1
        fi
    done
    log_success "All dependencies found"
}

setup_env() {
    log_step "Setting up environment..."
    if [ ! -f "$ENV_FILE" ]; then
        if [ -f "$ENV_EXAMPLE" ]; then
            cp "$ENV_EXAMPLE" "$ENV_FILE"
            log_success "Created .env from .env.example"
            log_warning "Please edit .env with your configuration"
        else
            log_error ".env.example not found"
            return 1
        fi
    else
        log_info ".env already exists"
    fi
}

build_images() {
    log_step "Building Docker images..."
    docker-compose -f "$COMPOSE_FILE" build --parallel
    log_success "Images built"
}

start_services() {
    local profile="${1:-cpu}"
    log_step "Starting services with profile: $profile"
    docker-compose -f "$COMPOSE_FILE" --profile "$profile" up -d
    log_success "Services started"
}

stop_services() {
    log_step "Stopping services..."
    docker-compose -f "$COMPOSE_FILE" down
    log_success "Services stopped"
}

restart_services() {
    stop_services
    start_services "$@"
}

logs() {
    local service="${1:-}"
    if [ -n "$service" ]; then
        docker-compose -f "$COMPOSE_FILE" logs -f "$service"
    else
        docker-compose -f "$COMPOSE_FILE" logs -f
    fi
}

status() {
    log_step "Service status:"
    docker-compose -f "$COMPOSE_FILE" ps
}

pull_models() {
    log_step "Pulling FreeLLM models..."
    docker-compose -f "$COMPOSE_FILE" exec freellm ollama pull codellama:34b
    docker-compose -f "$COMPOSE_FILE" exec freellm ollama pull deepseek-coder:33b
    docker-compose -f "$COMPOSE_FILE" exec freellm ollama pull qwen2.5-coder:32b
    docker-compose -f "$COMPOSE_FILE" exec freellm ollama pull starcoder2:15b
    docker-compose -f "$COMPOSE_FILE" exec freellm ollama pull codellama:13b
    docker-compose -f "$COMPOSE_FILE" exec freellm ollama pull deepseek-coder:6.7b
    log_success "Models pulled"
}

pull_models_light() {
    log_step "Pulling lightweight FreeLLM models..."
    docker-compose -f "$COMPOSE_FILE" exec freellm ollama pull codellama:13b
    docker-compose -f "$COMPOSE_FILE" exec freellm ollama pull deepseek-coder:6.7b
    docker-compose -f "$COMPOSE_FILE" exec freellm ollama pull qwen2.5-coder:7b
    log_success "Lightweight models pulled"
}

list_models() {
    docker-compose -f "$COMPOSE_FILE" exec freellm ollama list
}

run_migrations() {
    log_step "Running database migrations..."
    docker-compose -f "$COMPOSE_FILE" exec backend go run ./cmd/migrate
    log_success "Migrations complete"
}

seed_database() {
    log_step "Seeding database with initial data..."
    docker-compose -f "$COMPOSE_FILE" exec backend go run ./cmd/seed
    log_success "Database seeded"
}

sync_snippets() {
    log_step "Syncing snippets..."
    "$PROJECT_ROOT/scripts/sync-snippets.sh"
}

run_tests() {
    log_step "Running tests..."
    docker-compose -f "$COMPOSE_FILE" exec backend go test -v ./...
    docker-compose -f "$COMPOSE_FILE" exec frontend npm run test
    log_success "Tests complete"
}

lint() {
    log_step "Running linters..."
    docker-compose -f "$COMPOSE_FILE" exec backend golangci-lint run ./...
    docker-compose -f "$COMPOSE_FILE" exec frontend npm run lint
    log_success "Linting complete"
}

format_code() {
    log_step "Formatting code..."
    docker-compose -f "$COMPOSE_FILE" exec backend gofmt -w .
    docker-compose -f "$COMPOSE_FILE" exec frontend npm run fmt
    log_success "Formatting complete"
}

shell_backend() {
    docker-compose -f "$COMPOSE_FILE" exec backend sh
}

shell_frontend() {
    docker-compose -f "$COMPOSE_FILE" exec frontend sh
}

shell_db() {
    docker-compose -f "$COMPOSE_FILE" exec postgres psql -U codecrash -d codecrash
}

shell_redis() {
    docker-compose -f "$COMPOSE_FILE" exec redis redis-cli
}

clean() {
    log_warning "This will remove all containers, volumes, and images!"
    read -p "Are you sure? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker-compose -f "$COMPOSE_FILE" down -v --remove-orphans
        docker system prune -af --volumes
        log_success "Cleanup complete"
    else
        log_info "Cleanup cancelled"
    fi
}

reset_db() {
    log_warning "This will reset the database!"
    read -p "Are you sure? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker-compose -f "$COMPOSE_FILE" exec postgres psql -U codecrash -d codecrash -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
        run_migrations
        seed_database
        log_success "Database reset complete"
    else
        log_info "Reset cancelled"
    fi
}

generate_docs() {
    log_step "Generating API documentation..."
    docker-compose -f "$COMPOSE_FILE" exec backend go run github.com/swaggo/swag/cmd/swag@latest init -g ./cmd/api/main.go -o ./docs
    log_success "Documentation generated"
}

show_help() {
    cat <<EOF
CodeCrashReview Development Script

Usage: ./dev.sh [command]

Commands:
  setup              Initial setup (check deps, create .env, build images)
  start [profile]    Start services (profiles: cpu, gpu, sandbox)
  stop               Stop services
  restart [profile]  Restart services
  status             Show service status
  logs [service]     Show logs (optionally for specific service)

  pull-models        Pull all FreeLLM models (requires GPU)
  pull-models-light  Pull lightweight models only
  list-models        List available models

  migrate            Run database migrations
  seed               Seed database with initial data
  reset-db           Reset database (DANGEROUS)

  sync-snippets      Sync snippets from upstream repos
  test               Run all tests
  lint               Run linters
  format             Format code

  shell-backend      Open shell in backend container
  shell-frontend     Open shell in frontend container
  shell-db           Open PostgreSQL shell
  shell-redis        Open Redis CLI

  generate-docs      Generate API documentation
  clean              Clean up everything (DANGEROUS)
  help               Show this help

Examples:
  ./dev.sh setup
  ./dev.sh start gpu
  ./dev.sh logs backend
  ./dev.sh pull-models-light
  ./dev.sh test
EOF
}

# Main
main() {
    cd "$PROJECT_ROOT"

    case "${1:-help}" in
        setup)
            check_dependencies
            setup_env
            build_images
            log_success "Setup complete! Run './dev.sh start' to begin."
            ;;
        start)
            start_services "${2:-cpu}"
            ;;
        stop)
            stop_services
            ;;
        restart)
            restart_services "${2:-cpu}"
            ;;
        status)
            status
            ;;
        logs)
            logs "${2:-}"
            ;;
        pull-models)
            pull_models
            ;;
        pull-models-light)
            pull_models_light
            ;;
        list-models)
            list_models
            ;;
        migrate)
            run_migrations
            ;;
        seed)
            seed_database
            ;;
        reset-db)
            reset_db
            ;;
        sync-snippets)
            sync_snippets
            ;;
        test)
            run_tests
            ;;
        lint)
            lint
            ;;
        format)
            format_code
            ;;
        shell-backend)
            shell_backend
            ;;
        shell-frontend)
            shell_frontend
            ;;
        shell-db)
            shell_db
            ;;
        shell-redis)
            shell_redis
            ;;
        generate-docs)
            generate_docs
            ;;
        clean)
            clean
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"