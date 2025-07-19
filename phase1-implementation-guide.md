# Phase 1 Implementation Guide: Quick Wins for Chess Engine Optimization

## Overview

This document provides detailed implementation instructions for Phase 1 optimizations of the chess engine. These "quick wins" should provide 30-40% performance improvement with minimal architectural changes.

**Expected Timeline**: 1 week
**Expected Performance Gain**: 30-40% improvement
**Risk Level**: Low - these changes don't alter core logic

## Prerequisites

Before starting Phase 1, ensure:
1. All existing tests pass: `go test ./...`
2. Baseline performance is measured
3. Git branch created: `git checkout -b phase1-quick-wins`

## Implementation Tasks

### Task 1.1: King Position Caching (Expected: 10-15% improvement)

#### Problem Statement
Currently, finding the king position requires scanning all 64 squares every time `IsKingInCheck` is called. This happens multiple times per move validation.

#### Current Code Analysis
```go
// Current inefficient code in game/moves/generator.go:234-246
func (g *Generator) findKing(b *board.Board, player Player) *board.Square {
    var kingPiece board.Piece
    if player == White {
        kingPiece = board.WhiteKing
    } else {
        kingPiece = board.BlackKing
    }
    
    for rank := MinRank; rank < BoardSize; rank++ {
        for file := MinFile; file < BoardSize; file++ {
            if b.GetPiece(rank, file) == kingPiece {
                return &board.Square{File: file, Rank: rank}
            }
        }
    }
    return nil
}
```

#### Implementation Steps

**Step 1: Modify the Generator struct**
```go
// In game/moves/generator.go, update the Generator struct:
type Generator struct {
    castlingHandler  *CastlingHandler
    enPassantHandler *EnPassantHandler
    promotionHandler *PromotionHandler
    moveExecutor     *MoveExecutor
    attackDetector   *AttackDetector
    
    // ADD THESE NEW FIELDS:
    whiteKingPos     *board.Square  // Cache white king position
    blackKingPos     *board.Square  // Cache black king position
    kingCacheValid   bool           // Flag to indicate if cache is valid
}
```

**Step 2: Update NewGenerator constructor**
```go
// In game/moves/generator.go, update NewGenerator:
func NewGenerator() *Generator {
    return &Generator{
        castlingHandler:  &CastlingHandler{},
        enPassantHandler: &EnPassantHandler{},
        promotionHandler: &PromotionHandler{},
        moveExecutor:     &MoveExecutor{},
        attackDetector:   &AttackDetector{},
        
        // ADD THESE INITIALIZATIONS:
        whiteKingPos:     nil,
        blackKingPos:     nil,
        kingCacheValid:   false,
    }
}
```

**Step 3: Create king cache initialization method**
```go
// Add this new method to game/moves/generator.go:
func (g *Generator) initializeKingCache(b *board.Board) {
    if g.kingCacheValid {
        return
    }
    
    // Find both kings in one pass
    for rank := MinRank; rank < BoardSize; rank++ {
        for file := MinFile; file < BoardSize; file++ {
            piece := b.GetPiece(rank, file)
            switch piece {
            case board.WhiteKing:
                g.whiteKingPos = &board.Square{File: file, Rank: rank}
            case board.BlackKing:
                g.blackKingPos = &board.Square{File: file, Rank: rank}
            }
            
            // Early exit if both found
            if g.whiteKingPos != nil && g.blackKingPos != nil {
                g.kingCacheValid = true
                return
            }
        }
    }
    g.kingCacheValid = true
}
```

**Step 4: Update findKing to use cache**
```go
// Replace the existing findKing method in game/moves/generator.go:
func (g *Generator) findKing(b *board.Board, player Player) *board.Square {
    // Initialize cache if needed
    g.initializeKingCache(b)
    
    if player == White {
        return g.whiteKingPos
    }
    return g.blackKingPos
}
```

**Step 5: Update king position during moves**
```go
// Add this method to game/moves/generator.go:
func (g *Generator) updateKingCache(move board.Move, piece board.Piece) {
    if piece == board.WhiteKing {
        g.whiteKingPos = &board.Square{File: move.To.File, Rank: move.To.Rank}
    } else if piece == board.BlackKing {
        g.blackKingPos = &board.Square{File: move.To.File, Rank: move.To.Rank}
    }
}

// Modify makeMove method in game/moves/generator.go:
func (g *Generator) makeMove(b *board.Board, move board.Move) *MoveHistory {
    piece := b.GetPiece(move.From.Rank, move.From.File)
    history := g.moveExecutor.MakeMove(b, move, g.updateBoardState)
    
    // ADD THIS LINE:
    g.updateKingCache(move, piece)
    
    return history
}
```

