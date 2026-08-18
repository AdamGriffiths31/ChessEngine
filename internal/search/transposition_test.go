package search

import (
	"testing"
	"time"
	"unsafe"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
)

// TestTranspositionEntrySize pins the entry to 16 bytes so four entries fit in a
// 64-byte cache line. Probe latency is dominated by the DRAM fetch of the entry,
// so accidental padding growth (e.g. from field reordering) is a real regression.
func TestTranspositionEntrySize(t *testing.T) {
	const wantSize = 16
	if size := unsafe.Sizeof(TranspositionEntry{}); size != wantSize {
		t.Errorf("TranspositionEntry size = %d bytes, want %d", size, wantSize)
	}
}

// TestTranspositionTableEntryCount verifies the table allocates entries based on
// the actual entry size, so the requested memory budget is used, not overshot.
func TestTranspositionTableEntryCount(t *testing.T) {
	const sizeMB = 8
	tt := NewTranspositionTable(sizeMB)

	budgetBytes := uint64(sizeMB) * 1024 * 1024
	entrySize := uint64(unsafe.Sizeof(TranspositionEntry{}))
	usedBytes := tt.size * entrySize

	if usedBytes > budgetBytes {
		t.Errorf("table uses %d bytes, exceeds %dMB budget", usedBytes, sizeMB)
	}
	// The table is sized to the largest power of two within budget; anything
	// under half the budget means the entry-size constant is out of date.
	if usedBytes*2 <= budgetBytes {
		t.Errorf("table uses only %d of %d budget bytes; entry size constant likely stale", usedBytes, budgetBytes)
	}
}

// sq is a small helper for building board.Square literals tersely in table tests.
func sq(file, rank int) board.Square {
	return board.Square{File: file, Rank: rank}
}

// TestStoreProbeRoundTrip_AllEntryTypes verifies that a stored entry of each of
// the three EntryType values is faithfully returned by Probe: hash match,
// depth, score, type, and move all survive the round trip.
func TestStoreProbeRoundTrip_AllEntryTypes(t *testing.T) {
	tests := []struct {
		name      string
		entryType EntryType
	}{
		{"exact", EntryExact},
		{"lower bound", EntryLowerBound},
		{"upper bound", EntryUpperBound},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tt := NewTranspositionTable(1)
			hash := uint64(0xC0FFEE00 + i)
			depth := 7 + i
			score := eval.EvaluationScore(111 + i)
			move := board.Move{From: sq(i%8, 1), To: sq((i+2)%8, 3), Promotion: board.Empty, IsCapture: i%2 == 0}

			tt.Store(hash, depth, score, tc.entryType, move)

			entry, found := tt.Probe(hash)
			if !found {
				t.Fatalf("Probe(%d) did not find the entry just stored", hash)
			}
			if entry.Hash != hash {
				t.Errorf("Hash = %d, want %d", entry.Hash, hash)
			}
			if entry.GetDepth() != depth {
				t.Errorf("GetDepth() = %d, want %d", entry.GetDepth(), depth)
			}
			if entry.GetType() != tc.entryType {
				t.Errorf("GetType() = %v, want %v", entry.GetType(), tc.entryType)
			}
			if entry.Score != score {
				t.Errorf("Score = %d, want %d", entry.Score, score)
			}
			if got := entry.GetMove(); got.From != move.From || got.To != move.To || got.IsCapture != move.IsCapture {
				t.Errorf("GetMove() = %+v, want From/To/IsCapture matching %+v", got, move)
			}
		})
	}
}

