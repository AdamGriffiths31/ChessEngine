---
name: playing-on-lichess
description: Use when the user asks to play games on Lichess, run or stop GolangChessBot, or update the engine the bot uses.
argument-hint: "[games] [casual|rated|random] [time-control]"
disable-model-invocation: true
allowed-tools: AskUserQuestion, Read
---

# Playing on Lichess

The engine plays on Lichess as GolangChessBot through lichess-bot, which is installed outside this repo at `D:\Documents\VS Code\lichess-bot`. This skill runs the bot for a set number of games on the user's real account.

## Rules

- Start the bot only after the user has answered step 1 in this conversation. Their answers are the permission to launch it.
- Leave the token in `.env`. Never read, print or log that file.
- Run from the main checkout, because `.env` is at its root and worktrees do not have it.

## Steps

1. **Ask how to run.** Use any values already given as arguments. Ask for the rest with AskUserQuestion:

   | Question | Options | Default |
   |----------|---------|---------|
   | How many games? | 1 to 100 | 1 |
   | Which mode? | casual, rated, random | casual |
   | Which time control? | Keep the config's mix, `60+1`, `180+2`, or another `BASE+INC` | Keep the config's mix |

   Rated and random games change the bot's public rating, so state that when the user picks one.

2. **Build the engine** from this checkout, so every run uses a fresh build:

   ```powershell
   go build -o "D:\Documents\VS Code\lichess-bot\engines\chessengine.exe" .\cmd\uci
   ```

3. **Start the games** in the background, leaving out `-Mode` and `-TimeControl` when using the config's values:

   ```powershell
   powershell -NoProfile -File .claude\skills\playing-on-lichess\scripts\play-games.ps1 -Games <n> -BotDir "D:\Documents\VS Code\lichess-bot" -Mode <mode> -TimeControl <tc>
   ```

   The script counts finished games and stops the bot after the last one. Tell the user it can take a while, because the bot waits for a free opponent.

4. **Report** the wins, losses and draws the script prints, plus any lost games worth reviewing. To analyse them against Stockfish, see [reference.md](reference.md#reviewing-games).

## Common mistakes

| Mistake | Instead |
|---------|---------|
| Stopping the bot by hand mid-game | That abandons the game, which counts as a loss. Let the script finish |
| Asking for more than 100 games | Lichess allows a bot 100 games a day against other bots |
| Editing `config.yml` for one run | Pass `-Mode` and `-TimeControl`, which patch a generated copy |

See [reference.md](reference.md) for settings, logs and troubleshooting.
