package moveGen

import (
	"github.com/AdamGriffiths31/ChessEngine/attack"
	"github.com/AdamGriffiths31/ChessEngine/data"
)

func generateWhiteCastleMoves(pos *data.Board, moveList *data.MoveList) {
	if (pos.CastlePermission & data.WhiteKingCastle) != 0 {
		if pos.Pieces[data.F1] == data.Empty && pos.Pieces[data.G1] == data.Empty {
			if !attack.SquareAttacked(data.E1, data.Black, pos) && !attack.SquareAttacked(data.F1, data.Black, pos) {
				addQuiteMove(pos, MakeMoveInt(data.E1, data.G1, data.Empty, data.Empty, data.MFLAGGCA), moveList)
			}
		}
	}
	if (pos.CastlePermission & data.WhiteQueenCastle) != 0 {
		if pos.Pieces[data.D1] == data.Empty && pos.Pieces[data.C1] == data.Empty && pos.Pieces[data.B1] == data.Empty {
			if !attack.SquareAttacked(data.E1, data.Black, pos) && !attack.SquareAttacked(data.D1, data.Black, pos) {
				addQuiteMove(pos, MakeMoveInt(data.E1, data.C1, data.Empty, data.Empty, data.MFLAGGCA), moveList)
			}
		}
	}
}

func generateBlackCastleMoves(pos *data.Board, moveList *data.MoveList) {
	if (pos.CastlePermission & data.BlackKingCastle) != 0 {
		if pos.Pieces[data.F8] == data.Empty && pos.Pieces[data.G8] == data.Empty {
			if !attack.SquareAttacked(data.E8, data.White, pos) && !attack.SquareAttacked(data.F8, data.White, pos) {
				addQuiteMove(pos, MakeMoveInt(data.E8, data.G8, data.Empty, data.Empty, data.MFLAGGCA), moveList)
			}
		}
	}
	if (pos.CastlePermission & data.BlackQueenCastle) != 0 {
		if pos.Pieces[data.D8] == data.Empty && pos.Pieces[data.C8] == data.Empty && pos.Pieces[data.B8] == data.Empty {
			if !attack.SquareAttacked(data.E8, data.White, pos) && !attack.SquareAttacked(data.D8, data.White, pos) {
				addQuiteMove(pos, MakeMoveInt(data.E8, data.C8, data.Empty, data.Empty, data.MFLAGGCA), moveList)
			}
		}
	}
}
