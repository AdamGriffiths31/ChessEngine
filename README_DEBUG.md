# Chess Engine Debug Guide Implementation

This implementation provides comprehensive debugging tools for the two critical bugs identified in the chess engine:

1. **Disappearing Pawn Bug**: After playing e2e4, the white pawn disappears from the board
2. **Opening Book Failure**: Despite loading entries, the opening book fails to find standard positions

## Files Created/Modified

### New Debug Files
- `board/debug.go` - Board debugging utilities
- `game/openings/debug.go` - Opening book debugging utilities  
- `debug_main.go` - Main debug runner program

### Modified Files (Debug Output Added)
- `board/moves.go` - Added debug output to `MakeMove` function
- `ui/renderer.go` - Added debug output to `RenderBoard` function for e4 square
- `game/openings/hash.go` - Added debug output to `HashPosition` function
- `game/openings/book.go` - Added debug output to `FindBookMove` function

## Usage

### Run All Debug Tests
```bash
go run debug_main.go
# or
./debug_runner
```

### Run Specific Tests
```bash
# Board/move debugging only (disappearing pawn)
go run debug_main.go board

# Opening book debugging only
go run debug_main.go book  
```

## Debug Functions Available

### Board Debugging (`board/debug.go`)
- `DebugSquareIndices()` - Verifies square mapping (e2=12, e4=28)
- `DebugMoveParser()` - Tests move parsing for common moves
- `DebugBoardPosition()` - Shows detailed board state
- `DebugMoveExecution()` - Tracks a move step by step
- `RunBasicDebugTests()` - Runs all board debug tests

### Opening Book Debugging (`game/openings/debug.go`)
- `TestOpeningBook()` - Tests initial position and e2e4 position lookups
- `DebugOpeningBook()` - Displays first N entries from loaded books
- `DebugHashCalculation()` - Shows step-by-step hash calculation
- `CompareHashes()` - Compares hashes of two positions
- `RunOpeningBookDebugTests()` - Runs all opening book debug tests

## Expected Debug Output

### For Disappearing Pawn Bug
The debug output will show:
```
DEBUG: Moving piece 'P' from e2 to e4
DEBUG: After move - From square e2: '.', To square e4: 'P'
DEBUG: Square e4 (index 28) contains: 'P'
```

If the bug is present, you'll see:
```
DEBUG: ERROR - e4 should contain WhitePawn 'P' but contains '.'
```

### For Opening Book Bug  
The debug output will show:
```
DEBUG: Position hash: B128A9662B3EF1C0
DEBUG: Found 0 moves in books
ERROR: Position after e2e4 not found in opening book
```

## Implementation Details

### Debug Output Locations
- **MakeMove**: `board/moves.go:68` - Tracks piece movement step by step
- **RenderBoard**: `ui/renderer.go:10` - Special checks for e4 square contents
- **HashPosition**: `game/openings/hash.go:34` - Step-by-step hash calculation
- **FindBookMove**: `game/openings/book.go:95` - Book lookup process

### Key Debug Features
- **Square Index Verification**: Confirms e2=rank:1,file:4 and e4=rank:3,file:4
- **Move Execution Tracking**: Before/after piece positions for every move
- **Hash Calculation Details**: Shows contribution of each piece, castling, en passant
- **Book Lookup Process**: Hash generation, book search, move filtering

## Testing Sequence

The debug runner follows this sequence:

1. **Square Index Test**: Verify e2/e4 mapping
2. **Move Parser Test**: Verify "e2e4" parsing  
3. **Initial Position Setup**: Load starting position
4. **Execute e2e4**: Make the move with full debug output
5. **Board Display**: Show final board state with e4 verification
6. **Opening Book Test**: Test initial position and post-e2e4 lookups
7. **Hash Debug**: Show detailed hash calculations

## Removing Debug Output

To remove debug output after bug fixing:
- Remove `fmt.Printf("DEBUG: ...)` lines from modified files
- Keep the debug utility files for future debugging needs

## Notes

- Debug output is extensive - redirect to file if needed: `./debug_runner > debug.log 2>&1`
- Opening book tests require book files (*.bin) in the directory
- All debug functions are safe to run and don't modify game state permanently