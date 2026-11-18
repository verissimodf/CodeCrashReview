<#
.SYNOPSIS
    Verifica a integridade do repositório CodeCrashReview
.DESCRIPTION
    Valida se todos os arquivos e pastas esperados existem no repositório
.NOTES
    Execute na raiz do repositório: .\verify-repo.ps1
    Opções: -Verbose (mostra arquivos encontrados) -FixMissing (cria dirs ausentes)
#>

param(
    [switch]$Verbose,
    [switch]$FixMissing
)

# Cores para output
$Colors = @{
    Green  = [ConsoleColor]::Green
    Red    = [ConsoleColor]::Red
    Yellow = [ConsoleColor]::Yellow
    Cyan   = [ConsoleColor]::Cyan
    Gray   = [ConsoleColor]::DarkGray
}

function Write-Status {
    param($Message, $Color = 'Gray', $Prefix = '')
    $originalColor = $Host.UI.RawUI.ForegroundColor
    $Host.UI.RawUI.ForegroundColor = $Colors[$Color]
    Write-Host "$Prefix$Message"
    $Host.UI.RawUI.ForegroundColor = $originalColor
}

function Test-PathEx {
    param($Path, $Type = 'Any')
    $fullPath = Join-Path $PSScriptRoot $Path
    $exists = Test-Path $fullPath -PathType $Type
    return $exists
}

# ============================================
# ESTRUTURA ESPERADA DO REPOSITÓRIO
# ============================================

