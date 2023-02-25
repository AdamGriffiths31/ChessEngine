package moveGen

import (
	"math/bits"

	"github.com/AdamGriffiths31/ChessEngine/data"
)

func generateWhitePawnMoves(pos *data.Board, moveList *data.MoveList) {
	generateWhitePawnEnPassantMoves(pos, moveList)
	generateWhitePawnCaptureMoves(pos, moveList)
	generateWhitePawnQuietMoves(pos, moveList)
}

func generateBlackPawnMoves(pos *data.Board, moveList *data.MoveList) {
	generateBlackPawnQuietMoves(pos, moveList)
	generateBlackPawnCaptureMoves(pos, moveList)
	generateBlackPawnEnPassantMoves(pos, moveList)
}

func generateBlackPawnQuietMoves(pos *data.Board, moveList *data.MoveList) {
	pawns := pos.Pawns[data.Black]
	for pawns != 0 {
		sq := bits.TrailingZeros64(pawns)
		pawns ^= (1 << sq)
		sq = data.Sqaure64ToSquare120[sq]

		// Check for single pawn push
		to := sq - 10
		if pos.Pieces[to] == data.Empty {
			addBlackPawnMove(pos, sq, to, moveList)

			// Check for double pawn push
			if data.RanksBoard[sq] == data.Rank7 && pos.Pieces[to-10] == data.Empty {
				addQuiteMove(pos, MakeMoveInt(sq, to-10, data.Empty, data.Empty, data.MFLAGPS), moveList)
			}
		}
	}
}

func generateBlackPawnCaptureMoves(pos *data.Board, moveList *data.MoveList) {
	blackPawns := pos.Pawns[data.Black]
	whitePieces := pos.WhitePiecesBB

	// Generate attacks to the left
	leftAttacks := (blackPawns >> 9) &^ data.FileBBMask[data.FileH] & whitePieces
	for leftAttacks != 0 {
		sq := bits.TrailingZeros64(leftAttacks)
		leftAttacks &= leftAttacks - 1
		sq = data.Sqaure64ToSquare120[sq]
		addBlackPawnCaptureMove(pos, moveList, sq+11, sq, pos.Pieces[sq])
	}

	// Generate attacks to the right
	rightAttacks := (blackPawns >> 7) &^ data.FileBBMask[data.FileA] & whitePieces
	for rightAttacks != 0 {
		sq := bits.TrailingZeros64(rightAttacks)
		rightAttacks &= rightAttacks - 1
		sq = data.Sqaure64ToSquare120[sq]
		addBlackPawnCaptureMove(pos, moveList, sq+9, sq, pos.Pieces[sq])
	}
}

func generateBlackPawnEnPassantMoves(pos *data.Board, moveList *data.MoveList) {
	if pos.EnPas != data.NoSqaure {
		pawns := pos.Pawns[data.Black] &^ (data.RankBBMask[data.Rank1] | data.RankBBMask[data.Rank2])
		leftAttacks := (pawns >> 9) &^ data.FileBBMask[data.FileH] & data.SquareBB[data.Sqaure120ToSquare64[pos.EnPas]]
		rightAttacks := (pawns >> 7) &^ data.FileBBMask[data.FileA] & data.SquareBB[data.Sqaure120ToSquare64[pos.EnPas]]
		if leftAttacks != 0 {
			sq := bits.TrailingZeros64(leftAttacks)
			sq = data.Sqaure64ToSquare120[sq]
			addEnPasMove(pos, MakeMoveInt(sq+11, pos.EnPas, data.Empty, data.Empty, data.MFLAGEP), moveList)
		}

		if rightAttacks != 0 {
			sq := bits.TrailingZeros64(rightAttacks)
			sq = data.Sqaure64ToSquare120[sq]
			addEnPasMove(pos, MakeMoveInt(sq+9, pos.EnPas, data.Empty, data.Empty, data.MFLAGEP), moveList)
		}
	}
}

func generateWhitePawnQuietMoves(pos *data.Board, moveList *data.MoveList) {
	pawns := pos.Pawns[data.White]
	for pawns != 0 {
		sq := bits.TrailingZeros64(pawns)
		pawns ^= (1 << sq)
		sq = data.Sqaure64ToSquare120[sq]

		// Check for single pawn push
		to := sq + 10
		if pos.Pieces[to] == data.Empty {
			addWhitePawnMove(pos, sq, to, moveList)

			// Check for double pawn push
			if data.RanksBoard[sq] == data.Rank2 && pos.Pieces[to+10] == data.Empty {
				addQuiteMove(pos, MakeMoveInt(sq, to+10, data.Empty, data.Empty, data.MFLAGPS), moveList)
			}
		}
	}
}

func generateWhitePawnCaptureMoves(pos *data.Board, moveList *data.MoveList) {
	whitePawns := pos.Pawns[data.White]
	blackPieces := pos.ColoredPiecesBB

	// Generate attacks to the left
	leftAttacks := (whitePawns << 7) &^ data.FileBBMask[data.FileH] & blackPieces

	for leftAttacks != 0 {
		sq := bits.TrailingZeros64(leftAttacks)
		leftAttacks &= leftAttacks - 1
		sq = data.Sqaure64ToSquare120[sq]
		addWhitePawnCaptureMove(pos, moveList, sq-9, sq, pos.Pieces[sq])
	}

	// Generate attacks to the right
	rightAttacks := (whitePawns << 9) &^ data.FileBBMask[data.FileA] & blackPieces
	for rightAttacks != 0 {
		sq := bits.TrailingZeros64(rightAttacks)
		rightAttacks &= rightAttacks - 1
		sq = data.Sqaure64ToSquare120[sq]
		addWhitePawnCaptureMove(pos, moveList, sq-11, sq, pos.Pieces[sq])
	}
}

func generateWhitePawnEnPassantMoves(pos *data.Board, moveList *data.MoveList) {
	if pos.EnPas != data.NoSqaure {
		pawns := pos.Pawns[data.White] &^ (data.RankBBMask[data.Rank8] | data.RankBBMask[data.Rank7])
		leftAttacks := (pawns << 7) &^ data.FileBBMask[data.FileH] & data.SquareBB[data.Sqaure120ToSquare64[pos.EnPas]]
		rightAttacks := (pawns << 9) &^ data.FileBBMask[data.FileA] & data.SquareBB[data.Sqaure120ToSquare64[pos.EnPas]]

		if leftAttacks != 0 {
			sq := bits.TrailingZeros64(leftAttacks)
			sq = data.Sqaure64ToSquare120[sq]
			addEnPasMove(pos, MakeMoveInt(sq-9, pos.EnPas, data.Empty, data.Empty, data.MFLAGEP), moveList)
		}

		if rightAttacks != 0 {
			sq := bits.TrailingZeros64(rightAttacks)
			sq = data.Sqaure64ToSquare120[sq]
			addEnPasMove(pos, MakeMoveInt(sq-11, pos.EnPas, data.Empty, data.Empty, data.MFLAGEP), moveList)
		}
	}
}