// TestProbeTT_BoundBasedCutoffDecisions constructs TT entries of each type
// directly, then drives MinimaxEngine.probeTT with chosen alpha/beta windows
// to verify the cutoff/narrowing decisions match transposition.go's contract:
// EntryExact always cuts off; EntryLowerBound cuts off only when its score
// reaches beta (otherwise it may raise alpha); EntryUpperBound cuts off only
// when its score falls to alpha or below (otherwise it may lower beta). Each
// cutoff case also asserts the PVNodes/CutNodes/AllNodes classification a TT
// shortcut must produce -- the same classification the normal move-loop
// fallthrough at the bottom of negamax would have assigned had the node not
// been resolved early.
//
//nolint:gocyclo // table of cutoff/narrowing cases per entry type; splitting
// would scatter the bound-decision contract this test exists to pin down
func TestProbeTT_BoundBasedCutoffDecisions(t *testing.T) {
	newEngineWithEntry := func(depth int, score eval.EvaluationScore, entryType EntryType, move board.Move) (*MinimaxEngine, uint64) {
		engine := NewMinimaxEngine()
		engine.SetTranspositionTableSize(1)
		hash := uint64(42)
		engine.transpositionTable.Store(hash, depth, score, entryType, move)
		return engine, hash
	}
	storedMove := board.Move{From: sq(1, 1), To: sq(1, 3), Promotion: board.Empty}

	t.Run("exact entry always cuts off and leaves bounds untouched", func(t *testing.T) {
		engine, hash := newEngineWithEntry(4, 250, EntryExact, storedMove)

		ttMove, score, alphaOut, betaOut, done := engine.probeTT(hash, 4, 0, 100, 200)

		if !done {
			t.Fatal("expected EntryExact to cut off")
		}
		if score != 250 {
			t.Errorf("score = %d, want 250", score)
		}
		if alphaOut != 100 || betaOut != 200 {
			t.Errorf("bounds = (%d,%d), want unchanged (100,200)", alphaOut, betaOut)
		}
		if ttMove != storedMove {
			t.Errorf("ttMove = %+v, want %+v", ttMove, storedMove)
		}
		if got := engine.searchState.searchStats.TTCutoffs; got != 1 {
			t.Errorf("TTCutoffs = %d, want 1", got)
		}
		if got := engine.searchState.searchStats.PVNodes; got != 1 {
			t.Errorf("PVNodes = %d, want 1 (an exact TT entry resolves the node's true score)", got)
		}
	})

	t.Run("lower bound cuts off when score reaches beta", func(t *testing.T) {
		engine, hash := newEngineWithEntry(4, 300, EntryLowerBound, storedMove)

		_, score, _, _, done := engine.probeTT(hash, 4, 0, 100, 200)

		if !done {
			t.Fatal("expected EntryLowerBound with score >= beta to cut off")
		}
		if score != 300 {
			t.Errorf("score = %d, want 300", score)
		}
		if got := engine.searchState.searchStats.TTCutoffs; got != 1 {
			t.Errorf("TTCutoffs = %d, want 1 (lower-bound cutoff must count)", got)
		}
		if got := engine.searchState.searchStats.CutNodes; got != 1 {
			t.Errorf("CutNodes = %d, want 1 (a lower-bound cutoff is a fail-high)", got)
		}
	})

	t.Run("lower bound narrows alpha without cutting off when between bounds", func(t *testing.T) {
		engine, hash := newEngineWithEntry(4, 150, EntryLowerBound, storedMove)

		_, _, alphaOut, betaOut, done := engine.probeTT(hash, 4, 0, 100, 200)

		if done {
			t.Fatal("expected no cutoff when lower-bound score is below beta")
		}
		if alphaOut != 150 {
			t.Errorf("alphaOut = %d, want 150 (raised to the lower-bound score)", alphaOut)
		}
		if betaOut != 200 {
			t.Errorf("betaOut = %d, want unchanged 200", betaOut)
		}
	})

	t.Run("lower bound leaves alpha alone when score does not exceed it", func(t *testing.T) {
		engine, hash := newEngineWithEntry(4, 50, EntryLowerBound, storedMove)

		_, _, alphaOut, _, done := engine.probeTT(hash, 4, 0, 100, 200)

		if done {
			t.Fatal("expected no cutoff")
		}
		if alphaOut != 100 {
			t.Errorf("alphaOut = %d, want unchanged 100", alphaOut)
		}
	})

	t.Run("upper bound cuts off when score falls to alpha WITHOUT counting as a TTCutoff", func(t *testing.T) {
		// This is the documented asymmetry: EntryExact and EntryLowerBound
		// cutoffs increment TTCutoffs, but the EntryUpperBound cutoff path in
		// probeTT returns done=true without incrementing TTCutoffs.
		engine, hash := newEngineWithEntry(4, 50, EntryUpperBound, storedMove)

		_, score, _, _, done := engine.probeTT(hash, 4, 0, 100, 200)

		if !done {
			t.Fatal("expected EntryUpperBound with score <= alpha to cut off")
		}
		if score != 50 {
			t.Errorf("score = %d, want 50", score)
		}
		if got := engine.searchState.searchStats.TTCutoffs; got != 0 {
			t.Errorf("TTCutoffs = %d, want 0 (upper-bound cutoff is not counted, unlike exact/lower)", got)
		}
		if got := engine.searchState.searchStats.TTHits; got != 1 {
			t.Errorf("TTHits = %d, want 1", got)
		}
		if got := engine.searchState.searchStats.AllNodes; got != 1 {
			t.Errorf("AllNodes = %d, want 1 (an upper-bound cutoff is a fail-low, counted despite not being a TTCutoff)", got)
		}
	})

	t.Run("upper bound narrows beta without cutting off when between bounds", func(t *testing.T) {
		engine, hash := newEngineWithEntry(4, 150, EntryUpperBound, storedMove)

		_, _, alphaOut, betaOut, done := engine.probeTT(hash, 4, 0, 100, 200)

		if done {
			t.Fatal("expected no cutoff when upper-bound score is above alpha")
		}
		if betaOut != 150 {
			t.Errorf("betaOut = %d, want 150 (lowered to the upper-bound score)", betaOut)
		}
		if alphaOut != 100 {
			t.Errorf("alphaOut = %d, want unchanged 100", alphaOut)
		}
	})

	t.Run("upper bound leaves beta alone when score does not fall below it", func(t *testing.T) {
		engine, hash := newEngineWithEntry(4, 250, EntryUpperBound, storedMove)

		_, _, _, betaOut, done := engine.probeTT(hash, 4, 0, 100, 200)

		if done {
			t.Fatal("expected no cutoff")
		}
		if betaOut != 200 {
			t.Errorf("betaOut = %d, want unchanged 200", betaOut)
		}
	})

	t.Run("entry shallower than requested depth is not used for cutoff or narrowing", func(t *testing.T) {
		engine, hash := newEngineWithEntry(2, 300, EntryLowerBound, storedMove)

		ttMove, _, alphaOut, betaOut, done := engine.probeTT(hash, 5, 0, 100, 200)

		if done {
			t.Fatal("expected shallow entry not to cut off a deeper request")
		}
		if alphaOut != 100 || betaOut != 200 {
			t.Errorf("bounds = (%d,%d), want unchanged (100,200)", alphaOut, betaOut)
		}
		// The move is still handed back for move ordering even when the
		// score/bound is too shallow to trust.
		if ttMove != storedMove {
			t.Errorf("ttMove = %+v, want %+v (TT move still returned for ordering)", ttMove, storedMove)
		}
		if got := engine.searchState.searchStats.TTCutoffs; got != 0 {
			t.Errorf("TTCutoffs = %d, want 0", got)
		}
		if got := engine.searchState.searchStats.TTHits; got != 1 {
			t.Errorf("TTHits = %d, want 1 (still counts as a hit)", got)
		}
	})

	t.Run("invalid stored move is sanitized to the zero move", func(t *testing.T) {
		// board.Move{} has From == To, which board.Move.IsValid() rejects;
		// probeTT must not hand back a nonsensical move for ordering.
		engine, hash := newEngineWithEntry(4, 250, EntryExact, board.Move{})

		ttMove, _, _, _, _ := engine.probeTT(hash, 4, 0, 100, 200)

		if ttMove != (board.Move{}) {
			t.Errorf("ttMove = %+v, want zero value", ttMove)
		}
	})

	t.Run("miss leaves bounds untouched and reports not done", func(t *testing.T) {
		engine := NewMinimaxEngine()
		engine.SetTranspositionTableSize(1)

		ttMove, score, alphaOut, betaOut, done := engine.probeTT(0xDEADBEEF, 4, 0, 100, 200)

		if done {
			t.Fatal("expected a miss on an empty table")
		}
		if score != 0 {
			t.Errorf("score = %d, want 0 on a miss", score)
		}
		if alphaOut != 100 || betaOut != 200 {
			t.Errorf("bounds = (%d,%d), want unchanged (100,200)", alphaOut, betaOut)
		}
		if ttMove != (board.Move{}) {
			t.Errorf("ttMove = %+v, want zero value on a miss", ttMove)
		}
	})
}