$RequiredFiles = @(
    # Root files
    "README.md", "CLAUDE.md", "docker-compose.yml", "docker-compose.prod.yml",
    ".env.example", "Makefile", ".gitignore", "LICENSE", "PROJECT_SUMMARY.md",

    # Frontend
    "frontend/package.json", "frontend/svelte.config.js", "frontend/tailwind.config.js",
    "frontend/vite.config.ts", "frontend/tsconfig.json", "frontend/postcss.config.js",
    "frontend/Dockerfile", "frontend/nginx.conf", "frontend/vitest.config.ts",
    "frontend/src/app.html", "frontend/src/app.css", "frontend/src/app.d.ts",
    "frontend/src/test-setup.ts", "frontend/static/favicon.svg",

    # Frontend lib
    "frontend/src/lib/types/index.ts",
    "frontend/src/lib/stores/index.ts",
    "frontend/src/lib/hooks/index.ts",
    "frontend/src/lib/utils/index.ts",
    "frontend/src/lib/utils/index.test.ts",

    # UI Components
    "frontend/src/lib/components/ui/index.ts",
    "frontend/src/lib/components/ui/Button.svelte",
    "frontend/src/lib/components/ui/Card.svelte",
    "frontend/src/lib/components/ui/Input.svelte",
    "frontend/src/lib/components/ui/Badge.svelte",
    "frontend/src/lib/components/ui/Modal.svelte",
    "frontend/src/lib/components/ui/Toast.svelte",
    "frontend/src/lib/components/ui/Progress.svelte",
    "frontend/src/lib/components/ui/Avatar.svelte",
    "frontend/src/lib/components/ui/Dropdown.svelte",
    "frontend/src/lib/components/ui/Tabs.svelte",

    # Arena Components
    "frontend/src/lib/components/arena/AgentPanel.svelte",
    "frontend/src/lib/components/arena/Timeline.svelte",

    # Editor Components
    "frontend/src/lib/components/editor/index.ts",
    "frontend/src/lib/components/editor/CodeEditor.svelte",
    "frontend/src/lib/components/editor/diff.ts",
    "frontend/src/lib/components/editor/DiffView.svelte",

    # Challenges Components
    "frontend/src/lib/components/challenges/index.ts",
    "frontend/src/lib/components/challenges/ChallengeCard.svelte",
    "frontend/src/lib/components/challenges/ChallengeList.svelte",
    "frontend/src/lib/components/challenges/ChallengeForm.svelte",

    # Leaderboard Components
    "frontend/src/lib/components/leaderboard/index.ts",
    "frontend/src/lib/components/leaderboard/LeaderboardTable.svelte",
    "frontend/src/lib/components/leaderboard/LeaderboardCharts.svelte",

    # Snippets Components
    "frontend/src/lib/components/snippets/index.ts",
    "frontend/src/lib/components/snippets/SnippetCard.svelte",
    "frontend/src/lib/components/snippets/SnippetList.svelte",

    # Profile Components
    "frontend/src/lib/components/profile/index.ts",
    "frontend/src/lib/components/profile/ProfileStats.svelte",

    # Components Index
    "frontend/src/lib/components/index.ts",

    # Routes
    "frontend/src/routes/+layout.svelte",
    "frontend/src/routes/+page.svelte",
    "frontend/src/routes/arena/+page.svelte",
    "frontend/src/routes/arena/[matchId]/+page.svelte",
    "frontend/src/routes/challenges/+page.svelte",
    "frontend/src/routes/challenges/create/+page.svelte",
    "frontend/src/routes/leaderboard/+page.svelte",
    "frontend/src/routes/snippets/+page.svelte",
    "frontend/src/routes/profile/+page.svelte",
    "frontend/src/routes/auth/login/+page.svelte",
    "frontend/src/routes/auth/register/+page.svelte",

    # Backend
    "backend/go.mod", "backend/go.sum", "backend/Dockerfile", "backend/.air.toml", "backend/sqlc.yaml",
    "backend/cmd/api/main.go", "backend/cmd/migrate/main.go", "backend/cmd/seed/main.go",

    # Backend Internal
    "backend/internal/config/config.go",
    "backend/internal/database/database.go", "backend/internal/database/models.go",
    "backend/internal/database/repositories.go", "backend/internal/database/repositories_extended.go",
    "backend/internal/database/queries/user.sql", "backend/internal/database/queries/match.sql",
    "backend/internal/database/queries/challenge.sql", "backend/internal/database/queries/challenge_template.sql",
    "backend/internal/database/queries/snippet.sql", "backend/internal/database/queries/leaderboard.sql",
    "backend/internal/auth/auth.go", "backend/internal/matchmaking/matchmaking.go",
    "backend/internal/sandbox/sandbox.go", "backend/internal/agents/personas.go",
    "backend/internal/challenges/challenges.go", "backend/internal/snippets/snippets.go",
    "backend/internal/leaderboard/leaderboard.go", "backend/internal/websocket/hub.go",

    # Backend Packages
    "backend/pkg/ratelimit/ratelimit.go", "backend/pkg/docker/docker.go", "backend/pkg/freellm/client.go",

    # Migrations
    "backend/migrations/000001_initial_schema.sql",

    # Sandbox (9 languages)
    "sandbox/cpp/Dockerfile", "sandbox/python/Dockerfile", "sandbox/java/Dockerfile",
    "sandbox/javascript/Dockerfile", "sandbox/typescript/Dockerfile", "sandbox/rust/Dockerfile",
    "sandbox/shell/Dockerfile", "sandbox/html/Dockerfile", "sandbox/gherkin/Dockerfile",

    # Scripts
    "scripts/dev.sh", "scripts/sync-snippets.sh", "scripts/verify-repo.ps1",

    # Docs
    "docs/architecture.md", "docs/api-spec.md", "docs/agent-personas.md",
    "docs/challenge-format.md", "docs/sandbox-spec.md",

    # Monitoring
    "monitoring/prometheus.yml", "monitoring/loki.yml", "monitoring/promtail.yml",
    "monitoring/grafana/dashboards/codecrashreview.json",
    "monitoring/grafana/datasources/datasources.yml",

    # Nginx & CI/CD
    "nginx/nginx.conf", ".github/workflows/ci.yml",

    # Shared
    "shared/constants/constants.go"
)

$RequiredDirs = @(
    "frontend/src/lib/components/ui", "frontend/src/lib/components/arena",
    "frontend/src/lib/components/editor", "frontend/src/lib/components/challenges",
    "frontend/src/lib/components/leaderboard", "frontend/src/lib/components/snippets",
    "frontend/src/lib/components/profile", "frontend/src/routes/arena",
    "frontend/src/routes/challenges", "frontend/src/routes/leaderboard",
    "frontend/src/routes/snippets", "frontend/src/routes/profile", "frontend/src/routes/auth",
    "backend/cmd/api", "backend/cmd/migrate", "backend/cmd/seed",
    "backend/internal/config", "backend/internal/database", "backend/internal/database/queries",
    "backend/internal/auth", "backend/internal/matchmaking", "backend/internal/sandbox",
    "backend/internal/agents", "backend/internal/challenges", "backend/internal/snippets",
    "backend/internal/leaderboard", "backend/internal/websocket",
    "backend/pkg/ratelimit", "backend/pkg/docker", "backend/pkg/freellm",
    "backend/migrations", "sandbox/cpp", "sandbox/python", "sandbox/java",
    "sandbox/javascript", "sandbox/typescript", "sandbox/rust", "sandbox/shell",
    "sandbox/html", "sandbox/gherkin", "scripts", "docs",
    "monitoring/grafana/dashboards", "monitoring/grafana/datasources",
    "nginx", ".github/workflows", "shared/constants",
    "frontend/src/lib/components", "frontend/src/lib", "frontend/src/routes",
    "frontend/src", "frontend/static", "frontend", "backend/internal", "backend/pkg",
    "backend/cmd", "backend", "sandbox", "scripts", "docs", "monitoring",
    "monitoring/grafana", "shared"
)

