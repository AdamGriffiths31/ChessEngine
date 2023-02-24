package moveGen

import (
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/validate"
)

func generateSliderMoves(pos *data.Board, moveList *data.MoveList, includeQuite bool) {
	pieceIndex := data.LoopSlideIndex[pos.Side]
	piece := data.LoopSlidePiece[pieceIndex]
	pieceIndex++

	for piece != 0 {
		for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
			sq := pos.PieceList[piece][pieceNum]
			for i := 0; i < data.NumDir[piece]; i++ {
				dir := data.PieceDir[piece][i]
				tempSq := sq + dir

				for validate.SqaureOnBoard(tempSq) {
					if pos.Pieces[tempSq] != data.Empty {
						if data.PieceCol[pos.Pieces[tempSq]] == pos.Side^1 {
							addCaptureMove(pos, MakeMoveInt(sq, tempSq, pos.Pieces[tempSq], data.Empty, 0), moveList)
						}
						break
					}
					if includeQuite {
						addQuiteMove(pos, MakeMoveInt(sq, tempSq, data.Empty, data.Empty, 0), moveList)
					}
					tempSq += dir
				}
			}
		}
		piece = data.LoopSlidePiece[pieceIndex]
		pieceIndex++
	}
}
