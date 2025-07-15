# Golang CLI Chess Engine

## Phase 1: CLI Chess Board Renderer ✅ COMPLETED

### Overview

This module focuses on rendering the current board state to the command line using ASCII art. The output is clean, consistent, and testable — forming the visual foundation for human interaction with the chess engine.

### Objectives ✅

- Render an 8x8 chess board with:
  - Piece placement
  - File and rank labels (`a`–`h`, `1`–`8`)
- Use standard symbols:
  - Uppercase letters for White pieces (`P`, `N`, `B`, `R`, `Q`, `K`)
  - Lowercase letters for Black pieces (`p`, `n`, `b`, `r`, `q`, `k`)
  - Dots (`.`) for empty squares
- Output must be a multiline string suitable for terminal printing
- Designed for testability — rendering correctness validated via snapshot-style unit tests

---

## Phase 2: Game Mode 1 - Manual Play ✅ COMPLETED

### Overview

Game Mode 1 allows users to manually play both sides of a chess game through the CLI. This mode focuses on move input, board state management, and user interaction without move validation — assuming the user makes legal moves.

### Objectives ✅

- Interactive CLI game loop with move input
- Support coordinate notation moves (e.g., "e2e4", "o-o", "o-o-o")
- Display current board state after each move
- Track whose turn it is (White/Black)
- Handle basic commands:
  - Move input (e.g., "e2e4")
  - "quit" to exit game
  - "reset" to start new game
  - "fen" to display current FEN string
- Clean user interface with prompts and feedback
- No move validation (trust user input)

### File Structure

```
chess-cli/
├── game/
│   ├── modes/
│   │   ├── mode1.go          # Manual play game mode
│   │   └── mode1_test.go
│   ├── engine.go             # Game state management
│   ├── engine_test.go
│   └── moves.go              # Move parsing and application
├── ui/
│   ├── renderer.go
│   ├── renderer_test.go
│   ├── prompts.go            # User interaction
│   └── prompts_test.go
├── board/
│   ├── board.go
│   ├── board_test.go
│   └── moves.go              # Board move operations
```
### Testing Strategy

#### Game Mode 1 Testing Requirements:

- **Move Parsing Tests**: Validate algebraic notation parsing
  - Standard moves: "e2e4", "Nf3", "Qh5"
  - Castling: "O-O", "O-O-O"
  - Captures: "exd5", "Nxf7"
  - Promotions: "e8=Q", "a1=N"
  - Edge cases: Invalid format handling

- **Game State Tests**: Ensure proper state management
  - Turn tracking (White/Black alternation)
  - Board state updates after moves
  - Command handling (quit, reset, fen)
  - Game loop integration

- **UI Interaction Tests**: Validate user experience
  - Prompt display and formatting
  - Input handling and validation
  - Error message clarity
  - Board rendering integration

#### Test Structure:
- Use table-driven tests for move parsing scenarios
- Mock user input for game loop testing
- Golden file approach for complex game state scenarios
- Integration tests for full game flow

### Milestone Tasks

1. **Move System**: Implement algebraic notation parsing and board updates
2. **Game Engine**: Create game state management and turn tracking
3. **User Interface**: Build interactive prompts and command handling
4. **Game Mode 1**: Integrate all components into playable manual mode
5. **Testing Suite**: Comprehensive tests for all components
6. **Main Integration**: Update main.go to launch Game Mode 1

### Example User Experience

```
Chess Engine - Game Mode 1: Manual Play
========================================

Current turn: White

  a b c d e f g h
8 r n b q k b n r 8
7 p p p p p p p p 7
6 . . . . . . . . 6
5 . . . . . . . . 5
4 . . . . . . . . 4
3 . . . . . . . . 3
2 P P P P P P P P 2
1 R N B Q K B N R 1
  a b c d e f g h

Enter move (or 'quit', 'reset', 'fen'): e2e4

Current turn: Black

  a b c d e f g h
8 r n b q k b n r 8
7 p p p p p p p p 7
6 . . . . . . . . 6
5 . . . . . . . . 5
4 . . . . P . . . 4
3 . . . . . . . . 3
2 P P P P . P P P 2
1 R N B Q K B N R 1
  a b c d e f g h

Enter move (or 'quit', 'reset', 'fen'): 
```

---

## Phase 3: Move Generation and Validation

### Overview

Phase 3 introduces legal move generation and validation to the chess engine. Users can now type "moves" to see all valid moves available for the current position. This phase establishes the foundation for proper chess rule enforcement and AI gameplay.

Phase 3 is implemented in multiple sub-phases, each focusing on specific piece types:
- **Phase 3a**: Pawn move generation ✅ COMPLETED
- **Phase 3b**: Rook move generation ✅ COMPLETED
- **Phase 3c**: Bishop move generation ✅ COMPLETED
- **Phase 3d**: Knight move generation ✅ COMPLETED
- **Phase 3e**: Queen move generation ✅ COMPLETED
- **Phase 3f**: King move generation ✅ COMPLETED

