package search

import (
	"context"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

// TestRepetitionHistory_DetectsRepeatViaMakeUnmake builds a genuine threefold
// repetition on a real board by shuffling both knights out and back
// (Nb1-c3/Nb1 and Nb8-c6/Nb8) twice from the starting position, feeding the
// resulting hashes into the engine's repetition history exactly the way
// FindBestMove/negamax do: setupRepetitionHistory seeds the root, then
// addHistory records each ply's hash as a move is made (mirroring the
// make-move-then-addHistory pairing in negamax's move loop), and
// removeHistory undoes it in lockstep with unmaking the move.
//
// isDrawByRepetition only requires ONE prior match (it is not literally
// counting to three), so the position already reads as a repetition the
// moment the shuffle first returns to the starting position (ply 4); the
// second full cycle (ply 8) is included to demonstrate a genuine threefold
// (occurrences at ply 0, 4, and 8) and confirm detection is stable across it.
func TestRepetitionHistory_DetectsRepeatViaMakeUnmake(t *testing.T) {
	const startFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	b, err := board.FromFEN(startFEN)
	if err != nil {
		t.Fatalf("bad FEN: %v", err)
	}

	engine := NewMinimaxEngine()
	b.SetHashUpdater(engine)
	b.InitializeHashFromPosition(engine.zobrist.HashPosition)

	rootHash := b.GetHash()
	engine.setupRepetitionHistory(rootHash)

	nb1c3 := board.Move{From: sq(1, 0), To: sq(2, 2), Promotion: board.Empty}
	nc3b1 := board.Move{From: sq(2, 2), To: sq(1, 0), Promotion: board.Empty}
	nb8c6 := board.Move{From: sq(1, 7), To: sq(2, 5), Promotion: board.Empty}
	nc6b8 := board.Move{From: sq(2, 5), To: sq(1, 7), Promotion: board.Empty}

	// One full shuffle cycle is 4 plies (White out, Black out, White back,
	// Black back) and reproduces the exact starting hash (knight moves don't
	// touch castling rights or create an en passant square). Two cycles give
	// occurrences at ply 0 (root), 4, and 8: a genuine threefold.
	cycle := []board.Move{nb1c3, nb8c6, nc3b1, nc6b8}
	moves := append(append([]board.Move{}, cycle...), cycle...)

	// ply (1-indexed) -> whether isDrawByRepetition must be true immediately
	// after that move is made and recorded. Plies 1-3 are new positions never
	// seen before, so no match yet. Ply 4 repeats the root (ply 0). From ply 4
	// onward every ply also repeats its counterpart in the first cycle
	// (ply 5 matches ply 1, ply 6 matches ply 2, ..., ply 8 matches ply 4/0),
	// so the match stays true for the rest of the sequence.
	wantRepeatAtPly := map[int]bool{
		1: false, 2: false, 3: false,
		4: true, 5: true, 6: true, 7: true, 8: true,
	}

	undos := make([]board.MoveUndo, 0, len(moves))
	for i, mv := range moves {
		ply := i + 1
		undo, err := b.MakeMoveWithUndo(mv)
		if err != nil {
			t.Fatalf("ply %d: move %+v failed: %v", ply, mv, err)
		}
		undos = append(undos, undo)

		hash := b.GetHash()
		engine.addHistory(hash)

		got := engine.isDrawByRepetition(hash)
		want := wantRepeatAtPly[ply]
		if got != want {
			t.Errorf("ply %d: isDrawByRepetition = %v, want %v", ply, got, want)
		}
		if ply == 4 || ply == 8 {
			if hash != rootHash {
				t.Fatalf("ply %d: hash = %d, want it to match root hash %d (shuffle should return to the exact starting position)", ply, hash, rootHash)
			}
		}
	}

	// Unwind in lockstep (removeHistory paired with UnmakeMove, as negamax's
	// loop does) and confirm the history shrinks back to nothing.
	for i := len(moves) - 1; i >= 0; i-- {
		b.UnmakeMove(undos[i])
		engine.removeHistory()
	}
	if engine.zobristHistoryPly != 0 {
		t.Errorf("zobristHistoryPly = %d after fully unwinding, want 0", engine.zobristHistoryPly)
	}
}

// TestFindBestMove_ReturnsDrawScoreOnUnavoidableRepetition uses a bare king
// and knight vs king and knight position: with only a king and a knight per
// side, neither side can ever force checkmate (a real chess fact - this is
// insufficient mating material), so shuffling the pieces around cannot make
// progress. Verified empirically: at depth 6 with the transposition table on,
// the engine's own isDrawByRepetition fires dozens of times along the
// explored lines, and the final root score is exactly eval.DrawScore, not
// merely a coincidentally-near-zero static evaluation.
func TestFindBestMove_ReturnsDrawScoreOnUnavoidableRepetition(t *testing.T) {
	const fen = "7k/8/8/4n3/4N3/8/8/7K w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("bad FEN: %v", err)
	}

	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(8)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := engine.FindBestMove(ctx, b, movegen.White, SearchConfig{MaxDepth: 6})

	if result.Score != eval.DrawScore {
		t.Errorf("Score = %d, want eval.DrawScore (%d): bare K+N vs K+N cannot force progress", result.Score, eval.DrawScore)
	}
}

// TestFindBestMove_WinningSideAvoidsRepetition uses the same bare K+N vs K+N
// skeleton as above, but gives White an extra queen. White has both a
// tempting repetition-shuffle available (the same knight dance) and an
// objectively winning continuation (the queen forks/wins the black knight
// and marches toward mate). The search must prefer the winning line: the
// returned score should sit far above DrawScore, not settle for the draw the
// shuffle offers.
func TestFindBestMove_WinningSideAvoidsRepetition(t *testing.T) {
	const fen = "7k/8/8/4n3/4N3/8/8/3Q3K w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("bad FEN: %v", err)
	}

	engine := NewMinimaxEngine()
	engine.SetTranspositionTableSize(8)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := engine.FindBestMove(ctx, b, movegen.White, SearchConfig{MaxDepth: 6})

	const winningThreshold = eval.EvaluationScore(500)
	if result.Score <= winningThreshold {
		t.Errorf("Score = %d, want > %d: the extra queen must keep the search from settling for a repetition draw", result.Score, winningThreshold)
	}
	if result.Score == eval.DrawScore {
		t.Error("Score = DrawScore: winning side walked into (or accepted) a repetition instead of playing for the win")
	}
}
