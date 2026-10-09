# SPRT Testing Workflow

How to compare two builds of the engine (a baseline vs. a candidate change)
using `bench sprt`, the way this branch's bug-hunting sessions did it.

## 1. Build base (baseline) and dev (candidate)

`base` is normally the last commit (`HEAD`) — the known-good state before
your uncommitted change. `dev` is the current working tree, including
uncommitted changes.

```bash
SCRATCH="/path/to/scratch/dir"   # any writable temp dir

# base: extract HEAD into a clean temp tree and build it there, so
# uncommitted changes in the working tree don't leak into base.
rm -rf "$SCRATCH/base-tree"
mkdir -p "$SCRATCH/base-tree"
git archive HEAD | tar -x -C "$SCRATCH/base-tree"
go -C "$SCRATCH/base-tree" build -o "$SCRATCH/base-uci.exe" ./cmd/uci

# dev: build directly from the working tree (includes uncommitted changes).
go build -o "$SCRATCH/dev-uci.exe" ./cmd/uci
```

To compare two specific commits instead of "HEAD vs. working tree", use
`git archive <commit>` for both sides (e.g. `git archive HEAD~1` for base).

## 2. Run the match

```bash
go run ./cmd/bench sprt \
  -base "$SCRATCH/base-uci.exe" \
  -dev  "$SCRATCH/dev-uci.exe" \
  -tc 1+0.05 \
  -games 500
```

- `-tc 1+0.05` — fast time control (1s + 0.05s increment), used throughout
  this branch's testing for quick turnaround. Use a longer TC (e.g.
  `10+0.1`) for a more realistic/production-representative result at the
  cost of much longer wall-clock time.
- `-games 500` — safety cap on total games; SPRT can stop earlier if the
  bound is reached. Each invocation starts a **fresh** SPRT sequence — it
  does not resume or accumulate across separate runs. To get more games'
  worth of evidence, either raise `-games` in one run or run several
  same-sized batches and read them together (see below), not both as one
  continuous test.
- `-concurrency N` — number of games played in parallel (default 1).
  Raise this to roughly your physical core count to cut wall-clock time
  proportionally; e.g. on an 8-core machine:

  ```bash
  go run ./cmd/bench sprt -base ... -dev ... -tc 1+0.05 -games 500 -concurrency 8
  ```

  A 500-game batch that takes ~55 minutes at concurrency 1 finishes in
  ~7-10 minutes at 8. Keep `Threads=1` on the engines themselves
  (`-dev-options`/`-base-options`) when using concurrency, so the games
  share cores instead of fighting over them.
- Optional: `-dev-options "Skill Level=5,Threads=1"` / `-base-options ...`
  to test against a handicapped engine (e.g. Stockfish) instead of two
  ChessEngine builds — see `go run ./cmd/bench sprt -h` for all flags
  (elo0/elo1/alpha/beta bounds, openings file, etc.).

Output is also appended to `tools/results/sprt_history.md` /
`sprt_results.jsonl` automatically.

## 3. Read the result

Grep the log for the summary lines:

```bash
grep "Elo difference\|LOS" <logfile>
tail -3 <logfile>                          # final SPRT verdict line
```

- **SPRT verdict**: `H1 accepted` = dev is stronger (crossed the upper
  bound), `H0 accepted` = dev is not stronger (crossed the lower bound),
  `inconclusive` = games cap hit before either bound was reached.
- **Elo difference: X ± Y, LOS: Z%**: cutechess's own point estimate and
  confidence interval, useful even when SPRT itself says inconclusive — if
  the interval sits clearly away from zero (and LOS is high), that's a real
  signal even without a formal SPRT accept. If the interval straddles zero
  and LOS sits near 50%, treat it as noise.

## 4. Multiple batches don't average automatically — read them together

Because each `bench sprt` invocation is an independent SPRT sequence, running
several 500-game batches back to back gives you several independent
Elo-difference estimates, not one larger sample. To judge the overall
picture, look at whether the estimates cluster consistently on one side of
zero (a real effect) or scatter across both sides with overlapping error
bars (noise around ~0 Elo). If you want one clean, larger-sample answer
instead of stitching together small batches, run a single larger batch
(e.g. `-games 5000`) rather than many small ones.