### Phase 3a: Pawn Move Generation ✅ COMPLETED

The initial implementation focuses exclusively on pawn moves, covering all pawn-specific rules and edge cases.

### Objectives

- **Move Generation Command**: Type "moves" to display all legal moves for current player
- **Pawn Move Logic**: Complete implementation of pawn movement rules:
  - Forward moves (one square from any rank)
  - Initial two-square moves (from starting rank only)
  - Diagonal captures (when enemy piece present)
  - En passant captures (when conditions met)
  - Promotion moves (when reaching end rank)
- **Move Validation**: Validate user input against generated legal moves
- **Move Display**: Clean, organized presentation of available moves
- **Integration**: Seamless integration with existing Game Mode 1

### File Structure

```
chess-cli/
├── game/
│   ├── moves/
│   │   ├── generator.go        # Move generation engine
│   │   ├── generator_test.go
│   │   ├── pawn.go            # Pawn-specific move logic
│   │   ├── pawn_test.go
│   │   └── validation.go      # Move validation
│   ├── modes/
│   │   ├── mode1.go          # Updated with move generation
│   │   └── mode1_test.go
│   ├── engine.go             # Enhanced game state
│   ├── engine_test.go
│   └── moves.go              # Move parsing
├── ui/
│   ├── renderer.go
│   ├── renderer_test.go
│   ├── prompts.go            # Updated with move display
│   ├── prompts_test.go
│   └── moves_display.go      # Move formatting and display
├── board/
│   ├── board.go
│   ├── board_test.go
│   ├── moves.go              # Board operations
│   └── position.go           # Position analysis utilities
```

### Testing Strategy

#### Pawn Move Generation Tests:

- **Basic Forward Moves**:
  - Single square forward (any rank)
  - Two square initial move (rank 2/7 only)
  - Blocked path detection
  - Edge of board boundaries

- **Capture Moves**:
  - Diagonal captures (left/right)
  - No capture when no enemy piece
  - Cannot capture own pieces
  - Cannot capture empty squares

- **En Passant**:
  - Valid en passant scenarios
  - Previous move was pawn two-square advance
  - Capturing pawn on correct rank (5th/4th)
  - Target square availability

- **Promotion**:
  - Pawn reaching end rank (8th/1st)
  - All promotion piece options (Q, R, B, N)
  - Promotion on captures vs forward moves
  - Multiple promotion moves per position

- **Edge Cases**:
  - Pawns on starting rank but obstructed
  - Pawns near board edges
  - Multiple en passant possibilities
  - Promotion with captures

#### Move Generation Integration Tests:

- **"moves" Command**: Verify command parsing and execution
- **Move Display**: Validate formatting and completeness
- **Move Validation**: Ensure only legal moves accepted
- **Game Flow**: Integration with existing Game Mode 1

#### Test Structure:
- **Golden File Tests**: Board positions with expected move lists
- **Table-Driven Tests**: Comprehensive pawn scenarios
- **Integration Tests**: Full command and validation flow
- **Performance Tests**: Move generation efficiency

### Milestone Tasks

1. **Move Generation Framework**: Create base generator infrastructure
2. **Pawn Move Logic**: Implement all pawn movement rules
3. **Move Validation System**: Validate moves against generated legal moves
4. **UI Integration**: Add "moves" command and move display
5. **Comprehensive Testing**: Full test coverage for all pawn scenarios
6. **Game Mode Enhancement**: Integrate with existing manual play mode

### Example User Experience

```
Chess Engine - Game Mode 1: Manual Play
========================================

Current turn: White
Move: 1

  a b c d e f g h
8 r n b q k b n r 8
7 p p p p p p p p 7
6 . . . . . . . . 6
5 . . . . . . . . 5
4 . . . . . . . . 4
3 . . . . . . . . 3
2 P P P P P P P P 2
1 R N B Q K B N R 1
  a b c d e f g h

Enter move (or 'quit', 'reset', 'fen', 'moves'): moves

Available moves for White:
Pawn moves:
  a2a3, a2a4, b2b3, b2b4, c2c3, c2c4, d2d3, d2d4
  e2e3, e2e4, f2f3, f2f4, g2g3, g2g4, h2h3, h2h4

Enter move (or 'quit', 'reset', 'fen', 'moves'): e2e4

Move validated ✓

Current turn: Black
Move: 1

  a b c d e f g h
8 r n b q k b n r 8
7 p p p p p p p p 7
6 . . . . . . . . 6
5 . . . . . . . . 5
4 . . . . P . . . 4
3 . . . . . . . . 3
2 P P P P . P P P 2
1 R N B Q K B N R 1
  a b c d e f g h

Enter move (or 'quit', 'reset', 'fen', 'moves'): moves

Available moves for Black:
Pawn moves:
  a7a6, a7a5, b7b6, b7b5, c7c6, c7c5, d7d6, d7d5
  e7e6, e7e5, f7f6, f7f5, g7g6, g7g5, h7h6, h7h5

Enter move (or 'quit', 'reset', 'fen', 'moves'): d7d5

Move validated ✓
```

