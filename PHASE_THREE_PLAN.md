# Phase 3 Implementation Plan: Move Generation and Validation

## Overview
Phase 3 introduces legal move generation and validation, starting with pawn moves only (Phase 3a). Users can type "moves" to see all valid moves and the system validates input moves against legal possibilities.

## Implementation Steps

### Step 1: Create Move Generation Framework
**Objective**: Build the foundation for move generation system

**Tasks**:
1. Create `game/moves/` directory structure
2. Implement `generator.go` with base `MoveGenerator` interface
3. Create move list data structures
4. Implement position analysis utilities in `board/position.go`

**Files to Create**:
- `game/moves/generator.go`
- `game/moves/generator_test.go`
- `board/position.go`

**Key Components**:
```go
type MoveGenerator interface {
    GenerateAllMoves(board *board.Board, player game.Player) []board.Move
    GeneratePawnMoves(board *board.Board, player game.Player) []board.Move
}

type MoveList struct {
    Moves []board.Move
    Count int
}
```

### Step 2: Implement Pawn Move Logic
**Objective**: Complete pawn movement rule implementation

**Tasks**:
1. Create `pawn.go` with all pawn-specific move generation
2. Implement forward moves (1 and 2 squares)
3. Implement diagonal captures
4. Implement en passant detection and moves
5. Implement promotion move generation

**Files to Create**:
- `game/moves/pawn.go`
- `game/moves/pawn_test.go`

**Pawn Rules to Implement**:
- **Forward Moves**: 1 square from any rank, 2 squares from starting rank only
- **Captures**: Diagonal attacks when enemy piece present
- **En Passant**: Previous move was enemy pawn 2-square advance, capturing pawn on 5th/4th rank
- **Promotion**: Reaching end rank generates Q/R/B/N promotion options

### Step 3: Create Move Validation System
**Objective**: Validate user input against generated legal moves

**Tasks**:
1. Create `validation.go` with move validation logic
2. Implement move comparison and matching
3. Add validation to move parsing pipeline
4. Handle promotion move validation

**Files to Create**:
- `game/moves/validation.go`

**Key Functions**:
```go
func ValidateMove(move board.Move, legalMoves []board.Move) bool
func IsMoveLegal(board *board.Board, move board.Move, player game.Player) bool
```

### Step 4: Add UI Integration
**Objective**: Integrate move generation with user interface

**Tasks**:
1. Create `ui/moves_display.go` for move formatting
2. Add "moves" command to move parser
3. Update prompts to show "moves" option
4. Implement clean move list display

**Files to Create**:
- `ui/moves_display.go`
- `ui/moves_display_test.go`

**Updated Files**:
- `game/moves.go` (add "moves" command)
- `ui/prompts.go` (add move display methods)

### Step 5: Update Game Engine
**Objective**: Integrate move generation with game state

**Tasks**:
1. Add move generator to game engine
2. Track en passant state for move generation
3. Update move application to track en passant
4. Add move validation to engine

**Updated Files**:
- `game/engine.go`
- `board/board.go` (en passant tracking)

### Step 6: Enhance Game Mode 1
**Objective**: Integrate move generation with manual play

**Tasks**:
1. Add "moves" command handling to Mode 1
2. Implement move validation in game loop
3. Add "Move validated ✓" feedback
4. Update error messages for invalid moves

**Updated Files**:
- `game/modes/mode1.go`

### Step 7: Comprehensive Testing
**Objective**: Ensure robust testing coverage

**Tasks**:
1. Create golden file tests for move generation scenarios
2. Implement table-driven tests for all pawn cases
3. Add integration tests for "moves" command
4. Create performance benchmarks

**Test Categories**:
- **Unit Tests**: Individual pawn scenarios
- **Integration Tests**: Full command flow
- **Golden File Tests**: Complex board positions
- **Performance Tests**: Move generation speed

## Testing Strategy

### Golden File Structure
```json
[
  {
    "name": "initial_position_white_pawns",
    "fen": "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
    "player": "White",
    "expected_moves": [
      "a2a3", "a2a4", "b2b3", "b2b4", "c2c3", "c2c4",
      "d2d3", "d2d4", "e2e3", "e2e4", "f2f3", "f2f4",
      "g2g3", "g2g4", "h2h3", "h2h4"
    ]
  },
  {
    "name": "en_passant_scenario",
    "fen": "rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPP1PPP/RNBQKBNR w KQkq f6 0 3",
    "player": "White",
    "expected_moves": [
      "exf6", "e6", "a2a3", "a2a4", "b2b3", "b2b4",
      "c2c3", "c2c4", "d2d3", "d2d4", "f2f3", "f2f4",
      "g2g3", "g2g4", "h2h3", "h2h4"
    ]
  }
]
```

### Test Coverage Requirements
- **Pawn Forward Moves**: 95%+ coverage
- **Pawn Captures**: 100% coverage
- **En Passant**: 100% coverage
- **Promotion**: 100% coverage
- **Edge Cases**: 90%+ coverage

## Success Criteria

### Functional Requirements
✅ **Move Generation**: Generate all legal pawn moves for any position
✅ **Move Display**: Clean, organized presentation via "moves" command
✅ **Move Validation**: Accept only legal moves, reject invalid ones
✅ **En Passant**: Correct detection and generation of en passant moves
✅ **Promotion**: Generate all promotion options (Q/R/B/N)
✅ **Integration**: Seamless integration with existing Game Mode 1

### Quality Requirements
✅ **Test Coverage**: >90% code coverage for move generation
✅ **Performance**: <1ms move generation for typical positions
✅ **Robustness**: Handle all edge cases and malformed input
✅ **Maintainability**: Clean, modular code structure