# ============================================
# EXECUÇÃO
# ============================================

Write-Host "`n============================================"
Write-Host "  CodeCrashReview - Repository Verification"
Write-Host "============================================`n"

$missingFiles = @()
$missingDirs = @()
$foundFiles = 0
$foundDirs = 0

Write-Host "`n[1/3] Verificando arquivos obrigatórios...`n"
foreach ($file in $RequiredFiles) {
    if (Test-PathEx $file) { $foundFiles++; if ($Verbose) { Write-Status "  ✓ $file" 'Green' '  ' } }
    else { $missingFiles += $file; Write-Status "  ✗ $file (AUSENTE)" 'Red' '  ' }
}

Write-Host "`n[2/3] Verificando diretórios obrigatórios...`n"
foreach ($dir in $RequiredDirs) {
    if (Test-PathEx $dir 'Container') { $foundDirs++; if ($Verbose) { Write-Status "  ✓ $dir/" 'Green' '  ' } }
    else { $missingDirs += $dir; Write-Status "  ✗ $dir/ (AUSENTE)" 'Red' '  ' }
}

Write-Host "`n[3/3] Resumo da Verificação`n"
Write-Host "============================================"

$totalFiles = $RequiredFiles.Count; $totalDirs = $RequiredDirs.Count
$missingFilesCount = $missingFiles.Count; $missingDirsCount = $missingDirs.Count

Write-Host "Arquivos: $foundFiles / $totalFiles encontrados"
Write-Host "Diretórios: $foundDirs / $totalDirs encontrados"

if ($missingFilesCount -gt 0) {
    Write-Status "`nArquivos ausentes ($missingFilesCount):" 'Red'
    foreach ($f in $missingFiles) { Write-Status "  - $f" 'Red' }
}
if ($missingDirsCount -gt 0) {
    Write-Status "`nDiretórios ausentes ($missingDirsCount):" 'Red'
    foreach ($d in $missingDirs) { Write-Status "  - $d/" 'Red' }
}

if ($missingFilesCount -eq 0 -and $missingDirsCount -eq 0) {
    Write-Status "`n✅ TUDO OK! Repositório completo e íntegro.`n" 'Green'; $exitCode = 0
} else {
    Write-Status "`n❌ REPOSITÓRIO INCOMPLETO - $missingFilesCount arquivos e $missingDirsCount diretórios ausentes.`n" 'Red'; $exitCode = 1
}

if ($FixMissing -and ($missingDirsCount -gt 0)) {
    Write-Host "`n[FIX] Criando diretórios ausentes...`n" -ForegroundColor Yellow
    foreach ($d in $missingDirs) {
        $fullPath = Join-Path $PSScriptRoot $d
        try { New-Item -ItemType Directory -Path $fullPath -Force | Out-Null; Write-Status "  ✓ Criado: $d" 'Green' }
        catch { Write-Status "  ✗ Erro ao criar: $d" 'Red' }
    }
}

Write-Host "`n--- Git Status ---"
$gitStatus = git status --porcelain 2>$null
if ($LASTEXITCODE -eq 0) {
    if ($gitStatus) { Write-Status "  ⚠ Alterações não commitadas:" 'Yellow'; $gitStatus.Split("`n") | ForEach-Object { Write-Host "    $_" } }
    else { Write-Status "  ✓ Working tree limpo" 'Green' }
    $remote = git remote get-url origin 2>$null
    if ($remote) { Write-Status "  ✓ Remote origin: $remote" 'Green' } else { Write-Status "  ⚠ Remote 'origin' não configurado" 'Yellow' }
} else { Write-Status "  ❌ Não é um repositório Git" 'Red' }

Write-Host "`n============================================"
exit $exitCode