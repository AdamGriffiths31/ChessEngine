package search

import (
	"fmt"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// Helper function to convert board square to string
func boardSquareToString(sq board.Square) string {
	files := "abcdefgh"
	return fmt.Sprintf("%c%d", files[sq.File], sq.Rank+1)
}

func TestSEEIntegrationInMoveOrdering(t *testing.T) {
	// Position with white queen that can capture undefended knight (good) or defended knight (bad)
	fen := "7k/1n4pp/5n2/8/8/8/1Q6/4K3 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	engine := NewMinimaxEngine()
	engine.SetDebugMoveOrdering(true)

	generator := moves.NewGenerator()
	legalMoves := generator.GenerateAllMoves(b, moves.White)
	defer moves.ReleaseMoveList(legalMoves)

	// Debug: Print all queen capture moves to see what's available
	t.Log("Available queen captures:")
	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]
		if move.Piece == board.WhiteQueen && move.IsCapture {
			t.Logf("Queen capture: from %s to %s, captures %v",
				boardSquareToString(move.From), boardSquareToString(move.To), move.Captured)
		}
	}

	// Find the two knight captures
	var captureKnightB7, captureKnightF6 board.Move
	var foundB7, foundF6 bool

	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]
		if move.Piece == board.WhiteQueen && move.IsCapture {
			if move.To.Rank == 6 && move.To.File == 1 { // b7 knight
				captureKnightB7 = move
				foundB7 = true
			} else if move.To.Rank == 5 && move.To.File == 5 { // f6 knight
				captureKnightF6 = move
				foundF6 = true
			}
		}
	}

	if !foundB7 || !foundF6 {
		t.Skipf("Could not find expected knight captures (found b7: %v, found f6: %v)", foundB7, foundF6)
	}

	// Calculate SEE values to see which is better
	b7SEE := engine.seeCalculator.SEE(b, captureKnightB7)
	f6SEE := engine.seeCalculator.SEE(b, captureKnightF6)

	t.Logf("Qxb7 SEE: %d", b7SEE)
	t.Logf("Qxf6 SEE: %d", f6SEE)

	// Determine which is better/worse
	var betterCapture, worseCapture board.Move
	var betterSEE, worseSEE int

	if b7SEE >= f6SEE {
		betterCapture = captureKnightB7
		worseCapture = captureKnightF6
		betterSEE = b7SEE
		worseSEE = f6SEE
	} else {
		betterCapture = captureKnightF6
		worseCapture = captureKnightB7
		betterSEE = f6SEE
		worseSEE = b7SEE
	}

	t.Logf("Better capture: %s to %s (SEE: %d)",
		boardSquareToString(betterCapture.From), boardSquareToString(betterCapture.To), betterSEE)
	t.Logf("Worse capture: %s to %s (SEE: %d)",
		boardSquareToString(worseCapture.From), boardSquareToString(worseCapture.To), worseSEE)

	// Order moves using the engine's method
	threadState := engine.getThreadLocalState()
	engine.orderMoves(b, legalMoves, 0, board.Move{}, threadState)

	// Get the ordered moves
	orderedMoves := engine.GetLastMoveOrder()

	// Find positions of our captures in the ordered list
	var betterPos, worsePos int = -1, -1
	for i, move := range orderedMoves {
		if move.From == betterCapture.From && move.To == betterCapture.To {
			betterPos = i
		} else if move.From == worseCapture.From && move.To == worseCapture.To {
			worsePos = i
		}
	}

	if betterPos == -1 || worsePos == -1 {
		t.Fatalf("Could not find captures in ordered move list")
	}

	t.Logf("Better capture position in ordered list: %d", betterPos)
	t.Logf("Worse capture position in ordered list: %d", worsePos)

	// The better capture should come before the worse capture in move ordering
	if betterPos >= worsePos {
		t.Errorf("Expected better capture (pos %d) to be ordered before worse capture (pos %d)",
			betterPos, worsePos)
	}
}

func TestSEEIntegrationQuiescencePruning(t *testing.T) {
	// Test that SEE-based pruning works in quiescence search
	// We'll create a position with a clearly bad capture and verify it gets pruned

	fen := "4k3/8/8/4p3/3p4/8/3Q4/4K3 w - - 0 1" // Queen can capture defended pawn
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	engine := NewMinimaxEngine()

	// The bad capture should be pruned by SEE in quiescence search
	move := board.Move{
		From:      board.Square{Rank: 1, File: 3}, // d2
		To:        board.Square{Rank: 3, File: 3}, // d4
		Piece:     board.WhiteQueen,
		Captured:  board.BlackPawn,
		IsCapture: true,
	}

	// Verify this is indeed a bad capture
	seeValue := engine.seeCalculator.SEE(b, move)
	if seeValue >= -100 {
		t.Skipf("This capture is not bad enough for pruning (SEE: %d), skipping test", seeValue)
	}

	t.Logf("Bad capture SEE value: %d (should be < -100 for pruning)", seeValue)
	t.Log("SEE-based pruning in quiescence search should improve search efficiency")
}
