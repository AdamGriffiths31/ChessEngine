package openings

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestE2E4OpeningBookLookup(t *testing.T) {
	// Create a service with the real opening book
	config := DefaultBookConfig()
	config.BookFiles = []string{"testdata/performance.bin"}
	config.SelectionMode = SelectBest
	
	service := NewBookLookupService(config)
	
	// Load the books
	err := service.LoadBooks()
	if err != nil {
		t.Fatalf("Failed to load books: %v", err)
	}
	
	// Test with starting position
	startingBoard, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create starting board: %v", err)
	}
	
	// Look for a book move in starting position
	bookMove, err := service.FindBookMove(startingBoard)
	if err != nil {
		t.Logf("No opening book move found for starting position: %v", err)
		return
	}
	
	t.Logf("Found opening book move: %s%s-%s%s", 
		string('a'+rune(bookMove.From.File)), string('1'+rune(bookMove.From.Rank)),
		string('a'+rune(bookMove.To.File)), string('1'+rune(bookMove.To.Rank)))
	
	// Verify the move is valid
	if !isValidMove(*bookMove) {
		t.Errorf("Invalid book move returned: %+v", *bookMove)
	}
}

func TestE2E4ResponseFromBook(t *testing.T) {
	// Test that after white plays e2-e4, black gets a book response
	config := DefaultBookConfig()
	config.BookFiles = []string{"testdata/performance.bin"}
	config.SelectionMode = SelectBest
	
	service := NewBookLookupService(config)
	
	// Load the books
	err := service.LoadBooks()
	if err != nil {
		t.Fatalf("Failed to load books: %v", err)
	}
	
	// Create board position after 1.e4
	boardAfterE4, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")
	if err != nil {
		t.Fatalf("Failed to create board after e4: %v", err)
	}
	
	// Look for black's book response
	blackResponse, err := service.FindBookMove(boardAfterE4)
	if err != nil {
		t.Logf("No opening book response for black after e4: %v", err)
		return
	}
	
	t.Logf("Black's book response to e4: %s%s-%s%s", 
		string('a'+rune(blackResponse.From.File)), string('1'+rune(blackResponse.From.Rank)),
		string('a'+rune(blackResponse.To.File)), string('1'+rune(blackResponse.To.Rank)))
	
	// Common responses to e4 include e5, c5, e6, c6, d5, Nf6, etc.
	expectedResponses := []string{"e7e5", "c7c5", "e7e6", "c7c6", "d7d5", "g8f6", "b8c6"}
	moveStr := getMoveString(*blackResponse)
	
	found := false
	for _, expected := range expectedResponses {
		if moveStr == expected {
			found = true
			break
		}
	}
	
	if !found {
		t.Logf("Unexpected response to e4: %s (not necessarily wrong, just uncommon)", moveStr)
	}
	
	// Verify the move is valid
	if !isValidMove(*blackResponse) {
		t.Errorf("Invalid book move returned: %+v", *blackResponse)
	}
}

func TestE2E4SicilianResponse(t *testing.T) {
	// Test specifically for the Sicilian Defense (1.e4 c5)
	config := DefaultBookConfig()
	config.BookFiles = []string{"testdata/performance.bin"}
	config.SelectionMode = SelectWeightedRandom
	
	service := NewBookLookupService(config)
	
	// Load the books
	err := service.LoadBooks()
	if err != nil {
		t.Fatalf("Failed to load books: %v", err)
	}
	
	// Position after 1.e4 c5
	sicilianBoard, err := board.FromFEN("rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2")
	if err != nil {
		t.Fatalf("Failed to create Sicilian position: %v", err)
	}
	
	// Look for white's response to Sicilian
	whiteResponse, err := service.FindBookMove(sicilianBoard)
	if err != nil {
		t.Logf("No opening book response for white in Sicilian: %v", err)
		return
	}
	
	t.Logf("White's book response in Sicilian: %s%s-%s%s", 
		string('a'+rune(whiteResponse.From.File)), string('1'+rune(whiteResponse.From.Rank)),
		string('a'+rune(whiteResponse.To.File)), string('1'+rune(whiteResponse.To.Rank)))
	
	// Common responses in Sicilian include Nf3, Nc3, f4, Bc4, etc.
	moveStr := getMoveString(*whiteResponse)
	commonSicilianMoves := []string{"g1f3", "b1c3", "f2f4", "f1c4", "d2d4"}
	
	found := false
	for _, expected := range commonSicilianMoves {
		if moveStr == expected {
			found = true
			break
		}
	}
	
	if !found {
		t.Logf("Uncommon Sicilian response: %s", moveStr)
	}
	
	// Verify the move is valid
	if !isValidMove(*whiteResponse) {
		t.Errorf("Invalid book move returned: %+v", *whiteResponse)
	}
}

