package search

import (
	"sync"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

const (
	MaxHistoryScore    = 10000
	HistoryDecayFactor = 4
	HistoryBonus       = 1
)

// HistoryTable tracks the success rate of moves based on from/to square combinations
// Uses a butterfly table approach for better cache locality
type HistoryTable struct {
	table [64][64]int32
	mutex sync.RWMutex
	age   uint32
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

	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Add bonus based on depth - deeper searches give more bonus
	bonus := HistoryBonus * int32(depth+1)

	// Add the bonus but cap at maximum score
	h.table[from][to] += bonus
	if h.table[from][to] > MaxHistoryScore {
		h.table[from][to] = MaxHistoryScore
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

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.table[from][to]
}

// Clear resets all history scores to zero
func (h *HistoryTable) Clear() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for i := 0; i < 64; i++ {
		for j := 0; j < 64; j++ {
			h.table[i][j] = 0
		}
	}
	h.age = 0
}

// Age applies decay to all history scores to prevent them from growing too large
// and to give more weight to recent patterns
func (h *HistoryTable) Age() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.age++

	// Apply decay every few ages to prevent scores from growing too large
	if h.age%8 == 0 {
		for i := 0; i < 64; i++ {
			for j := 0; j < 64; j++ {
				h.table[i][j] /= HistoryDecayFactor
			}
		}
	}
}

// squareToIndex converts a board square to an index (0-63)
func squareToIndex(square board.Square) int {
	return int(square.Rank*8 + square.File)
}

