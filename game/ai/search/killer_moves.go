// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// isValidKillerDepth checks if depth is within valid bounds for killer move table
func isValidKillerDepth(depth int) bool {
	return depth >= 0 && depth < MaxKillerDepth
}

// isKillerMove checks if a move matches one of the stored killer moves at the given depth.
// Killer moves are quiet moves that previously caused beta cutoffs and are likely to be good
// in similar positions, improving move ordering without requiring expensive evaluation.
func (m *MinimaxEngine) isKillerMove(move board.Move, depth int) bool {
	if !isValidKillerDepth(depth) {
		return false
	}

	return (move.From == m.searchState.killerTable[depth][0].From &&
		move.To == m.searchState.killerTable[depth][0].To &&
		move.Promotion == m.searchState.killerTable[depth][0].Promotion) ||
		(move.From == m.searchState.killerTable[depth][1].From &&
			move.To == m.searchState.killerTable[depth][1].To &&
			move.Promotion == m.searchState.killerTable[depth][1].Promotion)
}

// storeKiller stores a killer move at the given depth using a two-slot replacement strategy.
// The most recent killer move replaces slot 0, pushing the previous slot 0 move to slot 1.
// This maintains the two most recent successful quiet moves at each depth for move ordering.
func (m *MinimaxEngine) storeKiller(move board.Move, depth int) {
	if !isValidKillerDepth(depth) {
		return
	}

	if m.isKillerMove(move, depth) {
		return
	}

	m.searchState.killerTable[depth][1] = m.searchState.killerTable[depth][0]
	m.searchState.killerTable[depth][0] = move
}
