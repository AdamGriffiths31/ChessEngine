package epd

import (
	"context"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// mockEngine implements ai.Engine for testing
type mockEngine struct {
	moveToReturn  board.Move
	scoreToReturn ai.EvaluationScore
}

func (m *mockEngine) FindBestMove(_ context.Context, _ *board.Board, _ moves.Player, _ ai.SearchConfig) ai.SearchResult {
	return ai.SearchResult{
		BestMove: m.moveToReturn,
		Score:    m.scoreToReturn,
		Stats:    ai.SearchStats{},
	}
}

func (m *mockEngine) SetEvaluator(_ ai.Evaluator) {}
func (m *mockEngine) GetName() string             { return "mock" }

func TestCalculateScore(t *testing.T) {
	engine := &mockEngine{}
	config := ai.SearchConfig{
		MaxDepth: 3,
		MaxTime:  time.Second,
	}
	scorer := NewSTSScorer(engine, config, false)

	tests := []struct {
		name          string
		bestMove      string
		avoidMove     string
		engineMove    board.Move
		expectedScore int
	}{
		{
			name:          "Exact match with best move",
			bestMove:      "e2e4",
			engineMove:    board.Move{From: board.Square{File: 4, Rank: 1}, To: board.Square{File: 4, Rank: 3}},
			expectedScore: 10,
		},
		{
			name:          "No match with best move",
			bestMove:      "e2e4",
			engineMove:    board.Move{From: board.Square{File: 3, Rank: 1}, To: board.Square{File: 3, Rank: 3}},
			expectedScore: 1, // Partial credit for legal move
		},
		{
			name:          "Avoided move played",
			avoidMove:     "e2e4",
			engineMove:    board.Move{From: board.Square{File: 4, Rank: 1}, To: board.Square{File: 4, Rank: 3}},
			expectedScore: 0, // No points for avoided move
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			position := &Position{
				BestMove:  tt.bestMove,
				AvoidMove: tt.avoidMove,
			}

			engine.moveToReturn = tt.engineMove
			score := scorer.calculateScore(position, tt.engineMove, moveToString(tt.engineMove), moveToString(tt.engineMove))

			if score != tt.expectedScore {
				t.Errorf("Expected score %d, got %d", tt.expectedScore, score)
			}
		})
	}
}

func TestCastlingMoveMatches(t *testing.T) {
	engine := &mockEngine{}
	config := ai.SearchConfig{
		MaxDepth: 3,
		MaxTime:  time.Second,
	}
	scorer := NewSTSScorer(engine, config, false)

	tests := []struct {
		name         string
		expectedMove string
		engineMove   board.Move
		shouldMatch  bool
	}{
		{
			name:         "White kingside castling",
			expectedMove: "e1g1",
			engineMove:   board.Move{From: board.Square{File: 4, Rank: 0}, To: board.Square{File: 6, Rank: 0}},
			shouldMatch:  true,
		},
		{
			name:         "White queenside castling",
			expectedMove: "e1c1",
			engineMove:   board.Move{From: board.Square{File: 4, Rank: 0}, To: board.Square{File: 2, Rank: 0}},
			shouldMatch:  true,
		},
		{
			name:         "O-O notation",
			expectedMove: "O-O",
			engineMove:   board.Move{From: board.Square{File: 4, Rank: 0}, To: board.Square{File: 6, Rank: 0}},
			shouldMatch:  true,
		},
		{
			name:         "Wrong castling side",
			expectedMove: "e1g1",
			engineMove:   board.Move{From: board.Square{File: 4, Rank: 0}, To: board.Square{File: 2, Rank: 0}},
			shouldMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := scorer.moveMatches(tt.engineMove, tt.expectedMove)
			if matches != tt.shouldMatch {
				t.Errorf("Expected moveMatches to return %v, got %v", tt.shouldMatch, matches)
			}
		})
	}
}

func TestScoreSuite(t *testing.T) {
	// Create a test board position
	testBoard, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create test board: %v", err)
	}

	positions := []*Position{
		{
			Board:    testBoard,
			BestMove: "e2e4",
		},
		{
			Board:    testBoard,
			BestMove: "d2d4",
		},
	}

	engine := &mockEngine{
		moveToReturn:  board.Move{From: board.Square{File: 4, Rank: 1}, To: board.Square{File: 4, Rank: 3}},
		scoreToReturn: 50,
	}

	config := ai.SearchConfig{
		MaxDepth: 3,
		MaxTime:  time.Second,
	}
	scorer := NewSTSScorer(engine, config, false)

	result := scorer.ScoreSuite(context.Background(), positions, "test_suite")

	if result.SuiteName != "test_suite" {
		t.Errorf("Expected suite name 'test_suite', got '%s'", result.SuiteName)
	}

	if result.PositionCount != 2 {
		t.Errorf("Expected 2 positions, got %d", result.PositionCount)
	}

	if result.MaxScore != 20 { // 2 positions * 10 points each
		t.Errorf("Expected max score 20, got %d", result.MaxScore)
	}

	// First position should get 10 points (exact match), second gets 1 (no match but legal)
	expectedTotalScore := 11
	if result.TotalScore != expectedTotalScore {
		t.Errorf("Expected total score %d, got %d", expectedTotalScore, result.TotalScore)
	}
}