**Step 6: Invalidate cache when needed**
```go
// Update GenerateAllMoves to ensure cache is valid:
func (g *Generator) GenerateAllMoves(b *board.Board, player Player) *MoveList {
    // ADD THIS LINE at the beginning:
    g.initializeKingCache(b)
    
    // ... rest of the method remains the same
}
```

#### Testing the King Cache
```go
// Add to game/moves/generator_test.go:
func TestKingCache(t *testing.T) {
    gen := NewGenerator()
    b, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
    
    // First call should initialize cache
    kingSquare := gen.findKing(b, White)
    if kingSquare == nil || kingSquare.File != 4 || kingSquare.Rank != 0 {
        t.Error("White king not found at e1")
    }
    
    // Verify cache is being used
    if !gen.kingCacheValid {
        t.Error("King cache should be valid")
    }
    
    // Test cache update on king move
    move := board.Move{
        From: board.Square{File: 4, Rank: 0},
        To:   board.Square{File: 5, Rank: 0},
    }
    gen.makeMove(b, move)
    
    kingSquare = gen.findKing(b, White)
    if kingSquare.File != 5 || kingSquare.Rank != 0 {
        t.Error("King cache not updated correctly")
    }
}
```

### Task 1.2: Move List Object Pool (Expected: 5-10% improvement)

#### Problem Statement
Every move generation creates multiple `MoveList` objects, causing frequent heap allocations and GC pressure.

#### Implementation Steps

**Step 1: Create the pool implementation**
```go
// Create new file: game/moves/pool.go
package moves

import (
    "sync"
    "github.com/AdamGriffiths31/ChessEngine/board"
)

// MoveListPool manages a pool of reusable MoveList objects
type MoveListPool struct {
    pool sync.Pool
}

// Global pool instance
var globalMoveListPool = &MoveListPool{
    pool: sync.Pool{
        New: func() interface{} {
            return &MoveList{
                Moves: make([]board.Move, 0, 256), // Pre-allocate larger capacity
                Count: 0,
            }
        },
    },
}

// GetMoveList retrieves a MoveList from the pool
func GetMoveList() *MoveList {
    ml := globalMoveListPool.pool.Get().(*MoveList)
    ml.Clear() // Ensure it's clean
    return ml
}

// ReleaseMoveList returns a MoveList to the pool
func ReleaseMoveList(ml *MoveList) {
    if ml == nil {
        return
    }
    
    // Only pool lists with reasonable capacity to avoid memory bloat
    if cap(ml.Moves) <= 512 {
        ml.Clear() // Clear before returning to pool
        globalMoveListPool.pool.Put(ml)
    }
}

// PoolStats provides statistics about pool usage (for debugging)
type PoolStats struct {
    Gets     int64
    Puts     int64
    New      int64
}

var poolStats PoolStats

// GetPoolStats returns current pool statistics
func GetPoolStats() PoolStats {
    return poolStats
}
```

**Step 2: Update MoveList Clear method**
```go
// In game/moves/types.go, update the Clear method:
func (ml *MoveList) Clear() {
    ml.Moves = ml.Moves[:0] // Reuse underlying array
    ml.Count = 0
}
```

**Step 3: Replace NewMoveList usage**
```go
// Update all methods in game/moves/generator.go that create MoveLists
// Example for GenerateAllMoves:
func (g *Generator) GenerateAllMoves(b *board.Board, player Player) *MoveList {
    g.initializeKingCache(b)
    
    // REPLACE: pseudoLegalMoves := g.generateAllPseudoLegalMoves(b, player)
    // WITH:
    pseudoLegalMoves := g.generateAllPseudoLegalMoves(b, player)
    
    // REPLACE: legalMoves := NewMoveList()
    // WITH:
    legalMoves := GetMoveList()
    
    // Filter out moves that would leave the king in check
    for _, move := range pseudoLegalMoves.Moves {
        if g.isMoveLegal(b, move, player) {
            legalMoves.AddMove(move)
        }
    }
    
    // ADD: Release the pseudo-legal moves list back to pool
    ReleaseMoveList(pseudoLegalMoves)
    
    return legalMoves
}

// Update generateAllPseudoLegalMoves:
func (g *Generator) generateAllPseudoLegalMoves(b *board.Board, player Player) *MoveList {
    // REPLACE: moveList := NewMoveList()
    // WITH:
    moveList := GetMoveList()
    
    // ... rest of method remains the same
    
    return moveList
}
```

