# ChessEngine

Go implementation of a UCI compatible chess engine.

## Rating

| Version | Time Control | Est Rating |
| ------- | ------------ | ---------- |
| 0.4     | 5 +6         | 2100       |
| 0.2     | 5 +6         | 1900       |

## Perft

### Perft testing depth 1-6

| Version | Count | Time    |
| ------- | ----- | ------- |
| 0.4     | 5     | 262.05s |
| 0.3     | 5     | 275.03s |

## Versions

### v0.5

- PolyGlot openning book

### v0.4 (done)

- Moved MoveGen to a bitboard system
- Transposition table
- Null move pruning

### v0.3 (done)

- Xboard intergration
- Console mode

### v0.2 (done)

- AB Logic
- Quiescence
- UCI intergration

### v0.1 (done)

- Perft works
- MoveGenerator
- Basic user input
