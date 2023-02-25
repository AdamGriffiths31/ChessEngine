package moveGen

import (
	"fmt"
	"math/bits"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
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

func generateRookMoves(pos *data.Board, moveList *data.MoveList, includeQuite bool) {
	for pieceNum := 0; pieceNum < pos.PieceNumber[data.WR]; pieceNum++ {
		sqPiece := pos.PieceList[data.WR][pieceNum]
		fmt.Printf("Square  %v\n", io.SqaureString(sqPiece))

		rookBB := data.GetRookAttacks(pos.PiecesBB, data.Sqaure120ToSquare64[sqPiece])

		captures := rookBB & pos.ColoredPiecesBB
		for captures != 0 {
			sq := bits.TrailingZeros64(captures)
			captures &= captures - 1
			sq = data.Sqaure64ToSquare120[sq]
			fmt.Printf("Capture on %v\n", io.SqaureString(sq))
			addCaptureMove(pos, MakeMoveInt(sqPiece, sq, pos.Pieces[sq], data.Empty, 0), moveList)
		}

		if includeQuite {
			quite := rookBB &^ pos.PiecesBB
			for quite != 0 {
				sq := bits.TrailingZeros64(quite)
				quite &= quite - 1
				sq = data.Sqaure64ToSquare120[sq]
				fmt.Printf("quite on %v\n", io.SqaureString(sq))
				addQuiteMove(pos, MakeMoveInt(sqPiece, sq, data.Empty, data.Empty, 0), moveList)
			}
		}
	}
}
