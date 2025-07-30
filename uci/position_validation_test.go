package uci

import (
	"testing"
	"strings"
	"github.com/AdamGriffiths31/ChessEngine/game"
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// TestPositionValidation validates our FEN interpretation matches expected position
func TestPositionValidation(t *testing.T) {
	// The exact FEN from the UCI logs where illegal move occurred
	fenPosition := "rn1qk2r/1b3ppp/1p2pn2/p2p4/PpPQPb2/5P1P/3K4/RNBQ1BNR w kq - 1 14"
	
	t.Logf("=== POSITION VALIDATION ===")
	t.Logf("FEN: %s", fenPosition)
	
	// Create engine and load position
	engine := game.NewEngine()
	err := engine.LoadFromFEN(fenPosition)
	if err != nil {
		t.Fatalf("Failed to load FEN: %v", err)
	}
	
	// Get board and validate key pieces
	gameState := engine.GetState()
	gameBoard := gameState.Board
	
	// Validate each rank manually
	t.Logf("\n=== RANK-BY-RANK VALIDATION ===")
	
	// Rank 8 (index 7): rn1qk2r
	validateRank(t, gameBoard, 7, "rn1qk2r", map[int]board.Piece{
		0: board.BlackRook, 1: board.BlackKnight, 2: board.Empty, 3: board.BlackQueen,
		4: board.BlackKing, 5: board.Empty, 6: board.Empty, 7: board.BlackRook,
	})
	
	// Rank 7 (index 6): 1b3ppp  
	validateRank(t, gameBoard, 6, "1b3ppp", map[int]board.Piece{
		0: board.Empty, 1: board.BlackBishop, 2: board.Empty, 3: board.Empty,
		4: board.Empty, 5: board.BlackPawn, 6: board.BlackPawn, 7: board.BlackPawn,
	})
	
	// Rank 6 (index 5): 1p2pn2
	validateRank(t, gameBoard, 5, "1p2pn2", map[int]board.Piece{
		0: board.Empty, 1: board.BlackPawn, 2: board.Empty, 3: board.Empty,
		4: board.BlackPawn, 5: board.BlackKnight, 6: board.Empty, 7: board.Empty,
	})
	
	// Rank 5 (index 4): p2p4
	validateRank(t, gameBoard, 4, "p2p4", map[int]board.Piece{
		0: board.BlackPawn, 1: board.Empty, 2: board.Empty, 3: board.BlackPawn,
		4: board.Empty, 5: board.Empty, 6: board.Empty, 7: board.Empty,
	})
	
	// Rank 4 (index 3): PpPQPb2 ← CRITICAL RANK
	t.Logf("\n*** CRITICAL RANK 4 ANALYSIS ***")
	validateRank(t, gameBoard, 3, "PpPQPb2", map[int]board.Piece{
		0: board.WhitePawn,   // a4
		1: board.BlackPawn,   // b4  
		2: board.WhitePawn,   // c4
		3: board.WhiteQueen,  // d4 ← QUEEN HERE
		4: board.WhitePawn,   // e4
		5: board.BlackBishop, // f4 ← BISHOP HERE
		6: board.Empty,       // g4
		7: board.Empty,       // h4
	})
	
	// Rank 3 (index 2): 5P1P
	validateRank(t, gameBoard, 2, "5P1P", map[int]board.Piece{
		0: board.Empty, 1: board.Empty, 2: board.Empty, 3: board.Empty,
		4: board.Empty, 5: board.WhitePawn, 6: board.Empty, 7: board.WhitePawn,
	})
	
	// Rank 2 (index 1): 3K4 ← CRITICAL RANK  
	t.Logf("\n*** CRITICAL RANK 2 ANALYSIS ***")
	validateRank(t, gameBoard, 1, "3K4", map[int]board.Piece{
		0: board.Empty, 1: board.Empty, 2: board.Empty, 3: board.WhiteKing, // d2 ← KING HERE
		4: board.Empty, 5: board.Empty, 6: board.Empty, 7: board.Empty,
	})
	
	// Rank 1 (index 0): RNBQ1BNR
	validateRank(t, gameBoard, 0, "RNBQ1BNR", map[int]board.Piece{
		0: board.WhiteRook, 1: board.WhiteKnight, 2: board.WhiteBishop, 3: board.WhiteQueen,
		4: board.Empty, 5: board.WhiteBishop, 6: board.WhiteKnight, 7: board.WhiteRook,
	})
	
	// Key position verification
	t.Logf("\n=== KEY PIECES VERIFICATION ===")
	whiteKing := gameBoard.GetPiece(1, 3) // d2
	whiteQueen := gameBoard.GetPiece(3, 3) // d4  
	blackBishop := gameBoard.GetPiece(3, 5) // f4
	
	t.Logf("White King on d2: %d (expected %d)", whiteKing, board.WhiteKing)
	t.Logf("White Queen on d4: %d (expected %d)", whiteQueen, board.WhiteQueen)
	t.Logf("Black Bishop on f4: %d (expected %d)", blackBishop, board.BlackBishop)
	
	if whiteKing != board.WhiteKing {
		t.Errorf("WRONG: King not on d2, found piece %d", whiteKing)
	}
	if whiteQueen != board.WhiteQueen {
		t.Errorf("WRONG: Queen not on d4, found piece %d", whiteQueen)
	}
	if blackBishop != board.BlackBishop {
		t.Errorf("WRONG: Bishop not on f4, found piece %d", blackBishop)
	}
	
	// Verify the diagonal attack  
	t.Logf("\n=== DIAGONAL ATTACK ANALYSIS ===")
	
	// f4 -> e3 -> d2 diagonal
	t.Logf("Diagonal f4 -> d2:")
	t.Logf("  f4 (5,3): %d", gameBoard.GetPiece(3, 5))
	t.Logf("  e3 (4,2): %d", gameBoard.GetPiece(2, 4))
	t.Logf("  d2 (3,1): %d", gameBoard.GetPiece(1, 3))
	
	// Check if king is in check before any moves
	whiteKingSquare := board.FileRankToSquare(3, 1) // d2
	isInCheck := gameBoard.IsSquareAttackedByColor(whiteKingSquare, board.BitboardBlack)
	t.Logf("White king in check: %v", isInCheck)
	
	if !isInCheck {
		t.Errorf("ERROR: King should be in check from f4 bishop!")
	}
	
	// Now test if d4e3 would block the check
	t.Logf("\n=== MOVE d4e3 BLOCKING TEST ===")
	
	// Simulate d4e3 move
	originalE3 := gameBoard.GetPiece(2, 4) // e3
	gameBoard.SetPiece(3, 3, board.Empty)    // Clear d4
	gameBoard.SetPiece(2, 4, board.WhiteQueen) // Queen to e3
	
	// Check if king still in check after move
	stillInCheck := gameBoard.IsSquareAttackedByColor(whiteKingSquare, board.BitboardBlack)
	t.Logf("After d4e3, king still in check: %v", stillInCheck)
	
	if stillInCheck {
		t.Logf("d4e3 does NOT resolve check - move should be ILLEGAL")
	} else {
		t.Logf("d4e3 RESOLVES check - move should be LEGAL")
	}
	
	// Restore position
	gameBoard.SetPiece(2, 4, originalE3)       // Restore e3
	gameBoard.SetPiece(3, 3, board.WhiteQueen) // Queen back to d4
	
	// Generate FEN and compare
	t.Logf("\n=== FEN ROUND-TRIP TEST ===")
	generatedFEN := engine.GetCurrentFEN()
	t.Logf("Original FEN:  %s", fenPosition)
	t.Logf("Generated FEN: %s", generatedFEN)
	
	if !strings.HasPrefix(generatedFEN, strings.Split(fenPosition, " ")[0]) {
		t.Errorf("FEN round-trip failed - position mismatch!")
	}
}

// validateRank checks if a rank matches expected piece placement
func validateRank(t *testing.T, b *board.Board, rank int, fenRank string, expectedPieces map[int]board.Piece) {
	t.Logf("Rank %d (%s): %s", rank+1, fenRank, getRankString(b, rank))
	
	for file := 0; file < 8; file++ {
		actualPiece := b.GetPiece(rank, file)
		expectedPiece := expectedPieces[file]
		
		if actualPiece != expectedPiece {
			t.Errorf("  MISMATCH at %c%d: expected %d, got %d", 
				'a'+file, rank+1, expectedPiece, actualPiece)
		}
	}
}

// getRankString returns a visual representation of a rank
func getRankString(b *board.Board, rank int) string {
	var result strings.Builder
	for file := 0; file < 8; file++ {
		piece := b.GetPiece(rank, file)
		if piece == board.Empty {
			result.WriteString(".")
		} else {
			result.WriteString(pieceToChar(piece))
		}
	}
	return result.String()
}

// pieceToChar converts piece to character representation
func pieceToChar(piece board.Piece) string {
	switch piece {
	case board.WhitePawn: return "P"
	case board.WhiteRook: return "R"
	case board.WhiteKnight: return "N"
	case board.WhiteBishop: return "B"
	case board.WhiteQueen: return "Q"
	case board.WhiteKing: return "K"
	case board.BlackPawn: return "p"
	case board.BlackRook: return "r"
	case board.BlackKnight: return "n"
	case board.BlackBishop: return "b"
	case board.BlackQueen: return "q"
	case board.BlackKing: return "k"
	default: return "?"
	}
}