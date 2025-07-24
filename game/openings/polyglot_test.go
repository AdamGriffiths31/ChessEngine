package openings

import (
	"bytes"
	"os"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestMoveEncoding(t *testing.T) {
	pb := NewPolyglotBook()
	
	// Create test board
	testBoard, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	tests := []struct {
		name string
		move board.Move
	}{
		{
			name: "simple pawn move",
			move: board.Move{
				From: board.Square{File: 4, Rank: 1}, // e2
				To:   board.Square{File: 4, Rank: 3}, // e4
			},
		},
		{
			name: "knight move",
			move: board.Move{
				From: board.Square{File: 1, Rank: 0}, // b1
				To:   board.Square{File: 2, Rank: 2}, // c3
			},
		},
		{
			name: "queen promotion",
			move: board.Move{
				From:      board.Square{File: 4, Rank: 6}, // e7
				To:        board.Square{File: 4, Rank: 7}, // e8
				Promotion: board.WhiteQueen,
			},
		},
		{
			name: "knight promotion",
			move: board.Move{
				From:      board.Square{File: 0, Rank: 6}, // a7
				To:        board.Square{File: 0, Rank: 7}, // a8
				Promotion: board.WhiteKnight,
			},
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Encode then decode
			encoded := pb.encodeMove(test.move)
			decoded, err := pb.decodeMove(encoded, testBoard)
			
			if err != nil {
				t.Fatalf("Failed to decode move: %v", err)
			}
			
			// Compare moves
			if !movesEqual(test.move, decoded) {
				t.Errorf("Move encoding/decoding failed:\nOriginal: %+v\nDecoded:  %+v", test.move, decoded)
			}
		})
	}
}

func TestMoveDecodingBitfields(t *testing.T) {
	pb := NewPolyglotBook()
	
	// Create test board
	testBoard, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	tests := []struct {
		name     string
		encoded  uint16
		expected board.Move
	}{
		{
			name:    "e2-e4 (pawn two squares)",
			encoded: 0x031C, // to=28 (e4), from=12 (e2) = 28 | (12 << 6) = 28 | 768 = 796 = 0x031C
			expected: board.Move{
				From: board.Square{File: 4, Rank: 1},
				To:   board.Square{File: 4, Rank: 3},
			},
		},
		{
			name:    "a7-a8=Q (queen promotion)",
			encoded: 0x4C38, // to=56 (a8), from=48 (a7), promo=4 = 56 | (48 << 6) | (4 << 12) = 0x4C38
			expected: board.Move{
				From:      board.Square{File: 0, Rank: 6},
				To:        board.Square{File: 0, Rank: 7},
				Promotion: board.WhiteQueen,
			},
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decoded, err := pb.decodeMove(test.encoded, testBoard)
			if err != nil {
				t.Fatalf("Failed to decode move: %v", err)
			}
			
			if !movesEqual(test.expected, decoded) {
				t.Errorf("Move decoding failed:\nExpected: %+v\nDecoded:  %+v", test.expected, decoded)
			}
		})
	}
}

func TestPolyglotEntryWriteRead(t *testing.T) {
	// Create test entry
	entry := PolyglotEntry{
		Hash:   0x123456789ABCDEF0,
		Move:   0x1234,
		Weight: 100,
		Learn:  42,
	}
	
	// Write to buffer
	var buf bytes.Buffer
	err := WriteEntry(&buf, entry)
	if err != nil {
		t.Fatalf("Failed to write entry: %v", err)
	}
	
	// Check buffer size
	if buf.Len() != PolyglotEntrySize {
		t.Errorf("Expected buffer size %d, got %d", PolyglotEntrySize, buf.Len())
	}
	
	// Create temporary file and write entry
	tmpFile := createTestBookFile(t, []PolyglotEntry{entry})
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()
	
	// Load book and verify entry
	book := NewPolyglotBook()
	err = book.LoadFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load book: %v", err)
	}
	
	if len(book.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(book.entries))
	}
	
	loadedEntry := book.entries[0]
	if loadedEntry != entry {
		t.Errorf("Entry mismatch:\nOriginal: %+v\nLoaded:   %+v", entry, loadedEntry)
	}
}