### User Experience Requirements
✅ **Intuitive Commands**: "moves" command easy to discover and use
✅ **Clear Display**: Move lists well-formatted and easy to read
✅ **Fast Feedback**: Immediate validation feedback for moves
✅ **Error Handling**: Helpful error messages for invalid moves

## Estimated Timeline
- **Step 1-3**: 2-3 days (Core move generation)
- **Step 4-6**: 1-2 days (UI and game integration) 
- **Step 7**: 1-2 days (Comprehensive testing)
- **Total**: 4-7 days for complete Phase 3a implementation

## Phase 3b: Rook Move Generation

### Objective
Implement complete rook movement rules including straight-line moves and castling preparation.

### Tasks
1. Create `rook.go` with rook-specific move generation
2. Implement horizontal and vertical moves
3. Handle blocking pieces and path validation
4. Add rook moves to main generator
5. Create comprehensive test suite

### Rook Rules to Implement
- **Straight Line Moves**: Horizontal and vertical movement until blocked
- **Path Checking**: Cannot jump over pieces
- **Capture Logic**: Can capture enemy pieces but not own pieces
- **Board Boundaries**: Respect board edges

### Files to Create/Update
- `game/moves/rook.go`
- `game/moves/rook_test.go`
- Update `game/moves/generator.go` to include rook moves

## Phase 3c: Bishop Move Generation

### Objective
Implement complete bishop movement rules including diagonal moves.

### Tasks
1. Create `bishop.go` with bishop-specific move generation
2. Implement diagonal moves in all four directions
3. Handle blocking pieces and path validation
4. Add bishop moves to main generator
5. Create comprehensive test suite

### Bishop Rules to Implement
- **Diagonal Moves**: Four diagonal directions until blocked
- **Path Checking**: Cannot jump over pieces
- **Capture Logic**: Can capture enemy pieces but not own pieces
- **Board Boundaries**: Respect board edges

### Files to Create/Update
- `game/moves/bishop.go`
- `game/moves/bishop_test.go`
- Update `game/moves/generator.go` to include bishop moves

## Phase 3d: Knight Move Generation

### Objective
Implement complete knight movement rules including L-shaped moves.

### Tasks
1. Create `knight.go` with knight-specific move generation
2. Implement all eight L-shaped moves
3. Handle board boundaries and piece capture
4. Add knight moves to main generator
5. Create comprehensive test suite

### Knight Rules to Implement
- **L-Shaped Moves**: Eight possible L-shaped moves (2+1 in all directions)
- **Jump Ability**: Can jump over other pieces
- **Capture Logic**: Can capture enemy pieces but not own pieces
- **Board Boundaries**: Respect board edges

### Files to Create/Update
- `game/moves/knight.go`
- `game/moves/knight_test.go`
- Update `game/moves/generator.go` to include knight moves

## Phase 3e: Queen Move Generation

### Objective
Implement complete queen movement rules combining rook and bishop moves.

### Tasks
1. Create `queen.go` with queen-specific move generation
2. Combine rook and bishop move logic
3. Handle blocking pieces and path validation
4. Add queen moves to main generator
5. Create comprehensive test suite

### Queen Rules to Implement
- **Combined Movement**: Both rook and bishop movement patterns
- **Straight Lines**: Horizontal, vertical, and diagonal moves
- **Path Checking**: Cannot jump over pieces
- **Capture Logic**: Can capture enemy pieces but not own pieces
- **Board Boundaries**: Respect board edges

### Files to Create/Update
- `game/moves/queen.go`
- `game/moves/queen_test.go`
- Update `game/moves/generator.go` to include queen moves

## Phase 3f: King Move Generation

### Objective
Implement complete king movement rules including single-square moves and castling.

### Tasks
1. Create `king.go` with king-specific move generation
2. Implement single-square moves in all directions
3. Implement castling logic (kingside and queenside)
4. Add king moves to main generator
5. Create comprehensive test suite

### King Rules to Implement
- **Single Square Moves**: One square in any direction
- **Castling**: Kingside and queenside castling when conditions met
- **Castling Conditions**: King and rook not moved, no pieces between, not in check
- **Capture Logic**: Can capture enemy pieces but not own pieces
- **Board Boundaries**: Respect board edges

### Files to Create/Update
- `game/moves/king.go`
- `game/moves/king_test.go`
- Update `game/moves/generator.go` to include king moves

## Testing Strategy for Each Piece Type

### Common Test Categories
For each piece type (3b-3f), implement the following test categories:

1. **Basic Movement Tests**
   - Normal moves from center of board
   - Moves from corners and edges
   - Multiple move options validation

2. **Blocking and Path Tests**
   - Own pieces blocking movement
   - Enemy pieces blocking movement
   - Path validation for sliding pieces (rook, bishop, queen)

3. **Capture Tests**
   - Valid captures of enemy pieces
   - Cannot capture own pieces
   - Capture vs non-capture move differentiation

4. **Edge Case Tests**
   - Board boundary respect
   - No valid moves scenarios
   - Complex board positions

5. **Integration Tests**
   - Pieces working together
   - Combined move generation
   - Performance with full piece set

### Estimated Timeline
- **Phase 3b (Rook)**: 1-2 days
- **Phase 3c (Bishop)**: 1-2 days  
- **Phase 3d (Knight)**: 1-2 days
- **Phase 3e (Queen)**: 1 day (reuses rook/bishop logic)
- **Phase 3f (King)**: 2-3 days (includes castling complexity)
- **Total**: 6-10 days for complete Phase 3 implementation

## Next Phases
- **Phase 4**: Add check detection and legal move filtering
- **Phase 5**: Add checkmate and stalemate detection
- **Phase 6**: Add special moves (en passant integration, castling validation)