# ChessEngine

Go implementation of a UCI compatible chess engine.

## Rating

| Version | File          | Time | Score      |
| ------- | ------------- | ---- | ---------- |
| 0.7     | benchmark.epd | 10s  | 439 of 948 |
| 0.6     | benchmark.epd | 10s  | 452 of 948 |

## Perft

### Perft testing depth 1-6

| Version | Count | Time     |
| ------- | ----- | -------- |
| 0.7     | 5     | 211.615s |
| 0.4     | 5     | 262.05s  |
| 0.3     | 5     | 275.03s  |

## Versions

### v0.8

- Delta Pruning
- Aspiration windows
- Reverse Futility Pruning
- Mate Distance Pruning
- Razoring

### v0.7 (done)

- Code refactor
- LazySMP

### v0.6 (done)

- Age Hashing
- Search optimisation

### v0.5 (done)

- PolyGlot opening book

### v0.4 (done)

- Moved MoveGen to a bitboard system
- Transposition table
- Null move pruning

### v0.3 (done)

- Xboard integration
- Console mode

### v0.2 (done)

- AB Logic
- Quiescence
- UCI integration

### v0.1 (done)

- Perft works
- MoveGenerator
- Basic user input