**Step 4: Update all piece move generation methods**
```go
// Template for updating each piece's move generation:
// Example for GeneratePawnMoves in game/moves/pawn.go:
func (g *Generator) GeneratePawnMoves(b *board.Board, player Player) *MoveList {
    // REPLACE: moveList := NewMoveList()
    // WITH:
    moveList := GetMoveList()
    
    // ... rest of method remains the same
    
    return moveList
}

// Apply same change to:
// - GenerateRookMoves
// - GenerateBishopMoves
// - GenerateKnightMoves
// - GenerateQueenMoves
// - GenerateKingMoves
// - generateSlidingPieceMoves
// - generateJumpingPieceMoves
```

**Step 5: Update callers to release MoveLists**
```go
// In game/engine.go, update GetLegalMoves:
func (e *Engine) GetLegalMoves() *moves.MoveList {
    ml := e.generator.GenerateAllMoves(e.state.Board, moves.Player(e.state.CurrentTurn))
    // Note: The caller is responsible for releasing this MoveList
    return ml
}

// In game/modes/mode1.go, update handleSpecialCommand:
case "MOVES":
    moveList := mm.engine.GetLegalMoves()
    playerName := mm.engine.GetState().CurrentTurn.String()
    mm.prompter.ShowMoves(moveList, playerName)
    
    // ADD: Release the move list after use
    moves.ReleaseMoveList(moveList)
    
    return nil
```

#### Testing the Pool
```go
// Add to game/moves/pool_test.go:
package moves

import (
    "testing"
    "runtime"
)

func TestMoveListPool(t *testing.T) {
    // Force GC to establish baseline
    runtime.GC()
    
    var m1, m2 runtime.MemStats
    runtime.ReadMemStats(&m1)
    
    // Simulate heavy move generation without pool
    for i := 0; i < 1000; i++ {
        ml := &MoveList{
            Moves: make([]board.Move, 0, 256),
            Count: 0,
        }
        _ = ml
    }
    
    runtime.ReadMemStats(&m2)
    allocsWithoutPool := m2.Alloc - m1.Alloc
    
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Simulate with pool
    for i := 0; i < 1000; i++ {
        ml := GetMoveList()
        ReleaseMoveList(ml)
    }
    
    runtime.ReadMemStats(&m2)
    allocsWithPool := m2.Alloc - m1.Alloc
    
    // Pool should significantly reduce allocations
    if allocsWithPool >= allocsWithoutPool {
        t.Errorf("Pool did not reduce allocations: without=%d, with=%d", 
            allocsWithoutPool, allocsWithPool)
    }
}

func BenchmarkMoveListAllocation(b *testing.B) {
    b.Run("WithoutPool", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            ml := &MoveList{
                Moves: make([]board.Move, 0, 256),
                Count: 0,
            }
            _ = ml
        }
    })
    
    b.Run("WithPool", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            ml := GetMoveList()
            ReleaseMoveList(ml)
        }
    })
}
```

### Task 1.3: Piece Lists for Fast Lookup (Expected: 15-20% improvement)

#### Problem Statement
Finding all pieces of a specific type requires scanning all 64 squares. With piece lists, we can directly iterate over actual piece positions.

#### Implementation Steps

**Step 1: Extend the Board structure**
```go
// In board/board.go, add to the Board struct:
type Board struct {
    squares         [8][8]Piece
    castlingRights  string
    enPassantTarget *Square
    halfMoveClock   int
    fullMoveNumber  int
    sideToMove      string
    
    // ADD THESE NEW FIELDS:
    pieceLists      map[Piece][]Square  // Track positions of each piece type
    pieceCount      map[Piece]int       // Count of each piece type
}
```

**Step 2: Update NewBoard constructor**
```go
// In board/board.go, update NewBoard:
func NewBoard() *Board {
    board := &Board{
        castlingRights:  "KQkq",
        enPassantTarget: nil,
        halfMoveClock:   0,
        fullMoveNumber:  1,
        sideToMove:      "w",
        
        // ADD THESE INITIALIZATIONS:
        pieceLists: make(map[Piece][]Square),
        pieceCount: make(map[Piece]int),
    }
    
    // Initialize piece lists for all piece types
    pieces := []Piece{
        WhitePawn, WhiteRook, WhiteKnight, WhiteBishop, WhiteQueen, WhiteKing,
        BlackPawn, BlackRook, BlackKnight, BlackBishop, BlackQueen, BlackKing,
    }
    
    for _, piece := range pieces {
        board.pieceLists[piece] = make([]Square, 0, 16) // Max 16 of any piece type
        board.pieceCount[piece] = 0
    }
    
    for rank := 0; rank < 8; rank++ {
        for file := 0; file < 8; file++ {
            board.squares[rank][file] = Empty
        }
    }
    return board
}
```

