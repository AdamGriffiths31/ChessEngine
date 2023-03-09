package moveGen

import (
	"math/bits"

	"github.com/AdamGriffiths31/ChessEngine/data"
)

func generateSliderMoves(pos *data.Board, moveList *data.MoveList, includeQuite bool) {
	pieceIndex := data.LoopSlideIndex[pos.Side]
	piece := data.LoopSlidePiece[pieceIndex]
	pieceIndex++

	for piece != 0 {
		for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
			sq := pos.PieceList[piece][pieceNum]
			//White
			if piece == data.WR || piece == data.WQ {
				attack := data.GetRookAttacks(pos.PiecesBB, data.Square120ToSquare64[sq])
				generateMoves(pos, moveList, includeQuite, sq, attack, pos.ColoredPiecesBB)
			}
			if piece == data.WB || piece == data.WQ {
				attack := data.GetBishopAttacks(pos.PiecesBB, data.Square120ToSquare64[sq])
				generateMoves(pos, moveList, includeQuite, sq, attack, pos.ColoredPiecesBB)
			}
			//Black
			if piece == data.BR || piece == data.BQ {
				attack := data.GetRookAttacks(pos.PiecesBB, data.Square120ToSquare64[sq])
				generateMoves(pos, moveList, includeQuite, sq, attack, pos.WhitePiecesBB)
			}
			if piece == data.BB || piece == data.BQ {
				attack := data.GetBishopAttacks(pos.PiecesBB, data.Square120ToSquare64[sq])
				generateMoves(pos, moveList, includeQuite, sq, attack, pos.WhitePiecesBB)
			}
		}
		piece = data.LoopSlidePiece[pieceIndex]
		pieceIndex++
	}
}

func generateMoves(pos *data.Board, moveList *data.MoveList, includeQuite bool, square int, attackBB, oppositeBB uint64) {
	captures := attackBB & oppositeBB
	for captures != 0 {
		sq := bits.TrailingZeros64(captures)
		captures &= captures - 1
		sq = data.Square64ToSquare120[sq]
		addCaptureMove(pos, MakeMoveInt(square, sq, pos.Pieces[sq], data.Empty, 0), moveList)
	}

	if includeQuite {
		quite := attackBB &^ pos.PiecesBB
		for quite != 0 {
			sq := bits.TrailingZeros64(quite)
			quite &= quite - 1
			sq = data.Square64ToSquare120[sq]
			addQuiteMove(pos, MakeMoveInt(square, sq, data.Empty, data.Empty, 0), moveList)
		}
	}
}
