package evaluate

import (
	"math"

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

var kingE = [64]int{
	-50, -10, 0, 0, 0, 0, -10, -50,
	-10, 0, 10, 10, 10, 10, 0, -10,
	0, 10, 20, 20, 20, 20, 10, 0,
	0, 10, 20, 40, 40, 20, 10, 0,
	0, 10, 20, 40, 40, 20, 10, 0,
	0, 10, 20, 20, 20, 20, 10, 0,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-50, -10, 0, 0, 0, 0, -10, -50,
}

var kingO = [64]int{
	0, 5, 5, -10, -10, 0, 10, 5,
	-30, -30, -30, -30, -30, -30, -30, -30,
	-50, -50, -50, -50, -50, -50, -50, -50,
	-70, -70, -70, -70, -70, -70, -70, -70,
	-70, -70, -70, -70, -70, -70, -70, -70,
	-70, -70, -70, -70, -70, -70, -70, -70,
	-70, -70, -70, -70, -70, -70, -70, -70,
	-70, -70, -70, -70, -70, -70, -70, -70,
}

var pawnIsolated = -10
var pawnPassed = [8]int{0, 5, 10, 20, 35, 60, 100, 200}
var rookOpenFile = 10
var rookSemiOpenFile = 5
var queenOpenFile = 5
var queenSemiOpenFile = 3
var bishopPair = 30

func EvalPosistion(pos *data.Board) int {

	if pos.PieceNumber[data.WP] == 0 && pos.PieceNumber[data.BP] == 0 && materialDraw(pos) {
		return 0
	}

	score := pos.Material[data.White] - pos.Material[data.Black]

	//Pawns
	piece := data.WP
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score += pawnTable[data.Square120ToSquare64[sq]]

		if data.IsolatedMask[data.Square120ToSquare64[sq]]&pos.Pawns[data.White] == 0 {
			score += pawnIsolated
		}

		if data.WhitePassedMask[data.Square120ToSquare64[sq]]&pos.Pawns[data.Black] == 0 {
			score += pawnPassed[data.RanksBoard[sq]]
		}
	}
	piece = data.BP
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score -= pawnTable[data.Mirror64[data.Square120ToSquare64[sq]]]

		if data.IsolatedMask[data.Square120ToSquare64[sq]]&pos.Pawns[data.Black] == 0 {
			score -= pawnIsolated
		}

		if data.BlackPassedMask[data.Square120ToSquare64[sq]]&pos.Pawns[data.White] == 0 {
			score -= pawnPassed[data.Rank8-data.RanksBoard[sq]]
		}
	}
	//Knights
	piece = data.WN
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score += knightTable[data.Square120ToSquare64[sq]]
	}
	piece = data.BN
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score -= knightTable[data.Mirror64[data.Square120ToSquare64[sq]]]
	}
	//Bishop
	piece = data.WB
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score += bishopTable[data.Square120ToSquare64[sq]]
	}
	piece = data.BB
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score -= bishopTable[data.Mirror64[data.Square120ToSquare64[sq]]]
	}

	if pos.PieceNumber[data.WB] >= 2 {
		score += bishopPair
	}
	if pos.PieceNumber[data.BB] >= 2 {
		score -= bishopPair
	}
	//Rook
	piece = data.WR
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score += rookTable[data.Square120ToSquare64[sq]]

		if pos.Pawns[data.Both]&data.FileBBMask[data.FilesBoard[sq]] == 0 {
			score += rookOpenFile
		} else if pos.Pawns[data.White]&data.FileBBMask[data.FilesBoard[sq]] == 0 {
			score += rookSemiOpenFile
		}
	}
	piece = data.BR
	for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
		sq := pos.PieceList[piece][pieceNum]
		score -= rookTable[data.Mirror64[data.Square120ToSquare64[sq]]]

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
	//King
	piece = data.WK
	sq := pos.PieceList[piece][0]
	if pos.Material[data.Black] <= isEndGame() {
		score += kingE[data.Square120ToSquare64[sq]]
	} else {
		score += kingO[data.Square120ToSquare64[sq]]
	}
	piece = data.BK
	sq = pos.PieceList[piece][0]
	if pos.Material[data.White] <= isEndGame() {
		score -= kingE[data.Mirror64[data.Square120ToSquare64[sq]]]
	} else {
		score -= kingO[data.Mirror64[data.Square120ToSquare64[sq]]]
	}

	if pos.Side == data.White {
		return score
	} else {
		return -score
	}
}

func materialDraw(pos *data.Board) bool {
	if pos.PieceNumber[data.WR] == 0 && pos.PieceNumber[data.BR] == 0 && pos.PieceNumber[data.WQ] == 0 && pos.PieceNumber[data.BQ] == 0 {
		if pos.PieceNumber[data.WB] == 0 && pos.PieceNumber[data.BB] == 0 {
			if pos.PieceNumber[data.WN] < 3 && pos.PieceNumber[data.BN] < 3 {
				return true
			}
		} else if pos.PieceNumber[data.WN] == 0 && pos.PieceNumber[data.BN] == 0 {
			if math.Abs(float64(pos.PieceNumber[data.WB]-pos.PieceNumber[data.BB])) < 2 {
				return true
			}
		} else if (pos.PieceNumber[data.WN] < 3 && pos.PieceNumber[data.WB] == 0) || (pos.PieceNumber[data.WB] == 1 && pos.PieceNumber[data.WN] == 0) {
			if (pos.PieceNumber[data.BN] < 3 && pos.PieceNumber[data.BB] == 0) || (pos.PieceNumber[data.BB] == 1 && pos.PieceNumber[data.BN] == 0) {
				return true
			}
		}
	} else if pos.PieceNumber[data.WQ] == 0 && pos.PieceNumber[data.BQ] == 0 {
		if pos.PieceNumber[data.WR] == 1 && pos.PieceNumber[data.BR] == 1 {
			if (pos.PieceNumber[data.WN]+pos.PieceNumber[data.WB]) < 2 && (pos.PieceNumber[data.BN]+pos.PieceNumber[data.BB]) < 2 {
				return true
			}
		} else if pos.PieceNumber[data.WR] == 1 && pos.PieceNumber[data.BR] == 0 {
			if (pos.PieceNumber[data.WN]+pos.PieceNumber[data.WB] == 0) && (((pos.PieceNumber[data.BN] + pos.PieceNumber[data.BB]) == 1) || ((pos.PieceNumber[data.BN] + pos.PieceNumber[data.BB]) == 2)) {
				return true
			}
		} else if pos.PieceNumber[data.BR] == 1 && pos.PieceNumber[data.WR] == 0 {
			if (pos.PieceNumber[data.BN]+pos.PieceNumber[data.BB] == 0) && (((pos.PieceNumber[data.WN] + pos.PieceNumber[data.WB]) == 1) || ((pos.PieceNumber[data.WN] + pos.PieceNumber[data.WB]) == 2)) {
				return true
			}
		}
	}
	return false
}

func isEndGame() int {
	return (1 * data.PieceVal[data.WR]) + (2 * data.PieceVal[data.WN]) + (2 * data.PieceVal[data.WP]) + data.PieceVal[data.WK]
}
