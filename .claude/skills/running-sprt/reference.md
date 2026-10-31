# SPRT reference

## Flags

Run `go run ./cmd/bench sprt -h` for the full list.

| Flag | Default | Meaning |
|------|---------|---------|
| `-tc` | `10+0.1` | `seconds+increment`. `1+0.05` is the fast setting; use `10+0.1` for a production-representative result |
| `-games` | `4000` | Cap on total games, played as `games/2` opening pairs. SPRT can stop earlier if a bound is crossed |
| `-concurrency` | `1` | Games in parallel. 500 games at `1+0.05` take about 55 minutes at 1 and 7-10 minutes at 8 |
| `-base-options`, `-dev-options` | none | Comma-separated UCI options such as `Hash=64`. Our engine has no `Threads` option |
| `-elo0`, `-elo1` | `0`, `2` | Bounds in normalized Elo. `[0, 2]` or `[0, 5]` for a gainer, `[-2, 0]` to show a refactor is not a regression |
| `-alpha`, `-beta` | `0.05` | Error rates |
| `-openings` | `testdata/8moves_v3.epd` | 34,700 balanced 8-move positions, picked at random |
| `-pgn` | none | File to save the games to, in UCI notation. Use it to trace an illegal move. fastchess appends, so give each run its own file |

## Worktrees

`tools/engines/` is not tracked, so a fresh worktree does not have it. `bench sprt` resolves the path from the repo root it runs in, and no flag overrides it. Link the main checkout's folder with a junction:

```powershell
New-Item -ItemType Junction -Path "<worktree>\tools\engines" -Target "<main checkout>\tools\engines"
```

Remove it with `cmd //c rmdir "<worktree>\tools\engines"` or `Remove-Item` on the junction. Never use `rm -rf` on it.

## Missing files

- **fastchess:** copy `fastchess.exe` from `fastchess-windows-x86-64.zip` on the [fastchess releases](https://github.com/Disservin/fastchess/releases) into `tools/engines/`.
- **Openings:** `testdata/8moves_v3.epd` is tracked in git (force-added, since `testdata/` is ignored), so restore it.

## Log lines

`Elo: X +/- Y, nElo: N +/- M` and `LOS: Z%` are the point estimate and its interval.

`Ptnml(0-2): [a, b, c, d, e]` counts game pairs by Dev's pair score (0, 0.5, 1, 1.5, 2).

The verdict line looks like this:

```
SPRT verdict: inconclusive (games cap reached) (llr=0.83, bounds=[-2.94, 2.94], W-L-D=216-168-116, games=500)
```

## History files

Every run appends to `tools/results/sprt_history.md` (tracked) and `tools/results/sprt_results.jsonl`. If `git status` shows the JSONL file as changed, restore it along with the history file. Never hand-edit older entries.

## Opening book

`bench sprt` passes `option.BookFile=<empty>`, so the book at `internal/book/testdata/performance.bin` is off during games. Builds without that option print a harmless `doesn't have option BookFile` warning and rely on running from their binary's directory, which is why built binaries stay outside the repo root.

## Failure modes

| Symptom | Cause | Fix |
|---------|-------|-----|
| `fastchess not found at .../tools/engines/fastchess.exe` | Worktree without the junction, or fastchess not installed | See [Worktrees](#worktrees) |
| Exit code 1 with no games played | Missing `.exe`, bad path or bad option string | Read the last 20 lines of the log |
| `expects key=value pairs` | Empty option value | Use `<empty>` |
| Result near 0 Elo with a huge `+/-` | Too few games | Raise `-games` |
| Games lose to timeouts | `-concurrency` too high | Lower it |
| `config.json` in the repo root | fastchess saved its tournament state | Ignored by git, so delete it |
