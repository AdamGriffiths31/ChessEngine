package search

import (
	"context"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

func testMove(toSquare int) board.Move {
	return board.Move{
		From: board.Square{File: 0, Rank: 0},
		To:   board.NewSquare(toSquare%8, toSquare/8),
	}
}

func buildMoveList(count int) *movegen.MoveList {
	ml := &movegen.MoveList{}
	for i := range count {
		ml.AddMove(testMove(i))
	}
	return ml
}

const unscored = -eval.MateScore - 1

func TestReorderRootMoves_UnscoredMovesSinkToBack(t *testing.T) {
	ml := buildMoveList(5)
	scores := []eval.EvaluationScore{unscored, 20, unscored, 40, -300}

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
	rng := rand.New(rand.NewSource(42))

	for trial := range 100 {
		count := 1 + rng.Intn(40)
		ml := buildMoveList(count)
		scores := make([]eval.EvaluationScore, count)

		wantScoreByMove := make(map[board.Move]eval.EvaluationScore, count)
		for i := range scores {
			scores[i] = eval.EvaluationScore(rng.Intn(21) - 10)
			wantScoreByMove[ml.Moves[i]] = scores[i]
		}

		type pair struct {
			move  board.Move
			score eval.EvaluationScore
		}
		ref := make([]pair, count)
		for i := range count {
			ref[i] = pair{ml.Moves[i], scores[i]}
		}
		sort.SliceStable(ref, func(a, b int) bool { return ref[a].score > ref[b].score })

		reorderRootMoves(ml, scores)

		for i := range count {
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
	ml := buildMoveList(6)
	scores := []eval.EvaluationScore{10, 30, 20}

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

func findBestMove(t *testing.T, fen string, player movegen.Player, depth int) board.Move {
	t.Helper()
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("bad FEN %q: %v", fen, err)
	}
	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(8)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := engine.FindBestMove(ctx, b, player, Config{MaxDepth: depth})
	return result.BestMove
}

func TestFindBestMove_TacticalCorrectnessAcrossIterations(t *testing.T) {
	tests := []struct {
		name   string
		fen    string
		player movegen.Player
		from   board.Square
		to     board.Square
	}{
		{
			name:   "white rook wins hanging queen",
			fen:    "4k3/8/8/4q3/8/8/4R3/4K3 w - - 0 1",
			player: movegen.White,
			from:   board.Square{File: 4, Rank: 1},
			to:     board.Square{File: 4, Rank: 4},
		},
		{
			name:   "black rook wins hanging queen",
			fen:    "4k3/3r4/8/8/3Q4/8/8/4K3 b - - 0 1",
			player: movegen.Black,
			from:   board.Square{File: 3, Rank: 6},
			to:     board.Square{File: 3, Rank: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findBestMove(t, tt.fen, tt.player, 5)
			if got.From != tt.from || got.To != tt.to {
				t.Errorf("expected %v->%v, got %v->%v", tt.from, tt.to, got.From, got.To)
			}
		})
	}
}
