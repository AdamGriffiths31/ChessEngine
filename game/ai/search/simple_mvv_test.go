package search

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/evaluation"
)

func TestMVVLVATiebreakerSimple(t *testing.T) {
	// Create a simple position to test MVV-LVA tiebreaker
	fen := "4k3/8/8/3qr3/8/8/8/4K2R w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	engine := NewMinimaxEngine()

	// Create two capture moves from the same piece (rook) to different victims
	captureQueen := board.Move{
		From:      board.Square{Rank: 0, File: 7}, // h1
		To:        board.Square{Rank: 4, File: 3}, // d5 
		Piece:     board.WhiteRook,
		Captured:  board.BlackQueen,
		IsCapture: true,
	}

	captureRook := board.Move{
		From:      board.Square{Rank: 0, File: 7}, // h1
		To:        board.Square{Rank: 4, File: 4}, // e5
		Piece:     board.WhiteRook,
		Captured:  board.BlackRook,
		IsCapture: true,
	}

	// Calculate SEE and scores
	queenSEE := engine.seeCalculator.SEE(b, captureQueen)
	rookSEE := engine.seeCalculator.SEE(b, captureRook)
	queenScore := engine.getCaptureScore(b, captureQueen)
	rookScore := engine.getCaptureScore(b, captureRook)

	t.Logf("Rook captures queen (Rxd5): SEE=%d, Score=%d", queenSEE, queenScore)
	t.Logf("Rook captures rook (Rxe5): SEE=%d, Score=%d", rookSEE, rookScore)

	// Both should be winning captures, but queen capture should score higher due to MVV-LVA
	if queenSEE <= 0 || rookSEE <= 0 {
		t.Skip("Test requires both captures to be winning (positive SEE)")
	}

	if queenScore <= rookScore {
		t.Errorf("❌ Queen capture (score %d) should score higher than rook capture (score %d)", 
			queenScore, rookScore)
	} else {
		t.Logf("✅ MVV-LVA tiebreaker working correctly")
	}

	// Calculate expected difference from victim values
	queenValue := evaluation.PieceValues[board.BlackQueen]
	if queenValue < 0 {
		queenValue = -queenValue
	}
	rookValue := evaluation.PieceValues[board.BlackRook]
	if rookValue < 0 {
		rookValue = -rookValue
	}

	expectedDifference := (queenValue - rookValue) / 100 // Divided by 100 as in our scoring function
	actualDifference := queenScore - rookScore

	t.Logf("Victim values: Queen=%d, Rook=%d", queenValue, rookValue)
	t.Logf("Score difference: %d (expected ~%d from MVV-LVA)", actualDifference, expectedDifference)
	
	// The difference should include the MVV-LVA component
	if actualDifference < expectedDifference {
		t.Errorf("❌ Score difference (%d) should include MVV-LVA component (%d)", 
			actualDifference, expectedDifference)
	}
}

func TestMVVLVAInEqualExchanges(t *testing.T) {
	// Test position where we have equal exchanges with different victim values
	fen := "4k3/8/8/3qr3/8/8/3QR3/4K3 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	engine := NewMinimaxEngine()

	// Queen takes queen (should be SEE = 0)
	queenTakesQueen := board.Move{
		From:      board.Square{Rank: 1, File: 3}, // d2
		To:        board.Square{Rank: 4, File: 3}, // d5
		Piece:     board.WhiteQueen,
		Captured:  board.BlackQueen,
		IsCapture: true,
	}

	// Rook takes rook (should be SEE = 0)  
	rookTakesRook := board.Move{
		From:      board.Square{Rank: 1, File: 4}, // e2
		To:        board.Square{Rank: 4, File: 4}, // e5
		Piece:     board.WhiteRook,
		Captured:  board.BlackRook,
		IsCapture: true,
	}

	queenSEE := engine.seeCalculator.SEE(b, queenTakesQueen)
	rookSEE := engine.seeCalculator.SEE(b, rookTakesRook)
	queenScore := engine.getCaptureScore(b, queenTakesQueen)
	rookScore := engine.getCaptureScore(b, rookTakesRook)

	t.Logf("Queen takes queen (Qxd5): SEE=%d, Score=%d", queenSEE, queenScore)
	t.Logf("Rook takes rook (Rxe5): SEE=%d, Score=%d", rookSEE, rookScore)

	// Both should be equal exchanges (SEE = 0), but queen exchange should score higher
	if queenSEE != 0 || rookSEE != 0 {
		t.Logf("Warning: Expected both exchanges to have SEE=0, got queen=%d, rook=%d", queenSEE, rookSEE)
		// Continue with test anyway to see tiebreaker effect
	}

	if queenScore <= rookScore {
		t.Errorf("❌ Queen exchange (score %d) should score higher than rook exchange (score %d) due to MVV-LVA", 
			queenScore, rookScore)
	} else {
		t.Logf("✅ MVV-LVA correctly prioritizes higher-value equal exchanges")
	}

	// For equal exchanges, the tiebreaker should be victim_value / 10
	expectedQueenBonus := 900 / 10  // Queen victim value / 10
	expectedRookBonus := 500 / 10   // Rook victim value / 10
	expectedDifference := expectedQueenBonus - expectedRookBonus

	actualDifference := queenScore - rookScore
	t.Logf("Expected MVV-LVA difference: %d, Actual: %d", expectedDifference, actualDifference)
}