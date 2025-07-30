package uci

import (
	"strings"
	"testing"
	
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// TestPositionTrace traces each move to find where the bug occurs
func TestPositionTrace(t *testing.T) {
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	t.Logf("=== POSITION TRACE TEST ===")
	
	// All moves in the sequence
	moves := []string{
		"c2c4", "g8f6", "g1f3", "a7a6", "b1c3", "a6a5", "a2a3", "a5a4", "d2d4", "b7b6",
		"c1f4", "b6b5", "c4b5", "c7c6", "e2e3", "c6c5", "f4b8", "c5c4", "a1c1", "d7d6", 
		"f1c4", "d6d5", "c3d5", "e7e6", "d5c7", // Add the problematic move
	}
	
	// Track critical pieces throughout the game
	criticalSquares := []string{"c3", "d5", "f6", "c7", "g8"}
	
	t.Logf("Starting position:")
	internalBoard := engine.engine.GetState().Board
	for _, square := range criticalSquares {
		piece := getTracePieceAtSquare(internalBoard, square)
		if piece != 0 {
			t.Logf("  %s: %c (%s %s)", square, piece, getTracePieceColorString(piece), traceDescribePiece(piece))
		}
	}
	
	// Apply moves one by one and track the critical pieces
	for i, move := range moves {
		t.Logf("\n--- Move %d: %s ---", i+1, move)
		
		// Build position command
		movesApplied := strings.Join(moves[:i+1], " ")
		posCmd := "position startpos moves " + movesApplied
		
		// Create fresh engine for this position
		testEngine := NewUCIEngine()
		testEngine.HandleCommand("uci")
		testEngine.HandleCommand("ucinewgame")
		testEngine.HandleCommand(posCmd)
		
		currentFEN := testEngine.engine.GetCurrentFEN()
		t.Logf("FEN after move: %s", currentFEN)
		
		// Check critical squares
		board := testEngine.engine.GetState().Board
		for _, square := range criticalSquares {
			piece := getTracePieceAtSquare(board, square)
			if piece != 0 {
				t.Logf("  %s: %c (%s %s)", square, piece, getTracePieceColorString(piece), traceDescribePiece(piece))
			}
		}
		
		// Special analysis for key moves
		if move == "c3d5" {
			t.Logf("🔍 CRITICAL: Knight should move from c3 to d5")
			c3Before := getTracePieceAtSquare(board, "c3")
			d5After := getTracePieceAtSquare(board, "d5")
			t.Logf("   c3 after move: %c (%s)", c3Before, traceDescribePiece(c3Before))
			t.Logf("   d5 after move: %c (%s)", d5After, traceDescribePiece(d5After))
			
			if d5After == 'N' {
				t.Logf("   ✅ WHITE knight correctly moved to d5")
			} else {
				t.Logf("   ❌ Expected WHITE knight on d5, got: %c", d5After)
			}
		}
		
		if move == "d5c7" {
			t.Logf("🚨 PROBLEMATIC MOVE: d5c7")
			d5Before := getTracePieceAtSquare(board, "d5")
			f6Before := getTracePieceAtSquare(board, "f6") 
			c7After := getTracePieceAtSquare(board, "c7")
			
			t.Logf("   d5 piece: %c (%s %s)", d5Before, getTracePieceColorString(d5Before), traceDescribePiece(d5Before))
			t.Logf("   f6 piece: %c (%s %s)", f6Before, getTracePieceColorString(f6Before), traceDescribePiece(f6Before))
			t.Logf("   c7 after: %c (%s %s)", c7After, getTracePieceColorString(c7After), traceDescribePiece(c7After))
			
			// This is the key question: is there actually a piece on d5 to move?
			if d5Before != 0 {
				t.Logf("   ✅ There IS a piece on d5 to move")
				if getTracePieceColorString(d5Before) == "White" {
					t.Logf("   🔍 It's a WHITE piece - this explains the behavior")
				} else {
					t.Logf("   🔍 It's a BLACK piece")
				}
			} else {
				t.Logf("   ❌ NO piece on d5 - this would be illegal!")
			}
		}
	}
}

// TestMoveConversionIsolated tests the UCI move conversion in isolation
func TestMoveConversionIsolated(t *testing.T) {
	t.Logf("=== UCI MOVE CONVERSION TEST ===")
	
	// Create the exact position before d5c7
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	moves := []string{
		"c2c4", "g8f6", "g1f3", "a7a6", "b1c3", "a6a5", "a2a3", "a5a4", "d2d4", "b7b6",
		"c1f4", "b6b5", "c4b5", "c7c6", "e2e3", "c6c5", "f4b8", "c5c4", "a1c1", "d7d6", 
		"f1c4", "d6d5", "c3d5", "e7e6",
	}
	
	movesApplied := strings.Join(moves, " ")
	posCmd := "position startpos moves " + movesApplied
	engine.HandleCommand(posCmd)
	
	board := engine.engine.GetState().Board
	
	t.Logf("Position before d5c7 move:")
	fen := engine.engine.GetCurrentFEN()
	t.Logf("FEN: %s", fen)
	
	// Test various UCI move conversions
	testMoves := []string{"d5c7", "f6d5", "f6e4", "f6h5"}
	
	converter := NewMoveConverter()
	
	for _, uciMove := range testMoves {
		t.Logf("\nTesting UCI move: %s", uciMove)
		
		move, err := converter.FromUCIWithLogging(uciMove, board, engine.debugLogger)
		if err != nil {
			t.Logf("  ❌ Conversion failed: %v", err)
			continue
		}
		
		t.Logf("  ✅ Conversion successful:")
		t.Logf("    From: %s", move.From.String())  
		t.Logf("    To: %s", move.To.String())
		t.Logf("    Piece: %c (%s %s)", move.Piece, getTracePieceColorString(byte(move.Piece)), traceDescribePiece(byte(move.Piece)))
		t.Logf("    Captured: %c", move.Captured)
		
		// Verify the piece exists at the from square
		fromPiece := getTracePieceAtSquare(board, move.From.String())
		if fromPiece == byte(move.Piece) {
			t.Logf("    ✅ Piece matches board state")
		} else {
			t.Logf("    ❌ Piece mismatch! Board has %c, move claims %c", fromPiece, move.Piece)
		}
	}
}

// TestExpectedVsActualPosition compares our position with what the debug logs suggest
func TestExpectedVsActualPosition(t *testing.T) {
	t.Logf("=== EXPECTED VS ACTUAL POSITION TEST ===")
	
	// From the original debug logs, the position before d5c7 should be:
	// "r1b1kb1r/2B2Bpp/5q2/1P2p3/p2P4/P3PN2/1P3PPP/2RQK2R b Kkq - 0 15"
	expectedFEN := "r1b1kb1r/2B2Bpp/5q2/1P2p3/p2P4/P3PN2/1P3PPP/2RQK2R b Kkq - 0 15"
	
	// Build our position
	engine := NewUCIEngine()
	engine.HandleCommand("uci")
	engine.HandleCommand("ucinewgame")
	
	moves := []string{
		"c2c4", "g8f6", "g1f3", "a7a6", "b1c3", "a6a5", "a2a3", "a5a4", "d2d4", "b7b6",
		"c1f4", "b6b5", "c4b5", "c7c6", "e2e3", "c6c5", "f4b8", "c5c4", "a1c1", "d7d6", 
		"f1c4", "d6d5", "c3d5", "e7e6",
	}
	
	movesApplied := strings.Join(moves, " ")
	posCmd := "position startpos moves " + movesApplied
	engine.HandleCommand(posCmd)
	
	actualFEN := engine.engine.GetCurrentFEN()
	
	t.Logf("Expected FEN: %s", expectedFEN)
	t.Logf("Actual FEN:   %s", actualFEN)
	
	if expectedFEN == actualFEN {
		t.Logf("✅ FEN positions match exactly!")
	} else {
		t.Logf("❌ FEN positions differ!")
		
		// Break down the differences
		t.Logf("\nAnalyzing differences:")
		
		// Parse expected position manually and compare key squares
		keySquares := []string{"d5", "f6", "c7", "b8", "f7"}
		
		for _, square := range keySquares {
			actualPiece := getTracePieceAtSquare(engine.engine.GetState().Board, square)
			t.Logf("  %s: actual=%c (%s)", square, actualPiece, traceDescribePiece(actualPiece))
			
			// From the expected FEN "r1b1kb1r/2B2Bpp/5q2/1P2p3/p2P4/P3PN2/1P3PPP/2RQK2R"
			// f6 should have 'q' (queen), but we expect it to have 'n' (knight)
			if square == "f6" {
				t.Logf("    Expected from debug logs: queen ('q')")
				t.Logf("    Our expectation: knight ('n')")
				if actualPiece == 'n' {
					t.Logf("    🔍 Our position has knight - this might be correct")
				} else if actualPiece == 'q' {
					t.Logf("    🔍 Our position matches debug logs exactly")
				}
			}
		}
	}
}

// Helper functions with unique names
func getTracePieceAtSquare(boardPtr *board.Board, square string) byte {
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

func getTracePieceColorString(piece byte) string {
	if piece >= 'A' && piece <= 'Z' {
		return "White"
	} else if piece >= 'a' && piece <= 'z' {
		return "Black"
	}
	return "Empty"
}

func traceDescribePiece(piece byte) string {
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