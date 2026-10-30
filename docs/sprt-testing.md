# SPRT Testing

Compare two builds of the engine (baseline vs. candidate) with `bench sprt`,
which drives [fastchess](https://github.com/Disservin/fastchess). Results use
the pentanomial model (game pairs, one per colour from each opening) and
normalized Elo bounds, as on fishtest and OpenBench. Written to be followed step by step by a human or
an AI agent: every step is a literal command plus the output to expect.

**Terms:** `base` = known-good build (usually `HEAD`). `dev` = candidate build
(usually the working tree, including uncommitted changes).

## Prerequisites (check first)

| Need | How to check | If missing |
|------|--------------|------------|
| `fastchess.exe` at `<repo root>/tools/engines/` | `ls tools/engines/fastchess.exe` | Download `fastchess-windows-x86-64.zip` from the fastchess GitHub releases and copy `fastchess.exe` there; in a worktree see "Git worktrees" below |
| Openings at `<repo root>/testdata/8moves_v3.epd` | `ls testdata/8moves_v3.epd` | Tracked in git (force-added; `testdata/` is ignored); restore it |
| Go toolchain | `go version` | Install Go |
| Free cores | `nproc` | Lower `-concurrency` |

`tools/engines/` is **not tracked by git**, so a fresh git worktree does not
have it. The tool resolves the path from the repo root it runs in, and there is
no flag to override it.

### Git worktrees

Link the main checkout's folder into the worktree with a junction (no copy,
nothing for git to track):

```powershell
New-Item -ItemType Junction -Path "<worktree>\tools\engines" -Target "<main checkout>\tools\engines"
```

Remove it when finished. Use `cmd //c rmdir "<worktree>\tools\engines"` (Git
Bash) or `Remove-Item` on the junction itself. `rmdir` on a junction deletes
only the link, never the target, but do not use `rm -rf`.

## 1. Build base and dev

Pick a scratch directory outside the repo. Use a literal path in each command;
do not build the executable path from a shell variable.

```bash
S="/path/to/scratch"

# base: HEAD extracted into a clean tree, so uncommitted changes cannot leak in.
rm -rf "$S/base-tree"
mkdir -p "$S/base-tree"
git archive HEAD | tar -x -C "$S/base-tree"
go -C "$S/base-tree" build -o "$S/base-uci.exe" ./cmd/uci

# dev: the working tree as it is now.
go build -o "$S/dev-uci.exe" ./cmd/uci
```

To compare two commits, run `git archive <commit>` for each side instead.

**Check:** both `.exe` files exist, and `git status` shows the changes you mean
to test. If `base` and `dev` are built from identical code the test is
meaningless.

## 2. Run the match

This takes minutes to hours. Run it in the background and wait for the
completion notice; do not poll or sleep.

```bash
go run ./cmd/bench sprt \
  -base "$S/base-uci.exe" \
  -dev  "$S/dev-uci.exe" \
  -tc 1+0.05 \
  -games 500 \
  -concurrency 8 \
  > "$S/sprt.log" 2>&1
```

| Flag | Default | Meaning |
|------|---------|---------|
| `-tc` | `10+0.1` | Time control, `seconds+increment`. `1+0.05` is the fast screening setting; use `10+0.1` for a production-representative result |
| `-games` | `4000` | Cap on total games, played as `games/2` opening pairs. SPRT can stop earlier if a bound is crossed |
| `-concurrency` | `1` | Games in parallel. Use about half the logical core count (8 on a 16-thread machine) |
| `-base-options`, `-dev-options` | none | Comma-separated UCI options, e.g. `Hash=64`. Our engine is single-threaded and has no `Threads` option; set `Threads=1` only for other engines such as Stockfish |
| `-elo0`, `-elo1` | `0`, `2` | SPRT hypotheses in **normalized** Elo (nElo). `[0, 2]` for a gainer; `[-2, 0]` to check a simplification or refactor is not a regression |
| `-alpha`, `-beta` | `0.05` each | Error rates |
| `-openings` | `testdata/8moves_v3.epd` | EPD opening book: 34,700 balanced 8-move positions (Stockfish `8moves_v3`, CC0), picked at random |

