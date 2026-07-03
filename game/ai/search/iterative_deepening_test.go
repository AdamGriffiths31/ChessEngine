package search

import (
	"context"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// testMove creates a distinct move identified by its destination square index
func testMove(toSquare int) board.Move {
	return board.Move{
		From: board.Square{File: 0, Rank: 0},
		To:   board.Square{File: toSquare % 8, Rank: toSquare / 8},
	}
}

func buildMoveList(count int) *moves.MoveList {
	ml := &moves.MoveList{}
	for i := 0; i < count; i++ {
		ml.AddMove(testMove(i))
	}
	return ml
}

const unscored = -ai.MateScore - 1

func TestReorderRootMoves_SortsByScoreDescending(t *testing.T) {
	ml := buildMoveList(4)
	scores := []ai.EvaluationScore{10, 50, 30, 20}

	reorderRootMoves(ml, scores)

	wantOrder := []int{1, 2, 3, 0} // original indices sorted by score desc
	wantScores := []ai.EvaluationScore{50, 30, 20, 10}
	for i, origIdx := range wantOrder {
		if ml.Moves[i] != testMove(origIdx) {
			t.Errorf("position %d: expected move %d, got %v", i, origIdx, ml.Moves[i].To)
		}
		if scores[i] != wantScores[i] {
			t.Errorf("position %d: expected score %d, got %d", i, wantScores[i], scores[i])
		}
	}
}

func TestReorderRootMoves_BestMoveEndsUpFirst(t *testing.T) {
	// The property iterative deepening relies on: the move with the highest
	// score from the completed iteration must be searched first at the next
	// depth, wherever it started in the list
	for bestIdx := 0; bestIdx < 10; bestIdx++ {
		ml := buildMoveList(10)
		scores := make([]ai.EvaluationScore, 10)
		for i := range scores {
			scores[i] = ai.EvaluationScore(i) // ascending, so last is best...
		}
		scores[bestIdx] = 1000 // ...except the designated best

		reorderRootMoves(ml, scores)

		if ml.Moves[0] != testMove(bestIdx) {
			t.Errorf("best move started at index %d but is not first after reorder", bestIdx)
		}
		if scores[0] != 1000 {
			t.Errorf("best score not first after reorder: got %d", scores[0])
		}
	}
}

func TestReorderRootMoves_StableForEqualScores(t *testing.T) {
	ml := buildMoveList(5)
	scores := []ai.EvaluationScore{20, 20, 50, 20, 20}

	reorderRootMoves(ml, scores)

	// Move 2 (score 50) first, then moves 0,1,3,4 in original order
	wantOrder := []int{2, 0, 1, 3, 4}
	for i, origIdx := range wantOrder {
		if ml.Moves[i] != testMove(origIdx) {
			t.Errorf("position %d: expected move %d (stable order), got %v", i, origIdx, ml.Moves[i].To)
		}
	}
}

func TestReorderRootMoves_UnscoredMovesSinkToBack(t *testing.T) {
	// Illegal or unsearched moves keep the sentinel score and must end up
	// behind every scored move, preserving their own relative order
	ml := buildMoveList(5)
	scores := []ai.EvaluationScore{unscored, 20, unscored, 40, -300}

	reorderRootMoves(ml, scores)

	wantOrder := []int{3, 1, 4, 0, 2}
	for i, origIdx := range wantOrder {
		if ml.Moves[i] != testMove(origIdx) {
			t.Errorf("position %d: expected move %d, got %v", i, origIdx, ml.Moves[i].To)
		}
	}
	if scores[3] != unscored || scores[4] != unscored {
		t.Errorf("unscored sentinels not at the back: %v", scores)
	}
}

func TestReorderRootMoves_KeepsMovesAndScoresAligned(t *testing.T) {
	// Property test against a reference stable sort: after reordering, every
	// move must still carry the exact score it earned. Misalignment here would
	// make the engine play a move other than the one it evaluated best.
	rng := rand.New(rand.NewSource(42))

	for trial := 0; trial < 100; trial++ {
		count := 1 + rng.Intn(40)
		ml := buildMoveList(count)
		scores := make([]ai.EvaluationScore, count)

		wantScoreByMove := make(map[board.Move]ai.EvaluationScore, count)
		for i := range scores {
			scores[i] = ai.EvaluationScore(rng.Intn(21) - 10) // few values, many ties
			wantScoreByMove[ml.Moves[i]] = scores[i]
		}

		// Reference: stable sort of (move, score) pairs by score descending
		type pair struct {
			move  board.Move
			score ai.EvaluationScore
		}
		ref := make([]pair, count)
		for i := 0; i < count; i++ {
			ref[i] = pair{ml.Moves[i], scores[i]}
		}
		sort.SliceStable(ref, func(a, b int) bool { return ref[a].score > ref[b].score })

		reorderRootMoves(ml, scores)

		for i := 0; i < count; i++ {
			if ml.Moves[i] != ref[i].move || scores[i] != ref[i].score {
				t.Fatalf("trial %d: position %d diverges from reference stable sort", trial, i)
			}
			if scores[i] != wantScoreByMove[ml.Moves[i]] {
				t.Fatalf("trial %d: move at %d lost its score: got %d, want %d",
					trial, i, scores[i], wantScoreByMove[ml.Moves[i]])
			}
		}
	}
}

func TestReorderRootMoves_ScoresShorterThanMoveList(t *testing.T) {
	// Defensive: must not panic or disturb moves beyond len(scores)
	ml := buildMoveList(6)
	scores := []ai.EvaluationScore{10, 30, 20}

	reorderRootMoves(ml, scores)

	if ml.Moves[0] != testMove(1) || ml.Moves[1] != testMove(2) || ml.Moves[2] != testMove(0) {
		t.Errorf("first three moves not sorted: %v", ml.Moves[:3])
	}
	for i := 3; i < 6; i++ {
		if ml.Moves[i] != testMove(i) {
			t.Errorf("move beyond scores length was disturbed at %d", i)
		}
	}
}

// findBestMove runs a search on a FEN and returns the engine's chosen move
func findBestMove(t *testing.T, fen string, player moves.Player, depth int) board.Move {
	t.Helper()
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("bad FEN %q: %v", fen, err)
	}
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(8)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := engine.FindBestMove(ctx, b, player, ai.SearchConfig{MaxDepth: depth})
	return result.BestMove
}

