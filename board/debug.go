package board

import "fmt"

// DebugSquareIndices verifies square index mapping conversions
func DebugSquareIndices() {
	fmt.Println("Square Index Debug:")
	
	// Test key squares mentioned in debug guide
	testSquares := []string{"e2", "e4", "a1", "h8", "d5", "f7"}
	
	for _, squareStr := range testSquares {
		square, err := ParseSquare(squareStr)
		if err != nil {
			fmt.Printf("ERROR: Failed to parse square %s: %v\n", squareStr, err)
			continue
		}
		
		// Convert to array index (rank*8 + file for bitboard index)
		bitboardIndex := square.Rank*8 + square.File
		
		fmt.Printf("%s = rank:%d file:%d (Square struct) -> bitboard index:%d\n", 
			squareStr, square.Rank, square.File, bitboardIndex)
		
		// Test reverse conversion
		reverseStr := square.String()
		fmt.Printf("  Reverse: rank:%d file:%d -> %s (should be %s)\n", 
			square.Rank, square.File, reverseStr, squareStr)
		
		if reverseStr != squareStr {
			fmt.Printf("  ERROR: Reverse conversion failed!\n")
		}
	}
	
	// Specifically test e2 and e4 as mentioned in debug guide
	fmt.Println("\nSpecific e2/e4 test:")
	e2, _ := ParseSquare("e2")
	e4, _ := ParseSquare("e4")
	fmt.Printf("e2 = %d,%d (rank,file) -> bitboard index %d\n", e2.Rank, e2.File, e2.Rank*8+e2.File)
	fmt.Printf("e4 = %d,%d (rank,file) -> bitboard index %d\n", e4.Rank, e4.File, e4.Rank*8+e4.File)
}

// DebugMoveParser tests move parsing for common moves
func DebugMoveParser() {
	fmt.Println("Move Parser Debug:")
	testMoves := []string{"e2e4", "e7e5", "g1f3", "b8c6", "d7d8Q", "O-O", "O-O-O"}
	
	for _, moveStr := range testMoves {
		move, err := ParseSimpleMove(moveStr)
		if err != nil {
			fmt.Printf("ERROR: Failed to parse move '%s': %v\n", moveStr, err)
			continue
		}
		
		fmt.Printf("Parsed '%s': From=%s (%d,%d), To=%s (%d,%d)", 
			moveStr, 
			move.From.String(), move.From.Rank, move.From.File,
			move.To.String(), move.To.Rank, move.To.File)
		
		if move.Promotion != Empty {
			fmt.Printf(", Promotion=%c", move.Promotion)
		}
		if move.IsCastling {
			fmt.Printf(", Castling=true")
		}
		fmt.Println()
	}
}

// DebugBoardPosition prints detailed board state for debugging
func DebugBoardPosition(b *Board, description string) {
	fmt.Printf("\n=== Board Debug: %s ===\n", description)
	
	// Check specific squares mentioned in the bug reports
	checkSquares := []string{"e2", "e4", "a7", "d1", "e1", "e8"}
	
	for _, squareStr := range checkSquares {
		square, err := ParseSquare(squareStr)
		if err != nil {
			continue
		}
		
		piece := b.GetPiece(square.Rank, square.File)
		fmt.Printf("Square %s (rank:%d file:%d): '%c' (%d)\n", 
			squareStr, square.Rank, square.File, piece, int(piece))
	}
	
	// Print game state
	fmt.Printf("Side to move: %s\n", b.GetSideToMove())
	fmt.Printf("Castling rights: %s\n", b.GetCastlingRights())
	if b.GetEnPassantTarget() != nil {
		fmt.Printf("En passant target: %s\n", b.GetEnPassantTarget().String())
	} else {
		fmt.Printf("En passant target: none\n")
	}
	fmt.Printf("Half move clock: %d\n", b.GetHalfMoveClock())
	fmt.Printf("Full move number: %d\n", b.GetFullMoveNumber())
}

