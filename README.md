# Chess Engine Performance Optimization Plan

## Problem Analysis

The kiwipete_depth6 perft test is taking 10+ minutes instead of seconds, indicating severe performance bottlenecks. The current implementation has several critical issues:

1. **Expensive Legal Move Validation**: Every pseudo-legal move requires makeMove → IsKingInCheck → unmakeMove
2. **Inefficient Board Scanning**: Full 8x8 board scans for each piece type repeatedly  
3. **Memory Allocation Overhead**: Frequent MoveList object creation
4. **Attack Detection Inefficiency**: Repeated IsSquareAttacked calls without caching
5. **No Incremental Updates**: Complete board state recalculation after each move

Expected nodes for kiwipete_depth6: **8,031,647,685** (should complete in seconds, not minutes)

## Stage 1: Core Performance Infrastructure (Priority: Critical)
**Target**: Reduce move generation time by 10-50x

### 1.1 Bitboard System Implementation
- Replace array-based board representation with 64-bit bitboards
- Create bitboards for each piece type (whitePawns, blackPawns, whiteRooks, etc.)
- Implement bitwise operations for piece movements and attacks
- Add utility functions for bitboard manipulation (LSB, MSB, popcount)

**Files to modify:**
- `board/bitboard.go` (new)
- `board/board.go` (major refactor)
- `game/moves/types.go` (add bitboard types)

### 1.2 Piece Lists and Incremental Updates
- Maintain piece lists to avoid board scanning (max 16 pieces per type)
- Update piece lists incrementally during make/unmake moves
- Cache king positions for fast check detection
- Implement zobrist hashing for position identification

**Files to modify:**
- `game/moves/piece_lists.go` (new)
- `game/moves/generator.go` (major refactor)
- `game/moves/board_utils.go` (update MoveExecutor)

### 1.3 Pre-computed Attack Tables
- Generate attack tables for knights (64 entries)
- Generate attack tables for kings (64 entries)  
- Create ray tables for sliding pieces
- Implement distance and direction lookup tables

**Files to modify:**
- `game/moves/attack_tables.go` (new)
- `game/moves/attacks.go` (major refactor)

## Stage 2: Move Generation Optimization (Priority: High)
**Target**: Achieve 1M+ nodes/second in perft tests

### 2.1 Magic Bitboards for Sliding Pieces
- Implement magic bitboards for rooks and bishops
- Pre-calculate magic numbers and attack databases
- Optimize queen moves as combination of rook + bishop attacks
- Add blocking piece detection using bitboard intersections

**Files to modify:**
- `game/moves/magic_bitboards.go` (new)
- `game/moves/sliding_pieces.go` (new)
- `game/moves/generator.go` (update sliding piece generation)

### 2.2 Optimized Pseudo-Legal Move Generation
- Use bitboard operations for piece movement calculations
- Implement bulk move generation (generate all moves of a type at once)
- Add move ordering hints during generation
- Optimize pawn move generation with bitboard shifts

**Files to modify:**
- `game/moves/pawn.go` (major refactor)
- `game/moves/knight.go` (major refactor)
- `game/moves/king.go` (major refactor)
- `game/moves/generator.go` (major refactor)

### 2.3 Fast Legal Move Filtering
- Implement pinned piece detection using bitboards
- Add discovered check detection
- Optimize king safety checks
- Cache attack information between moves

**Files to modify:**
- `game/moves/pins.go` (new)
- `game/moves/legal_moves.go` (new)
- `game/moves/generator.go` (update legal move filtering)

## Stage 3: Memory and Algorithm Optimization (Priority: Medium)
**Target**: Reduce memory allocations and improve cache efficiency

### 3.1 Memory Management
- Implement object pooling for MoveList and frequently allocated objects
- Use stack-based move history instead of dynamic allocation
- Pre-allocate move arrays with sufficient capacity
- Implement memory-efficient move representation

**Files to modify:**
- `game/moves/memory_pool.go` (new)
- `game/moves/types.go` (optimize move representation)
- `game/moves/generator.go` (use pooled objects)

### 3.2 Algorithm Optimizations
- Add bulk operations to reduce function call overhead
- Implement compiler hints for hot paths (inline, likely/unlikely)
- Optimize critical loops with manual unrolling where beneficial
- Add SIMD optimizations for bitboard operations

**Files to modify:**
- `game/moves/optimizations.go` (new)
- `game/moves/bitboard_ops.go` (new)
- All move generation files (add inlining hints)

### 3.3 Cache-Friendly Data Structures
- Organize data structures for better cache locality
- Minimize pointer indirection in hot paths
- Use bit-packed structures where appropriate
- Implement copy-make for shallow position copies

**Files to modify:**
- `game/moves/types.go` (optimize data layout)
- `board/board.go` (cache-friendly board representation)
- `game/moves/generator.go` (minimize pointer dereferencing)

## Stage 4: Advanced Optimizations (Priority: Low)
**Target**: Achieve competitive engine performance (5M+ nodes/second)

### 4.1 Parallel Processing
- Implement parallel perft using goroutines
- Add work-stealing for load balancing
- Optimize for multi-core systems
- Add parallel move generation for complex positions

