package engine

//SqaureOnBoard Check the sq is not off board
func SqaureOnBoard(sq int) bool {
	return FilesBoard[sq] != OffBoard
}

//SideValid Check the side is either white or black
func SideValid(side int) bool {
	return side == White || side == Black
}

//FileRankValid Check the file or rank is valid
func FileRankValid(fileRank int) bool {
	return fileRank >= 0 && fileRank <= 7
}

//PieceValidOrEmpty Check piece is valid or empty
func PieceValidOrEmpty(piece int) bool {
	return piece >= Empty && piece <= BK
}

//PieceValid Checks the piece is valid
func PieceValid(piece int) bool {
	return piece >= WP && piece <= BK
}