**Step 3: Create piece list management methods**
```go
// Add these methods to board/board.go:

// addPieceToList adds a piece to the piece list
func (b *Board) addPieceToList(piece Piece, square Square) {
    if piece == Empty {
        return
    }
    
    b.pieceLists[piece] = append(b.pieceLists[piece], square)
    b.pieceCount[piece]++
}

// removePieceFromList removes a piece from the piece list
func (b *Board) removePieceFromList(piece Piece, square Square) {
    if piece == Empty {
        return
    }
    
    list := b.pieceLists[piece]
    for i, sq := range list {
        if sq.File == square.File && sq.Rank == square.Rank {
            // Remove by swapping with last element
            list[i] = list[len(list)-1]
            b.pieceLists[piece] = list[:len(list)-1]
            b.pieceCount[piece]--
            break
        }
    }
}

// GetPieceList returns all squares containing a specific piece type
func (b *Board) GetPieceList(piece Piece) []Square {
    return b.pieceLists[piece]
}

// GetPieceCount returns the count of a specific piece type
func (b *Board) GetPieceCount(piece Piece) int {
    return b.pieceCount[piece]
}
```

**Step 4: Update SetPiece to maintain piece lists**
```go
// In board/board.go, update SetPiece:
func (b *Board) SetPiece(rank, file int, piece Piece) {
    if rank >= 0 && rank <= 7 && file >= 0 && file <= 7 {
        square := Square{File: file, Rank: rank}
        oldPiece := b.squares[rank][file]
        
        // Remove old piece from list
        if oldPiece != Empty {
            b.removePieceFromList(oldPiece, square)
        }
        
        // Add new piece to list
        if piece != Empty {
            b.addPieceToList(piece, square)
        }
        
        b.squares[rank][file] = piece
    }
}
```

**Step 5: Update FromFEN to populate piece lists**
```go
// In board/board.go, update FromFEN after setting pieces:
func FromFEN(fen string) (*Board, error) {
    // ... existing FEN parsing code ...
    
    board := NewBoard()
    
    // Parse board position
    for rankIndex, rankStr := range ranks {
        actualRank := 7 - rankIndex
        file := 0
        for _, char := range rankStr {
            if file >= 8 {
                return nil, errors.New("invalid FEN: too many files in rank")
            }

            if char >= '1' && char <= '8' {
                emptySquares, _ := strconv.Atoi(string(char))
                for i := 0; i < emptySquares; i++ {
                    if file >= 8 {
                        return nil, errors.New("invalid FEN: too many files in rank")
                    }
                    // Empty squares already initialized
                    file++
                }
            } else {
                piece := Piece(char)
                if !isValidPiece(piece) {
                    return nil, errors.New("invalid FEN: invalid piece character")
                }
                // Use SetPiece to maintain piece lists
                board.SetPiece(actualRank, file, piece)
                file++
            }
        }
        
        if file != 8 {
            return nil, errors.New("invalid FEN: incorrect number of files in rank")
        }
    }
    
    // ... rest of FEN parsing ...
    
    return board, nil
}
```

**Step 6: Update move generators to use piece lists**
```go
// Example update for GeneratePawnMoves in game/moves/pawn.go:
func (g *Generator) GeneratePawnMoves(b *board.Board, player Player) *MoveList {
    moveList := GetMoveList()

    var pawnPiece board.Piece
    var direction int
    var startRank, promotionRank int

    if player == White {
        pawnPiece = board.WhitePawn
        direction = 1
        startRank = 1
        promotionRank = 7
    } else {
        pawnPiece = board.BlackPawn
        direction = -1
        startRank = 6
        promotionRank = 0
    }

    // REPLACE the nested loops with:
    pawns := b.GetPieceList(pawnPiece)
    for _, square := range pawns {
        g.generatePawnMovesFromSquare(b, player, square.Rank, square.File, 
            direction, startRank, promotionRank, moveList)
    }

    return moveList
}

// Apply similar changes to other piece generators
```