// DebugMoveExecution tracks a move step by step
func DebugMoveExecution(b *Board, moveStr string) error {
	fmt.Printf("\n=== Move Execution Debug: %s ===\n", moveStr)
	
	// Parse move
	move, err := ParseSimpleMove(moveStr)
	if err != nil {
		fmt.Printf("ERROR: Failed to parse move: %v\n", err)
		return err
	}
	
	// Print move details
	fmt.Printf("Move parsed: From=%s (%d,%d), To=%s (%d,%d)\n",
		move.From.String(), move.From.Rank, move.From.File,
		move.To.String(), move.To.Rank, move.To.File)
	
	// Check pieces before move
	fromPiece := b.GetPiece(move.From.Rank, move.From.File)
	toPiece := b.GetPiece(move.To.Rank, move.To.File)
	fmt.Printf("Before move: From square (%s) = '%c', To square (%s) = '%c'\n", 
		move.From.String(), fromPiece, move.To.String(), toPiece)
	
	if fromPiece == Empty {
		fmt.Printf("ERROR: No piece at from square %s!\n", move.From.String())
		return fmt.Errorf("no piece at from square")
	}
	
	// Execute move
	fmt.Printf("Executing move...\n")
	err = b.MakeMove(move)
	if err != nil {
		fmt.Printf("ERROR: Move execution failed: %v\n", err)
		return err
	}
	
	// Check pieces after move
	fromPieceAfter := b.GetPiece(move.From.Rank, move.From.File)
	toPieceAfter := b.GetPiece(move.To.Rank, move.To.File)
	fmt.Printf("After move: From square (%s) = '%c', To square (%s) = '%c'\n", 
		move.From.String(), fromPieceAfter, move.To.String(), toPieceAfter)
	
	// Validation
	if fromPieceAfter != Empty {
		fmt.Printf("WARNING: From square should be empty but contains '%c'\n", fromPieceAfter)
	}
	if toPieceAfter != fromPiece {
		fmt.Printf("WARNING: To square should contain '%c' but contains '%c'\n", fromPiece, toPieceAfter)
	}
	
	fmt.Printf("Move execution completed\n")
	return nil
}

// RunBasicDebugTests runs the fundamental debug tests mentioned in the guide
func RunBasicDebugTests() {
	fmt.Println("=== Chess Engine Debug Tests ===")
	
	// Test 1: Square indices
	DebugSquareIndices()
	fmt.Println()
	
	// Test 2: Move parser
	DebugMoveParser()
	fmt.Println()
	
	// Test 3: Initial board state
	board, err := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		fmt.Printf("ERROR: Failed to set up initial position: %v\n", err)
		return
	}
	
	DebugBoardPosition(board, "Initial Position")
	
	// Test 4: Execute e2e4 move with debugging
	fmt.Printf("\n=== Testing e2e4 Move (The Disappearing Pawn Bug) ===\n")
	err = DebugMoveExecution(board, "e2e4")
	if err != nil {
		fmt.Printf("Move execution failed, stopping debug tests\n")
		return
	}
	
	DebugBoardPosition(board, "After e2e4")
	
	fmt.Println("\n=== Debug Tests Complete ===")
}


// DebugMove provides detailed make/unmake operation tracing
func (b *Board) DebugMove(move Move, operation string) {
	fmt.Printf("\n=== %s Move Debug ===\n", operation)
	fmt.Printf("Move: %s -> %s\n", move.From.String(), move.To.String())
	fmt.Printf("Piece in move struct: %c\n", move.Piece)
	fmt.Printf("Captured: %c\n", move.Captured)
	fmt.Printf("Promotion: %c\n", move.Promotion)
	fmt.Printf("Flags: IsCapture=%t, IsCastling=%t, IsEnPassant=%t\n", 
		move.IsCapture, move.IsCastling, move.IsEnPassant)
	
	// Show affected squares
	fromPiece := b.GetPiece(move.From.Rank, move.From.File)
	toPiece := b.GetPiece(move.To.Rank, move.To.File)
	fmt.Printf("From square %s: %c\n", move.From.String(), fromPiece)
	fmt.Printf("To square %s: %c\n", move.To.String(), toPiece)
}

