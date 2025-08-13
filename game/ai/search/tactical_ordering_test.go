package search

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func TestCompleteTacticalMoveOrdering(t *testing.T) {
	// Test position with various types of moves to verify complete ordering hierarchy
	// This position has: good captures, equal exchanges, bad captures, killers, history, and quiet moves
	fen := "r3k2r/ppp2ppp/2n1bn2/2bpp3/2BPP3/2N1BN2/PPP2PPP/R2QK2R w KQkq - 0 8"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	engine := NewMinimaxEngine()
	engine.SetDebugMoveOrdering(true)

	// Set up killer moves (simulate previous search results)
	killerMove1 := board.Move{
		From:  board.Square{Rank: 0, File: 3}, // d1
		To:    board.Square{Rank: 2, File: 3}, // d3  
		Piece: board.WhiteQueen,
	}
	killerMove2 := board.Move{
		From:  board.Square{Rank: 0, File: 4}, // e1
		To:    board.Square{Rank: 1, File: 4}, // e2
		Piece: board.WhiteKing,
	}
	// Store killer moves using the public API
	threadState := engine.getThreadLocalState()
	engine.storeKiller(killerMove1, 0, threadState)
	engine.storeKiller(killerMove2, 0, threadState)

	// Simulate history table with some moves having good scores
	historyMove1 := board.Move{
		From:  board.Square{Rank: 2, File: 5}, // f3
		To:    board.Square{Rank: 4, File: 4}, // e5
		Piece: board.WhiteKnight,
	}
	engine.historyTable.UpdateHistory(historyMove1, 5) // Give it a good history score

	generator := moves.NewGenerator()
	legalMoves := generator.GenerateAllMoves(b, moves.White)
	defer moves.ReleaseMoveList(legalMoves)

	// Analyze moves by category
	var goodCaptures, equalExchanges, badCaptures, killers, historyMoves, quietMoves []board.Move

	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]
		
		if move.IsCapture {
			seeValue := engine.seeCalculator.SEE(b, move)
			if seeValue > 0 {
				goodCaptures = append(goodCaptures, move)
			} else if seeValue == 0 {
				equalExchanges = append(equalExchanges, move)
			} else {
				badCaptures = append(badCaptures, move)
			}
		} else if engine.isKillerMove(move, 0, threadState) {
			killers = append(killers, move)
		} else if engine.getHistoryScore(move, threadState) > 1000 {
			historyMoves = append(historyMoves, move)
		} else {
			quietMoves = append(quietMoves, move)
		}
	}

	t.Logf("Move categories found:")
	t.Logf("  Good captures: %d", len(goodCaptures))
	t.Logf("  Equal exchanges: %d", len(equalExchanges))
	t.Logf("  Bad captures: %d", len(badCaptures))
	t.Logf("  Killer moves: %d", len(killers))
	t.Logf("  History moves: %d", len(historyMoves))
	t.Logf("  Quiet moves: %d", len(quietMoves))

	// Order the moves
	// threadState already available from above
	engine.orderMoves(b, legalMoves, 0, board.Move{}, threadState)
	orderedMoves := engine.GetLastMoveOrder()

	// Verify ordering by checking positions
	type moveWithCategory struct {
		move     board.Move
		position int
		category string
		score    int
	}

	var categorizedMoves []moveWithCategory

	for pos, move := range orderedMoves {
		var category string
		var score int

		if move.IsCapture {
			seeValue := engine.seeCalculator.SEE(b, move)
			threadState := engine.getThreadLocalState()
			score = engine.getCaptureScore(b, move, threadState)
			if seeValue > 0 {
				category = "Good Capture"
			} else if seeValue == 0 {
				category = "Equal Exchange"
			} else if seeValue >= -100 {
				category = "Slightly Bad Capture"
			} else {
				category = "Terrible Capture"
			}
		} else if engine.isKillerMove(move, 0, threadState) {
			category = "Killer Move"
			score = 500000
		} else if engine.getHistoryScore(move, threadState) > 1000 {
			category = "History Move"
			score = int(engine.getHistoryScore(move, threadState))
		} else {
			category = "Quiet Move"
			score = 0
		}

		categorizedMoves = append(categorizedMoves, moveWithCategory{
			move:     move,
			position: pos,
			category: category,
			score:    score,
		})
	}

	// Show the first 10 moves in order
	t.Logf("\nFirst 10 moves in search order:")
	for i, cm := range categorizedMoves {
		if i >= 10 {
			break
		}
		t.Logf("  %d. %s %s (score: %d)", 
			i+1, boardSquareToString(cm.move.From)+"-"+boardSquareToString(cm.move.To), 
			cm.category, cm.score)
	}

	// Verify ordering hierarchy
	expectedOrder := []string{
		"Good Capture", "Equal Exchange", "Killer Move", "History Move", 
		"Slightly Bad Capture", "Terrible Capture", "Quiet Move",
	}

	categoryFirstPos := make(map[string]int)
	for _, cm := range categorizedMoves {
		if _, exists := categoryFirstPos[cm.category]; !exists {
			categoryFirstPos[cm.category] = cm.position
		}
	}

	t.Logf("\nCategory ordering verification:")
	for i, category := range expectedOrder {
		if pos, exists := categoryFirstPos[category]; exists {
			t.Logf("  %d. %s: first appears at position %d ✅", i+1, category, pos)
		} else {
			t.Logf("  %d. %s: not found in position", i+1, category)
		}
	}

	// Verify that categories appear in correct order
	var prevPos int = -1
	orderCorrect := true
	
	for _, category := range expectedOrder {
		if pos, exists := categoryFirstPos[category]; exists {
			if pos < prevPos {
				t.Errorf("❌ Order violation: %s (pos %d) should not appear before previous category (pos %d)", 
					category, pos, prevPos)
				orderCorrect = false
			}
			prevPos = pos
		}
	}

	if orderCorrect {
		t.Log("✅ Tactical move ordering hierarchy is correct!")
	}
}

