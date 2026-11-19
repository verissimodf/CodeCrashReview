#!/usr/bin/env bash
# ============================================
# CodeCrashReview - Repository Verification Script (Bash)
# ============================================
# Validates if all expected files and directories exist in the repository
# Usage: ./verify-repo.sh [--verbose] [--fix-missing]
# ============================================

set -euo pipefail

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
GRAY='\033[0;90m'
NC='\033[0m' # No Color

VERBOSE=false
FIX_MISSING=false

# Parse arguments
for arg in "$@"; do
    case $arg in
        --verbose|-v) VERBOSE=true ;;
        --fix-missing|-f) FIX_MISSING=true ;;
        --help|-h)
            echo "Usage: $0 [--verbose|-v] [--fix-missing|-f]"
            echo "  --verbose, -v       Show all found files/directories"
            echo "  --fix-missing, -f   Create missing directories"
            echo "  --help, -h          Show this help"
            exit 0
            ;;
    esac
done

# Colors helper
status() {
    local msg="$1"
    local color="${2:-$NC}"
    local prefix="${3:-}"
    echo -e "${color}${prefix}${msg}${NC}"
}

# Check if path exists
check_path() {
    local path="$1"
    local type="${2:-any}"
    local full_path="$(pwd)/$path"

    case $type in
        file)   [[ -f "$full_path" ]] ;;
        dir)    [[ -d "$full_path" ]] ;;
        *)      [[ -e "$full_path" ]] ;;
    esac
}

# ============================================
# EXPECTED REPOSITORY STRUCTURE
# ============================================

