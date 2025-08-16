// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"sync/atomic"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// History table configuration constants
const (
	MaxHistoryScore    = 10000
	HistoryDecayFactor = 4
	HistoryBonus       = 1
)

// HistoryTable tracks the success rate of moves based on from/to square combinations
// Uses a butterfly table approach for better cache locality and atomic operations for thread safety
type HistoryTable struct {
	table [64][64]atomic.Int32
	age   atomic.Uint32
}

// NewHistoryTable creates a new history table
func NewHistoryTable() *HistoryTable {
	return &HistoryTable{}
}

// UpdateHistory increases the history score for a move that caused a beta cutoff
// The move is considered successful and should be tried earlier in future searches
func (h *HistoryTable) UpdateHistory(move board.Move, depth int) {
	if move.From.File < 0 || move.From.File > 7 || move.From.Rank < 0 || move.From.Rank > 7 {
		return
	}
	if move.To.File < 0 || move.To.File > 7 || move.To.Rank < 0 || move.To.Rank > 7 {
		return
	}

	from := squareToIndex(move.From)
	to := squareToIndex(move.To)

	bonus := HistoryBonus * int32(depth+1) // #nosec G115 - depth is small, intentional conversion

	for {
		current := h.table[from][to].Load()
		newValue := current + bonus
		if newValue > MaxHistoryScore {
			newValue = MaxHistoryScore
		}
		if h.table[from][to].CompareAndSwap(current, newValue) {
			break
		}
	}
}

// GetHistoryScore returns the history score for a move
// Higher scores indicate moves that have been more successful in the past
func (h *HistoryTable) GetHistoryScore(move board.Move) int32 {
	if move.From.File < 0 || move.From.File > 7 || move.From.Rank < 0 || move.From.Rank > 7 {
		return 0
	}
	if move.To.File < 0 || move.To.File > 7 || move.To.Rank < 0 || move.To.Rank > 7 {
		return 0
	}

	from := squareToIndex(move.From)
	to := squareToIndex(move.To)

	return h.table[from][to].Load()
}

// Clear resets all history scores to zero
func (h *HistoryTable) Clear() {
	for i := 0; i < 64; i++ {
		for j := 0; j < 64; j++ {
			h.table[i][j].Store(0)
		}
	}
	h.age.Store(0)
}

// Age applies decay to all history scores to prevent them from growing too large
// and to give more weight to recent patterns
func (h *HistoryTable) Age() {
	currentAge := h.age.Add(1)

	if currentAge%8 == 0 {
		for i := 0; i < 64; i++ {
			for j := 0; j < 64; j++ {
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
}

// GetMaxScore returns the maximum history score currently in the table
// Used for normalizing history scores for LMR reduction calculations
func (h *HistoryTable) GetMaxScore() int32 {
	return MaxHistoryScore
}

// squareToIndex converts a board square to an index (0-63)
func squareToIndex(square board.Square) int {
	return square.Rank*8 + square.File
}
