# ChessEngine

A high-performance chess engine written in Go, featuring advanced search algorithms, comprehensive evaluation functions, and UCI protocol support for integration with chess GUIs.

## Features

### Core Engine Capabilities
- **Minimax Search** with alpha-beta pruning and aspiration windows
- **Advanced Pruning** including null move pruning and late move reductions (LMR)
- **Transposition Tables** with Zobrist hashing for position caching
- **Opening Book** support via Polyglot format
- **Time Management** with iterative deepening

### Board Representation
- **Hybrid Architecture** combining bitboards and mailbox representation
- **Magic Bitboards** for efficient sliding piece move generation
- **12 Piece Bitboards** (6 piece types × 2 colors) for fast position queries
- **Incremental Updates** for hash keys and position evaluation

### Move Generation & Validation
- **Legal Move Generation** with check detection and pinned piece handling
- **Attack Detection** using pre-computed attack tables
- **Move Ordering** with MVV-LVA (Most Valuable Victim - Least Valuable Attacker)
- **Killer Moves** and history heuristics for enhanced move ordering

### Position Evaluation
- **Material Balance** with piece-square tables
- **Pawn Structure** analysis including passed pawns and pawn chains
- **King Safety** evaluation with shelter and storm patterns
- **Piece Mobility** and control evaluation
- **Static Exchange Evaluation (SEE)** for capture analysis

### Performance Optimizations
- **Object Pooling** for move lists to reduce garbage collection
- **Pre-computed Tables** for piece attacks, knight mobility, and evaluation
- **Incremental Updates** for Zobrist hashing, material balance, and piece-square table scores
- **Two-bucket Transposition Table** with collision resolution and age-based replacement
- **32-bit Packed Moves** in transposition table (compressed from 80 to 32 bits)
- **Pawn Hash Table** for caching expensive pawn structure evaluations
- **Lazy Evaluation** with early cutoffs at 1000cp and 500cp thresholds
- **SEE-based Move Ordering** for accurate capture evaluation
- **Tactical Move Bonuses** for attacks on valuable pieces and king zones
- **Sampled TT Statistics** to reduce overhead (updated every 256th probe)

## Architecture

### Package Structure
```
cmd/
  ├── gchess/     # Benchmark launcher entry point
  │   └── modes/  # Benchmark mode implementations (STS, Elo)
  ├── uci/        # UCI engine binary for chess GUIs
  └── bench/      # Bench multi-tool: sts/profile/report subcommands
internal/
  ├── board/      # Board representation and bitboard operations
  ├── movegen/    # Legal move generation with magic bitboards and object pooling
  ├── search/     # Negamax, iterative deepening, quiescence, transposition tables
  ├── eval/       # Position evaluation with pawn hash table
  │   └── values/ # Piece values and piece-square tables
  ├── book/       # Polyglot opening book support
  ├── player/     # Player implementations, including the computer player
  ├── epd/        # EPD file parsing and STS scoring
  ├── bench/      # Benchmark infrastructure for STS/Elo comparison
  ├── uci/        # UCI protocol implementation
  ├── game/       # Game state management and move parsing
  ├── ui/         # Board rendering and user interface
  └── testutil/   # Shared test helpers
```

### Key Components
- **Search Engine** (`internal/search/`) - Negamax with alpha-beta, LMR, null move pruning, quiescence search
- **Transposition Table** (`internal/search/transposition.go`) - Two-bucket TT with 32-bit packed moves
- **Evaluator** (`internal/eval/`) - Lazy evaluation with pawn hash table
- **Move Generator** (`internal/movegen/`) - Legal move generation with magic bitboards and object pooling
- **Move Ordering** (`internal/search/move_ordering.go`) - SEE-based ordering with killer moves and history heuristic
- **Opening Book** (`internal/book/`) - Polyglot opening book integration
- **UCI Interface** (`internal/uci/`) - Universal Chess Interface protocol support

## Usage

### UCI Mode (Chess GUI Integration)
```bash
# Build UCI executable
go build -o chessengine-uci ./cmd/uci

# Use with chess GUIs like Arena, ChessBase, or Cute Chess
./chessengine-uci
```