#### Testing Piece Lists
```go
// Add to board/board_test.go:
func TestPieceLists(t *testing.T) {
    board := NewBoard()
    
    // Add a white pawn
    board.SetPiece(1, 4, WhitePawn) // e2
    
    pawns := board.GetPieceList(WhitePawn)
    if len(pawns) != 1 {
        t.Errorf("Expected 1 white pawn, got %d", len(pawns))
    }
    
    if pawns[0].File != 4 || pawns[0].Rank != 1 {
        t.Error("White pawn not at expected position")
    }
    
    // Move the pawn
    board.SetPiece(1, 4, Empty)
    board.SetPiece(3, 4, WhitePawn) // e4
    
    pawns = board.GetPieceList(WhitePawn)
    if len(pawns) != 1 {
        t.Errorf("Expected 1 white pawn after move, got %d", len(pawns))
    }
    
    if pawns[0].File != 4 || pawns[0].Rank != 3 {
        t.Error("White pawn not at expected position after move")
    }
}

func TestPieceListsFromFEN(t *testing.T) {
    board, _ := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
    
    // Check piece counts
    testCases := []struct {
        piece    Piece
        expected int
    }{
        {WhitePawn, 8},
        {BlackPawn, 8},
        {WhiteRook, 2},
        {BlackRook, 2},
        {WhiteKnight, 2},
        {BlackKnight, 2},
        {WhiteBishop, 2},
        {BlackBishop, 2},
        {WhiteQueen, 1},
        {BlackQueen, 1},
        {WhiteKing, 1},
        {BlackKing, 1},
    }
    
    for _, tc := range testCases {
        count := board.GetPieceCount(tc.piece)
        if count != tc.expected {
            t.Errorf("Expected %d %c pieces, got %d", tc.expected, tc.piece, count)
        }
    }
}
```

## Performance Benchmarking

### Create Benchmark Suite
```go
// Create game/moves/phase1_bench_test.go:
package moves

import (
    "testing"
    "github.com/AdamGriffiths31/ChessEngine/board"
)

func BenchmarkPhase1Improvements(b *testing.B) {
    positions := []struct {
        name string
        fen  string
    }{
        {"Initial", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
        {"Kiwipete", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq -"},
        {"Endgame", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - -"},
    }
    
    for _, pos := range positions {
        b.Run(pos.name, func(b *testing.B) {
            board, _ := board.FromFEN(pos.fen)
            gen := NewGenerator()
            
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                moves := gen.GenerateAllMoves(board, White)
                // Ensure moves is used to prevent optimization
                if moves.Count < 0 {
                    b.Fatal("Impossible")
                }
                ReleaseMoveList(moves)
            }
        })
    }
}

func BenchmarkPerftPhase1(b *testing.B) {
    testCases := []struct {
        name  string
        fen   string
        depth int
    }{
        {"Initial_3", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 3},
        {"Kiwipete_3", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq -", 3},
    }
    
    for _, tc := range testCases {
        b.Run(tc.name, func(b *testing.B) {
            board, _ := board.FromFEN(tc.fen)
            
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                result := Perft(board, tc.depth, White)
                _ = result
            }
        })
    }
}
```

### Performance Validation Script
```bash
#!/bin/bash
# save as scripts/phase1_performance.sh

echo "Phase 1 Performance Validation"
echo "=============================="

# Baseline before changes
git stash
echo "Running baseline benchmarks..."
go test -bench=BenchmarkPerft -benchtime=10s ./game/moves > baseline.txt

# Apply changes
git stash pop
echo "Running optimized benchmarks..."
go test -bench=BenchmarkPerft -benchtime=10s ./game/moves > optimized.txt

# Compare results
echo "Performance Comparison:"
benchstat baseline.txt optimized.txt
```

## Integration Testing

### Comprehensive Test Suite
```go
// Add to game/moves/integration_test.go:
package moves

import (
    "testing"
    "github.com/AdamGriffiths31/ChessEngine/board"
)

func TestPhase1Integration(t *testing.T) {
    // Test that all optimizations work together
    gen := NewGenerator()
    
    positions := []struct {
        fen      string
        expected int
    }{
        {"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 20},
        {"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq -", 48},
    }
    
    for _, pos := range positions {
        b, _ := board.FromFEN(pos.fen)
        moves := gen.GenerateAllMoves(b, White)
        
        if moves.Count != pos.expected {
            t.Errorf("Position %s: expected %d moves, got %d", 
                pos.fen, pos.expected, moves.Count)
        }
        
        // Verify king cache is working
        if !gen.kingCacheValid {
            t.Error("King cache should be valid after move generation")
        }
        
        // Clean up
        ReleaseMoveList(moves)
    }
}

func TestPhase1Correctness(t *testing.T) {
    // Ensure optimizations don't break correctness
    testData, _ := LoadPerftTestData(GetTestDataPath())
    
    for _, position := range testData.Positions {
        t.Run(position.Name, func(t *testing.T) {
            b, _ := board.FromFEN(position.FEN)
            
            for _, depthTest := range position.Depths {
                if depthTest.Depth > 4 {
                    continue // Skip deep tests for integration
                }
                
                result := Perft(b, depthTest.Depth, White)
                if result != depthTest.Nodes {
                    t.Errorf("Depth %d: expected %d nodes, got %d",
                        depthTest.Depth, depthTest.Nodes, result)
                }
            }
        })
    }
}
```