// TestScoreToFromTT_MateDistanceRoundTrip verifies scoreToTT/scoreFromTT are
// exact inverses (for a matching ply) at plies 0, 1, 5, and 30, for both mate
// signs (about to deliver mate, and about to be mated), and confirms the
// concrete arithmetic: positive mate scores gain `ply` on the way into the
// table and lose it on the way out; negative mate scores do the reverse.
func TestScoreToFromTT_MateDistanceRoundTrip(t *testing.T) {
	plies := []int{0, 1, 5, 30}
	mateScores := []eval.EvaluationScore{
		eval.MateScore,      // mate delivered this move
		eval.MateScore - 1,  // mate in 2
		eval.MateScore - 49, // still within the mate band (threshold is 1000)
		-eval.MateScore,     // about to be mated this move
		-eval.MateScore + 1,
		-eval.MateScore + 49,
	}

	for _, ply := range plies {
		for _, s := range mateScores {
			stored := scoreToTT(s, ply)
			back := scoreFromTT(stored, ply)
			if back != s {
				t.Errorf("ply=%d score=%d: round trip gave %d via stored=%d", ply, s, back, stored)
			}

			if s > 0 {
				if want := s + eval.EvaluationScore(ply); stored != want {
					t.Errorf("ply=%d score=%d: scoreToTT = %d, want %d (score+ply)", ply, s, stored, want)
				}
			} else {
				if want := s - eval.EvaluationScore(ply); stored != want {
					t.Errorf("ply=%d score=%d: scoreToTT = %d, want %d (score-ply)", ply, s, stored, want)
				}
			}
		}
	}
}

