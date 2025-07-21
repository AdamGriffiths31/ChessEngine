# Chess Engine: Player vs Computer Mode Implementation Plan

## Overview
This document outlines the plan for implementing a Player vs Computer mode for the chess engine, with a focus on creating a modular, testable architecture that allows for easy swapping of search algorithms and evaluation functions.

## Architecture Design

### Core Components

#### 1. AI Engine Interface (`game/ai/engine.go`)
- **Purpose**: Define a common interface for all AI implementations
- **Key Methods**:
  - `FindBestMove(board, timeLimit) -> Move`
  - `SetDifficulty(level)`
  - `Stop()` - for interrupting long searches
  - `GetStatistics() -> SearchStats`

#### 2. Search Algorithms (`game/ai/search/`)
- **Minimax** (`minimax.go`)
  - Basic minimax with configurable depth
  - Alpha-beta pruning variant
- **Negamax** (`negamax.go`)
  - Simplified minimax variant
- **Future Options**:
  - Iterative deepening
  - MTD(f)
  - Monte Carlo Tree Search

#### 3. Position Evaluation (`game/ai/evaluation/`)
- **Interface** (`evaluator.go`)
  - `Evaluate(board, player) -> score`
- **Basic Material Evaluator** (`material.go`)
  - Piece values only
- **Positional Evaluator** (`positional.go`)
  - Piece-square tables
  - Mobility
  - King safety
- **Composite Evaluator** (`composite.go`)
  - Combines multiple evaluators with weights

#### 4. Move Ordering (`game/ai/ordering/`)
- **Purpose**: Order moves for better alpha-beta pruning
- **Strategies**:
  - MVV-LVA (Most Valuable Victim - Least Valuable Attacker)
  - History heuristic
  - Killer moves
  - Hash move (when transposition table is added)

#### 5. Time Management (`game/ai/time/`)
- **Simple Time Control** (`basic.go`)
  - Fixed time per move
  - Percentage of remaining time
- **Advanced Time Control** (`advanced.go`)
  - Dynamic allocation based on position complexity
  - Sudden death handling

### Game Mode Implementation

#### 6. Computer Player (`game/ai/computer_player.go`)
- Wraps AI engine with game-specific logic
- Handles move validation and execution
- Manages thinking time display

#### 7. Mode 2: Player vs Computer (`game/modes/mode2.go`)
- Extends base game mode functionality
- Manages turn alternation between human and computer
- Handles difficulty selection
- Provides UI for computer thinking feedback

## Folder Structure

```
game/
├── ai/
│   ├── engine.go              # AI engine interface
│   ├── computer_player.go     # Computer player implementation
│   ├── types.go              # Common types (SearchStats, Config, etc.)
│   ├── search/
│   │   ├── search.go         # Search interface
│   │   ├── minimax.go        # Minimax implementation
│   │   ├── negamax.go        # Negamax implementation
│   │   ├── alphabeta.go      # Alpha-beta pruning
│   │   └── search_test.go
│   ├── evaluation/
│   │   ├── evaluator.go      # Evaluation interface
│   │   ├── material.go       # Material-only evaluation
│   │   ├── positional.go     # Positional evaluation
│   │   ├── composite.go      # Composite evaluator
│   │   ├── tables.go         # Piece-square tables
│   │   └── evaluation_test.go
│   ├── ordering/
│   │   ├── orderer.go        # Move ordering interface
│   │   ├── mvv_lva.go        # MVV-LVA implementation
│   │   ├── history.go        # History heuristic
│   │   └── ordering_test.go
│   └── time/
│       ├── controller.go     # Time control interface
│       ├── basic.go          # Basic time management
│       └── time_test.go
├── modes/
│   ├── mode1.go              # Existing manual mode
│   ├── mode2.go              # NEW: Player vs Computer
│   └── mode_test.go
└── engine.go                  # Core game engine
```

## Implementation Steps

### Phase 1: Foundation (Week 1)
1. **Create AI interfaces**
   - Define `AIEngine` interface
   - Define `Evaluator` interface
   - Define `SearchAlgorithm` interface
   - Create basic types (`SearchConfig`, `SearchStats`, etc.)

2. **Implement basic material evaluation**
   - Piece values (Q=9, R=5, B=3, N=3, P=1)
   - Simple board evaluation summing material

3. **Implement basic minimax**
   - Fixed depth search (start with depth 3-4)
   - No optimizations initially
   - Return first legal move if search fails

### Phase 2: Core Search (Week 2)
1. **Add alpha-beta pruning**
   - Extend minimax with alpha-beta bounds
   - Add move ordering interface
   - Implement basic move ordering (captures first)

2. **Create computer player wrapper**
   - Implement `ComputerPlayer` struct
   - Handle move selection and execution
   - Add thinking feedback mechanism

3. **Implement Mode 2**
   - Create `ComputerMode` struct
   - Handle game flow for PvC
   - Add difficulty selection UI