func TestE2E4OpeningVariations(t *testing.T) {
	// Test multiple variations after 1.e4
	config := DefaultBookConfig()
	config.BookFiles = []string{"testdata/performance.bin"}
	config.SelectionMode = SelectRandom
	
	service := NewBookLookupService(config)
	
	// Load the books
	err := service.LoadBooks()
	if err != nil {
		t.Fatalf("Failed to load books: %v", err)
	}
	
	variations := []struct {
		name string
		fen  string
		desc string
	}{
		{
			name: "Kings Pawn Game",
			fen:  "rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2",
			desc: "After 1.e4 e5",
		},
		{
			name: "French Defense",
			fen:  "rnbqkbnr/pppp1ppp/4p3/8/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 2",
			desc: "After 1.e4 e6",
		},
		{
			name: "Caro-Kann Defense",
			fen:  "rnbqkbnr/pp1ppppp/2p5/8/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 2",
			desc: "After 1.e4 c6",
		},
		{
			name: "Scandinavian Defense",
			fen:  "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2",
			desc: "After 1.e4 d5",
		},
	}
	
	for _, variation := range variations {
		t.Run(variation.name, func(t *testing.T) {
			board, err := board.FromFEN(variation.fen)
			if err != nil {
				t.Fatalf("Failed to create board for %s: %v", variation.name, err)
			}
			
			move, err := service.FindBookMove(board)
			if err != nil {
				t.Logf("No book move found for %s: %v", variation.desc, err)
				return
			}
			
			t.Logf("%s - Book move: %s%s-%s%s", 
				variation.desc,
				string('a'+rune(move.From.File)), string('1'+rune(move.From.Rank)),
				string('a'+rune(move.To.File)), string('1'+rune(move.To.Rank)))
			
			if !isValidMove(*move) {
				t.Errorf("Invalid book move for %s: %+v", variation.name, *move)
			}
		})
	}
}

func TestBookMoveSelectionModes(t *testing.T) {
	// Test different selection modes with the same position
	startingBoard, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	modes := []struct {
		mode SelectionMode
		name string
	}{
		{SelectBest, "Best"},
		{SelectRandom, "Random"},
		{SelectWeightedRandom, "WeightedRandom"},
	}
	
	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			config := DefaultBookConfig()
			config.BookFiles = []string{"testdata/performance.bin"}
			config.SelectionMode = mode.mode
			
			service := NewBookLookupService(config)
			
			// Load the books
			err := service.LoadBooks()
			if err != nil {
				t.Fatalf("Failed to load books: %v", err)
			}
			
			// Try to get a move multiple times to see variation (for random modes)
			moves := make([]*board.Move, 5)
			for i := 0; i < 5; i++ {
				move, err := service.FindBookMove(startingBoard)
				if err != nil {
					t.Logf("No book move found with %s mode: %v", mode.name, err)
					return
				}
				moves[i] = move
			}
			
			// Log the moves found
			t.Logf("%s mode moves:", mode.name)
			for i, move := range moves {
				if move != nil {
					t.Logf("  %d: %s", i+1, getMoveString(*move))
				}
			}
			
			// For SelectBest, all moves should be the same
			if mode.mode == SelectBest {
				for i := 1; i < len(moves); i++ {
					if moves[i] != nil && moves[0] != nil {
						if getMoveString(*moves[i]) != getMoveString(*moves[0]) {
							t.Errorf("SelectBest should return consistent moves, got %s and %s",
								getMoveString(*moves[0]), getMoveString(*moves[i]))
						}
					}
				}
			}
		})
	}
}

func TestBookWeightThreshold(t *testing.T) {
	// Test weight threshold filtering
	config := DefaultBookConfig()
	config.BookFiles = []string{"testdata/performance.bin"}
	config.WeightThreshold = 100 // Only consider moves with weight >= 100
	
	service := NewBookLookupService(config)
	
	// Load the books
	err := service.LoadBooks()
	if err != nil {
		t.Fatalf("Failed to load books: %v", err)
	}
	
	startingBoard, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	move, err := service.FindBookMove(startingBoard)
	if err != nil {
		t.Logf("No book move found with weight threshold 100: %v", err)
		return
	}
	
	t.Logf("Move with weight >= 100: %s", getMoveString(*move))
	
	// Test with very high threshold
	config.WeightThreshold = 10000
	service = NewBookLookupService(config)
	
	_, err = service.FindBookMove(startingBoard)
	if err == nil {
		t.Log("Found move even with very high threshold")
	} else {
		t.Logf("No moves found with high threshold (expected): %v", err)
	}
}

// Helper functions

func isValidMove(move board.Move) bool {
	return move.From.File >= 0 && move.From.File <= 7 &&
		   move.From.Rank >= 0 && move.From.Rank <= 7 &&
		   move.To.File >= 0 && move.To.File <= 7 &&
		   move.To.Rank >= 0 && move.To.Rank <= 7
}

func getMoveString(move board.Move) string {
	fromFile := string('a' + rune(move.From.File))
	fromRank := string('1' + rune(move.From.Rank))
	toFile := string('a' + rune(move.To.File))
	toRank := string('1' + rune(move.To.Rank))
	
	return fromFile + fromRank + toFile + toRank
}

// Benchmark for e2e4 book lookups
func BenchmarkE2E4BookLookup(b *testing.B) {
	config := DefaultBookConfig()
	config.BookFiles = []string{"testdata/performance.bin"}
	config.SelectionMode = SelectBest
	
	service := NewBookLookupService(config)
	
	// Load the books
	err := service.LoadBooks()
	if err != nil {
		b.Fatalf("Failed to load books: %v", err)
	}
	
	startingBoard, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.FindBookMove(startingBoard)
	}
}

func BenchmarkSicilianBookLookup(b *testing.B) {
	config := DefaultBookConfig()
	config.BookFiles = []string{"testdata/performance.bin"}
	config.SelectionMode = SelectWeightedRandom
	
	service := NewBookLookupService(config)
	
	// Load the books
	err := service.LoadBooks()
	if err != nil {
		b.Fatalf("Failed to load books: %v", err)
	}
	
	sicilianBoard, _ := board.FromFEN("rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.FindBookMove(sicilianBoard)
	}
}