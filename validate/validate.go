package validate

import (
	"github.com/AdamGriffiths31/ChessEngine/data"
)

// SquareOnBoard Check the sq is not off board
func SquareOnBoard(sq int) bool {
	return data.FilesBoard[sq] != data.OffBoard
}

// SideValid Check the side is either white or black
func SideValid(side int) bool {
	return side == data.White || side == data.Black
}

// FileRankValid Check the file or rank is valid
func FileRankValid(fileRank int) bool {
	return fileRank >= 0 && fileRank <= 7
}

// PieceValidOrEmpty Check piece is valid or empty
func PieceValidOrEmpty(piece int) bool {
	return piece >= data.Empty && piece <= data.BK
}

// PieceValid Checks the piece is valid
func PieceValid(piece int) bool {
	return piece >= data.WP && piece <= data.BK
}
