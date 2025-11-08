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
board/          # Board representation and bitboard operations
game/
  ├── ai/         # Search algorithms and evaluation
  │   ├── search/     # Negamax, iterative deepening, quiescence, transposition tables
  │   └── evaluation/ # Position evaluation with pawn hash table
  ├── moves/      # Move generation and validation with object pooling
  ├── openings/   # Polyglot opening book support
  └── modes/      # Game mode implementations (manual, vs AI, benchmark)
uci/            # UCI protocol implementation
epd/            # EPD file parsing and STS scoring
benchmark/      # Benchmark infrastructure and engine comparison
ui/             # Board rendering and user interface
cmd/            # Command-line applications
  ├── uci/        # UCI interface for chess GUIs
  ├── benchmark/  # Performance benchmarking with/without TT comparison
  ├── sts/        # Strategic Test Suite runner with scoring
  └── profile/    # CPU and memory profiling tool
```

### Key Components
- **Search Engine** (`game/ai/search/`) - Negamax with alpha-beta, LMR, null move pruning, quiescence search
- **Transposition Table** (`game/ai/search/transposition.go`) - Two-bucket TT with 32-bit packed moves
- **Evaluator** (`game/ai/evaluation/`) - Lazy evaluation with pawn hash table
- **Move Generator** (`game/moves/`) - Legal move generation with magic bitboards and object pooling
- **Move Ordering** (`game/ai/search/move_ordering.go`) - SEE-based ordering with killer moves and history heuristic
- **Opening Book** (`game/openings/`) - Polyglot opening book integration
- **UCI Interface** (`uci/`) - Universal Chess Interface protocol support

## Usage

### UCI Mode (Chess GUI Integration)
```bash
# Build UCI executable
go build -o chessengine-uci ./cmd/uci

# Use with chess GUIs like Arena, ChessBase, or Cute Chess
./chessengine-uci
```

### Interactive Play
```bash
# Build and run interactive mode
go build -o chessengine .
./chessengine

# Choose game mode:
# 1. Manual Play (Player vs Player)
# 2. Player vs Computer
```

### Benchmarking
```bash
# Run performance benchmarks (with/without TT comparison)
go build -o benchmark ./cmd/benchmark
./benchmark

# Run Strategic Test Suite (STS)
go build -o sts ./cmd/sts
./sts -file testdata/STS1.epd -timeout 5 -max 100

# Run CPU/memory profiling
go build -o profile ./cmd/profile
./profile
```

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
go test -run TestPerft ./game/moves
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
- **Benchmark History** - See [history.md](history.md) for detailed match results against various opponents
- **STS Performance** - See [sts_history.md](sts_history.md) for Strategic Test Suite validation results over time
