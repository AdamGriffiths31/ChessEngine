package engine

var pawnTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	10, 10, 0, -10, -10, 0, 10, 10,
	5, 0, 0, 5, 5, 0, 0, 5,
	0, 0, 10, 20, 20, 10, 0, 0,
	5, 5, 5, 10, 10, 5, 5, 5,
	10, 10, 10, 20, 20, 10, 10, 10,
	20, 20, 20, 30, 30, 20, 20, 20,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightTable = [64]int{
	0, -10, 0, 0, 0, 0, -10, 0,
	0, 0, 0, 5, 5, 0, 0, 0,
	0, 0, 10, 10, 10, 10, 0, 0,
	0, 0, 10, 20, 20, 10, 5, 0,
	5, 10, 15, 20, 20, 15, 10, 5,
	5, 10, 10, 20, 20, 10, 10, 5,
	0, 0, 5, 10, 10, 5, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var bishopTable = [64]int{
	0, 0, -10, 0, 0, -10, 0, 0,
	0, 0, 0, 10, 10, 0, 0, 0,
	0, 0, 10, 15, 15, 10, 0, 0,
	0, 10, 15, 20, 20, 15, 10, 0,
	0, 10, 15, 20, 20, 15, 10, 0,
	0, 0, 10, 15, 15, 10, 0, 0,
	0, 0, 0, 10, 10, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var rookTable = [64]int{
	0, 0, 5, 10, 10, 5, 0, 0,
	0, 0, 5, 10, 10, 5, 0, 0,
	0, 0, 5, 10, 10, 5, 0, 0,
	0, 0, 5, 10, 10, 5, 0, 0,
	0, 0, 5, 10, 10, 5, 0, 0,
	0, 0, 5, 10, 10, 5, 0, 0,
	25, 25, 25, 25, 25, 25, 25, 25,
	0, 0, 5, 10, 10, 5, 0, 0,
}

var mirror64 = [64]int{
	56, 57, 58, 59, 60, 61, 62, 63,
	48, 49, 50, 51, 52, 53, 54, 55,
	40, 41, 42, 43, 44, 45, 46, 47,
	32, 33, 34, 35, 36, 37, 38, 39,
	24, 25, 26, 27, 28, 29, 30, 31,
	16, 17, 18, 19, 20, 21, 22, 23,
	8, 9, 10, 11, 12, 13, 14, 15,
	0, 1, 2, 3, 4, 5, 6, 7,
}

func EvalPosistion(pos *Board) int {
	score := pos.Material[White] - pos.Material[Black]

	//Pawns
	piece := WP
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score += pawnTable[Sqaure120ToSquare64[sq]]
	}
	piece = BP
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score -= pawnTable[mirror64[Sqaure120ToSquare64[sq]]]
	}
	//Knights
	piece = WN
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score += knightTable[Sqaure120ToSquare64[sq]]
	}
	piece = BN
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score -= knightTable[mirror64[Sqaure120ToSquare64[sq]]]
	}
	//Bishop
	piece = WB
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score += bishopTable[Sqaure120ToSquare64[sq]]
	}
	piece = BB
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score -= bishopTable[mirror64[Sqaure120ToSquare64[sq]]]
	}
	//Rook
	piece = WR
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score += rookTable[Sqaure120ToSquare64[sq]]
	}
	piece = BR
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score -= rookTable[mirror64[Sqaure120ToSquare64[sq]]]
	}

	if pos.Side == White {
		return score
	} else {
		return -score
	}
}
