$ErrorActionPreference = 'Stop'
param([string]$Dir = "$PWD\CodeCrashReview-validation", [string]$FreeLlmUrl = 'http://host.docker.internal:8082/v1', [string]$FreeLlmKey = 'local-validation')
docker version | Out-Null; docker compose version | Out-Null
if (-not (Test-Path "$Dir\.git")) { git clone --branch Develop --depth 1 https://github.com/verissimodf/CodeCrashReview.git $Dir }
Set-Location $Dir
if (-not (Test-Path .env.validation)) {
  $rng = [Convert]::ToHexString((1..32 | ForEach-Object { Get-Random -Maximum 256 }))
  @("POSTGRES_USER=ccr_validation", "POSTGRES_DB=ccr_validation", "POSTGRES_PASSWORD=$rng", "REDIS_PASSWORD=$rng", "JWT_SECRET=$rng", "FREELLM_BASE_URL=$FreeLlmUrl", "FREELLM_API_KEY=$FreeLlmKey", "DEFAULT_MODEL=claude-3-freellm-no-thinking/ollama/qwen2.5:7b", "VALIDATOR_IMAGE=golang@sha256:ecb67f987aa1a57794c91d19ade7ae0209f2f0f5496aaef9e6ae85ddf92d02ab", "VALIDATOR_PYTHON_IMAGE=python@sha256:9ef1a599f863231c3d91e941389bfdc4a2d6070557243d54df1a4e57015b960e", "VALIDATOR_NODE_IMAGE=node@sha256:1314fb30ec37592d5845e5a160d522c0e680972aa534549940b8d73bde88481a") | Set-Content .env.validation
}
docker compose --env-file .env.validation -f compose.validation.yml up -d --build
Write-Host "CodeCrashReview disponível em http://127.0.0.1:18080"
