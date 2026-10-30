---
name: running-sprt
description: Use when checking whether an engine change gains or loses Elo, comparing two ChessEngine builds, running bench sprt or fastchess matches, or reading SPRT results such as LLR, LOS or Ptnml.
argument-hint: "[base-ref] [dev-ref]"
---

# Running SPRT

An SPRT plays two engine builds against each other until the result is statistically clear. `bench sprt` drives fastchess with the pentanomial model and normalized Elo.

`base` is the known-good build, `HEAD` by default. `dev` is the candidate, the working tree by default.

## Rules

- Run the match in the background and wait for the notification.
- Use `bench sprt` as the only match runner. `run.ps1` starts the Lichess bot with a real token, so never launch it for testing.
- Commit or push results only when the user asks.

## Steps

1. **Check prerequisites.** `tools/engines/fastchess.exe` and `testdata/8moves_v3.epd` must exist. In a worktree, link `tools/engines` first (see [reference.md](reference.md#worktrees)).

2. **Build both engines** into a scratch directory outside the repo:

   ```bash
   bash .claude/skills/running-sprt/scripts/build-engines.sh <scratch> [base-ref] [dev-ref]
   ```

   The script prints both bench signatures. Matching signatures mean search did not change, so use bounds `-2`/`0` or skip the match.

3. **Run the match** in the background:

   ```bash
   go run ./cmd/bench sprt -base <scratch>/base-uci.exe -dev <scratch>/dev-uci.exe \
     -tc 1+0.05 -games 4000 -concurrency 8 -elo0 0 -elo1 5 > <scratch>/sprt.log 2>&1
   ```

   This takes over an hour. Set `-concurrency` to about half the logical cores.

4. **Read the result:**

   ```bash
   grep -E "^(Elo|LOS|Ptnml|SPRT verdict)" <scratch>/sprt.log | tail -n 4
   grep -iE "illegal|disconnect|stall" <scratch>/sprt.log
   ```

5. **Report** the verdict line, the Elo line and the game count, plus any flagged games. A flagged game is a bug even when the Elo looks good.

6. **Keep or discard the history entry.** Each run appends to `tools/results/sprt_history.md`. Keep it for a real strength change. For a sanity check or a failed run, restore the file with `git checkout tools/results/sprt_history.md`.

7. **Remove the worktree junction** if you added one.

## Reading results

| Result | Meaning |
|--------|---------|
| `H1 accepted` | Dev is stronger |
| `H0 accepted` | Dev is not stronger |
| `inconclusive (games cap reached)` | Neither bound was crossed, so the Elo estimate is only a hint |

An interval clear of 0 with a high LOS is probably real. An interval straddling 0 with LOS near 50% is noise.

## Common mistakes

| Mistake | Instead |
|---------|---------|
| Trusting 500-1000 games | LOS of 80-99% is routinely noise here. Confirm with 2000+ games |
| Chaining short runs | Runs never resume or average. Raise `-games` in one run |
| Passing an empty UCI option value | Fastchess rejects it. Use `<empty>` |
| Building into the repo root | Older builds would load the opening book from there |

See [reference.md](reference.md) for flags, worktree setup and failure modes.
