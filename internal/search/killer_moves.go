package search

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func isValidKillerDepth(depth int) bool {
	return depth >= 0 && depth < MaxKillerDepth
}

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