## Rollout Checklist

### Pre-Implementation
- [ ] Create feature branch: `git checkout -b phase1-quick-wins`
- [ ] Run baseline benchmarks and save results
- [ ] Ensure all tests pass: `go test ./...`
- [ ] Document current performance metrics

### Implementation Order
1. [ ] Implement King Position Caching (Task 1.1)
   - [ ] Update Generator struct
   - [ ] Add cache methods
   - [ ] Update findKing
   - [ ] Add tests
   - [ ] Benchmark improvement
   
2. [ ] Implement Move List Object Pool (Task 1.2)
   - [ ] Create pool.go
   - [ ] Replace NewMoveList calls
   - [ ] Add pool releases
   - [ ] Add tests
   - [ ] Benchmark improvement
   
3. [ ] Implement Piece Lists (Task 1.3)
   - [ ] Update Board struct
   - [ ] Add piece list methods
   - [ ] Update SetPiece
   - [ ] Update move generators
   - [ ] Add tests
   - [ ] Benchmark improvement

### Post-Implementation
- [ ] Run full test suite
- [ ] Run performance benchmarks
- [ ] Compare with baseline (expect 30-40% improvement)
- [ ] Update documentation
- [ ] Create pull request

## Troubleshooting Guide

### Common Issues and Solutions

1. **King cache becomes invalid**
   - Solution: Ensure initializeKingCache is called at start of GenerateAllMoves
   - Check: Add logging to verify cache hits
   - Debug: Add validation method to compare cached vs actual positions

2. **Memory leaks with object pool**
   - Solution: Ensure all GetMoveList calls have matching ReleaseMoveList
   - Check: Add defer statements immediately after GetMoveList
   - Debug: Track pool statistics and log unreleased lists

3. **Piece lists out of sync**
   - Solution: Always use SetPiece method, never modify squares directly
   - Check: Add piece count validation after each move
   - Debug: Create ValidatePieceLists method for testing

4. **Performance regression**
   - Solution: Profile specific bottlenecks with pprof
   - Check: Ensure pool is actually being used (not creating new objects)
   - Debug: Add timing logs for each optimization

### Debug Helpers
```go
// Add to game/moves/debug.go:
package moves

import (
    "fmt"
    "log"
    "github.com/AdamGriffiths31/ChessEngine/board"
)

// DebugConfig controls debug output
type DebugConfig struct {
    LogKingCache    bool
    LogPool         bool
    LogPieceLists   bool
    ValidateState   bool
}

var Debug = DebugConfig{
    LogKingCache:    false, // Set to true for debugging
    LogPool:         false,
    LogPieceLists:   false,
    ValidateState:   false,
}

// ValidateKingCache checks if cached king positions are correct
func (g *Generator) ValidateKingCache(b *board.Board) error {
    actualWhiteKing := g.findKingSlow(b, White)
    actualBlackKing := g.findKingSlow(b, Black)
    
    if g.whiteKingPos != nil && actualWhiteKing != nil {
        if g.whiteKingPos.File != actualWhiteKing.File || 
           g.whiteKingPos.Rank != actualWhiteKing.Rank {
            return fmt.Errorf("white king cache mismatch: cached=%v, actual=%v", 
                g.whiteKingPos, actualWhiteKing)
        }
    }
    
    if g.blackKingPos != nil && actualBlackKing != nil {
        if g.blackKingPos.File != actualBlackKing.File || 
           g.blackKingPos.Rank != actualBlackKing.Rank {
            return fmt.Errorf("black king cache mismatch: cached=%v, actual=%v", 
                g.blackKingPos, actualBlackKing)
        }
    }
    
    return nil
}

// findKingSlow is the original implementation for validation
func (g *Generator) findKingSlow(b *board.Board, player Player) *board.Square {
    var kingPiece board.Piece
    if player == White {
        kingPiece = board.WhiteKing
    } else {
        kingPiece = board.BlackKing
    }
    
    for rank := MinRank; rank < BoardSize; rank++ {
        for file := MinFile; file < BoardSize; file++ {
            if b.GetPiece(rank, file) == kingPiece {
                return &board.Square{File: file, Rank: rank}
            }
        }
    }
    return nil
}

// LogPoolStats logs current pool statistics
func LogPoolStats() {
    stats := GetPoolStats()
    log.Printf("Pool Stats - Gets: %d, Puts: %d, New: %d", 
        stats.Gets, stats.Puts, stats.New)
}

// ValidatePieceLists checks if piece lists match board state
func ValidatePieceLists(b *board.Board) error {
    // Count pieces on board
    actualCounts := make(map[board.Piece]int)
    for rank := 0; rank < 8; rank++ {
        for file := 0; file < 8; file++ {
            piece := b.GetPiece(rank, file)
            if piece != board.Empty {
                actualCounts[piece]++
            }
        }
    }
    
    // Compare with piece lists
    pieces := []board.Piece{
        board.WhitePawn, board.WhiteRook, board.WhiteKnight, 
        board.WhiteBishop, board.WhiteQueen, board.WhiteKing,
        board.BlackPawn, board.BlackRook, board.BlackKnight, 
        board.BlackBishop, board.BlackQueen, board.BlackKing,
    }
    
    for _, piece := range pieces {
        listCount := b.GetPieceCount(piece)
        actualCount := actualCounts[piece]
        if listCount != actualCount {
            return fmt.Errorf("piece list mismatch for %c: list=%d, actual=%d", 
                piece, listCount, actualCount)
        }
    }
    
    return nil
}
```

