package search

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func TestMVVLVATiebreaker(t *testing.T) {
	testCases := []struct {
		name        string
		description string
		fen         string
		expectedOrder []string // List of expected captures in order (from-to format)
	}{
		{
			name:        "Equal exchange tiebreaker",  
			description: "When multiple equal exchanges (SEE=0), prefer capturing higher value piece",
			fen:         "4k3/8/8/3qr3/8/8/3QR3/4K3 w - - 0 1", // Qxd5, Rxe5 both SEE=0
			expectedOrder: []string{"d2d5", "e2e5"}, // Queen capture (SEE=0) before rook capture (SEE=0)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := board.FromFEN(tc.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			engine := NewMinimaxEngine()
			engine.SetDebugMoveOrdering(true)

			player := moves.White
			if b.GetSideToMove() == "b" {
				player = moves.Black
			}

			generator := moves.NewGenerator()
			legalMoves := generator.GenerateAllMoves(b, player)
			defer moves.ReleaseMoveList(legalMoves)

			t.Logf("%s", tc.description)

			// Find the captures we're testing
			var testCaptures []board.Move
			for i := 0; i < legalMoves.Count; i++ {
				move := legalMoves.Moves[i]
				if move.IsCapture {
					moveStr := boardSquareToString(move.From) + boardSquareToString(move.To)
					for _, expected := range tc.expectedOrder {
						if moveStr == expected {
							testCaptures = append(testCaptures, move)
							
							seeValue := engine.seeCalculator.SEE(b, move)
							threadState := engine.getThreadLocalState()
					score := engine.getCaptureScore(b, move, threadState)
							victimValue := move.Captured
							
							t.Logf("  %s: SEE=%d, Score=%d, Victim=%s", 
								moveStr, seeValue, score, string(victimValue))
							break
						}
					}
				}
			}

			if len(testCaptures) != len(tc.expectedOrder) {
				t.Fatalf("Expected %d test captures, found %d", len(tc.expectedOrder), len(testCaptures))
			}

			// Order moves
			threadState := engine.getThreadLocalState()
			engine.orderMoves(b, legalMoves, 0, board.Move{}, threadState)
			orderedMoves := engine.GetLastMoveOrder()

			// Find positions of our test captures
			capturePositions := make(map[string]int)
			for pos, move := range orderedMoves {
				if move.IsCapture {
					moveStr := boardSquareToString(move.From) + boardSquareToString(move.To)
					capturePositions[moveStr] = pos
				}
			}

			t.Logf("  Actual order in search:")
			for i, expected := range tc.expectedOrder {
				if pos, exists := capturePositions[expected]; exists {
					t.Logf("    %d. %s at position %d", i+1, expected, pos)
				}
			}

			// Verify order
			for i := 0; i < len(tc.expectedOrder)-1; i++ {
				current := tc.expectedOrder[i]
				next := tc.expectedOrder[i+1]
				
				currentPos, currentExists := capturePositions[current]
				nextPos, nextExists := capturePositions[next]
				
				if !currentExists || !nextExists {
					t.Errorf("Missing expected captures in ordered moves")
					continue
				}
				
				if currentPos >= nextPos {
					t.Errorf("❌ MVV-LVA tiebreaker failed: %s (pos %d) should come before %s (pos %d)", 
						current, currentPos, next, nextPos)
				} else {
					t.Logf("✅ %s correctly ordered before %s", current, next)
				}
			}
		})
	}
}

func TestSEEWithMVVLVAScoring(t *testing.T) {
	// Test that scores correctly incorporate both SEE and victim value
	fen := "4k3/8/8/3qr3/3Q4/8/8/4K3 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	engine := NewMinimaxEngine()

	// Test moves: Qxd5 (capture queen) and Qxe5 (capture rook)
	captureQueen := board.Move{
		From:      board.Square{Rank: 3, File: 3}, // d4
		To:        board.Square{Rank: 4, File: 3}, // d5
		Piece:     board.WhiteQueen,
		Captured:  board.BlackQueen,
		IsCapture: true,
	}

	captureRook := board.Move{
		From:      board.Square{Rank: 3, File: 3}, // d4
		To:        board.Square{Rank: 4, File: 4}, // e5
		Piece:     board.WhiteQueen,
		Captured:  board.BlackRook,
		IsCapture: true,
	}

	queenSEE := engine.seeCalculator.SEE(b, captureQueen)
	rookSEE := engine.seeCalculator.SEE(b, captureRook)
	threadState := engine.getThreadLocalState()
	queenScore := engine.getCaptureScore(b, captureQueen, threadState)
	rookScore := engine.getCaptureScore(b, captureRook, threadState)

	t.Logf("Qxd5 (capture queen): SEE=%d, Score=%d", queenSEE, queenScore)
	t.Logf("Qxe5 (capture rook): SEE=%d, Score=%d", rookSEE, rookScore)

	// Both should be good captures, but queen capture should have higher score due to MVV-LVA
	if queenScore <= rookScore {
		t.Errorf("❌ Queen capture (score %d) should have higher score than rook capture (score %d)", 
			queenScore, rookScore)
	} else {
		t.Logf("✅ MVV-LVA tiebreaker working: queen capture (%d) > rook capture (%d)", 
			queenScore, rookScore)
	}

	scoreDifference := queenScore - rookScore
	expectedDifference := (900 - 500) / 100 // Queen (900) vs Rook (500), divided by 100
	
	if scoreDifference != expectedDifference {
		t.Logf("Score difference: %d (expected ~%d from victim value tiebreaker)", 
			scoreDifference, expectedDifference)
	}
}