// DebugMakeUnmake performs a complete make/unmake cycle with debugging
func (b *Board) DebugMakeUnmake(move Move, label string) error {
	fmt.Printf("\n=== Debug Make/Unmake: %s ===\n", label)
	
	// Save initial state
	initialFEN := b.ToFEN()
	fmt.Printf("Initial FEN: %s\n", initialFEN)
	
	b.DebugPieceCounts("Before Make")
	b.DebugMove(move, "MAKE")
	
	// Make the move
	undo, err := b.MakeMoveWithUndo(move)
	if err != nil {
		fmt.Printf("❌ MakeMove failed: %v\n", err)
		return err
	}
	
	b.DebugPieceCounts("After Make")
	fmt.Printf("\n--- Undo Information ---\n")
	fmt.Printf("Undo.Move.Piece: %c\n", undo.Move.Piece)
	fmt.Printf("Undo.CapturedPiece: %c\n", undo.CapturedPiece)
	fmt.Printf("Undo.Move.Promotion: %c\n", undo.Move.Promotion)
	
	// Unmake the move
	fmt.Printf("\n--- Unmaking Move ---\n")
	b.UnmakeMove(undo)
	
	b.DebugPieceCounts("After Unmake")
	
	// Verify we're back to original state
	finalFEN := b.ToFEN()
	fmt.Printf("Final FEN: %s\n", finalFEN)
	
	if initialFEN != finalFEN {
		fmt.Printf("❌ CRITICAL: FEN mismatch after make/unmake!\n")
		fmt.Printf("   Initial: %s\n", initialFEN)
		fmt.Printf("   Final:   %s\n", finalFEN)
		return fmt.Errorf("make/unmake cycle failed")
	} else {
		fmt.Printf("✅ Make/Unmake cycle successful\n")
	}
	
	return nil
}

// PrintBoard prints a visual representation of the board
func (b *Board) PrintBoard(label string) {
	fmt.Printf("\n=== Board State: %s ===\n", label)
	fmt.Println("  a b c d e f g h")
	for rank := 7; rank >= 0; rank-- {
		fmt.Printf("%d ", rank+1)
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece == Empty {
				fmt.Print(". ")
			} else {
				fmt.Printf("%c ", piece)
			}
		}
		fmt.Printf("%d\n", rank+1)
	}
	fmt.Println("  a b c d e f g h")
}

// DebugReproduceBug reproduces the specific bug described in the issue
func DebugReproduceBug() {
	fmt.Println("\n=== Reproducing Missing Pawn Bug ===\n")
	
	// Start with initial position
	board, err := FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	
	board.PrintBoard("Initial Position")
	board.DebugPieceCounts("Initial")
	
	// Make e2e4
	move1, _ := ParseSimpleMove("e2e4")
	move1.Piece = board.GetPiece(move1.From.Rank, move1.From.File)
	err = board.DebugMakeUnmake(move1, "e2e4")
	if err != nil {
		fmt.Printf("Failed at e2e4: %v\n", err)
		return
	}
	
	board.PrintBoard("After e2e4")
	
	// Make a7a6 (computer move)
	move2, _ := ParseSimpleMove("a7a6")
	move2.Piece = board.GetPiece(move2.From.Rank, move2.From.File)
	err = board.DebugMakeUnmake(move2, "a7a6")
	if err != nil {
		fmt.Printf("Failed at a7a6: %v\n", err)
		return
	}
	
	board.PrintBoard("After a7a6 - Check for Missing Pawns!")
	board.DebugPieceCounts("After Both Moves")
	
	fmt.Println("\n=== Bug Reproduction Complete ===\n")
}