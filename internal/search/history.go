package search

import (
	"sync/atomic"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// History table configuration constants
const (
	// HistoryMaxScore bounds every entry via gravity clamping: updates follow
	// entry += bonus - entry*|bonus|/HistoryMaxScore, so scores asymptotically
	// approach +-HistoryMaxScore instead of hitting a hard clamp. This keeps
	// frequently-repeated bonuses meaningful while allowing negative (malus)
	// values, which the linear bonus/hard-cap scheme could never produce.
	HistoryMaxScore = 16384

	// HistoryDecayFactor halves all entries once per search (called from
	// FindBestMove), so stale signal fades within a few moves instead of
	// persisting across eight.
	HistoryDecayFactor = 2
)

// HistoryTable tracks the success rate of moves based on from/to square combinations
// Uses a butterfly table approach for better cache locality
type HistoryTable struct {
	table [64][64]atomic.Int32
}

// NewHistoryTable creates a new history table
func NewHistoryTable() *HistoryTable {
	return &HistoryTable{}
}

// historyBonus returns the per-update magnitude: quadratic in remaining depth,
// so cutoffs found deeper in the tree (more evidence of a good move) shift
// history more than shallow ones. Matches modern engine practice.
func historyBonus(depth int) int32 {
	if depth < 0 {
		return 0
	}
	return int32(depth * depth) // #nosec G115 - depth is small, intentional conversion
}

// Bonus credits a quiet move that caused a beta cutoff.
func (h *HistoryTable) Bonus(move board.Move, depth int) {
	h.apply(move, historyBonus(depth))
}

// Malus penalizes a quiet move that was searched but did not cause the
// cutoff. Negative entries are what make the LMR "reduce bad-history moves"
// threshold (params.HistoryLowThreshold) reachable at all.
func (h *HistoryTable) Malus(move board.Move, depth int) {
	h.apply(move, -historyBonus(depth))
}

// apply adjusts one entry by bonus using the gravity formula above, bounding
// the result to (-HistoryMaxScore, +HistoryMaxScore) without a hard clamp.
func (h *HistoryTable) apply(move board.Move, bonus int32) {
	if !isValidSquare(move.From) || !isValidSquare(move.To) || bonus == 0 {
		return
	}

	from := squareToIndex(move.From)
	to := squareToIndex(move.To)

	for {
		current := h.table[from][to].Load()
		newValue := current + bonus - (current*bonus)/HistoryMaxScore
		if h.table[from][to].CompareAndSwap(current, newValue) {
			break
		}
	}
}

// GetHistoryScore returns the history score for a move
// Higher scores indicate moves that have been more successful in the past;
// negative scores mark moves that repeatedly failed to cause cutoffs.
func (h *HistoryTable) GetHistoryScore(move board.Move) int32 {
	if !isValidSquare(move.From) || !isValidSquare(move.To) {
		return 0
	}

	from := squareToIndex(move.From)
	to := squareToIndex(move.To)

	return h.table[from][to].Load()
}

// Clear resets all history scores to zero
func (h *HistoryTable) Clear() {
	for i := range 64 {
		for j := range 64 {
			h.table[i][j].Store(0)
		}
	}
}

// Age applies decay to all history scores to prevent them from growing too large
// and to give more weight to recent patterns
func (h *HistoryTable) Age() {
	for i := range 64 {
		for j := range 64 {
			for {
				current := h.table[i][j].Load()
				newValue := current / HistoryDecayFactor
				if h.table[i][j].CompareAndSwap(current, newValue) {
					break
				}
			}
		}
	}
}

func isValidSquare(square board.Square) bool {
	return square.File >= 0 && square.File <= 7 && square.Rank >= 0 && square.Rank <= 7
}

func squareToIndex(square board.Square) int {
	return square.Rank*8 + square.File
}