REQUIRED_FILES=(
    # Root files
    "README.md" "CLAUDE.md" "docker-compose.yml" "docker-compose.prod.yml"
    ".env.example" "Makefile" ".gitignore" "LICENSE" "PROJECT_SUMMARY.md"

    # Frontend root
    "frontend/package.json" "frontend/svelte.config.js" "frontend/tailwind.config.js"
    "frontend/vite.config.ts" "frontend/tsconfig.json" "frontend/postcss.config.js"
    "frontend/Dockerfile" "frontend/nginx.conf" "frontend/vitest.config.ts"
    "frontend/src/app.html" "frontend/src/app.css" "frontend/src/app.d.ts"
    "frontend/src/test-setup.ts" "frontend/static/favicon.svg"

    # Frontend lib
    "frontend/src/lib/types/index.ts"
    "frontend/src/lib/stores/index.ts"
    "frontend/src/lib/hooks/index.ts"
    "frontend/src/lib/utils/index.ts"
    "frontend/src/lib/utils/index.test.ts"

    # UI Components
    "frontend/src/lib/components/ui/index.ts"
    "frontend/src/lib/components/ui/Button.svelte"
    "frontend/src/lib/components/ui/Card.svelte"
    "frontend/src/lib/components/ui/Input.svelte"
    "frontend/src/lib/components/ui/Badge.svelte"
    "frontend/src/lib/components/ui/Modal.svelte"
    "frontend/src/lib/components/ui/Toast.svelte"
    "frontend/src/lib/components/ui/Progress.svelte"
    "frontend/src/lib/components/ui/Avatar.svelte"
    "frontend/src/lib/components/ui/Dropdown.svelte"
    "frontend/src/lib/components/ui/Tabs.svelte"

    # Arena Components
    "frontend/src/lib/components/arena/AgentPanel.svelte"
    "frontend/src/lib/components/arena/Timeline.svelte"

    # Editor Components
    "frontend/src/lib/components/editor/index.ts"
    "frontend/src/lib/components/editor/CodeEditor.svelte"
    "frontend/src/lib/components/editor/diff.ts"
    "frontend/src/lib/components/editor/DiffView.svelte"

    # Challenges Components
    "frontend/src/lib/components/challenges/index.ts"
    "frontend/src/lib/components/challenges/ChallengeCard.svelte"
    "frontend/src/lib/components/challenges/ChallengeList.svelte"
    "frontend/src/lib/components/challenges/ChallengeForm.svelte"

    # Leaderboard Components
    "frontend/src/lib/components/leaderboard/index.ts"
    "frontend/src/lib/components/leaderboard/LeaderboardTable.svelte"
    "frontend/src/lib/components/leaderboard/LeaderboardCharts.svelte"

    # Snippets Components
    "frontend/src/lib/components/snippets/index.ts"
    "frontend/src/lib/components/snippets/SnippetCard.svelte"
    "frontend/src/lib/components/snippets/SnippetList.svelte"

    # Profile Components
    "frontend/src/lib/components/profile/index.ts"
    "frontend/src/lib/components/profile/ProfileStats.svelte"

    # Components Index
    "frontend/src/lib/components/index.ts"

    # Routes
    "frontend/src/routes/+layout.svelte"
    "frontend/src/routes/+page.svelte"
    "frontend/src/routes/arena/+page.svelte"
    "frontend/src/routes/arena/[matchId]/+page.svelte"
    "frontend/src/routes/challenges/+page.svelte"
    "frontend/src/routes/challenges/create/+page.svelte"
    "frontend/src/routes/leaderboard/+page.svelte"
    "frontend/src/routes/snippets/+page.svelte"
    "frontend/src/routes/profile/+page.svelte"
    "frontend/src/routes/auth/login/+page.svelte"
    "frontend/src/routes/auth/register/+page.svelte"

    # Backend
    "backend/go.mod" "backend/go.sum" "backend/Dockerfile" "backend/.air.toml" "backend/sqlc.yaml"
    "backend/cmd/api/main.go" "backend/cmd/migrate/main.go" "backend/cmd/seed/main.go"

    # Backend Internal
    "backend/internal/config/config.go"
    "backend/internal/database/database.go" "backend/internal/database/models.go"
    "backend/internal/database/repositories.go" "backend/internal/database/repositories_extended.go"
    "backend/internal/database/queries/user.sql" "backend/internal/database/queries/match.sql"
    "backend/internal/database/queries/challenge.sql" "backend/internal/database/queries/challenge_template.sql"
    "backend/internal/database/queries/snippet.sql" "backend/internal/database/queries/leaderboard.sql"
    "backend/internal/auth/auth.go" "backend/internal/matchmaking/matchmaking.go"
    "backend/internal/sandbox/sandbox.go" "backend/internal/agents/personas.go"
    "backend/internal/challenges/challenges.go" "backend/internal/snippets/snippets.go"
    "backend/internal/leaderboard/leaderboard.go" "backend/internal/websocket/hub.go"

    # Backend Packages
    "backend/pkg/ratelimit/ratelimit.go" "backend/pkg/docker/docker.go" "backend/pkg/freellm/client.go"

    # Migrations
    "backend/migrations/000001_initial_schema.sql"

    # Sandbox (9 languages)
    "sandbox/cpp/Dockerfile" "sandbox/python/Dockerfile" "sandbox/java/Dockerfile"
    "sandbox/javascript/Dockerfile" "sandbox/typescript/Dockerfile" "sandbox/rust/Dockerfile"
    "sandbox/shell/Dockerfile" "sandbox/html/Dockerfile" "sandbox/gherkin/Dockerfile"

    # Scripts
    "scripts/dev.sh" "scripts/sync-snippets.sh" "scripts/verify-repo.sh"

    # Docs
    "docs/architecture.md" "docs/api-spec.md" "docs/agent-personas.md"
    "docs/challenge-format.md" "docs/sandbox-spec.md"

    # Monitoring
    "monitoring/prometheus.yml" "monitoring/loki.yml" "monitoring/promtail.yml"
    "monitoring/grafana/dashboards/codecrashreview.json"
    "monitoring/grafana/datasources/datasources.yml"

    # Nginx & CI/CD
    "nginx/nginx.conf" ".github/workflows/ci.yml"

    # Shared
    "shared/constants/constants.go"
)

REQUIRED_DIRS=(
    "frontend/src/lib/components/ui" "frontend/src/lib/components/arena"
    "frontend/src/lib/components/editor" "frontend/src/lib/components/challenges"
    "frontend/src/lib/components/leaderboard" "frontend/src/lib/components/snippets"
    "frontend/src/lib/components/profile" "frontend/src/routes/arena"
    "frontend/src/routes/challenges" "frontend/src/routes/leaderboard"
    "frontend/src/routes/snippets" "frontend/src/routes/profile" "frontend/src/routes/auth"
    "backend/cmd/api" "backend/cmd/migrate" "backend/cmd/seed"
    "backend/internal/config" "backend/internal/database" "backend/internal/database/queries"
    "backend/internal/auth" "backend/internal/matchmaking" "backend/internal/sandbox"
    "backend/internal/agents" "backend/internal/challenges" "backend/internal/snippets"
    "backend/internal/leaderboard" "backend/internal/websocket"
    "backend/pkg/ratelimit" "backend/pkg/docker" "backend/pkg/freellm"
    "backend/migrations" "sandbox/cpp" "sandbox/python" "sandbox/java"
    "sandbox/javascript" "sandbox/typescript" "sandbox/rust" "sandbox/shell"
    "sandbox/html" "sandbox/gherkin" "scripts" "docs"
    "monitoring/grafana/dashboards" "monitoring/grafana/datasources"
    "nginx" ".github/workflows" "shared/constants"
    "frontend/src/lib/components" "frontend/src/lib" "frontend/src/routes"
    "frontend/src" "frontend/static" "frontend" "backend/internal" "backend/pkg"
    "backend/cmd" "backend" "sandbox" "scripts" "docs" "monitoring"
    "monitoring/grafana" "shared"
)

