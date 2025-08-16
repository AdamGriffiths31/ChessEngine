# ChessEngine

A high-performance chess engine written in Go, featuring advanced search algorithms, comprehensive evaluation functions, and UCI protocol support for integration with chess GUIs.

## Features

### Core Engine Capabilities
- **Minimax Search** with alpha-beta pruning and aspiration windows
- **Advanced Pruning** including null move pruning and late move reductions (LMR)
- **Transposition Tables** with Zobrist hashing for position caching
- **Multi-threading Support** (1-32 threads) with Lazy SMP for parallel search
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
- **Pre-computed Tables** for piece attacks and evaluation
- **Lazy Evaluation** with early cutoffs
- **Memory-efficient** transposition table with aging

## Architecture

### Package Structure
```
board/          # Board representation and bitboard operations
game/
  ├── ai/         # Search algorithms and evaluation
  ├── moves/      # Move generation and validation
  └── openings/   # Opening book support
uci/            # UCI protocol implementation
cmd/            # Command-line applications
├── uci/        # UCI interface
├── benchmark/  # Performance benchmarking
└── sts/        # Strategic Test Suite runner
```

### Key Components
- **Search Engine** (`game/ai/search/`) - Minimax with advanced pruning techniques
- **Evaluator** (`game/ai/evaluation/`) - Comprehensive position evaluation
- **Move Generator** (`game/moves/`) - Legal move generation with magic bitboards
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
# Run performance benchmarks
go build -o benchmark ./cmd/benchmark
./benchmark

# Run Strategic Test Suite (STS)
go build -o sts ./cmd/sts
./sts
```

## Testing & Validation

### Comprehensive Test Suite
- **Unit Tests** for all major components with >80% coverage
- **Perft Tests** for move generation validation at various depths
- **Integration Tests** for UCI protocol and game flow
- **Strategic Test Suite (STS)** for positional evaluation validation

### Performance Testing
- **Benchmark Suite** comparing against established engines
- **Profiling Tools** for performance optimization
- **Memory Usage** analysis and optimization

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
- **Node rate**: ~500K-1M nodes/second (hardware dependent)
- **Search depth**: Typically 6-12 ply in tournament time controls
- **Memory usage**: ~50-200MB for transposition tables
- **Multi-threading**: Linear speedup up to 8 threads

### Engine Strength
- Validated against standard test suites (STS, EPD)
- Competitive performance against engines of similar complexity
- Estimated playing strength: ~1800-2200 ELO (hardware dependent)

## Technical Details

### Search Algorithm Features
- **Iterative Deepening** with time management
- **Principal Variation Search** (PVS)
- **Late Move Reductions** (LMR) for non-critical moves
- **Aspiration Windows** for efficient root search
- **Null Move Pruning** for forward pruning

### Evaluation Features
- **Tapered Evaluation** blending middle game and endgame scores
- **Piece-Square Tables** for positional evaluation
- **King Safety** patterns and pawn shields
- **Mobility** evaluation for all piece types
- **Passed Pawn** evaluation with advancement bonuses

## Performance History

### Development Progress
- **Benchmark History** - See [history.md](history.md) for detailed match results against various opponents
- **STS Performance** - See [sts_history.md](sts_history.md) for Strategic Test Suite validation results over time
