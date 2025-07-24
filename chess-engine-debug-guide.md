# Chess Engine Debug Guide

## Bug Report Summary

Two critical bugs identified during gameplay:

1. **Disappearing Pawn Bug**: After playing e2e4, the white pawn disappears from the board
2. **Opening Book Failure**: Despite loading 92,954 entries, the opening book fails to find standard positions

## Bug 1: Disappearing Pawn (Critical)

### Symptoms
- White plays e2e4
- Board shows empty square at e4 (should show white pawn 'P')
- The pawn seems to vanish completely

### Debug Steps

#### 1. Check Move Execution (`MakeMove` function)
```go
// In your MakeMove function, add debug prints:
func (b *Board) MakeMove(move Move) {
    fmt.Printf("DEBUG: Moving piece %v from %s to %s\n", 
        b.squares[move.From], SquareToString(move.From), SquareToString(move.To))
    
    // Your move logic here
    b.squares[move.To] = b.squares[move.From]
    b.squares[move.From] = Empty
    
    fmt.Printf("DEBUG: After move - From square: %v, To square: %v\n", 
        b.squares[move.From], b.squares[move.To])
}
```

#### 2. Verify Square Indexing
```go
// Add this debug function to verify square indices:
func DebugSquareIndices() {
    fmt.Println("Square Index Debug:")
    fmt.Printf("e2 = %d (should be 12)\n", StringToSquare("e2"))
    fmt.Printf("e4 = %d (should be 28)\n", StringToSquare("e4"))
    
    // Test reverse conversion
    fmt.Printf("Square 12 = %s (should be e2)\n", SquareToString(12))
    fmt.Printf("Square 28 = %s (should be e4)\n", SquareToString(28))
}
```

#### 3. Check Board Display Function
```go
// In your board display/print function:
func (b *Board) Display() {
    fmt.Println("  a b c d e f g h")
    for rank := 7; rank >= 0; rank-- {
        fmt.Printf("%d ", rank+1)
        for file := 0; file < 8; file++ {
            square := rank*8 + file
            piece := b.squares[square]
            
            // Debug: print square index for e4
            if square == 28 {
                fmt.Printf("DEBUG: Square e4 (index %d) contains: %v\n", square, piece)
            }
            
            // Your display logic here
            fmt.Printf("%c ", PieceToChar(piece))
        }
        fmt.Printf(" %d\n", rank+1)
    }
    fmt.Println("  a b c d e f g h")
}
```

#### 4. Add Move Validation Debug
```go
// Before and after move validation:
func ValidateMove(board *Board, move Move) bool {
    fmt.Printf("DEBUG: Validating move from %s to %s\n", 
        SquareToString(move.From), SquareToString(move.To))
    fmt.Printf("DEBUG: Piece at from square: %v\n", board.squares[move.From])
    
    // Your validation logic
    
    return true
}
```

## Bug 2: Opening Book Failure

### Symptoms
- Opening book loads successfully (92,954 entries)
- Position hash: B128A9662B3EF1C0
- Lookup fails for standard position after e2e4
- Computer plays unusual move a7a6

### Debug Steps

#### 1. Verify Hash Calculation
```go
// Add debug output to your hash calculation:
func (b *Board) Hash() uint64 {
    var hash uint64
    
    // Debug: print initial position hash
    if b.IsInitialPosition() {
        fmt.Println("DEBUG: Calculating hash for initial position")
    }
    
    for square := 0; square < 64; square++ {
        piece := b.squares[square]
        if piece != Empty {
            // Your hash logic
            hash ^= zobristTable[piece][square]
            
            // Debug first few pieces
            if square < 16 || square > 47 {
                fmt.Printf("DEBUG: Square %s, Piece %v, Hash contribution: %X\n", 
                    SquareToString(square), piece, zobristTable[piece][square])
            }
        }
    }
    
    fmt.Printf("DEBUG: Final position hash: %X\n", hash)
    return hash
}
```

#### 2. Test Opening Book Lookup
```go
// Add test function for opening book:
func TestOpeningBook(book *OpeningBook) {
    // Test 1: Initial position
    initialBoard := NewBoard()
    initialHash := initialBoard.Hash()
    fmt.Printf("DEBUG: Initial position hash: %X\n", initialHash)
    
    if move, found := book.Lookup(initialHash); found {
        fmt.Printf("Found opening move for initial position: %s\n", move.String())
    } else {
        fmt.Println("ERROR: Initial position not found in opening book!")
    }
    
    // Test 2: After e2e4
    initialBoard.MakeMove(ParseMove("e2e4"))
    afterE4Hash := initialBoard.Hash()
    fmt.Printf("DEBUG: Position after e2e4 hash: %X\n", afterE4Hash)
    
    if move, found := book.Lookup(afterE4Hash); found {
        fmt.Printf("Found response to e2e4: %s\n", move.String())
    } else {
        fmt.Println("ERROR: Position after e2e4 not found in opening book!")
    }
}
```

#### 3. Verify Opening Book Format
```go
// Check what's actually in the opening book:
func DebugOpeningBook(book *OpeningBook, limit int) {
    fmt.Printf("DEBUG: First %d entries in opening book:\n", limit)
    count := 0
    
    for hash, moves := range book.positions {
        if count >= limit {
            break
        }
        
        fmt.Printf("Hash: %X, Moves: ", hash)
        for _, move := range moves {
            fmt.Printf("%s ", move.String())
        }
        fmt.Println()
        count++
    }
}
```

#### 4. Check Move Parsing
```go
// Ensure moves are parsed correctly:
func DebugMoveParser() {
    testMoves := []string{"e2e4", "e7e5", "g1f3", "b8c6"}
    
    for _, moveStr := range testMoves {
        move := ParseMove(moveStr)
        fmt.Printf("Parsed '%s': From=%d (%s), To=%d (%s)\n", 
            moveStr, move.From, SquareToString(move.From), 
            move.To, SquareToString(move.To))
    }
}
```

## Quick Fix Attempts

### For Disappearing Pawn:
1. Check if the board array is 0-indexed vs 1-indexed
2. Verify the piece constants (Empty should be 0)
3. Check if there's a board flip happening between moves

### For Opening Book:
1. Try using a different opening book format (PGN-based)
2. Verify Zobrist hash initialization is consistent
3. Check if the book uses different move notation

## Testing Sequence

Run these tests in order:
1. `DebugSquareIndices()` - Verify square mapping
2. `DebugMoveParser()` - Verify move parsing
3. Play e2e4 with debug output enabled
4. `TestOpeningBook(book)` - Check if standard positions exist
5. `DebugOpeningBook(book, 10)` - See what's actually in the book

## Expected Output

After implementing debug code, you should see:
```
DEBUG: Moving piece WhitePawn from e2 to e4
DEBUG: After move - From square: Empty, To square: WhitePawn
DEBUG: Square e4 (index 28) contains: WhitePawn
```

If you see different output, that will pinpoint the exact location of the bug.