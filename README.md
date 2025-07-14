# Phase 1: CLI Chess Board Renderer

## Overview

This module is the first step in building the Golang CLI Chess Engine. It focuses on rendering the current board state to the command line using ASCII art. The output must be clean, consistent, and testable — forming the visual foundation for human interaction with the chess engine.

---

## Objectives

- Render an 8x8 chess board with:
  - Piece placement
  - File and rank labels (`a`–`h`, `1`–`8`)
- Use standard symbols:
  - Uppercase letters for White pieces (`P`, `N`, `B`, `R`, `Q`, `K`)
  - Lowercase letters for Black pieces (`p`, `n`, `b`, `r`, `q`, `k`)
  - Dots (`.`) for empty squares
- Output must be a multiline string suitable for terminal printing
- Designed for testability — rendering correctness must be validated via snapshot-style unit tests

---

## File Structure

```
chess-cli/
├── ui/
│   ├── renderer.go
│   └── renderer_test.go
├── board/
│   └── board.go
```
---

## Testing Strategy

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