func TestPolyglotBookLookup(t *testing.T) {
	// Create test entries with same hash but different moves
	// Note: entries must be sorted by hash for PolyGlot format
	hash := uint64(0x123456789ABCDEF0)
	entries := []PolyglotEntry{
		{Hash: hash, Move: 0x1234, Weight: 100, Learn: 10},     // Same hash (d2-h5)
		{Hash: hash, Move: 0x0678, Weight: 200, Learn: 20},     // Same hash (b4-a8, no promotion)
		{Hash: hash + 1, Move: 0x9ABC, Weight: 50, Learn: 5},   // Different hash (sorted after)
	}
	
	tmpFile := createTestBookFile(t, entries)
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()
	
	book := NewPolyglotBook()
	err := book.LoadFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load book: %v", err)
	}
	
	// Create test board
	testBoard, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	// Lookup moves for the hash
	bookMoves, err := book.LookupMove(hash, testBoard)
	if err != nil {
		t.Fatalf("Failed to lookup moves: %v", err)
	}
	
	if len(bookMoves) != 2 {
		t.Errorf("Expected 2 moves for hash %x, got %d", hash, len(bookMoves))
	}
	
	// Verify weights
	expectedWeights := []uint16{100, 200}
	for i, bookMove := range bookMoves {
		if bookMove.Weight != expectedWeights[i] {
			t.Errorf("Expected weight %d, got %d", expectedWeights[i], bookMove.Weight)
		}
	}
	
	// Lookup non-existent hash
	_, err = book.LookupMove(0x999, testBoard)
	if err != ErrPositionNotFound {
		t.Errorf("Expected ErrPositionNotFound for non-existent hash, got %v", err)
	}
}

func TestPolyglotBookInfo(t *testing.T) {
	entries := []PolyglotEntry{
		{Hash: 1, Move: 0x1234, Weight: 100, Learn: 10},
		{Hash: 2, Move: 0x5678, Weight: 200, Learn: 20},
	}
	
	tmpFile := createTestBookFile(t, entries)
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()
	
	book := NewPolyglotBook()
	
	// Check before loading
	if book.IsLoaded() {
		t.Error("Book should not be loaded initially")
	}
	
	err := book.LoadFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load book: %v", err)
	}
	
	// Check after loading
	if !book.IsLoaded() {
		t.Error("Book should be loaded after LoadFromFile")
	}
	
	info := book.GetBookInfo()
	if info.Filename != tmpFile.Name() {
		t.Errorf("Expected filename %s, got %s", tmpFile.Name(), info.Filename)
	}
	
	if info.EntryCount != 2 {
		t.Errorf("Expected entry count 2, got %d", info.EntryCount)
	}
	
	expectedSize := int64(2 * PolyglotEntrySize)
	if info.FileSize != expectedSize {
		t.Errorf("Expected file size %d, got %d", expectedSize, info.FileSize)
	}
}

func TestInvalidPolyglotFiles(t *testing.T) {
	book := NewPolyglotBook()
	
	// Test non-existent file
	err := book.LoadFromFile("/nonexistent/file.bin")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	
	// Test invalid file size (not multiple of 16)
	tmpFile := createInvalidSizeFile(t)
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()
	
	err = book.LoadFromFile(tmpFile.Name())
	if err != ErrInvalidBookFile {
		t.Errorf("Expected ErrInvalidBookFile for invalid size, got %v", err)
	}
}

// Helper functions for testing
// Note: movesEqual is now defined in book.go

func createTestBookFile(t *testing.T, entries []PolyglotEntry) *os.File {
	// Create a real temporary file
	tmpFile, err := os.CreateTemp("", "test_book_*.bin")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	// Write entries to the file
	for _, entry := range entries {
		err := WriteEntry(tmpFile, entry)
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			t.Fatalf("Failed to write entry: %v", err)
		}
	}
	
	// Close and reopen for reading
	tmpFile.Close()
	
	// Return a file handle for reading
	readFile, err := os.Open(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to reopen temp file: %v", err)
	}
	
	return readFile
}

func createInvalidSizeFile(t *testing.T) *os.File {
	// Create a file with invalid size (not multiple of 16)
	tmpFile, err := os.CreateTemp("", "invalid_book_*.bin")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	// Write 17 bytes (invalid for PolyGlot format)
	data := make([]byte, 17)
	for i := range data {
		data[i] = byte(i)
	}
	
	_, err = tmpFile.Write(data)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to write invalid data: %v", err)
	}
	
	tmpFile.Close()
	
	// Return a file handle for reading
	readFile, err := os.Open(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to reopen temp file: %v", err)
	}
	
	return readFile
}

// Benchmark tests
func BenchmarkMoveEncoding(b *testing.B) {
	pb := NewPolyglotBook()
	move := board.Move{
		From: board.Square{File: 4, Rank: 1},
		To:   board.Square{File: 4, Rank: 3},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pb.encodeMove(move)
	}
}

func BenchmarkMoveDecoding(b *testing.B) {
	pb := NewPolyglotBook()
	encoded := uint16(0x01CC)
	
	// Create test board
	testBoard, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pb.decodeMove(encoded, testBoard)
	}
}