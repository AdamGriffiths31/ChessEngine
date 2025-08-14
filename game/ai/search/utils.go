// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
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

// moveToDebugString converts a move to string for debugging (handles invalid moves)
func moveToDebugString(move board.Move) string {
	if move.From.File < 0 || move.From.File > 7 || move.From.Rank < 0 || move.From.Rank > 7 ||
		move.To.File < 0 || move.To.File > 7 || move.To.Rank < 0 || move.To.Rank > 7 {
		return "INVALID"
	}
	from := string('a'+rune(move.From.File)) + string('1'+rune(move.From.Rank))
	to := string('a'+rune(move.To.File)) + string('1'+rune(move.To.Rank))
	return from + to
}
