# ChessEngine

A chess engine written in Go. It speaks the UCI protocol, so it runs in chess GUIs, and it plays on Lichess as a bot.

## Live on Lichess

The engine plays on Lichess as the bot account [GolangChessBot](https://lichess.org/@/GolangChessBot) through
[lichess-bot](https://github.com/lichess-bot-devs/lichess-bot). While it runs, it challenges other bots, accepts
challenges, and plays rated and casual games.

Current bullet rating: 2393 (October 2026).

### What the bot does

- One setting chooses rated games, casual games, or a mix of both.
- A Polyglot opening book built into the engine plays the opening moves, so those moves cost almost no clock time.
- Every search writes a JSON log entry with the depth, nodes, score, principal variation, best move at each iteration,
  and the reason the search stopped.
- [`review_log.py`](tools/lichess/review_log.py) replays the log against Stockfish and flags the moves that lost the
  most centipawns, next to the engine's own score at each depth.

## Features

### Core engine

- Negamax search with alpha-beta pruning and aspiration windows
- Null move pruning and late move reductions (LMR)
- Transposition tables with Zobrist hashing
- Polyglot opening book support
- Iterative deepening with time management

### Board representation

- Bitboards combined with a mailbox array
- Magic bitboards for sliding piece move generation
- 12 piece bitboards (6 piece types for each of 2 colors)
- Incremental updates for hash keys and position evaluation

### Move generation and ordering

- Legal move generation with check detection and pinned piece handling
- Pre-computed attack tables
- MVV-LVA (most valuable victim, least valuable attacker) capture ordering
- Static exchange evaluation (SEE) for scoring captures
- Killer moves and the history heuristic

### Position evaluation

- Material balance and piece-square tables
- Pawn structure, including passed and connected pawns
- King safety from open files near the king and enemy pieces in the king zone
- Piece mobility

### Performance optimizations

- Object pools for move lists, which reduce garbage collection
- Pre-computed tables for piece attacks and evaluation
- Incremental updates for Zobrist hashes, material balance, and piece-square scores
- Two-bucket transposition table with age-based replacement
- Moves packed from 80 to 32 bits in the transposition table
- Pawn hash table that caches pawn structure scores
- Lazy evaluation, which skips pawn and king terms above 1000cp and piece activity above 500cp
- Transposition table statistics sampled on every 256th hash value

## Architecture

### Package structure

```
cmd/
  ├── gchess/     # Benchmark launcher entry point
  │   └── modes/  # Benchmark mode implementations (STS)
  ├── uci/        # UCI engine binary for chess GUIs
  └── bench/      # Bench multi-tool: sts/profile/report subcommands
internal/
  ├── board/      # Board representation and bitboard operations
  ├── movegen/    # Legal move generation with magic bitboards and object pooling
  ├── search/     # Negamax, iterative deepening, quiescence, transposition tables
  ├── eval/       # Position evaluation with pawn hash table
  │   └── values/ # Piece values and piece-square tables
  ├── book/       # Polyglot opening book support
  ├── epd/        # EPD file parsing and STS scoring
  ├── bench/      # Benchmark infrastructure for STS comparison
  ├── uci/        # UCI protocol implementation
  ├── game/       # Game state management and move parsing
  ├── ui/         # Board rendering and user interface
  └── testutil/   # Shared test helpers
```

### Key components

- `internal/search/` runs negamax with alpha-beta, LMR, null move pruning, and quiescence search.
- `internal/search/transposition.go` holds the two-bucket transposition table with 32-bit packed moves.
- `internal/eval/` scores positions with lazy evaluation and a pawn hash table.
- `internal/movegen/` generates legal moves with magic bitboards and object pools.
- `internal/search/move_ordering.go` orders moves with SEE, killer moves, and the history heuristic.
- `internal/book/` reads Polyglot opening books.
- `internal/uci/` implements the Universal Chess Interface protocol.

## Usage

### UCI mode (chess GUIs)

```bash
# Build UCI executable
go build -o chessengine-uci ./cmd/uci

# Use with chess GUIs like Arena, ChessBase, or Cute Chess
./chessengine-uci
```

### Benchmark launcher

```bash
# Build and run the gchess benchmark launcher
go build -o chessengine ./cmd/gchess
./chessengine

# Choose mode:
# 1. STS Benchmark
```

### Benchmarking

```bash
# Build the bench multi-tool (sts/profile/report/sprt/mates subcommands)
go build -o bench ./cmd/bench

# Run Strategic Test Suite (STS)
./bench sts -file testdata/STS1.epd -timeout 5 -max 100

# Run the matetrack mate-finding suite (testdata/matetrack.epd, not tracked:
# download matetrack.epd from github.com/vondele/matetrack, GPL-3.0)
./bench mates -time 200ms -concurrency 8

# Run CPU/memory profiling
./bench profile -file testdata/STS1.epd
```

### Bench signature

`uci bench` (or `bench` inside a UCI session) searches a fixed position set to depth 9 and prints
`Bench: <nodes> nodes <nps> nps`. The node count is deterministic, so put it in commit messages (`Bench: 1890275`). A
change that alters the count changes search behaviour.

### SPRT regression testing

SPRT runs a statistical A/B test between two engine builds through fastchess, using the pentanomial model and normalized
Elo. See [the running-sprt skill](.claude/skills/running-sprt/SKILL.md).

### Go benchmarks

`make bench-save` runs the Go benchmark suite for `internal/movegen`, `internal/eval`, and `internal/search` with
`-count 10` and saves the output to `tools/results/bench_<git-sha>.txt`. `make bench-compare OLD=<file> NEW=<file>`
compares two saves with [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat).

## Testing and validation

### Test suite

- Unit tests in 54 test files
- Perft tests that check move generation at depths 1-6
- Integration tests for the UCI protocol and game flow
- Strategic Test Suite (STS) with 600 positions that test positional evaluation
- EPD support for custom test position files

### Performance testing

- Benchmarks that compare search with and without transposition tables
- CPU and memory profiling in pprof format
- STS rating with an approximate Elo estimate
- Node statistics, including NPS, effective branching factor, and transposition table hit rate

### Run tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Run performance benchmarks
go test -bench=. ./...

# Validate move generation with Perft
go test -run TestPerft ./internal/movegen
```

## Performance

### Search performance

- Node rate: 1.0-1.4M nodes/second in recent STS runs (hardware dependent)
- Search depth: typically 8-12 ply at tournament time controls
- Memory: 128MB transposition table by default (set with the UCI `Hash` option), about 256KB for the pawn hash table
- Transposition table hit rate: about 70-80% in middlegame positions

### Engine strength

- Lichess bullet rating: 2393 (see [Live on Lichess](#live-on-lichess))
- STS score: 426/600 (71%), STS rating about 3000, from 10 positions per file across STS1-6

## Technical details

Comment conventions are in [docs/comment-guidelines.md](docs/comment-guidelines.md).

### Search algorithm

- Iterative deepening with time management
- Principal variation search (PVS)
- Late move reductions from a pre-calculated table, adjusted by history scores
- Aspiration windows that widen on a fail high or fail low
- Null move pruning with R of 2 or 3, depending on depth
- Razoring at depth 1 with a 125cp margin
- Check extensions
- Quiescence search with delta pruning and SEE capture filtering

### Evaluation

- Tapered evaluation that blends middlegame and endgame piece-square scores by game phase
- Piece-square tables with incremental updates
- Pawn hash table with 16K entries
- Pawn structure terms for passed, isolated, doubled, connected, and backward pawns
- King safety terms for open files near the king and enemy pieces in the king zone
- Piece mobility from attacked squares inside a pawn-safe mobility area
- Knight outposts on defended squares in enemy territory
- Bishop pair bonus
- Rook bonuses for open files and the 7th rank
- Passed pawn bonuses that grow exponentially with rank
- Lazy evaluation with early cutoffs

## Performance history

- [history.md](tools/results/history.md) has match results against various opponents.
- [sts_history.md](tools/results/sts_history.md) has Strategic Test Suite results over time.