### Benchmark Launcher
```bash
# Build and run the gchess benchmark launcher
go build -o chessengine ./cmd/gchess
./chessengine

# Choose mode:
# 1. STS Benchmark
# 2. Elo Benchmark
```

### Benchmarking
```bash
# Build the bench multi-tool (sts/profile/report subcommands)
go build -o bench ./cmd/bench

# Run Strategic Test Suite (STS)
./bench sts -file testdata/STS1.epd -timeout 5 -max 100

# Run CPU/memory profiling
./bench profile -file testdata/STS1.epd
```

**Dev note:** `make bench-save` runs the Go benchmark suite for
`internal/movegen`/`internal/eval`/`internal/search` (`-count 10`) and saves
it to `tools/results/bench_<git-sha>.txt`; `make bench-compare OLD=<file>
NEW=<file>` diffs two such saves with
[benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat).

## Testing & Validation

### Comprehensive Test Suite
- **Unit Tests** for all major components (32 test files, 133+ test functions)
- **Perft Tests** for move generation validation at depths 1-6
- **Integration Tests** for UCI protocol and game flow
- **Strategic Test Suite (STS)** with 600 test positions for positional evaluation validation
- **EPD Test Support** for custom test position files

### Performance Testing
- **Benchmark Suite** comparing search with/without transposition tables
- **Profiling Tools** for CPU and memory profiling (pprof format)
- **STS Rating System** with approximate ELO estimation
- **Node Performance Analysis** including NPS, effective branching factor, and TT hit rates

### Run Tests
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

### Search Performance
- **Node rate**: ~1.3-1.6M nodes/second (hardware dependent, verified via STS benchmarks)
- **Search depth**: Typically 8-12 ply in tournament time controls
- **Memory usage**: ~50-200MB for transposition tables, ~256KB for pawn hash table
- **TT hit rate**: ~70-80% in mid-game positions

### Engine Strength
- **STS Score**: 424/600 (71%) - STS Rating: ~3000
- **Positional Understanding**: IM+ level (based on STS categories)
- **Estimated Playing Strength**: ~2000-2200 ELO (hardware dependent)
- **Match Results**: 33-66% win rate against weak Stockfish variants in bullet (2+2) time controls
- Validated against Strategic Test Suite (STS1-6) with comprehensive positional tests

## Technical Details

### Search Algorithm Features
- **Iterative Deepening** with time management
- **Principal Variation Search** (PVS) at root level
- **Late Move Reductions** (LMR) with pre-calculated table and history heuristic adjustments
- **Aspiration Windows** with dynamic widening for efficient root search
- **Null Move Pruning** with adaptive R (2-3 based on depth)
- **Razoring** at depth 1 with conservative 125cp margin
- **Check Extensions** for automatic depth extension when in check
- **Quiescence Search** with delta pruning and SEE-based capture filtering

### Evaluation Features
- **Phase-based Evaluation** with endgame detection (< 14 pieces threshold)
- **Piece-Square Tables** with incremental updates for positional evaluation
- **Pawn Hash Table** with 16K entries for caching pawn structure evaluations
- **Pawn Structure** analysis including passed, isolated, doubled, connected, and backward pawns
- **King Safety** with castling bonus, pawn shelter, open file detection, and king zone threat evaluation
- **Piece Mobility** with pre-computed tables for knights and attack-based evaluation for bishops, rooks, and queens
- **Knight Outposts** with defended square detection in enemy territory
- **Bishop Pair** bonus
- **Rook Evaluation** including open files and 7th rank bonus
- **Passed Pawn** evaluation with exponential advancement bonuses by rank
- **Lazy Evaluation** with early cutoffs to avoid expensive calculations

## Performance History

### Development Progress
- **Benchmark History** - See [history.md](tools/results/history.md) for detailed match results against various opponents
- **STS Performance** - See [sts_history.md](tools/results/sts_history.md) for Strategic Test Suite validation results over time