// TestScoreToFromTT_NonMateScoresPassThrough verifies ordinary evaluation
// scores (far outside the mate band) are untouched by ply adjustment in
// either direction, at every tested ply.
func TestScoreToFromTT_NonMateScoresPassThrough(t *testing.T) {
	plies := []int{0, 1, 5, 30}
	scores := []eval.EvaluationScore{0, 1, -1, 500, -500, eval.DrawScore}

	for _, ply := range plies {
		for _, s := range scores {
			if got := scoreToTT(s, ply); got != s {
				t.Errorf("scoreToTT(%d, ply=%d) = %d, want unchanged %d", s, ply, got, s)
			}
			if got := scoreFromTT(s, ply); got != s {
				t.Errorf("scoreFromTT(%d, ply=%d) = %d, want unchanged %d", s, ply, got, s)
			}
		}
	}
}

// TestShouldReplace_SameHashGatedByDepthOnly verifies that when a new store
// targets the same position already occupying a slot (Hash match), replacement
// is governed purely by depth: a shallower re-store is rejected, a deeper one
// is accepted, and age plays no part.
func TestShouldReplace_SameHashGatedByDepthOnly(t *testing.T) {
	tt := NewTranspositionTable(1)
	hash := uint64(777)

	tt.Store(hash, 3, 100, EntryExact, board.Move{})

	// Shallower re-store of the same position must not replace.
	tt.Store(hash, 2, 999, EntryExact, board.Move{})
	entry, found := tt.Probe(hash)
	if !found || entry.Score != 100 || entry.GetDepth() != 3 {
		t.Fatalf("shallower same-hash store replaced a deeper entry: entry=%+v found=%v", entry, found)
	}

	// Deeper re-store of the same position must replace.
	tt.Store(hash, 5, 555, EntryExact, board.Move{})
	entry, found = tt.Probe(hash)
	if !found || entry.Score != 555 || entry.GetDepth() != 5 {
		t.Fatalf("deeper same-hash store did not replace: entry=%+v found=%v", entry, found)
	}
}