Run `go run ./cmd/bench sprt -h` for the full list.

**Rough duration at `-tc 1+0.05`:** 500 games takes about 55 min at
concurrency 1 and about 7-10 min at concurrency 8.

**Each invocation is a fresh SPRT sequence.** It never resumes or accumulates.
For more evidence, raise `-games` in one run; do not chain small runs.

**Side effect:** every run appends to `tools/results/sprt_history.md` (tracked
by git) and `tools/results/sprt_results.jsonl`.

**Keep or discard the history entry:**
- Keep it when the run tested a real change in playing strength, or the result
  is something you want on record (a clear H1/H0 verdict, a confirmed
  regression).
- Delete it when the change should have **no impact** on play (a refactor,
  renamed symbols, stats or logging changes) and the run was only a sanity
  check, or when the run was a failed or aborted attempt. Such entries add
  noise to the history.

To discard, revert the tracked file:

```bash
git checkout tools/results/sprt_history.md
```

Do this before committing so the entry never lands in the diff. If the run
appended to `sprt_results.jsonl` and that file is tracked too, revert it the
same way; check with `git status`. Never delete entries from runs you want to
keep, and never hand-edit older history entries.

## 3. Read the result

```bash
grep "^Elo:\|^LOS:\|^Ptnml" "$S/sprt.log" | tail -3   # point estimate
tail -3 "$S/sprt.log"                                # SPRT verdict line
```

The verdict line looks like:

```
SPRT verdict: inconclusive (games cap reached) (llr=0.83, bounds=[-2.94, 2.94], W-L-D=216-168-116, games=500)
```

| Verdict | Meaning |
|---------|---------|
| `H1 accepted` | LLR crossed the upper bound: dev is stronger |
| `H0 accepted` | LLR crossed the lower bound: dev is not stronger |
| `inconclusive (games cap reached)` | Cap hit before either bound; use the Elo estimate only as a hint |

`Elo: X +/- Y, nElo: N +/- M` and `LOS: Z%` are the point estimate with its
interval. `Ptnml(0-2): [a, b, c, d, e]` counts game pairs by Dev's pair score
(0, 0.5, 1, 1.5, 2); the history table records it.
Judge it by these rules:

- Interval clearly away from 0 and high LOS: probably real.
- Interval straddles 0 and LOS near 50%: noise.
- **Do not act on a short run.** At 500-1000 games an LOS of 80-99% is
  routinely noise at this time control. A 500-game run of a change that cannot
  alter play (a pure refactor) still showed +33 +/- 27 Elo. Confirm any
  promising signal with 2000+ games before keeping a change.

## 4. Several small runs do not average

Separate batches are independent estimates, not one bigger sample. If you must
compare them, check whether they land consistently on one side of zero. For a
single clean answer, run one large batch (`-games 2000` or more).

## Failure modes

| Symptom | Cause | Fix |
|---------|-------|-----|
| `fastchess not found at .../tools/engines/fastchess.exe` | Running from a worktree without the junction, or fastchess not installed | See "Prerequisites" and "Git worktrees" |
| Exit code 1 with no games played | Missing `.exe`, bad path, or bad option string | Read the last 20 lines of the log |
| Result near 0 Elo with huge `+/-` | Too few games | Raise `-games` |
| Games lose to timeouts | `-concurrency` too high for the core count | Lower it |
| `config.json` appears in the repo root | fastchess saves its tournament state at the end of a match | Ignored by git; delete it |

The engine's opening book (`internal/book/testdata/performance.bin`, loaded
relative to the working directory) is **not** used during SPRT games.
`bench sprt` passes `option.BookFile=<empty>` to both engines. Builds without
that option print a harmless `doesn't have option BookFile` warning and rely on
running from their binary's directory, so keep the built binaries outside the
repo root, as step 1 does.

## Rules for AI agents

- Run the match in the background and wait for the notification. Do not poll.
- Never launch `run.ps1` (the Lichess launcher uses a real token) as part of
  testing. `bench sprt` is the only match runner needed here.
- Do not commit or push results unless the user asks.
- Report the verdict line, the Elo line, and the game count. State clearly
  when the run is too short to trust.
