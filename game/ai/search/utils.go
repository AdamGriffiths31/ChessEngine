// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// oppositePlayer returns the opposite player
func oppositePlayer(player moves.Player) moves.Player {
	if player == moves.White {
		return moves.Black
	}
	return moves.White
}