### Phase 3: Enhancements (Week 3)
1. **Improve evaluation**
   - Add piece-square tables
   - Implement positional factors
   - Create composite evaluator

2. **Optimize move ordering**
   - Implement MVV-LVA
   - Add killer move heuristic
   - Track search statistics

3. **Add time management**
   - Implement basic time controller
   - Add iterative deepening
   - Handle time pressure

### Phase 4: Testing & Refinement (Week 4)
1. **Comprehensive testing**
   - Unit tests for each component
   - Integration tests for full games
   - Performance benchmarks

2. **Difficulty levels**
   - Beginner (depth 2-3, basic eval)
   - Intermediate (depth 4-5, positional eval)
   - Advanced (depth 6+, full features)

3. **Polish and optimization**
   - Profile and optimize hot paths
   - Add opening book (optional)
   - Implement pondering (optional)

## Testing Strategy

### Unit Tests
1. **Evaluation Tests** (`evaluation_test.go`)
   - Known positions with expected scores
   - Symmetry tests (mirrored positions)
   - Incremental update tests

2. **Search Tests** (`search_test.go`)
   - Tactical positions (mate in N)
   - Known best moves
   - Performance benchmarks

3. **Move Ordering Tests** (`ordering_test.go`)
   - Verify correct ordering
   - Performance impact measurements

### Integration Tests
1. **Self-play Tests**
   - Computer vs Computer games
   - Verify game completion
   - Check for illegal moves

2. **Position Tests**
   - Standard test positions
   - Bratko-Kopec test suite
   - Win At Chess positions

3. **Time Management Tests**
   - Ensure moves within time limit
   - Test sudden death scenarios

### Performance Tests
1. **Search Benchmarks**
   - Nodes per second
   - Time to depth
   - Move ordering efficiency

2. **Memory Usage**
   - Peak memory during search
   - Allocation patterns

## Configuration & Extensibility

### AI Configuration
```go
type AIConfig struct {
    SearchDepth      int
    TimePerMove      time.Duration
    UseOpeningBook   bool
    UseEndgameTable  bool
    EvaluationWeights map[string]float64
}
```

### Difficulty Presets
```go
var DifficultyPresets = map[string]AIConfig{
    "beginner": {
        SearchDepth: 3,
        TimePerMove: 1 * time.Second,
    },
    "intermediate": {
        SearchDepth: 5,
        TimePerMove: 5 * time.Second,
    },
    "advanced": {
        SearchDepth: 7,
        TimePerMove: 15 * time.Second,
    },
}
```

## Future Enhancements

### Near-term
1. **Transposition Tables**
   - Zobrist hashing
   - Memory-bounded hash table
   - Move ordering from hash

2. **Opening Book**
   - ECO classifications
   - Popular opening lines
   - Book learning from games

3. **Endgame Tablebases**
   - Basic endgames (KQ vs K, KR vs K)
   - Syzygy tablebase support

### Long-term
1. **Neural Network Evaluation**
   - Train on master games
   - NNUE-style architecture
   - GPU acceleration

2. **Parallel Search**
   - Split search across cores
   - Lazy SMP
   - YBWC algorithm

3. **Advanced Search**
   - Null move pruning
   - Late move reductions
   - Singular extensions

## Success Criteria

### Functional Requirements
- [ ] Computer makes legal moves
- [ ] Games complete without errors
- [ ] Difficulty levels provide appropriate challenge
- [ ] Response time under 10 seconds per move

### Performance Requirements
- [ ] Search 1M+ nodes/second
- [ ] Reach depth 6 in under 5 seconds
- [ ] Memory usage under 100MB during search

### Quality Requirements
- [ ] 90%+ test coverage
- [ ] No race conditions
- [ ] Clean, modular architecture
- [ ] Easy to add new search algorithms

## Development Guidelines

### Code Organization
1. **Single Responsibility**: Each component has one clear purpose
2. **Interface Segregation**: Small, focused interfaces
3. **Dependency Injection**: Pass dependencies explicitly
4. **Testability First**: Design for testing from the start

### Testing Approach
1. **Test-Driven Development**: Write tests first
2. **Mock Dependencies**: Use interfaces for easy mocking
3. **Benchmark Critical Paths**: Measure performance regularly
4. **Integration Tests**: Test component interactions

### Documentation
1. **Interface Documentation**: Clear contract definitions
2. **Algorithm Explanations**: Document non-obvious logic
3. **Performance Notes**: Document optimization decisions
4. **Usage Examples**: Show how to use each component

## Timeline Summary

- **Week 1**: Foundation - Interfaces, basic evaluation, simple minimax
- **Week 2**: Core Search - Alpha-beta, computer player, game mode
- **Week 3**: Enhancements - Better evaluation, move ordering, time control
- **Week 4**: Polish - Testing, difficulty tuning, optimizations

Total estimated time: 4 weeks for fully functional Player vs Computer mode with swappable components and comprehensive testing.