### Phase 3b-3f: Additional Piece Move Generation

The remaining phases implement move generation for all other piece types, following the same comprehensive approach as Phase 3a:

#### Phase 3b: Rook Move Generation
- **Straight Line Movement**: Horizontal and vertical moves until blocked
- **Path Validation**: Cannot jump over pieces
- **Capture Logic**: Enemy piece captures
- **Board Boundary Respect**: Proper edge handling

#### Phase 3c: Bishop Move Generation  
- **Diagonal Movement**: Four diagonal directions until blocked
- **Path Validation**: Cannot jump over pieces
- **Capture Logic**: Enemy piece captures
- **Board Boundary Respect**: Proper edge handling

#### Phase 3d: Knight Move Generation
- **L-Shaped Moves**: Eight possible L-shaped moves
- **Jump Ability**: Can jump over other pieces
- **Capture Logic**: Enemy piece captures
- **Board Boundary Respect**: Proper edge handling

#### Phase 3e: Queen Move Generation
- **Combined Movement**: Rook and bishop patterns combined
- **Path Validation**: Cannot jump over pieces (straight/diagonal)
- **Capture Logic**: Enemy piece captures
- **Board Boundary Respect**: Proper edge handling

#### Phase 3f: King Move Generation
- **Single Square Moves**: One square in any direction
- **Castling Logic**: Kingside and queenside castling
- **Castling Conditions**: King/rook unmoved, clear path, not in check
- **Capture Logic**: Enemy piece captures

### Testing Strategy for All Piece Types

Each piece type follows the same rigorous testing approach:

1. **Basic Movement Tests**: Normal moves from various board positions
2. **Blocking and Path Tests**: Piece interference and path validation
3. **Capture Tests**: Valid/invalid capture scenarios
4. **Edge Case Tests**: Board boundaries and complex positions
5. **Integration Tests**: Multi-piece interactions and performance

### Implementation Timeline

- **Phase 3a (Pawn)**: ✅ COMPLETED
- **Phase 3b (Rook)**: ✅ COMPLETED
- **Phase 3c (Bishop)**: ✅ COMPLETED
- **Phase 3d (Knight)**: ✅ COMPLETED
- **Phase 3e (Queen)**: ✅ COMPLETED
- **Phase 3f (King)**: ✅ COMPLETED (includes castling)
- **Phase 3 Complete**: ✅ ALL CHESS PIECES IMPLEMENTED

---

## Phase 1 Testing Strategy (Completed)

The renderer will be validated using Go's built-in testing framework with a thorough suite of unit tests.

### Testing Requirements:


### FEN-based Testing Requirements:

All FEN-based renderer tests will use a golden file approach. Each test case will be defined in a JSON file containing:

- The FEN string representing the board state
- The expected multiline string output from the renderer


Example JSON structure (showing one test case):
```json
[
  {
    "fen": "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
    "expected": "<expected board rendering as multiline string>"
  }
]
```

Test scenarios to be included (all will be present in the golden file):
- Initial board position (standard setup):  
    - `rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1`
- Completely empty board:  
    - `8/8/8/8/8/8/8/8 w - - 0 1`
- Custom midgame positions (e.g., pinned piece, checks, multiple captures):  
    - `r1bqkbnr/pppp1ppp/2n5/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 2 4`
- Edge rank/file alignment:  
    - `8/8/8/8/8/8/8/R3K2R w KQ - 0 1`
- Promotion scenarios:  
    - `8/P7/8/8/8/8/7p/8 w - - 0 1`
- Minimal piece endgame (e.g., King vs King):  
    - `8/8/8/8/8/8/8/4K3 w - - 0 1`
- Inconsistent/malformed FEN (robustness):  
    - `8/8/8/8/8/8/8/8/8 w - - 0 1`  (extra rank, invalid FEN)
- Single piece boards:
    - `8/8/8/8/8/8/8/4Q3 w - - 0 1`  (only a white queen)
    - `8/8/8/8/8/8/8/4p3 w - - 0 1`  (only a black pawn)
- Non-standard piece placement:
    - `8/8/8/8/8/8/P7/8 w - - 0 1`  (white pawn on 2nd rank)
    - `8/8/8/8/8/8/8/p7 w - - 0 1`  (black pawn on 1st rank)

Each test will:
- Use a FEN string from the JSON file to define the board state
- Compare the renderer's output to the expected multiline string from the JSON

This golden file strategy ensures rendering correctness and guards against regressions.

## Milestone Tasks

1. Define minimal `Board` and `Piece` types (mock or real)
2. Implement `RenderBoard(*Board) string`
3. Write exhaustive test suite
4. Refactor rendering for readability and single-responsibility
5. Approve rendering outputs as snapshots

---

## Best Practices

- Keep rendering logic pure and deterministic
- Decouple board formatting from game or engine logic
- Run renderer tests on each CI build to avoid visual regressions