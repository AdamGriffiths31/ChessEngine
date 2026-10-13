# Starts lichess-bot with the token read from the repo's .env file.
# Also turns on the engine's structured move log (JSON lines, one record per move
# plus the UCI transcript), appended to <BotDir>\logs\engine.jsonl by default.
# Usage: .\run.ps1 [-BotDir <lichess-bot checkout>] [-EnvFile <path>] [-EngineLog <path>]
param(
    [string]$BotDir = $env:LICHESS_BOT_DIR,
    [string]$EnvFile = (Join-Path $PSScriptRoot "..\..\.env"),
    [string]$EngineLog
)

if (-not $BotDir) { Write-Error "Pass -BotDir or set LICHESS_BOT_DIR to your lichess-bot checkout"; exit 1 }
if (-not (Test-Path $EnvFile)) { Write-Error "Env file not found: $EnvFile"; exit 1 }

if (-not $EngineLog) { $EngineLog = Join-Path $BotDir "logs\engine.jsonl" }
New-Item -ItemType Directory -Force -Path (Split-Path $EngineLog) | Out-Null
$env:CHESS_DEBUG_LOG = $EngineLog

$line = Get-Content $EnvFile | Where-Object { $_ -match '^\s*LICHESS_TOKEN\s*=' } | Select-Object -First 1
if (-not $line) { Write-Error "LICHESS_TOKEN not found in $EnvFile"; exit 1 }
$env:LICHESS_BOT_TOKEN = ($line -split '=', 2)[1].Trim()

# lichess-bot resolves ./config.yml and ./engines/ against the working directory.
Push-Location $BotDir
try {
    & (Join-Path $BotDir ".venv\Scripts\python.exe") (Join-Path $BotDir "lichess-bot.py") @args
} finally {
    Pop-Location
}
