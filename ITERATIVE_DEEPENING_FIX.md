# Iterative Deepening Fix for Chess Engine

## Problem Description

The chess engine was exhibiting critical tactical blindness, choosing obviously bad moves like e2e3 that hang pieces. Analysis revealed the root cause: **lack of iterative deepening**.

### Specific Issue
In position: `rnbqkb1r/pp4pp/2pp4/5pN1/2P1p1n1/PPNP4/4PPPP/R1BQKB1R w KQkq f6 0 8`

**Before Fix:**
- With timeout: Engine chooses e2e3 (score -300, hangs knight on g5)
- Reason: Engine searches first move to depth 6, times out, returns that move
- Never evaluates other moves like g5f7 (score +5) or g5h3 (score -130)

**Root Cause:**
```go
// Old problematic logic in FindBestMove
for i := 0; i < legalMoves.Count; i++ {
    move := legalMoves.Moves[i]
    // Search THIS move to full depth (e.g., depth 6)
    score := -m.minimaxWithDepthTracking(ctx, b, oppositePlayer(player), config.MaxDepth-1, ...)
    // TIMEOUT occurs here on first move, never tries others!
}
```

## Solution: Iterative Deepening as Default

### Design Principles
1. **No Configuration**: Behavior automatically determined by context type
2. **Backward Compatible**: Existing tests and fixed-depth analysis unchanged
3. **Smart Defaults**: Use iterative deepening when timeout exists, fixed-depth otherwise

### Implementation Strategy

#### Context-Based Behavior Selection
```go
func (m *MinimaxEngine) FindBestMove(ctx context.Context, b *board.Board, player moves.Player, config ai.SearchConfig) ai.SearchResult {
    // ... opening book logic unchanged ...
    
    // Automatically choose algorithm based on context
    if _, hasDeadline := ctx.Deadline(); hasDeadline {
        return m.findBestMoveIterative(ctx, b, player, config)
    } else {
        return m.findBestMoveFixedDepth(ctx, b, player, config)
    }
}
```

#### Iterative Deepening Algorithm
```go
func (m *MinimaxEngine) findBestMoveIterative(ctx context.Context, b *board.Board, player moves.Player, config ai.SearchConfig) ai.SearchResult {
    var bestMove board.Move
    var bestScore ai.EvaluationScore = MinEval
    var bestMoveFound bool
    
    // Search progressively deeper: 1, 2, 3, ..., MaxDepth
    for currentDepth := 1; currentDepth <= config.MaxDepth; currentDepth++ {
        // Try ALL moves at current depth
        for each legal move {
            score := evaluate move at currentDepth
            if score > bestScore {
                bestScore = score
                bestMove = move
                bestMoveFound = true
            }
        }
        
        // Check timeout BETWEEN depths, not during move evaluation
        if ctx timeout occurred {
            break // Return best move found so far
        }
    }
    
    return SearchResult{BestMove: bestMove, Score: bestScore, ...}
}
```

#### Fixed-Depth Preservation
```go
func (m *MinimaxEngine) findBestMoveFixedDepth(ctx context.Context, b *board.Board, player moves.Player, config ai.SearchConfig) ai.SearchResult {
    // Current implementation (renamed from FindBestMove)
    // Used for deterministic testing and analysis
    // No timeout handling - searches exactly to MaxDepth
}
```

## Benefits

### 1. Tactical Reliability
- Engine will always evaluate multiple moves before timing out
- No more hanging pieces due to first-move timeout
- Better move selection under time pressure

### 2. Time Management
- Always has a legal move ready (even at depth 1)
- Progressively improves move quality with available time
- Graceful degradation under time pressure

### 3. Backward Compatibility
- Existing tests use `context.Background()` → fixed-depth behavior unchanged
- APIs unchanged
- No breaking changes to existing code

### 4. Performance Characteristics
- **With timeout**: Iterative deepening ensures best move within time limit
- **Without timeout**: Fixed-depth search for deterministic analysis
- Minimal overhead from depth progression

## Expected Outcomes

### Before Fix
```
UCI Debug Log:
Position: rnbqkb1r/pp4pp/2pp4/5pN1/2P1p1n1/PPNP4/4PPPP/R1BQKB1R w KQkq f6 0 8
Move: e2e3 (bestmove e2e3) score=-300 depth=6 nodes=1295041 time=0.997s
Result: Hangs knight on g5 ❌
```

### After Fix
```
UCI Debug Log:
Position: rnbqkb1r/pp4pp/2pp4/5pN1/2P1p1n1/PPNP4/4PPPP/R1BQKB1R w KQkq f6 0 8
Depth 1: best=g5f7 score=+50
Depth 2: best=g5f7 score=+5  
Depth 3: best=g5f7 score=+15
...
Move: g5f7 (bestmove g5f7) score=+15 depth=3 nodes=45000 time=0.997s
Result: Knight takes pawn, good move ✅
```

## Implementation Files

### Modified Files
1. **`game/ai/search/minimax.go`**
   - Refactored `FindBestMove` with context detection
   - Added `findBestMoveIterative` function
   - Added `findBestMoveFixedDepth` function (renamed current logic)

2. **`game/ai/search/debug_test.go`**
   - Added test for hanging knight position with timeout
   - Verified better move selection under time constraints

### Created Files
1. **`ITERATIVE_DEEPENING_FIX.md`** (this file)
   - Documentation of problem, solution, and implementation

## Testing Strategy

### Regression Testing
- All existing tests should pass unchanged
- Tests using `context.Background()` get deterministic fixed-depth behavior
- Tests using timeout contexts get iterative deepening

### Specific Test Cases
1. **Hanging Knight Position**: Verify engine no longer chooses e2e3
2. **Time Pressure**: Test various timeout scenarios
3. **Depth Progression**: Verify moves improve with deeper search
4. **Performance**: Compare node counts and search efficiency

## Future Enhancements

This fix provides the foundation for additional optimizations:

1. **Principal Variation Table**: Store best moves from previous iterations for better move ordering
2. **Transposition Table**: Cache position evaluations across depth iterations  
3. **Aspiration Windows**: Narrow search windows based on previous iteration scores
4. **Time Management**: Allocate search time based on position complexity

## Conclusion

This implementation fixes the critical tactical blindness while maintaining full backward compatibility. The engine will now behave like a modern chess engine, always providing reasonable moves even under severe time pressure, while preserving the existing deterministic behavior for testing and analysis.