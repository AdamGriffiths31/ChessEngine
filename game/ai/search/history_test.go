package search

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestNewHistoryTable(t *testing.T) {
	h := NewHistoryTable()
	if h == nil {
		t.Fatal("NewHistoryTable should not return nil")
	}

	// Test that initial history scores are zero
	move := board.Move{
		From: board.Square{File: 0, Rank: 0},
		To:   board.Square{File: 1, Rank: 1},
	}
	score := h.GetHistoryScore(move)
	if score != 0 {
		t.Errorf("Expected initial history score to be 0, got %d", score)
	}
}

func TestUpdateHistory(t *testing.T) {
	h := NewHistoryTable()
	move := board.Move{
		From: board.Square{File: 0, Rank: 0},
		To:   board.Square{File: 1, Rank: 1},
	}

	// Update history for depth 1
	h.UpdateHistory(move, 1)
	score := h.GetHistoryScore(move)
	expectedScore := HistoryBonus * (1 + 1) // depth + 1
	if score != int32(expectedScore) {
		t.Errorf("Expected history score to be %d after one update, got %d", expectedScore, score)
	}

	// Update again at depth 2
	h.UpdateHistory(move, 2)
	newScore := h.GetHistoryScore(move)
	expectedNewScore := int32(expectedScore) + (HistoryBonus * (2 + 1))
	if newScore != expectedNewScore {
		t.Errorf("Expected history score to be %d after second update, got %d", expectedNewScore, newScore)
	}
}

func TestUpdateHistoryInvalidSquares(t *testing.T) {
	h := NewHistoryTable()

	// Test invalid from square
	invalidMove := board.Move{
		From: board.Square{File: -1, Rank: 0},
		To:   board.Square{File: 1, Rank: 1},
	}
	h.UpdateHistory(invalidMove, 1)
	score := h.GetHistoryScore(invalidMove)
	if score != 0 {
		t.Errorf("Expected history score to remain 0 for invalid move, got %d", score)
	}

	// Test invalid to square
	invalidMove2 := board.Move{
		From: board.Square{File: 0, Rank: 0},
		To:   board.Square{File: 8, Rank: 1}, // File 8 is out of bounds
	}
	h.UpdateHistory(invalidMove2, 1)
	score = h.GetHistoryScore(invalidMove2)
	if score != 0 {
		t.Errorf("Expected history score to remain 0 for invalid move, got %d", score)
	}
}

func TestHistoryMaxScore(t *testing.T) {
	h := NewHistoryTable()
	move := board.Move{
		From: board.Square{File: 0, Rank: 0},
		To:   board.Square{File: 1, Rank: 1},
	}

	// Update many times to try to exceed max score
	for i := 0; i < 1000; i++ {
		h.UpdateHistory(move, 10) // High depth for large bonuses
	}

	score := h.GetHistoryScore(move)
	if score > MaxHistoryScore {
		t.Errorf("Expected history score to be capped at %d, got %d", MaxHistoryScore, score)
	}
	if score != MaxHistoryScore {
		t.Errorf("Expected history score to reach maximum %d, got %d", MaxHistoryScore, score)
	}
}

func TestHistoryClear(t *testing.T) {
	h := NewHistoryTable()
	move1 := board.Move{
		From: board.Square{File: 0, Rank: 0},
		To:   board.Square{File: 1, Rank: 1},
	}
	move2 := board.Move{
		From: board.Square{File: 2, Rank: 2},
		To:   board.Square{File: 3, Rank: 3},
	}

	// Update history for both moves
	h.UpdateHistory(move1, 1)
	h.UpdateHistory(move2, 2)

	// Verify scores are non-zero
	if h.GetHistoryScore(move1) == 0 {
		t.Error("Expected non-zero score for move1 before clear")
	}
	if h.GetHistoryScore(move2) == 0 {
		t.Error("Expected non-zero score for move2 before clear")
	}

	// Clear and verify scores are zero
	h.Clear()
	if h.GetHistoryScore(move1) != 0 {
		t.Error("Expected zero score for move1 after clear")
	}
	if h.GetHistoryScore(move2) != 0 {
		t.Error("Expected zero score for move2 after clear")
	}
}

func TestHistoryAge(t *testing.T) {
	h := NewHistoryTable()
	move := board.Move{
		From: board.Square{File: 0, Rank: 0},
		To:   board.Square{File: 1, Rank: 1},
	}

	// Update to get a good score
	h.UpdateHistory(move, 5)
	originalScore := h.GetHistoryScore(move)

	// Age several times, but not enough to trigger decay
	for i := 0; i < 7; i++ {
		h.Age()
	}
	scoreAfterAging := h.GetHistoryScore(move)
	if scoreAfterAging != originalScore {
		t.Errorf("Expected score to remain %d after aging (no decay yet), got %d", originalScore, scoreAfterAging)
	}

	// Age one more time to trigger decay (8th age, multiple of 8)
	h.Age()
	scoreAfterDecay := h.GetHistoryScore(move)
	expectedAfterDecay := originalScore / HistoryDecayFactor
	if scoreAfterDecay != expectedAfterDecay {
		t.Errorf("Expected score to be %d after decay, got %d", expectedAfterDecay, scoreAfterDecay)
	}
}

func TestSquareToIndex(t *testing.T) {
	tests := []struct {
		square   board.Square
		expected int
	}{
		{board.Square{File: 0, Rank: 0}, 0},  // a1
		{board.Square{File: 7, Rank: 0}, 7},  // h1
		{board.Square{File: 0, Rank: 1}, 8},  // a2
		{board.Square{File: 7, Rank: 7}, 63}, // h8
		{board.Square{File: 4, Rank: 3}, 28}, // e4
	}

	for _, test := range tests {
		result := squareToIndex(test.square)
		if result != test.expected {
			t.Errorf("squareToIndex(%+v) = %d, expected %d", test.square, result, test.expected)
		}
	}
}