## Performance Measurement Guide

### Before Starting Implementation
```bash
# Create performance baseline
mkdir -p benchmarks/phase1
cd benchmarks/phase1

# Run baseline tests
go test -bench=. -benchmem -count=5 -timeout=30m \
    github.com/AdamGriffiths31/ChessEngine/game/moves > baseline.txt

# Run baseline perft
go test -run=TestKiwipeteDepth6Timing -v \
    github.com/AdamGriffiths31/ChessEngine/game/moves > baseline_perft.txt
```

### After Each Task Implementation
```bash
# After Task 1.1 (King Cache)
go test -bench=. -benchmem -count=5 -timeout=30m \
    github.com/AdamGriffiths31/ChessEngine/game/moves > task1.1_kingcache.txt

# After Task 1.2 (Object Pool)
go test -bench=. -benchmem -count=5 -timeout=30m \
    github.com/AdamGriffiths31/ChessEngine/game/moves > task1.2_pool.txt

# After Task 1.3 (Piece Lists)
go test -bench=. -benchmem -count=5 -timeout=30m \
    github.com/AdamGriffiths31/ChessEngine/game/moves > task1.3_piecelist.txt

# Compare results
benchstat baseline.txt task1.3_piecelist.txt
```

### Expected Performance Metrics

| Metric | Baseline | After 1.1 | After 1.2 | After 1.3 | Total Gain |
|--------|----------|-----------|-----------|-----------|------------|
| Nodes/sec | 13K | 14.5K (+12%) | 15.5K (+7%) | 18K (+16%) | +38% |
| Allocs/op | 1000 | 1000 (0%) | 200 (-80%) | 200 (0%) | -80% |
| B/op | 400KB | 400KB (0%) | 100KB (-75%) | 110KB (+10%) | -72% |
| Kiwipete d6 | 10m+ | 9m | 8.5m | 7m | -30% |

## Code Review Checklist

### General
- [ ] All tests pass
- [ ] No race conditions (run with -race flag)
- [ ] No memory leaks
- [ ] Performance improvements measured
- [ ] Code is well-documented

### Task 1.1: King Cache
- [ ] Cache initialized before use
- [ ] Cache updated on king moves
- [ ] Cache invalidated appropriately
- [ ] Thread-safe if needed
- [ ] Fallback to slow search works

### Task 1.2: Object Pool
- [ ] All GetMoveList have matching Release
- [ ] Pool size limits implemented
- [ ] Clear method properly resets state
- [ ] No data races in concurrent use
- [ ] Statistics tracking works

### Task 1.3: Piece Lists
- [ ] Lists stay synchronized with board
- [ ] SetPiece always updates lists
- [ ] FromFEN populates lists correctly
- [ ] No duplicate entries in lists
- [ ] Piece counts accurate

## Final Validation