// TestShouldReplace_CollisionGatedByAgeNotDepth verifies the two-bucket
// collision path: a colliding entry from the SAME search generation (age) is
// never evicted regardless of depth, but IncrementAge crossing the generation
// boundary makes a stale-generation entry replaceable by a shallower colliding
// entry. hashB and hashC are constructed to land in the exact same first
// bucket as hashA (same low bits, mod table size) so the collision path in
// shouldReplace, not the second-bucket fallback, is what's under test.
func TestShouldReplace_CollisionGatedByAgeNotDepth(t *testing.T) {
	tt := NewTranspositionTable(1)
	tableSize := tt.mask + 1

	hashA := uint64(5)
	hashB := hashA + tableSize   // same first-bucket index as hashA
	hashC := hashA + 2*tableSize // same first-bucket index as hashA

	tt.Store(hashA, 5, 100, EntryExact, board.Move{}) // deep entry, generation 0

	// hashB collides with hashA in the first bucket, same generation (age 0):
	// must NOT evict hashA even though shouldReplace is asked, and instead
	// falls through to the empty second bucket.
	tt.Store(hashB, 1, 200, EntryExact, board.Move{})

	entryA, foundA := tt.Probe(hashA)
	if !foundA || entryA.Score != 100 {
		t.Fatalf("hashA was evicted by a same-generation collision: found=%v entry=%+v", foundA, entryA)
	}
	entryB, foundB := tt.Probe(hashB)
	if !foundB || entryB.Score != 200 {
		t.Fatalf("hashB (same-generation collider) was not stored anywhere: found=%v entry=%+v", foundB, entryB)
	}
	if tt.collisions == 0 {
		t.Fatalf("expected the hashA/hashB collision to be counted, collisions=%d", tt.collisions)
	}

	// Cross the age-generation boundary.
	tt.IncrementAge()

	// hashC collides with hashA (still occupying the first bucket) but is now
	// from a NEW generation and has a shallower depth (1 < 5). Per
	// shouldReplace, a cross-generation collision replaces regardless of
	// depth, so hashC must evict hashA from the first bucket.
	tt.Store(hashC, 1, 300, EntryExact, board.Move{})

	if _, found := tt.Probe(hashA); found {
		t.Error("hashA should have been evicted by a new-generation colliding store, regardless of its shallower depth")
	}
	entryC, foundC := tt.Probe(hashC)
	if !foundC || entryC.Score != 300 {
		t.Fatalf("hashC (new-generation collider) did not replace hashA: found=%v entry=%+v", foundC, entryC)
	}
}

// TestUnpackMove_RoundTrip exercises packMove/unpackMove for every move shape
// the packed format distinguishes. Piece/Captured are never packed (derived
// from board position by the caller), so they are intentionally absent from
// both input and expected output; fields packMove does not preserve for a
// given move type (e.g. IsCapture on a promotion) are expected to come back
// zeroed, per packMove's documented behavior.
func TestUnpackMove_RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input board.Move
		want  board.Move
	}{
		{
			name:  "plain quiet move",
			input: board.Move{From: sq(4, 1), To: sq(4, 3), Promotion: board.Empty},
			want:  board.Move{From: sq(4, 1), To: sq(4, 3), Promotion: board.Empty},
		},
		{
			name:  "plain capture (not en passant)",
			input: board.Move{From: sq(0, 0), To: sq(7, 7), Promotion: board.Empty, IsCapture: true},
			want:  board.Move{From: sq(0, 0), To: sq(7, 7), Promotion: board.Empty, IsCapture: true},
		},
		{
			name:  "en passant capture",
			input: board.Move{From: sq(4, 4), To: sq(3, 5), Promotion: board.Empty, IsCapture: true, IsEnPassant: true},
			want:  board.Move{From: sq(4, 4), To: sq(3, 5), Promotion: board.Empty, IsCapture: true, IsEnPassant: true},
		},
		{
			name:  "kingside castling (king two-square move)",
			input: board.Move{From: sq(4, 0), To: sq(6, 0), Promotion: board.Empty, IsCastling: true},
			want:  board.Move{From: sq(4, 0), To: sq(6, 0), Promotion: board.Empty, IsCastling: true},
		},
		{
			name:  "queenside castling (king two-square move)",
			input: board.Move{From: sq(4, 7), To: sq(2, 7), Promotion: board.Empty, IsCastling: true},
			want:  board.Move{From: sq(4, 7), To: sq(2, 7), Promotion: board.Empty, IsCastling: true},
		},
		{
			name:  "white promotion to queen",
			input: board.Move{From: sq(1, 6), To: sq(1, 7), Promotion: board.WhiteQueen},
			want:  board.Move{From: sq(1, 6), To: sq(1, 7), Promotion: board.WhiteQueen},
		},
		{
			name:  "white promotion to rook",
			input: board.Move{From: sq(1, 6), To: sq(1, 7), Promotion: board.WhiteRook},
			want:  board.Move{From: sq(1, 6), To: sq(1, 7), Promotion: board.WhiteRook},
		},
		{
			name:  "white promotion to bishop",
			input: board.Move{From: sq(1, 6), To: sq(1, 7), Promotion: board.WhiteBishop},
			want:  board.Move{From: sq(1, 6), To: sq(1, 7), Promotion: board.WhiteBishop},
		},
		{
			name:  "white promotion to knight (capturing)",
			input: board.Move{From: sq(1, 6), To: sq(2, 7), Promotion: board.WhiteKnight, IsCapture: true},
			// packMove does not encode capture for promotions (derived from
			// board position instead), so IsCapture is expected to be lost.
			want: board.Move{From: sq(1, 6), To: sq(2, 7), Promotion: board.WhiteKnight},
		},
		{
			name:  "black promotion to queen",
			input: board.Move{From: sq(1, 1), To: sq(1, 0), Promotion: board.BlackQueen},
			want:  board.Move{From: sq(1, 1), To: sq(1, 0), Promotion: board.BlackQueen},
		},
		{
			name:  "black promotion to rook",
			input: board.Move{From: sq(1, 1), To: sq(1, 0), Promotion: board.BlackRook},
			want:  board.Move{From: sq(1, 1), To: sq(1, 0), Promotion: board.BlackRook},
		},
		{
			name:  "black promotion to bishop",
			input: board.Move{From: sq(1, 1), To: sq(1, 0), Promotion: board.BlackBishop},
			want:  board.Move{From: sq(1, 1), To: sq(1, 0), Promotion: board.BlackBishop},
		},
		{
			name:  "black promotion to knight",
			input: board.Move{From: sq(1, 1), To: sq(1, 0), Promotion: board.BlackKnight},
			want:  board.Move{From: sq(1, 1), To: sq(1, 0), Promotion: board.BlackKnight},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			packed := packMove(tc.input)
			got := unpackMove(packed)

			if got.From != tc.want.From {
				t.Errorf("From = %+v, want %+v", got.From, tc.want.From)
			}
			if got.To != tc.want.To {
				t.Errorf("To = %+v, want %+v", got.To, tc.want.To)
			}
			if got.IsCapture != tc.want.IsCapture {
				t.Errorf("IsCapture = %v, want %v", got.IsCapture, tc.want.IsCapture)
			}
			if got.IsCastling != tc.want.IsCastling {
				t.Errorf("IsCastling = %v, want %v", got.IsCastling, tc.want.IsCastling)
			}
			if got.IsEnPassant != tc.want.IsEnPassant {
				t.Errorf("IsEnPassant = %v, want %v", got.IsEnPassant, tc.want.IsEnPassant)
			}
			if got.Promotion != tc.want.Promotion {
				t.Errorf("Promotion = %v, want %v", got.Promotion, tc.want.Promotion)
			}
		})
	}
}

