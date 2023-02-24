package evaluate

import (
	"github.com/AdamGriffiths31/ChessEngine/data"
)

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

var pawnIsolated = -10
var pawnPassed = [8]int{0, 5, 10, 20, 35, 60, 100, 200}
var rookOpenFile = 10
var rookSemiOpenFile = 5
var queenOpenFile = 5
var queenSemiOpenFile = 3

func EvalPosistion(pos *data.Board) int {
	score := pos.Material[data.White] - pos.Material[data.Black]

	//Pawns
	piece := data.WP
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score += pawnTable[data.Sqaure120ToSquare64[sq]]

		if data.IsolatedMask[data.Sqaure120ToSquare64[sq]]&pos.Pawns[data.White] == 0 {
			score += pawnIsolated
		}

		if data.WhitePassedMask[data.Sqaure120ToSquare64[sq]]&pos.Pawns[data.Black] == 0 {
			score += pawnPassed[data.RanksBoard[sq]]
		}
	}
	piece = data.BP
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score -= pawnTable[data.Mirror64[data.Sqaure120ToSquare64[sq]]]

		if data.IsolatedMask[data.Sqaure120ToSquare64[sq]]&pos.Pawns[data.Black] == 0 {
			score -= pawnIsolated
		}

		if data.BlackPassedMask[data.Sqaure120ToSquare64[sq]]&pos.Pawns[data.White] == 0 {
			score -= pawnPassed[data.Rank8-data.RanksBoard[sq]]
		}
	}
	//Knights
	piece = data.WN
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score += knightTable[data.Sqaure120ToSquare64[sq]]
	}
	piece = data.BN
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score -= knightTable[data.Mirror64[data.Sqaure120ToSquare64[sq]]]
	}
	//Bishop
	piece = data.WB
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score += bishopTable[data.Sqaure120ToSquare64[sq]]
	}
	piece = data.BB
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score -= bishopTable[data.Mirror64[data.Sqaure120ToSquare64[sq]]]
	}
	//Rook
	piece = data.WR
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score += rookTable[data.Sqaure120ToSquare64[sq]]

		if pos.Pawns[data.Both]&data.FileBBMask[data.FilesBoard[sq]] == 0 {
			score += rookOpenFile
		} else if pos.Pawns[data.White]&data.FileBBMask[data.FilesBoard[sq]] == 0 {
			score += rookSemiOpenFile
		}
	}
	piece = data.BR
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score -= rookTable[data.Mirror64[data.Sqaure120ToSquare64[sq]]]

		if pos.Pawns[data.Both]&data.FileBBMask[data.FilesBoard[sq]] == 0 {
			score -= rookOpenFile
		} else if pos.Pawns[data.Black]&data.FileBBMask[data.FilesBoard[sq]] == 0 {
			score -= rookSemiOpenFile
		}
	}
	//Queen
	piece = data.WQ
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		if pos.Pawns[data.Both]&data.FileBBMask[data.FilesBoard[sq]] == 0 {
			score += queenOpenFile
		} else if pos.Pawns[data.White]&data.FileBBMask[data.FilesBoard[sq]] == 0 {
			score += queenSemiOpenFile
		}
	}

	piece = data.BQ
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		if pos.Pawns[data.Both]&data.FileBBMask[data.FilesBoard[sq]] == 0 {
			score -= queenOpenFile
		} else if pos.Pawns[data.Black]&data.FileBBMask[data.FilesBoard[sq]] == 0 {
			score -= queenSemiOpenFile
		}
	}

	if pos.Side == data.White {
		return score
	} else {
		return -score
	}
}
