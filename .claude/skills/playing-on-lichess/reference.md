# Lichess bot reference

## Files

- `tools/lichess/config.yml`: the lichess-bot configuration, which is copied to `<lichess-bot>/config.yml`. It holds no secrets. Changes take effect on the next start.
- `tools/lichess/run.ps1`: reads `LICHESS_TOKEN` from `.env`, passes it on as `LICHESS_BOT_TOKEN`, and starts the bot. It sets `CHESS_DEBUG_LOG` so the engine logs to `<lichess-bot>/logs/engine.jsonl`.
- `tools/lichess/review_log.py`: replays that log against Stockfish.

## Stopping cleanly

lichess-bot has no game limit. `play-games.ps1` counts new files in `<lichess-bot>/game_records/`, because the bot writes one PGN per finished game, then kills the bot process tree. A new game could start in the seconds between polls, and killing the bot would abandon it. The default 5 second poll and the 1 minute `challenge_timeout` make that unlikely, but an incoming challenge can still get in.

## Settings

| Setting | Value | Meaning |
|---------|-------|---------|
| `engine.working_dir` | this repo's path | The opening book loads relative to it |
| `engine.uci_options` | `Hash: 128` | Transposition table size in MB |
| `challenge.modes` | see `config.yml` | Modes accepted from incoming challenges |
| `challenge.concurrency` | `1` | One game at a time |
| `matchmaking.allow_matchmaking` | `true` | Challenges other bots after `challenge_timeout` idle minutes |
| `matchmaking.challenge_timeout` | `1` | Idle minutes before sending a challenge |
| `matchmaking.opponent_min_rating`, `opponent_max_rating` | `1800`, `2600` | Opponent band |
| `matchmaking.opponent_rating_difference` | `null` | Must stay `null`, or it overrides the band |
| `matchmaking.challenge_mode` | `casual` | Mode of challenges the bot sends |

## Healthy start

```
Engine configuration OK
Welcome <bot name>!
You're now connected to https://lichess.org/ and awaiting challenges.
```

`play-games.ps1` writes the bot's output to `<lichess-bot>/logs/play-games.log` and `play-games.err.log`.

## Reviewing games

The engine log has one JSON record per move. See the header of `tools/lichess/review_log.py` for its arguments.

## Troubleshooting

| Symptom | Cause and fix |
|---------|---------------|
| `Env file not found` | Running from a worktree. Run from the main checkout |
| `LICHESS_TOKEN not found` | `.env` is missing the line, or it is misspelled |
| `FileNotFoundError: './config.yml'` | Started from the wrong directory. `run.ps1` changes into `-BotDir`, so use it |
| `... is not a bot account` | The token belongs to an account that is not a bot account |
| No book moves | `engine.working_dir` does not point at this repo |
| `... played 100 games against other bots today` | That is the opponent's daily limit, so the bot picks someone else |
| Opponents far too strong | Narrow the rating band, keeping `opponent_rating_difference: null` |

## Token safety

The token lives only in `.env` and an environment variable. If it is ever pasted into a chat, a log or a commit, revoke it at <https://lichess.org/account/oauth/token> and create a new one. Never put a real token in `config.yml`.
