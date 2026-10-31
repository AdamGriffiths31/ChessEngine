package search

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func (m *MinimaxEngine) isKillerMove(move board.Move, depth int) bool {
	killers := &m.searchState.killerTable[depth]
	return sameMove(&move, &killers[0]) || sameMove(&move, &killers[1])
}

func (m *MinimaxEngine) storeKiller(move board.Move, depth int) {
	if m.isKillerMove(move, depth) {
		return
	}

	m.searchState.killerTable[depth][1] = m.searchState.killerTable[depth][0]
	m.searchState.killerTable[depth][0] = move
}
