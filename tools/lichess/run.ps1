# Starts lichess-bot with the token read from the repo's .env file.
# Also turns on the engine's structured move log (JSON lines, one record per move
# plus the UCI transcript), appended to <BotDir>\logs\engine.jsonl by default.
#
# -Mode and -TimeControl patch a generated copy of <BotDir>\config.yml (written to
# <BotDir>\logs\config.generated.yml) and launch with --config, so config.yml itself is
# never edited. With neither switch the bot runs on config.yml unchanged.
#   -Mode casual|rated|random   what the bot sends (matchmaking.challenge_mode) and accepts
#                               (challenge.modes); random sends and accepts both
#   -TimeControl BASE+INC       matchmaking challenges use only this control, in seconds, e.g. 60+1
#
# Usage: .\run.ps1 [-BotDir <lichess-bot checkout>] [-EnvFile <path>] [-EngineLog <path>]
#                  [-Mode casual|rated|random] [-TimeControl 60+1]
param(
    [string]$BotDir = $env:LICHESS_BOT_DIR,
    [string]$EnvFile = (Join-Path $PSScriptRoot "..\..\.env"),
    [string]$EngineLog,
    [ValidateSet("casual", "rated", "random")]
    [string]$Mode,
    [ValidatePattern('^\d+\+\d+$')]
    [string]$TimeControl
)

if (-not $BotDir) { Write-Error "Pass -BotDir or set LICHESS_BOT_DIR to your lichess-bot checkout"; exit 1 }
if (-not (Test-Path $EnvFile)) { Write-Error "Env file not found: $EnvFile"; exit 1 }

if (-not $EngineLog) { $EngineLog = Join-Path $BotDir "logs\engine.jsonl" }
New-Item -ItemType Directory -Force -Path (Split-Path $EngineLog) | Out-Null
$env:CHESS_DEBUG_LOG = $EngineLog

$line = Get-Content $EnvFile | Where-Object { $_ -match '^\s*LICHESS_TOKEN\s*=' } | Select-Object -First 1
if (-not $line) { Write-Error "LICHESS_TOKEN not found in $EnvFile"; exit 1 }
$env:LICHESS_BOT_TOKEN = ($line -split '=', 2)[1].Trim()

# Rewrites the list items under "<key>:" with the given values, keeping the key line.
function Set-ConfigList([string[]]$Lines, [string]$Key, [string[]]$Values) {
    $out = New-Object System.Collections.Generic.List[string]
    $skipping = $false
    $found = $false
    foreach ($l in $Lines) {
        if ($skipping -and $l -match '^\s*-\s') { continue }
        $skipping = $false
        $out.Add($l)
        if ($l -match "^(\s*)${Key}:") {
            $found = $true
            $skipping = $true
            $indent = $Matches[1] + "  "
            foreach ($v in $Values) { $out.Add("$indent- $v") }
        }
    }
    if (-not $found) { Write-Error "config key '$Key' not found"; exit 1 }
    return , $out.ToArray()
}

$configArgs = @()
if ($Mode -or $TimeControl) {
    $baseConfig = Join-Path $BotDir "config.yml"
    if (-not (Test-Path $baseConfig)) { Write-Error "Config not found: $baseConfig"; exit 1 }
    $lines = [string[]](Get-Content $baseConfig)

    if ($Mode) {
        $accepted = @{ casual = @("casual"); rated = @("rated"); random = @("casual", "rated") }[$Mode]
        $lines = Set-ConfigList $lines "modes" $accepted
        $lines = $lines | ForEach-Object { $_ -replace '^(\s*challenge_mode:\s*)"[^"]*"', "`$1`"$Mode`"" }
    }
    if ($TimeControl) {
        $base, $inc = $TimeControl -split '\+'
        $lines = Set-ConfigList $lines "challenge_initial_time" @($base)
        $lines = Set-ConfigList $lines "challenge_increment" @($inc)
    }

    $generated = Join-Path $BotDir "logs\config.generated.yml"
    Set-Content -Path $generated -Value $lines -Encoding utf8
    $configArgs = @("--config", $generated)
    Write-Host "lichess-bot config: $generated (mode=$(if ($Mode) { $Mode } else { 'config.yml' }), timecontrol=$(if ($TimeControl) { $TimeControl } else { 'config.yml' }))"
}

# lichess-bot resolves ./config.yml and ./engines/ against the working directory.
Push-Location $BotDir
try {
    & (Join-Path $BotDir ".venv\Scripts\python.exe") (Join-Path $BotDir "lichess-bot.py") @configArgs @args
} finally {
    Pop-Location
}
