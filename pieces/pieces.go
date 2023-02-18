package pieces

// Piece represents a chess piece.
type Piece string

var display = map[Piece]string{
	"":  " ",
	"B": "♝",
	"K": "♚",
	"N": "♞",
	"P": "♟",
	"Q": "♛",
	"R": "♜",
	"b": "♗",
	"k": "♔",
	"n": "♘",
	"p": "♙",
	"q": "♕",
	"r": "♖",
}

// Display returns the ASCII representation of the piece.
func (p Piece) Display() string {
	return display[p]
}

func DisplayRow(row [8]string) [8]Piece {
	var pieces [8]Piece
	for i, s := range row {
		pieces[i] = Piece(s)
	}
	return pieces
}