// TestStats_Correctness verifies Stats() derives its reported fields from
// the table's counters. Probe only samples hits/misses (1-in-256, to avoid
// counter overhead on every lookup - see Probe), so counters are set
// directly here rather than via a specific number of Probe calls.
func TestStats_Correctness(t *testing.T) {
	tt := NewTranspositionTable(1)
	tt.hits = 2
	tt.misses = 1
	tt.collisions = 4
	tt.secondBucketUse = 3
	tt.totalStores = 6

	stats := tt.Stats()
	if stats.Hits != 2 {
		t.Errorf("Hits = %d, want 2", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Misses = %d, want 1", stats.Misses)
	}
	if stats.Collisions != 4 {
		t.Errorf("Collisions = %d, want 4", stats.Collisions)
	}
	if want := float64(2) / float64(3) * 100; stats.HitRate != want {
		t.Errorf("HitRate = %v, want %v", stats.HitRate, want)
	}
	if stats.SecondBucketUse != 3 {
		t.Errorf("SecondBucketUse = %d, want 3", stats.SecondBucketUse)
	}
	if want := float64(3) / float64(6) * 100; stats.SecondBucketRate != want {
		t.Errorf("SecondBucketRate = %v, want %v", stats.SecondBucketRate, want)
	}
}

// TestStats_DoesNotScanTable is a regression test for a performance bug: a
// prior refactor merged Stats() with a fill-rate/average-depth calculation
// that walked every table entry (an O(table size) scan), and the UCI adapter
// calls Stats() unconditionally on every completed search. With a realistic
// hash size that scan is millions of entries per move. Stats() must stay
// O(1) regardless of table size - this allocates a large table and asserts
// Stats() completes near-instantly, which a full scan would not.
func TestStats_DoesNotScanTable(t *testing.T) {
	tt := NewTranspositionTable(128) // ~8M entries, matching the UCI default Hash size

	start := time.Now()
	tt.Stats()
	elapsed := time.Since(start)

	if elapsed > 5*time.Millisecond {
		t.Errorf("Stats() took %v on a 128MB table; expected O(1), suggests it is scanning entries again", elapsed)
	}
}