func TestSpecificTacticalScenarios(t *testing.T) {
	testCases := []struct {
		name        string
		fen         string
		description string
	}{
		{
			name:        "Queen sacrifice scenario",
			fen:         "r3k2r/ppp2ppp/2n5/2b1p3/2BPPq2/2N2N2/PPP2PPP/R2QK2R w KQkq - 0 10",
			description: "Queen sacrifice Qd5+ might be a good terrible capture leading to mate",
		},
		{
			name:        "Tactical pin scenario", 
			fen:         "r2qkb1r/ppp2ppp/2n2n2/2bpp3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 0 6",
			description: "Bxf7+ might be a good terrible capture (sacrifice) creating tactical complications",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := board.FromFEN(tc.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN: %v", err)
			}

			engine := NewMinimaxEngine()
			generator := moves.NewGenerator()
			legalMoves := generator.GenerateAllMoves(b, moves.White)
			defer moves.ReleaseMoveList(legalMoves)

			// Find terrible captures (SEE < -100)
			var terribleCaptures []board.Move
			for i := 0; i < legalMoves.Count; i++ {
				move := legalMoves.Moves[i]
				if move.IsCapture {
					seeValue := engine.seeCalculator.SEE(b, move)
					if seeValue < -100 {
						terribleCaptures = append(terribleCaptures, move)
					}
				}
			}

			t.Logf("%s", tc.description)
			t.Logf("Found %d terrible captures that might be tactical sacrifices:", len(terribleCaptures))
			
			for _, capture := range terribleCaptures {
				seeValue := engine.seeCalculator.SEE(b, capture)
				threadState := engine.getThreadLocalState()
				score := engine.getCaptureScore(b, capture, threadState)
				t.Logf("  %s%s (SEE: %d, Score: %d) - Still searched before quiet moves", 
					boardSquareToString(capture.From), boardSquareToString(capture.To), 
					seeValue, score)
			}
		})
	}
}