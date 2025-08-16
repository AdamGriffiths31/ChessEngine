package board

// Position analysis utilities for move generation and validation

// IsSquareEmpty checks if a square is empty
func (b *Board) IsSquareEmpty(rank, file int) bool {
	return b.GetPiece(rank, file) == Empty
}

// IsSquareOccupied checks if a square is occupied by any piece
func (b *Board) IsSquareOccupied(rank, file int) bool {
	return b.GetPiece(rank, file) != Empty
}

// IsSquareOnBoard checks if coordinates are within board boundaries
func IsSquareOnBoard(rank, file int) bool {
	return rank >= 0 && rank <= 7 && file >= 0 && file <= 7
}

// IsWhitePiece checks if a piece belongs to white
func IsWhitePiece(piece Piece) bool {
	return piece >= 'A' && piece <= 'Z'
}

// IsBlackPiece checks if a piece belongs to black
func IsBlackPiece(piece Piece) bool {
	return piece >= 'a' && piece <= 'z'
}

// IsPawn checks if a piece is a pawn
func IsPawn(piece Piece) bool {
	return piece == WhitePawn || piece == BlackPawn
}

// IsKing checks if a piece is a king
func IsKing(piece Piece) bool {
	return piece == WhiteKing || piece == BlackKing
}

// GetPieceColor returns the color of a piece
func GetPieceColor(piece Piece) PieceColor {
	if piece == Empty {
		return NoneColor
	}
	if IsWhitePiece(piece) {
		return WhiteColor
	}
	return BlackColor
}

// PieceColor represents the color of a piece
type PieceColor int

const (
	// NoneColor represents no piece color (empty squares)
	NoneColor PieceColor = iota
	// WhiteColor represents white pieces
	WhiteColor
	// BlackColor represents black pieces
	BlackColor
)

// String returns string representation of piece color
func (pc PieceColor) String() string {
	switch pc {
	case WhiteColor:
		return "White"
	case BlackColor:
		return "Black"
	default:
		return "None"
	}
}

// CountPieces counts total pieces on the board
func (b *Board) CountPieces() int {
	count := 0
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			if b.IsSquareOccupied(rank, file) {
				count++
			}
		}
	}
	return count
}

// CountPiecesByColor counts pieces by color
func (b *Board) CountPiecesByColor(color PieceColor) int {
	count := 0
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if GetPieceColor(piece) == color {
				count++
			}
		}
	}
	return count
}

// FindKing finds the position of the king for a given color
func (b *Board) FindKing(color PieceColor) (rank, file int, found bool) {
	var kingPiece Piece
	if color == WhiteColor {
		kingPiece = WhiteKing
	} else {
		kingPiece = BlackKing
	}

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			if b.GetPiece(rank, file) == kingPiece {
				return rank, file, true
			}
		}
	}
	return 0, 0, false
}

// GetAllPiecesOfType returns all squares containing a specific piece type
func (b *Board) GetAllPiecesOfType(piece Piece) []Square {
	var squares []Square
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			if b.GetPiece(rank, file) == piece {
				squares = append(squares, Square{File: file, Rank: rank})
			}
		}
	}
	return squares
}

// GetAllPiecesOfColor returns all squares containing pieces of a specific color
func (b *Board) GetAllPiecesOfColor(color PieceColor) []Square {
	var squares []Square
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if GetPieceColor(piece) == color {
				squares = append(squares, Square{File: file, Rank: rank})
			}
		}
	}
	return squares
}
