---
name: running-sts
description: Use when running the Strategic Test Suite, checking the engine's STS score or rating, or recording a new row in sts_history.md.
---

# Running STS

The standard STS run tests the first 10 positions of each of the 6 STS files at 5 seconds per move: 60 positions, about 5 minutes. It uses the `gchess` runner because that is what writes the combined row to `tools/results/sts_history.md`.

## Steps

1. **Stop other CPU-heavy work.** STS is timed, so a running SPRT or build lowers the score.

2. **Run it in the background** from the repo root and wait for the notification:

   ```bash
   printf '1\n5\n10\nn\ny\n' | go run ./cmd/gchess > <scratch>/sts.log 2>&1
   ```

   The answers select STS mode, 5 seconds, 10 positions per file, no verbose output, and proceed.

3. **Read the result** from the new last row of the history file:

   ```bash
   tail -n 1 tools/results/sts_history.md
   ```

   The row gives the score out of 600, percent, STS rating, correct moves, NPS and average depth. If the row is missing, read the end of `<scratch>/sts.log`.

4. **Report** the score, rating, correct moves and commit, then compare against earlier rows using the rules below.

5. **Keep or discard the row.** Keep it for a commit worth recording. For an aborted or uncommitted experiment, run `git checkout tools/results/sts_history.md`.

## Comparing runs

| Check | Rule |
|-------|------|
| Score difference under about 25 points | Noise at this size. Same-commit runs have differed by 27 points |
| Baseline | Confirm with `git log` that the earlier row is the true previous state |
| Rows from before `24e3f68` | Not comparable, because that commit corrected STS move matching and lowered many scores |
| A strength claim | STS cannot prove one. Use the running-sprt skill |

For a smaller error bar, enter `30` instead of `10` positions. That takes about 15 minutes.

## Common mistakes

| Mistake | Instead |
|---------|---------|
| Using `bench sts` for this | It runs one file, defaults to depth 5 and adds no history row |
| Running from a directory without `testdata/` | Run from the repo root |
