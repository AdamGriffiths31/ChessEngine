package movegen

import "github.com/AdamGriffiths31/ChessEngine/internal/board"

// Player represents a chess player (White or Black).
type Player int

const (
	// White represents the white player
	White Player = iota
	// Black represents the black player
	Black
)

// String returns string representation of player
func (p Player) String() string {
	if p == White {
		return "White"
	}
	return "Black"
}

// MoveList represents a collection of chess moves with efficient storage.
// Should be obtained from GetMoveList() and released with ReleaseMoveList() for optimal performance.
type MoveList struct {
	Moves []board.Move
	Count int
}

// AddMove adds a move to the list
func (ml *MoveList) AddMove(move board.Move) {
	ml.Moves = append(ml.Moves, move)
	ml.Count++
}

// Clear empties the move list
func (ml *MoveList) Clear() {
	ml.Moves = ml.Moves[:0]
	ml.Count = 0
}
