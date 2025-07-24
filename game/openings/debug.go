package openings

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// TestOpeningBook runs comprehensive tests on the opening book system
func TestOpeningBook(bls *BookLookupService) {
	fmt.Println("=== Opening Book Debug Tests ===")
	
	if !bls.IsEnabled() {
		fmt.Println("ERROR: Opening book is disabled!")
		return
	}
	
	// Test 1: Initial position
	fmt.Println("\n--- Test 1: Initial Position ---")
	initialBoard, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		fmt.Printf("ERROR: Failed to create initial board: %v\n", err)
		return
	}
	
	initialHash := bls.zobrist.HashPosition(initialBoard)
	fmt.Printf("Initial position hash: %016X\n", initialHash)
	
	move, err := bls.FindBookMove(initialBoard)
	if err != nil {
		fmt.Printf("ERROR: Initial position not found in opening book: %v\n", err)
	} else {
		fmt.Printf("Found opening move for initial position: %s -> %s\n", 
			move.From.String(), move.To.String())
	}
	
	// Test 2: After e2e4
	fmt.Println("\n--- Test 2: After e2e4 ---")
	e2e4Move, err := board.ParseSimpleMove("e2e4")
	if err != nil {
		fmt.Printf("ERROR: Failed to parse e2e4: %v\n", err)
		return
	}
	
	err = initialBoard.MakeMove(e2e4Move)
	if err != nil {
		fmt.Printf("ERROR: Failed to make e2e4 move: %v\n", err)
		return
	}
	
	afterE4Hash := bls.zobrist.HashPosition(initialBoard)
	fmt.Printf("Position after e2e4 hash: %016X\n", afterE4Hash)
	
	move, err = bls.FindBookMove(initialBoard)
	if err != nil {
		fmt.Printf("ERROR: Position after e2e4 not found in opening book: %v\n", err)
		fmt.Printf("This is the main bug - the opening book should have responses to e2e4!\n")
	} else {
		fmt.Printf("Found response to e2e4: %s -> %s\n", 
			move.From.String(), move.To.String())
	}
	
	// Test 3: Try a few more common moves
	fmt.Println("\n--- Test 3: Other Common Positions ---")
	testPositions := []string{
		"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1", // After 1.e4
		"rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2", // After 1.e4 e5
		"rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2", // After 1.e4 e5 2.Nf3
	}
	
	for i, fen := range testPositions {
		fmt.Printf("\nTest position %d: %s\n", i+1, fen)
		testBoard, err := board.FromFEN(fen)
		if err != nil {
			fmt.Printf("ERROR: Failed to parse FEN: %v\n", err)
			continue
		}
		
		hash := bls.zobrist.HashPosition(testBoard)
		fmt.Printf("Hash: %016X\n", hash)
		
		move, err := bls.FindBookMove(testBoard)
		if err != nil {
			fmt.Printf("No book move found: %v\n", err)
		} else {
			fmt.Printf("Book move: %s -> %s\n", move.From.String(), move.To.String())
		}
	}
}

// DebugOpeningBook displays the first N entries from loaded opening books
func DebugOpeningBook(bls *BookLookupService, limit int) {
	fmt.Printf("\n=== Opening Book Contents (first %d entries) ===\n", limit)
	
	if !bls.IsEnabled() {
		fmt.Println("Opening book is disabled")
		return
	}
	
	books := bls.GetLoadedBooks()
	if len(books) == 0 {
		fmt.Println("No books loaded")
		return
	}
	
	for i, bookInfo := range books {
		fmt.Printf("\nBook %d: %s (%d entries)\n", i+1, bookInfo.Filename, bookInfo.EntryCount)
		
		// Note: The actual book entries are not directly accessible through the current API
		// This would require modifying the BookManager or PolyglotBook to expose entries
		fmt.Printf("File size: %d bytes\n", bookInfo.FileSize)
		if bookInfo.Filename != "" {
			fmt.Printf("Filename: %s\n", bookInfo.Filename)
		}
	}
	
	fmt.Printf("\nTotal books loaded: %d\n", len(books))
}

// DebugHashCalculation shows step-by-step hash calculation for a position
func DebugHashCalculation(bls *BookLookupService, b *board.Board, description string) {
	fmt.Printf("\n=== Hash Calculation Debug: %s ===\n", description)
	
	// Show board state
	fmt.Printf("Side to move: %s\n", b.GetSideToMove())
	fmt.Printf("Castling rights: %s\n", b.GetCastlingRights())
	if b.GetEnPassantTarget() != nil {
		fmt.Printf("En passant target: %s\n", b.GetEnPassantTarget().String())
	} else {
		fmt.Printf("En passant target: none\n")
	}
	
	// Show piece positions for first and last ranks
	fmt.Println("\nPiece positions:")
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece != board.Empty {
				squareStr := board.Square{File: file, Rank: rank}.String()
				fmt.Printf("  %s: '%c'\n", squareStr, piece)
			}
		}
	}
	
	// Calculate and display hash
	hash := bls.zobrist.HashPosition(b)
	fmt.Printf("\nFinal hash: %016X\n", hash)
}

// CompareHashes compares hashes of two positions to help debug differences
func CompareHashes(bls *BookLookupService, b1, b2 *board.Board, desc1, desc2 string) {
	fmt.Printf("\n=== Hash Comparison: %s vs %s ===\n", desc1, desc2)
	
	hash1 := bls.zobrist.HashPosition(b1)
	hash2 := bls.zobrist.HashPosition(b2)
	
	fmt.Printf("%s hash: %016X\n", desc1, hash1)
	fmt.Printf("%s hash: %016X\n", desc2, hash2)
	
	if hash1 == hash2 {
		fmt.Println("Hashes are IDENTICAL")
	} else {
		fmt.Printf("Hashes are DIFFERENT (XOR: %016X)\n", hash1^hash2)
	}
}

// RunOpeningBookDebugTests runs all opening book debug tests
func RunOpeningBookDebugTests(bls *BookLookupService) {
	fmt.Println("\n=== Opening Book Debug Tests ===")
	
	// Load books first
	err := bls.LoadBooks()
	if err != nil {
		fmt.Printf("ERROR: Failed to load books: %v\n", err)
		return
	}
	
	// Show loaded books
	DebugOpeningBook(bls, 10)
	
	// Test positions
	TestOpeningBook(bls)
	
	// Debug hash calculation for initial position
	initialBoard, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err == nil {
		DebugHashCalculation(bls, initialBoard, "Initial Position")
		
		// Make e2e4 and test again
		e2e4Move, err := board.ParseSimpleMove("e2e4")
		if err == nil {
			err = initialBoard.MakeMove(e2e4Move)
			if err == nil {
				DebugHashCalculation(bls, initialBoard, "After e2e4")
			}
		}
	}
	
	fmt.Println("\n=== Opening Book Debug Tests Complete ===")
}