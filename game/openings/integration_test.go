package openings

import (
	"os"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestPerformanceBinIntegration(t *testing.T) {
	// Test with the actual performance.bin file from the repository
	bookPath := "testdata/performance.bin"
	
	// Check if file exists
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skipf("Skipping integration test: %s not found", bookPath)
		return
	}
	
	// Load the book
	book := NewPolyglotBook()
	err := book.LoadFromFile(bookPath)
	if err != nil {
		t.Fatalf("Failed to load performance.bin: %v", err)
	}
	
	// Verify book info
	info := book.GetBookInfo()
	t.Logf("Loaded book: %s with %d entries (%d bytes)", info.Filename, info.EntryCount, info.FileSize)
	
	if !book.IsLoaded() {
		t.Error("Book should be loaded")
	}
	
	if info.EntryCount == 0 {
		t.Error("Book should contain entries")
	}
	
	// Test with BookLookupService
	config := DefaultBookConfig()
	config.BookFiles = []string{bookPath}
	
	service := NewBookLookupService(config)
	err = service.LoadBooks()
	if err != nil {
		t.Fatalf("Failed to load books in service: %v", err)
	}
	
	// Test with starting position
	startingBoard := board.NewBoard()
	
	// Try to set up the starting position properly
	err = setUpStartingPosition(startingBoard)
	if err != nil {
		t.Fatalf("Failed to set up starting position: %v", err)
	}
	
	// Look for book moves
	bookMove, err := service.FindBookMove(startingBoard)
	if err == ErrPositionNotFound {
		t.Logf("Starting position not found in book (this is OK)")
	} else if err != nil {
		t.Errorf("Unexpected error looking up book move: %v", err)
	} else {
		t.Logf("Found book move: %+v", bookMove)
	}
	
	// Test book info
	loadedBooks := service.GetLoadedBooks()
	if len(loadedBooks) != 1 {
		t.Errorf("Expected 1 loaded book, got %d", len(loadedBooks))
	}
	
	t.Logf("Successfully tested performance.bin with %d entries", info.EntryCount)
}

func setUpStartingPosition(b *board.Board) error {
	// Set up the standard starting position
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	newBoard, err := board.FromFEN(fen)
	if err != nil {
		return err
	}
	
	// Copy the position to our board
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := newBoard.GetPiece(rank, file)
			b.SetPiece(rank, file, piece)
		}
	}
	
	b.SetSideToMove("w")
	b.SetCastlingRights("KQkq")
	b.SetEnPassantTarget(nil)
	b.SetHalfMoveClock(0)
	b.SetFullMoveNumber(1)
	
	return nil
}

func TestPolyglotBookLookupRealFile(t *testing.T) {
	bookPath := "testdata/performance.bin"
	
	// Check if file exists
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Skipf("Skipping test: %s not found", bookPath)
		return
	}
	
	book := NewPolyglotBook()
	err := book.LoadFromFile(bookPath)
	if err != nil {
		t.Fatalf("Failed to load book: %v", err)
	}
	
	// Try looking up some positions by hash
	// These are random hashes - we expect most to not be found
	testHashes := []uint64{
		0x0000968b7fcb1868, // First entry hash from hex dump
		0x0000da48997503d0, // Second entry hash
		0x0001a6f6d7e63f5f, // Fourth entry hash
		0x1234567890abcdef, // Random hash (likely not found)
	}
	
	foundCount := 0
	for i, hash := range testHashes {
		moves, err := book.LookupMove(hash)
		if err == ErrPositionNotFound {
			t.Logf("Hash %d (0x%016x): not found (expected)", i+1, hash)
		} else if err != nil {
			t.Errorf("Hash %d (0x%016x): unexpected error: %v", i+1, hash, err)
		} else {
			foundCount++
			t.Logf("Hash %d (0x%016x): found %d moves", i+1, hash, len(moves))
			for j, move := range moves {
				t.Logf("  Move %d: weight=%d, learn=%d", j+1, move.Weight, move.Learn)
			}
		}
	}
	
	t.Logf("Found %d/%d test hashes in book", foundCount, len(testHashes))
}

// Benchmark with real book file
func BenchmarkRealBookLookup(b *testing.B) {
	bookPath := "testdata/performance.bin"
	
	// Check if file exists
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		b.Skipf("Skipping benchmark: %s not found", bookPath)
		return
	}
	
	book := NewPolyglotBook()
	err := book.LoadFromFile(bookPath)
	if err != nil {
		b.Fatalf("Failed to load book: %v", err)
	}
	
	// Use a hash that exists in the book (from hex dump)
	testHash := uint64(0x0000968b7fcb1868)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := book.LookupMove(testHash)
		if err != nil && err != ErrPositionNotFound {
			b.Fatalf("Lookup failed: %v", err)
		}
	}
}