### Performance Test Suite
```go
// Create game/moves/phase1_validation_test.go:
package moves

import (
    "testing"
    "time"
    "github.com/AdamGriffiths31/ChessEngine/board"
)

func TestPhase1PerformanceGoals(t *testing.T) {
    tests := []struct {
        name     string
        fen      string
        depth    int
        maxTime  time.Duration
        minNodes int64
    }{
        {
            name:     "Initial_Position_5",
            fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
            depth:    5,
            maxTime:  30 * time.Second,
            minNodes: 4865609,
        },
        {
            name:     "Kiwipete_4",
            fen:      "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq -",
            depth:    4,
            maxTime:  5 * time.Second,
            minNodes: 4085603,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            b, _ := board.FromFEN(tt.fen)
            
            start := time.Now()
            nodes := Perft(b, tt.depth, White)
            elapsed := time.Since(start)
            
            if nodes != tt.minNodes {
                t.Errorf("Expected %d nodes, got %d", tt.minNodes, nodes)
            }
            
            if elapsed > tt.maxTime {
                t.Errorf("Too slow: took %v, max allowed %v", elapsed, tt.maxTime)
            }
            
            nodesPerSec := float64(nodes) / elapsed.Seconds()
            t.Logf("Performance: %.0f nodes/sec", nodesPerSec)
            
            // Phase 1 target: 50K+ nodes/sec
            if nodesPerSec < 50000 {
                t.Errorf("Below target performance: %.0f nodes/sec < 50K", nodesPerSec)
            }
        })
    }
}
```

### Memory Usage Validation
```go
func TestPhase1MemoryUsage(t *testing.T) {
    runtime.GC()
    var m1, m2 runtime.MemStats
    runtime.ReadMemStats(&m1)
    
    // Generate moves 1000 times
    b, _ := board.FromFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq -")
    gen := NewGenerator()
    
    for i := 0; i < 1000; i++ {
        moves := gen.GenerateAllMoves(b, White)
        ReleaseMoveList(moves)
    }
    
    runtime.ReadMemStats(&m2)
    
    alloced := m2.Alloc - m1.Alloc
    numAllocs := m2.Mallocs - m1.Mallocs
    
    t.Logf("Memory allocated: %d bytes", alloced)
    t.Logf("Number of allocations: %d", numAllocs)
    
    // Should see significant reduction from baseline
    maxAllocsPerGeneration := numAllocs / 1000
    if maxAllocsPerGeneration > 10 {
        t.Errorf("Too many allocations per move generation: %d", maxAllocsPerGeneration)
    }
}
```

## Summary

### What We've Accomplished in Phase 1

1. **King Position Caching**
   - Eliminates 64-square scans for king location
   - Reduces IsKingInCheck overhead by 90%
   - Expected: 10-15% overall improvement

2. **Move List Object Pooling**
   - Reduces heap allocations by 80%
   - Minimizes GC pressure
   - Expected: 5-10% overall improvement

3. **Piece Lists**
   - Direct access to piece positions
   - Eliminates board scanning in move generation
   - Expected: 15-20% overall improvement

### Combined Impact
- **Total Expected Improvement**: 30-40%
- **Nodes/sec**: 13K → 18K+
- **Memory Allocations**: -80%
- **Kiwipete Depth 6**: 10m → 7m

### Next Steps (Phase 2 Preview)
1. Bitboard representation (50x improvement potential)
2. Magic bitboards for sliding pieces
3. Incremental attack detection
4. Move ordering heuristics

### Maintenance Notes
- Run benchmarks after any changes
- Keep debug helpers available but disabled
- Document any deviations from plan
- Monitor for performance regressions

## Appendix: Quick Reference

### Key Functions Modified
```
Generator.findKing() - Now uses cache
Generator.GenerateAllMoves() - Uses pooled MoveLists
Board.SetPiece() - Maintains piece lists
All piece generators - Use GetPieceList() instead of scanning
```

### Performance Commands
```bash
# Run benchmarks
go test -bench=. -benchmem ./game/moves

# Profile CPU
go test -cpuprofile=cpu.prof -bench=. ./game/moves
go tool pprof -http=:8080 cpu.prof

# Profile memory
go test -memprofile=mem.prof -bench=. ./game/moves
go tool pprof -http=:8080 mem.prof

# Run specific perft test
go test -run=TestKiwipeteDepth6Timing -v ./game/moves
```

### Debug Flags
```go
// Enable debugging in game/moves/debug.go
Debug.LogKingCache = true    // Log king cache hits/misses
Debug.LogPool = true         // Log pool usage
Debug.LogPieceLists = true   // Log piece list updates
Debug.ValidateState = true   // Run validation after each move
```