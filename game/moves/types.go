package moves

import "github.com/AdamGriffiths31/ChessEngine/board"

// Player represents a chess player (White or Black).
type Player int

const (
	White Player = iota
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

// NewMoveList creates a new empty move list with pre-allocated capacity.
// Consider using GetMoveList() from the pool for better performance.
func NewMoveList() *MoveList {
	return &MoveList{
		Moves: make([]board.Move, 0, InitialMoveListCapacity), // Pre-allocate for performance
		Count: 0,
	}
}

// AddMove adds a move to the list
func (ml *MoveList) AddMove(move board.Move) {
	ml.Moves = append(ml.Moves, move)
	ml.Count++
}

// Contains checks if the move list contains a specific move
func (ml *MoveList) Contains(move board.Move) bool {
	for _, m := range ml.Moves {
		if MovesEqual(m, move) {
			return true
		}
	}
	return false
}

// Clear empties the move list
func (ml *MoveList) Clear() {
	ml.Moves = ml.Moves[:0]
	ml.Count = 0
}

// MoveHistory stores information needed to undo a move
type MoveHistory struct {
	Move             board.Move
	CapturedPiece    board.Piece
	CastlingRights   string
	EnPassantTarget  *board.Square
	HalfMoveClock    int
	FullMoveNumber   int
	WasEnPassant     bool
	WasCastling      bool
}

// MovesEqual compares two moves for equality
func MovesEqual(a, b board.Move) bool {
	return a.From.File == b.From.File &&
		a.From.Rank == b.From.Rank &&
		a.To.File == b.To.File &&
		a.To.Rank == b.To.Rank &&
		a.Promotion == b.Promotion
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}