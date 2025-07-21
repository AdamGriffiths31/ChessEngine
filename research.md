# Polyglot Opening Book Implementation Research

## Overview
This document outlines the research findings and implementation plan for adding Polyglot opening book support to our chess engine.

## Current Codebase Analysis

### Engine Architecture
- **Main Entry**: `main.go` - Simple CLI interface with manual and computer modes
- **AI Engine**: `game/ai/engine.go` - Interface-based engine with `FindBestMove()` method
- **Board Representation**: `board/` package with bitboard implementation
- **Move Generation**: `game/moves/` package with comprehensive move generation
- **Search**: `game/ai/search/minimax.go` - Basic minimax implementation
- **Evaluation**: `game/ai/evaluation/material.go` - Material-based evaluation

### Integration Points
The AI engine interface in `game/ai/engine.go` provides the perfect integration point:
```go
FindBestMove(ctx context.Context, b *board.Board, player moves.Player, config SearchConfig) SearchResult
```

## Polyglot Opening Book Format

### Binary File Structure
- **File Format**: Binary file with 16-byte entries
- **Entry Structure**:
  - 8 bytes: Position hash (Zobrist hash)
  - 2 bytes: Move encoding
  - 2 bytes: Move weight
  - 4 bytes: Learning data
- **Sorting**: Entries sorted by hash for binary search
- **Endianness**: Big-endian format

### Move Encoding (16-bit)
- **Bits 0-5**: Destination square (0-63)
- **Bits 6-11**: Origin square (0-63)  
- **Bits 12-14**: Promotion piece (1=Knight, 2=Bishop, 3=Rook, 4=Queen)
- **Special Cases**: Castling uses "king captures rook" representation

### Position Hashing
- Uses Zobrist hashing algorithm
- 64-bit hash values
- Compatible with standard Polyglot hash implementation

## Implementation Plan

### Phase 1: Core Components (2-3 days)

#### 1.1 Create `game/openings/` package
```
game/openings/
├── book.go          # Main book interface and manager
├── polyglot.go      # Polyglot binary format reader  
├── hash.go          # Zobrist hash generation
└── types.go         # Data structures
```

#### 1.2 Key Interfaces
```go
type OpeningBook interface {
    LookupMove(hash uint64) ([]BookMove, error)
    LoadFromFile(filename string) error
}

type BookMove struct {
    Move   moves.Move
    Weight uint16
    Learn  uint32
}
```

### Phase 2: Integration (1-2 days)

#### 2.1 Modify AI Engine
- Update `game/ai/engine.go` to consult opening book before search
- Add book configuration to `SearchConfig`
- Implement weighted move selection from book entries

#### 2.2 Engine Integration Flow
1. Generate position hash
2. Query opening book
3. If book moves found, select based on weights
4. If no book moves, fall back to search algorithm

### Phase 3: File Format Implementation (1 day)

#### 3.1 Binary File Reading
```go
type PolyglotEntry struct {
    Hash   uint64
    Move   uint16  
    Weight uint16
    Learn  uint32
}
```

#### 3.2 Binary Search Implementation
- Load entire file into memory (typical books are <100MB)
- Use Go's `sort.Search()` for efficient lookup
- Handle multiple entries for same position

### Phase 4: Testing Strategy (2-3 days)

#### 4.1 Unit Tests
- Hash generation correctness
- Move encoding/decoding
- Binary file parsing
- Search functionality

#### 4.2 Integration Tests  
- End-to-end book consultation
- Performance with different book sizes
- Fallback behavior when book unavailable

#### 4.3 Test Data
- Use standard opening book files (e.g., ProDeo.bin)
- Create minimal test book for unit tests
- Verify against known positions and moves

#### 4.4 Performance Benchmarks
- Book lookup speed
- Memory usage
- Impact on overall engine performance

### Phase 5: Configuration & Documentation (1 day)

#### 5.1 Configuration Options
- Book file paths
- Book selection probability
- Multiple book support
- Book learning updates

#### 5.2 Game Mode Integration
- Add book options to computer vs player mode
- Configuration file support
- Command-line book selection

## Technical Considerations

### Memory Management
- Load book files at startup
- Keep books in memory for fast access
- Support multiple concurrent books

### Error Handling
- Graceful fallback when book files missing
- Invalid book format handling
- Partial book loading on memory constraints

### Performance Optimizations
- Cache frequently accessed positions
- Lazy loading of large books
- Background book loading

### Future Enhancements
- Book learning from played games
- Custom book creation from PGN files
- Book statistics and analysis
- Multiple format support (CTG, Arena, etc.)

## Dependencies

### External Libraries
- Go standard library `binary` package for file I/O
- No additional dependencies required for basic implementation

### Test Books
- Download standard Polyglot books for testing
- Create minimal test books for unit tests
- Verify compatibility with popular chess GUIs

## Timeline

| Phase | Duration | Description |
|-------|----------|-------------|
| 1 | 2-3 days | Core components and interfaces |
| 2 | 1-2 days | AI engine integration |
| 3 | 1 day | Binary format implementation |
| 4 | 2-3 days | Comprehensive testing |
| 5 | 1 day | Documentation and configuration |

**Total Estimated Time: 7-10 days**

## Success Criteria

1. Engine correctly consults opening books before search
2. Supports standard Polyglot .bin format
3. Performance impact < 5ms per move
4. Comprehensive test coverage (>90%)
5. Graceful fallback when books unavailable
6. Integration with existing game modes
7. Documentation and configuration complete

## Risk Mitigation

- **File Format Issues**: Test with multiple standard book files
- **Performance Impact**: Benchmark and optimize hot paths
- **Integration Complexity**: Start with simple integration, expand gradually
- **Hash Compatibility**: Verify against known implementations
- **Memory Usage**: Monitor and optimize for large books

## References

- [Polyglot Opening Book Format](https://www.chessprogramming.org/PolyGlot)
- [Opening Book Implementation](https://www.chessprogramming.org/Opening_Book)
- [Zobrist Hashing](https://www.chessprogramming.org/Zobrist_Hashing)
- [Go Chess Libraries](https://github.com/notnil/chess)
- [Donna Chess Engine](https://github.com/michaeldv/donna) - Go implementation with Polyglot support