# ============================================
# MAIN EXECUTION
# ============================================

echo -e "\n============================================"
echo -e "  CodeCrashReview - Repository Verification"
echo -e "============================================\n"

missing_files=()
missing_dirs=()
found_files=0
found_dirs=0

# Check files
echo -e "\n[1/3] Checking required files...\n"
for file in "${REQUIRED_FILES[@]}"; do
    if check_path "$file" "file"; then
        ((found_files++))
        [[ "$VERBOSE" == true ]] && status "  ✓ $file" "$GREEN" "  "
    else
        missing_files+=("$file")
        status "  ✗ $file (MISSING)" "$RED" "  "
    fi
done

# Check directories
echo -e "\n[2/3] Checking required directories...\n"
for dir in "${REQUIRED_DIRS[@]}"; do
    if check_path "$dir" "dir"; then
        ((found_dirs++))
        [[ "$VERBOSE" == true ]] && status "  ✓ $dir/" "$GREEN" "  "
    else
        missing_dirs+=("$dir")
        status "  ✗ $dir/ (MISSING)" "$RED" "  "
    fi
done

# Summary
echo -e "\n[3/3] Verification Summary\n"
echo "============================================"

total_files=${#REQUIRED_FILES[@]}
total_dirs=${#REQUIRED_DIRS[@]}
missing_files_count=${#missing_files[@]}
missing_dirs_count=${#missing_dirs[@]}

echo "Files: $found_files / $total_files found"
echo "Directories: $found_dirs / $total_dirs found"

if [[ $missing_files_count -gt 0 ]]; then
    status "\nMissing files ($missing_files_count):" "$RED"
    for f in "${missing_files[@]}"; do
        status "  - $f" "$RED"
    done
fi

if [[ $missing_dirs_count -gt 0 ]]; then
    status "\nMissing directories ($missing_dirs_count):" "$RED"
    for d in "${missing_dirs[@]}"; do
        status "  - $d/" "$RED"
    done
fi

exit_code=0
if [[ $missing_files_count -eq 0 && $missing_dirs_count -eq 0 ]]; then
    status "\n✅ ALL GOOD! Repository is complete and intact.\n" "$GREEN"
    exit_code=0
else
    status "\n❌ REPOSITORY INCOMPLETE - $missing_files_count files and $missing_dirs_count directories missing.\n" "$RED"
    exit_code=1
fi

# Fix missing directories if requested
if [[ "$FIX_MISSING" == true && $missing_dirs_count -gt 0 ]]; then
    echo -e "\n[FIX] Creating missing directories...\n"
    for d in "${missing_dirs[@]}"; do
        mkdir -p "$d"
        if [[ -d "$d" ]]; then
            status "  ✓ Created: $d" "$GREEN"
        else
            status "  ✗ Failed to create: $d" "$RED"
        fi
    done
fi

# Git status check
echo -e "\n--- Git Status ---"
if git rev-parse --git-dir > /dev/null 2>&1; then
    git_status=$(git status --porcelain 2>/dev/null)
    if [[ -n "$git_status" ]]; then
        status "  ⚠ Uncommitted changes:" "$YELLOW"
        echo "$git_status" | while IFS= read -r line; do
            echo "    $line"
        done
    else
        status "  ✓ Working tree clean" "$GREEN"
    fi

    remote_url=$(git remote get-url origin 2>/dev/null || true)
    if [[ -n "$remote_url" ]]; then
        status "  ✓ Remote origin: $remote_url" "$GREEN"
    else
        status "  ⚠ Remote 'origin' not configured" "$YELLOW"
    fi
else
    status "  ❌ Not a Git repository" "$RED"
fi

echo -e "\n============================================"
exit $exit_code
