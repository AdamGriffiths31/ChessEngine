package search

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

const (
	// HistoryMaxScore bounds every entry via gravity clamping: updates follow
	// entry += bonus - entry*|bonus|/HistoryMaxScore, so scores asymptotically
	// approach +-HistoryMaxScore instead of hitting a hard clamp.
	HistoryMaxScore = 16384

	// HistoryDecayFactor halves all entries once per search.
	HistoryDecayFactor = 2
)

// HistoryTable tracks the success rate of moves based on from/to square combinations
type HistoryTable struct {
	table [64][64]int32
}

// NewHistoryTable creates a new history table
func NewHistoryTable() *HistoryTable {
	return &HistoryTable{}
}

// Bonus credits a quiet move that caused a beta cutoff.
func (h *HistoryTable) Bonus(move board.Move, depth int) {
	h.apply(move, int32(depth*depth))
}

// Malus penalizes a quiet move that was searched but did not cause the cutoff.
func (h *HistoryTable) Malus(move board.Move, depth int) {
	h.apply(move, -int32(depth*depth))
}

func (h *HistoryTable) apply(move board.Move, bonus int32) {
	entry := h.entry(move)
	*entry += bonus - (*entry*bonus)/HistoryMaxScore
}

func (h *HistoryTable) entry(move board.Move) *int32 {
	from := move.From.Index()
	to := move.To.Index()
	return &h.table[from][to]
}

// GetHistoryScore returns the history score for a move
// Higher scores indicate moves that have been more successful in the past;
// negative scores mark moves that repeatedly failed to cause cutoffs.
func (h *HistoryTable) GetHistoryScore(move board.Move) int32 {
	return *h.entry(move)
}

// Clear resets all history scores to zero
func (h *HistoryTable) Clear() {
	h.table = [64][64]int32{}
}

// Age applies decay to all history scores to prevent them from growing too large
// and to give more weight to recent patterns
func (h *HistoryTable) Age() {
	for i := range h.table {
		for j := range h.table[i] {
			h.table[i][j] /= HistoryDecayFactor
		}
	}
}
