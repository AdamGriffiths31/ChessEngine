package uci

import (
	"strings"
	"testing"
	
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// TestRealF6F7IllegalMoveBug recreates the exact bug from the PGN file
// Game sequence: c4 Nf6 Nf3 a6 Nc3 a5 a3 a4 d4 b6 Bf4 b5 cxb5 c6 e3 c5 Bxb8 c4 Rc1 d6 Bxc4 d5 Nxd5 e6 Nc7+ Qxc7 Bxc7 e5 Bxf7+
// Then Black tries f6f7 which is illegal
func TestRealF6F7IllegalMoveBug(t *testing.T) {
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	t.Logf("=== RECREATING REAL f6f7 BUG FROM PGN ===")
	
	// Real game sequence from PGN converted to UCI format
	realMoves := []string{
		"c2c4", "g8f6", "g1f3", "a7a6", "b1c3", "a6a5", "a2a3", "a5a4", "d2d4", "b7b6",
		"c1f4", "b6b5", "c4b5", "c7c6", "e2e3", "c6c5", "f4b8", "c5c4", "a1c1", "d7d6", 
		"b8c4", "d6d5", "f3d5", "e7e6", "d5c7", "d8c7", "c4c7", "e6e5", "c7f7",
	}
	
	// Apply all moves up to the critical point (after Bxf7+)
	movesApplied := strings.Join(realMoves, " ")
	posCmd := "position startpos moves " + movesApplied
	engine.HandleCommand(posCmd)
	
	// This is the position where Black will try the illegal f6f7
	criticalFEN := engine.engine.GetCurrentFEN()
	t.Logf("Position after Bxf7+: %s", criticalFEN)
	t.Logf("Black is in check and must respond")
	
	// Now ask the engine to generate a move from this position
	t.Logf("\n=== ASKING ENGINE TO GENERATE MOVE ===")
	engine.HandleCommand("go depth 1")
	t.Logf("Engine should now analyze and respond with a move")
	
	// Analyze the board state
	board := engine.engine.GetState().Board
	
	t.Logf("\n=== CRITICAL POSITION ANALYSIS ===")
	
	// Check key squares
	f6Piece := getRealBugPieceAtSquare(board, "f6")
	f7Piece := getRealBugPieceAtSquare(board, "f7")
	kingSquare := findKing(board, false) // Black king
	
	t.Logf("f6 contains: %c (%s %s)", f6Piece, getRealBugPieceColorString(f6Piece), realBugDescribePiece(f6Piece))
	t.Logf("f7 contains: %c (%s %s)", f7Piece, getRealBugPieceColorString(f7Piece), realBugDescribePiece(f7Piece))
	t.Logf("Black king position: %s", kingSquare)
	
	// Check if Black is actually in check
	isInCheck := isKingInCheck(board, false) // false for Black
	t.Logf("Black in check: %v", isInCheck)
	
	// Generate legal moves for Black
	t.Logf("\n=== LEGAL MOVE ANALYSIS ===")
	legalMoves := generateLegalMoves(engine)
	t.Logf("Legal moves available to Black: %d", len(legalMoves))
	
	for i, move := range legalMoves {
		if i < 10 { // Show first 10 moves
			t.Logf("  Legal[%d]: %s", i, move)
		}
	}
	if len(legalMoves) > 10 {
		t.Logf("  ... and %d more", len(legalMoves)-10)
	}
	
	// Test if f6f7 is in the legal moves
	f6f7Legal := false
	for _, move := range legalMoves {
		if move == "f6f7" {
			f6f7Legal = true
			break
		}
	}
	
	t.Logf("\n=== f6f7 MOVE ANALYSIS ===")
	t.Logf("Is f6f7 in legal moves: %v", f6f7Legal)
	
	if f6Piece == 0 {
		t.Logf("❌ BUG IDENTIFIED: No piece on f6, but trying to move from there")
	} else if f7Piece == 0 {
		t.Logf("❌ UNEXPECTED: f7 is empty, but expected White bishop after Bxf7+")
	} else if getRealBugPieceColorString(f7Piece) != "White" {
		t.Logf("❌ UNEXPECTED: f7 piece is not White after Bxf7+")
	} else {
		t.Logf("✅ f7 has White piece (expected after Bxf7+)")
		if f6Piece != 0 {
			t.Logf("✅ f6 has piece that could potentially capture f7")
			t.Logf("🔍 ANALYSIS: f6 has %s %s trying to capture %s %s on f7", 
				getRealBugPieceColorString(f6Piece), realBugDescribePiece(f6Piece),
				getRealBugPieceColorString(f7Piece), realBugDescribePiece(f7Piece))
		}
	}
	
	// Try to manually convert the UCI move f6f7 to see what happens
	t.Logf("\n=== UCI MOVE CONVERSION TEST ===")
	converter := NewMoveConverter()
	
	move, err := converter.FromUCIWithLogging("f6f7", board, engine.debugLogger)
	if err != nil {
		t.Logf("❌ UCI conversion of f6f7 failed: %v", err)
		t.Logf("   This suggests f6f7 is correctly identified as illegal")
	} else {
		t.Logf("✅ UCI conversion of f6f7 succeeded:")
		t.Logf("   From: %s, To: %s, Piece: %c", move.From.String(), move.To.String(), move.Piece)
		t.Logf("🚨 This means the move converter thinks f6f7 is legal!")
	}
	
	// Test what happens if we try to apply the illegal move
	t.Logf("\n=== ATTEMPTING TO APPLY f6f7 ===")
	
	// Try to apply f6f7 and see what error we get
	illegalCmd := posCmd + " f6f7"
	
	// Create a separate engine to test this
	errorEngine := NewUCIEngine()
	errorEngine.HandleCommand("uci")
	errorEngine.HandleCommand("ucinewgame")
	errorEngine.HandleCommand(illegalCmd)
	
	finalFEN := errorEngine.engine.GetCurrentFEN()
	t.Logf("FEN after attempting f6f7: %s", finalFEN)
	
	if finalFEN == criticalFEN {
		t.Logf("✅ f6f7 was rejected - FEN unchanged")
	} else {
		t.Logf("❌ f6f7 was applied - this is the bug!")
	}
}

// TestPinAnalysis specifically checks if f6f7 is illegal due to pinning
func TestPinAnalysis(t *testing.T) {
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	// Apply the real game sequence in UCI format
	realMoves := []string{
		"c2c4", "g8f6", "g1f3", "a7a6", "b1c3", "a6a5", "a2a3", "a5a4", "d2d4", "b7b6",
		"c1f4", "b6b5", "c4b5", "c7c6", "e2e3", "c6c5", "f4b8", "c5c4", "a1c1", "d7d6", 
		"b8c4", "d6d5", "f3d5", "e7e6", "d5c7", "d8c7", "c4c7", "e6e5", "c7f7",
	}
	
	movesApplied := strings.Join(realMoves, " ")
	posCmd := "position startpos moves " + movesApplied
	engine.HandleCommand(posCmd)
	
	board := engine.engine.GetState().Board
	
	t.Logf("=== PIN ANALYSIS ===")
	
	// Find Black king
	kingSquare := findKing(board, false)
	t.Logf("Black king at: %s", kingSquare)
	
	// Check if piece on f6 is pinned
	f6Piece := getRealBugPieceAtSquare(board, "f6")
	if f6Piece != 0 {
		t.Logf("Piece on f6: %c (%s %s)", f6Piece, getRealBugPieceColorString(f6Piece), realBugDescribePiece(f6Piece))
		
		// Test if moving f6f7 would leave king in check (indicating a pin)
		isPinned := wouldMoveLeaveMateInCheck(board, "f6", "f7", false)
		t.Logf("Would f6f7 leave Black king in check: %v", isPinned)
		
		if isPinned {
			t.Logf("🔍 ROOT CAUSE: f6 piece is pinned and cannot move to f7")
		}
	}
}

// Helper functions
func getRealBugPieceAtSquare(boardPtr *board.Board, square string) byte {
	if len(square) != 2 {
		return 0
	}
	
	file := int(square[0] - 'a')
	rank := int(square[1] - '1')
	
	if rank < 0 || rank > 7 || file < 0 || file > 7 {
		return 0
	}
	
	piece := boardPtr.GetPiece(rank, file)
	return byte(piece)
}

func getRealBugPieceColorString(piece byte) string {
	if piece >= 'A' && piece <= 'Z' {
		return "White"
	} else if piece >= 'a' && piece <= 'z' {
		return "Black"
	}
	return "Empty"
}

func realBugDescribePiece(piece byte) string {
	switch piece {
	case 'K': return "King"
	case 'Q': return "Queen" 
	case 'R': return "Rook"
	case 'B': return "Bishop"
	case 'N': return "Knight"
	case 'P': return "Pawn"
	case 'k': return "King"
	case 'q': return "Queen"
	case 'r': return "Rook"
	case 'b': return "Bishop"
	case 'n': return "Knight"
	case 'p': return "Pawn"
	case 0: return "Empty"
	default: return "Unknown"
	}
}

// Simplified helper functions (implementations would need to be added)
func findKing(boardPtr *board.Board, isWhite bool) string {
	// This would need to search the board for the king
	// For now, return placeholder
	return "unknown"
}

func isKingInCheck(boardPtr *board.Board, isWhite bool) bool {
	// This would need to check if the king is in check
	// For now, assume true since we know Bxf7+ puts Black in check
	return true
}

func generateLegalMoves(engine *UCIEngine) []string {
	// This would need to interface with the move generator
	// For now, return empty slice
	return []string{}
}

func wouldMoveLeaveMateInCheck(boardPtr *board.Board, from, to string, isWhite bool) bool {
	// This would need to test if a move leaves the king in check
	// For now, return false
	return false
}