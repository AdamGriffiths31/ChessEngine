package moveGen

import (
	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/validate"
)

func generateWhitePawnMoves(pos *data.Board, moveList *data.MoveList) {
	for pieceNum := 0; pieceNum < pos.PieceNumber[data.WP]; pieceNum++ {
		sq := pos.PieceList[data.WP][pieceNum]
		generateWhitePawnQuietMoves(pos, sq, moveList)
		generateWhitePawnCaptureMoves(pos, sq, moveList)
		generateWhitePawnEnPassantMoves(pos, sq, moveList)
	}
}

func generateBlackPawnMoves(pos *data.Board, moveList *data.MoveList) {
	for pieceNum := 0; pieceNum < pos.PieceNumber[data.BP]; pieceNum++ {
		sq := pos.PieceList[data.BP][pieceNum]
		generateBlackPawnQuietMoves(pos, sq, moveList)
		generateBlackPawnCaptureMoves(pos, sq, moveList)
		generateBlackPawnEnPassantMoves(pos, sq, moveList)
	}
}

func generateWhitePawnQuietMoves(pos *data.Board, sq int, moveList *data.MoveList) {
	if pos.Pieces[sq+10] == data.Empty {
		addWhitePawnMove(pos, sq, sq+10, moveList)
		if data.RanksBoard[sq] == data.Rank2 && pos.Pieces[sq+20] == data.Empty {
			addQuiteMove(pos, MakeMoveInt(sq, sq+20, data.Empty, data.Empty, data.MFLAGPS), moveList)
		}
	}
}

func generateWhitePawnCaptureMoves(pos *data.Board, sq int, moveList *data.MoveList) {
	if validate.SqaureOnBoard(sq+9) && data.PieceCol[pos.Pieces[sq+9]] == data.Black {
		addWhitePawnCaptureMove(pos, moveList, sq, sq+9, pos.Pieces[sq+9])
	}

	if validate.SqaureOnBoard(sq+11) && data.PieceCol[pos.Pieces[sq+11]] == data.Black {
		addWhitePawnCaptureMove(pos, moveList, sq, sq+11, pos.Pieces[sq+11])
	}
}

func generateWhitePawnEnPassantMoves(pos *data.Board, sq int, moveList *data.MoveList) {
	if pos.EnPas != data.NoSqaure {
		if sq+9 == pos.EnPas {
			addEnPasMove(pos, MakeMoveInt(sq, sq+9, data.Empty, data.Empty, data.MFLAGEP), moveList)
		}

		if sq+11 == pos.EnPas {
			addEnPasMove(pos, MakeMoveInt(sq, sq+11, data.Empty, data.Empty, data.MFLAGEP), moveList)
		}
	}
}

func generateBlackPawnQuietMoves(pos *data.Board, sq int, moveList *data.MoveList) {
	if pos.Pieces[sq-10] == data.Empty {
		addBlackPawnMove(pos, sq, sq-10, moveList)
		if data.RanksBoard[sq] == data.Rank7 && pos.Pieces[sq-20] == data.Empty {
			addQuiteMove(pos, MakeMoveInt(sq, sq-20, data.Empty, data.Empty, data.MFLAGPS), moveList)
		}
	}
}

func generateBlackPawnCaptureMoves(pos *data.Board, sq int, moveList *data.MoveList) {
	if validate.SqaureOnBoard(sq-9) && data.PieceCol[pos.Pieces[sq-9]] == data.White {
		addBlackPawnCaptureMove(pos, moveList, sq, sq-9, pos.Pieces[sq-9])
	}

	if validate.SqaureOnBoard(sq-11) && data.PieceCol[pos.Pieces[sq-11]] == data.White {
		addBlackPawnCaptureMove(pos, moveList, sq, sq-11, pos.Pieces[sq-11])
	}
}

func generateBlackPawnEnPassantMoves(pos *data.Board, sq int, moveList *data.MoveList) {
	if pos.EnPas != data.NoSqaure {
		if sq-9 == pos.EnPas {
			addEnPasMove(pos, MakeMoveInt(sq, sq-9, data.Empty, data.Empty, data.MFLAGEP), moveList)
		}

		if sq-11 == pos.EnPas {
			addEnPasMove(pos, MakeMoveInt(sq, sq-11, data.Empty, data.Empty, data.MFLAGEP), moveList)
		}
	}
}