**Files to modify:**
- `game/moves/parallel_perft.go` (new)
- `game/moves/perft.go` (add parallel option)

### 4.2 Transposition Tables
- Implement hash tables for position caching
- Add move ordering based on hash moves
- Implement replacement schemes (always replace, depth preferred)
- Add collision detection and handling

**Files to modify:**
- `game/moves/transposition_table.go` (new)
- `game/moves/zobrist.go` (new)
- `game/moves/perft.go` (add transposition table usage)

### 4.3 Assembly Optimizations
- Use CPU-specific optimizations for critical functions
- Implement fast bit manipulation routines
- Add vectorized operations for multiple piece calculations
- Optimize for specific processor architectures

**Files to modify:**
- `game/moves/asm_amd64.s` (new)
- `game/moves/bitboard_ops.go` (add assembly calls)

## Implementation Strategy

### Phase 1: Foundation (Week 1-2)
1. **Day 1-3**: Implement basic bitboard system and utilities
2. **Day 4-6**: Create piece lists and incremental updates
3. **Day 7-10**: Add pre-computed attack tables
4. **Day 11-14**: Ensure all existing tests pass with new system

### Phase 2: Core Optimizations (Week 3-4)
1. **Day 15-18**: Implement magic bitboards for sliding pieces
2. **Day 19-22**: Optimize move generation algorithms
3. **Day 23-26**: Add fast legal move filtering
4. **Day 27-28**: Target 10x performance improvement

### Phase 3: Refinement (Week 5-6)
1. **Day 29-32**: Implement memory optimizations
2. **Day 33-36**: Add cache-friendly data structures
3. **Day 37-40**: Optimize critical algorithms
4. **Day 41-42**: Target 50x performance improvement

### Phase 4: Advanced Features (Week 7-8)
1. **Day 43-46**: Add parallel processing
2. **Day 47-50**: Implement transposition tables
3. **Day 51-54**: Add assembly optimizations
4. **Day 55-56**: Target 100x+ performance improvement

## Success Metrics

### Performance Targets
- **kiwipete_depth6**: Complete in < 10 seconds (currently 10+ minutes)
- **initial_position_depth6**: Complete in < 30 seconds
- **Nodes per second**: Achieve 1M+ nodes/second minimum
- **Memory usage**: Reduce allocation overhead by 90%

### Benchmarking Command
```bash
go test -run=TestKiwipeteDepth6Timing -v
```

### Expected Results by Stage
- **Stage 1**: 10x improvement (1-2 minutes → 6-12 seconds)
- **Stage 2**: 50x improvement (10+ minutes → 10-20 seconds)
- **Stage 3**: 100x improvement (10+ minutes → 5-10 seconds)
- **Stage 4**: 200x+ improvement (10+ minutes → 1-5 seconds)

## Code Quality Standards

### Testing Requirements
- Maintain 100% test coverage for existing functionality
- Add comprehensive performance benchmarks
- Run full perft test suite after each major change
- Use property-based testing for move generation

### Documentation Standards
- Document all performance-critical sections
- Add inline comments explaining optimization techniques
- Maintain clear separation between optimization and logic
- Include performance analysis for each optimization

## Risk Mitigation

### Technical Risks
- **Complexity**: Implement incrementally with continuous testing
- **Correctness**: Maintain rigorous test suite throughout development
- **Performance**: Profile at each stage to ensure improvements
- **Maintainability**: Keep clear separation between optimization and logic

### Testing Strategy
- Run full perft test suite after each optimization
- Add performance regression tests
- Implement continuous benchmarking
- Use property-based testing for move generation

## Current Baseline Performance

### Perft Test Results (as of 2025-07-17)
- **kiwipete_depth6**: DNF (10+ minutes) - Target: 8,031,647,685 nodes
- **initial_position_depth6**: Not tested - Target: 119,060,324 nodes

### Critical Bottlenecks Identified
1. **Move Generation**: `generator.go:37-49` - Full board scanning
2. **Legal Move Validation**: `generator.go:262-305` - Make/unmake for each move
3. **Attack Detection**: `attacks.go` - Repeated square attack calculations
4. **Memory Allocation**: `types.go` - Frequent MoveList creation

## Expected Outcomes

### Short-term (1-2 weeks)
- 10x performance improvement in perft tests
- Solid foundation for future optimizations
- All existing functionality preserved

### Medium-term (1-2 months)
- 50-100x performance improvement
- Competitive chess engine performance
- Scalable architecture for AI features

### Long-term (2+ months)
- Foundation for advanced chess AI
- Tournament-ready performance
- Extensible optimization framework

## Getting Started

### Prerequisites
- Go 1.19+ for latest performance features
- Benchmarking tools: `go test -bench=.`
- Profiling tools: `go tool pprof`

### First Steps
1. Run current performance baseline: `go test -run=TestKiwipeteDepth6Timing -v`
2. Begin with Stage 1.1: Bitboard system implementation
3. Set up continuous benchmarking pipeline
4. Document performance improvements at each stage

---

**Note**: This plan represents a complete transformation from a functional chess engine to a high-performance chess engine. Each stage builds upon the previous one, with careful attention to maintaining correctness while dramatically improving performance.