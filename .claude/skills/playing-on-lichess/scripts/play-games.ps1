# Plays a set number of games on Lichess, then stops the bot and prints the score.
# lichess-bot has no game limit, so this counts the PGN file it writes per finished game.
#
# Usage: play-games.ps1 -Games <1-100> -BotDir <lichess-bot checkout>
#                       [-Mode casual|rated|random] [-TimeControl 60+1]
param(
    [Parameter(Mandatory)][ValidateRange(1, 100)][int]$Games,
    [string]$BotDir = $env:LICHESS_BOT_DIR,
    [ValidateSet("casual", "rated", "random")][string]$Mode,
    [ValidatePattern('^\d+\+\d+$')][string]$TimeControl,
    [string]$RunScript,
    [string]$BotName = "GolangChessBot",
    [int]$PollSeconds = 5
)

if (-not $BotDir) { Write-Error "Pass -BotDir or set LICHESS_BOT_DIR"; exit 1 }
if (-not $RunScript) { $RunScript = Join-Path $PSScriptRoot "..\..\..\..\tools\lichess\run.ps1" }

$records = Join-Path $BotDir "game_records"
$logs = Join-Path $BotDir "logs"
New-Item -ItemType Directory -Force -Path $records, $logs | Out-Null

function Get-PgnNames { @(Get-ChildItem -Path $records -Filter *.pgn | ForEach-Object { $_.Name }) }

$before = Get-PgnNames

$arguments = @("-NoProfile", "-ExecutionPolicy", "Bypass", "-File", "`"$RunScript`"", "-BotDir", "`"$BotDir`"")
if ($Mode) { $arguments += @("-Mode", $Mode) }
if ($TimeControl) { $arguments += @("-TimeControl", $TimeControl) }

$outLog = Join-Path $logs "play-games.log"
$errLog = Join-Path $logs "play-games.err.log"
$bot = Start-Process -FilePath "powershell.exe" -ArgumentList $arguments -PassThru -WindowStyle Hidden `
    -RedirectStandardOutput $outLog -RedirectStandardError $errLog

Write-Host "Bot started (pid $($bot.Id)), playing $Games game(s)"

$finished = @()
while ($finished.Count -lt $Games) {
    if ($bot.HasExited) {
        Write-Host "Bot exited early after $($finished.Count) game(s). Last output:"
        Get-Content $outLog, $errLog -Tail 20 -ErrorAction SilentlyContinue
        exit 1
    }
    Start-Sleep -Seconds $PollSeconds
    $finished = @(Get-PgnNames | Where-Object { $before -notcontains $_ })
}

& taskkill.exe /PID $bot.Id /T /F | Out-Null
Write-Host "Bot stopped after $($finished.Count) game(s)"

$wins = 0; $losses = 0; $draws = 0
foreach ($name in $finished) {
    $headers = Get-Content -LiteralPath (Join-Path $records $name) -TotalCount 30
    $white = ($headers | Select-String '^\[White "(.*)"\]').Matches.Groups[1].Value
    $result = ($headers | Select-String '^\[Result "(.*)"\]').Matches.Groups[1].Value
    $botIsWhite = $white -eq $BotName

    switch ($result) {
        "1/2-1/2" { $draws++ }
        "1-0" { if ($botIsWhite) { $wins++ } else { $losses++ } }
        "0-1" { if ($botIsWhite) { $losses++ } else { $wins++ } }
    }
    Write-Host "  $name  $result"
}

Write-Host "Score: +$wins -$losses =$draws"
