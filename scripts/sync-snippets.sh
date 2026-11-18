#!/bin/bash
# Sync snippets from upstream repositories
# Usage: ./sync-snippets.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SNIPPETS_DIR="$PROJECT_ROOT/frontend/static/snippets"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${BLUE}[INFO]${NC} $*"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $*"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# Create snippets directory
mkdir -p "$SNIPPETS_DIR"

# Repository configurations
declare -A REPOS=(
    ["awesome-algorithms"]="https://github.com/tayllan/awesome-algorithms.git"
    ["cp-algorithms"]="https://github.com/cp-algorithms/cp-algorithms.git"
    ["programming-idioms"]="https://github.com/programming-idioms/idioms.git"
    ["leetcode-patterns"]="https://github.com/haoel/leetcode.git"
)

# Language extensions mapping
declare -A LANG_EXT=(
    ["cpp"]="cpp cc cxx h hpp"
    ["python"]="py pyw"
    ["java"]="java"
    ["javascript"]="js mjs cjs"
    ["typescript"]="ts tsx"
    ["rust"]="rs"
    ["go"]="go"
    ["shell"]="sh bash zsh fish"
)

# Clone or update repository
sync_repo() {
    local name=$1
    local url=$2
    local repo_dir="$SNIPPETS_DIR/$name"

    log_info "Syncing $name from $url"

    if [ -d "$repo_dir/.git" ]; then
        cd "$repo_dir"
        git fetch --quiet origin
        git reset --quiet --hard origin/main 2>/dev/null || git reset --quiet --hard origin/master
        log_success "Updated $name"
    else
        git clone --quiet --depth 1 "$url" "$repo_dir" 2>/dev/null || {
            log_error "Failed to clone $name"
            return 1
        }
        log_success "Cloned $name"
    fi
}

# Extract snippets from repository
extract_snippets() {
    local name=$1
    local repo_dir="$SNIPPETS_DIR/$name"
    local output_dir="$SNIPPETS_DIR/extracted/$name"

    log_info "Extracting snippets from $name"

    mkdir -p "$output_dir"

    # Find code files
    find "$repo_dir" -type f \( \
        -name "*.cpp" -o -name "*.cc" -o -name "*.cxx" -o -name "*.h" -o -name "*.hpp" \
        -o -name "*.py" -o -name "*.java" \
        -o -name "*.js" -o -name "*.ts" -o -name "*.jsx" -o -name "*.tsx" \
        -o -name "*.rs" -o -name "*.go" \
        -o -name "*.sh" -o -name "*.bash" \
    \) | while read -r file; do
        # Get relative path
        rel_path="${file#$repo_dir/}"

        # Determine language
        ext="${file##*.}"
        lang=""
        for l in "${!LANG_EXT[@]}"; do
            if [[ " ${LANG_EXT[$l]} " =~ " ${ext} " ]]; then
                lang=$l
                break
            fi
        done

        if [ -z "$lang" ]; then
            continue
        fi

        # Create snippet metadata
        snippet_name=$(basename "$file" ."$ext")
        snippet_file="$output_dir/${snippet_name}.json"

        # Read file content (limit size)
        content=$(head -c 50000 "$file")

        # Create JSON snippet
        cat > "$snippet_file" <<EOF
{
    "id": "$(echo -n "$name/$rel_path" | sha256sum | cut -c1-12)",
    "title": "$snippet_name",
    "description": "From $name: $rel_path",
    "language": "$lang",
    "code": $(echo "$content" | jq -Rs .),
    "tags": ["$name", "$lang", "algorithm"],
    "complexity": "intermediate",
    "author": "$name",
    "source_url": "https://github.com/$(echo "$name" | sed 's/-algorithms//')/blob/main/$rel_path",
    "rating": 0,
    "usage_count": 0,
    "is_official": true,
    "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "updated_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
    done

    log_success "Extracted snippets from $name to $output_dir"
}

# Generate index
generate_index() {
    local index_file="$SNIPPETS_DIR/index.json"

    log_info "Generating snippet index"

    echo '{"snippets": [' > "$index_file"
    first=true

    find "$SNIPPETS_DIR/extracted" -name "*.json" | while read -r file; do
        if [ "$first" = true ]; then
            first=false
        else
            echo "," >> "$index_file"
        fi
        cat "$file" >> "$index_file"
    done

    echo ']}' >> "$index_file"

    log_success "Generated index at $index_file"
}

# Main
main() {
    log_info "Starting snippet sync..."

    # Sync repositories
    for name in "${!REPOS[@]}"; do
        sync_repo "$name" "${REPOS[$name]}"
    done

    # Extract snippets
    for name in "${!REPOS[@]}"; do
        extract_snippets "$name"
    done

    # Generate index
    generate_index

    log_success "Snippet sync complete!"
    log_info "Snippets available at: $SNIPPETS_DIR/extracted"
    log_info "Index at: $SNIPPETS_DIR/index.json"
}

# Check dependencies
for cmd in git jq sha256sum; do
    if ! command -v "$cmd" &> /dev/null; then
        log_error "Required command not found: $cmd"
        exit 1
    fi
done

main "$@"