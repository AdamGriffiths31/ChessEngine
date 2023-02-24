package moveGen

import (
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/validate"
)

func generateNonSliderMoves(pos *data.Board, moveList *data.MoveList, includeQuite bool) {
	pieceIndex := data.LoopNonSlideIndex[pos.Side]
	piece := data.LoopNonSlidePiece[pieceIndex]
	pieceIndex++

	for piece != 0 {
		for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
			sq := pos.PieceList[piece][pieceNum]
			for i := 0; i < data.NumDir[piece]; i++ {
				dir := data.PieceDir[piece][i]
				tempSq := sq + dir

				if !validate.SqaureOnBoard(tempSq) {
					continue
				}

				if pos.Pieces[tempSq] != data.Empty {
					if data.PieceCol[pos.Pieces[tempSq]] == pos.Side^1 {
						addCaptureMove(pos, MakeMoveInt(sq, tempSq, pos.Pieces[tempSq], data.Empty, 0), moveList)
					}
					continue
				}
				if includeQuite {
					addQuiteMove(pos, MakeMoveInt(sq, tempSq, data.Empty, data.Empty, 0), moveList)
				}
			}
		}
		piece = data.LoopNonSlidePiece[pieceIndex]
		pieceIndex++
	}
}