// TestFindBestMove_TacticalCorrectnessAcrossIterations verifies end to end
// that iterative deepening (with root reordering active between iterations)
// still returns the objectively best move. A move/score misalignment in the
// reordering would surface here as a wrong best move.
func TestFindBestMove_TacticalCorrectnessAcrossIterations(t *testing.T) {
	tests := []struct {
		name   string
		fen    string
		player moves.Player
		from   board.Square
		to     board.Square
	}{
		{
			name:   "white rook wins hanging queen",
			fen:    "4k3/8/8/4q3/8/8/4R3/4K3 w - - 0 1",
			player: moves.White,
			from:   board.Square{File: 4, Rank: 1}, // e2
			to:     board.Square{File: 4, Rank: 4}, // e5
		},
		{
			name:   "black rook wins hanging queen",
			fen:    "4k3/3r4/8/8/3Q4/8/8/4K3 b - - 0 1",
			player: moves.Black,
			from:   board.Square{File: 3, Rank: 6}, // d7
			to:     board.Square{File: 3, Rank: 3}, // d4
		},
		{
			name:   "white mates in one",
			fen:    "r1bqkbnr/pppp1ppp/2n5/4p2Q/2B1P3/8/PPPP1PPP/RNB1K1NR w KQkq - 4 4",
			player: moves.White,
			from:   board.Square{File: 7, Rank: 4}, // h5
			to:     board.Square{File: 5, Rank: 6}, // f7
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Depth 5 guarantees several completed iterations, so the root
			// list has been reordered multiple times before the final answer
			got := findBestMove(t, tt.fen, tt.player, 5)
			if got.From != tt.from || got.To != tt.to {
				t.Errorf("expected %v->%v, got %v->%v", tt.from, tt.to, got.From, got.To)
			}
		})
	}
}
