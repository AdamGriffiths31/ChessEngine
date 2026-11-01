package movegen

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

func filterLegalMovesInPlace(b *board.Board, side *sideInfo, moves *MoveList) {
	ourKingBitboard := b.GetPieceBitboard(side.king)
	if ourKingBitboard == 0 {
		moves.Clear()
		return
	}
	kingSquare := ourKingBitboard.LSB()

	enemy := &sides[side.enemy]
	pinnedPieces := calculatePinnedPieces(b, kingSquare, enemy)
	checkers := b.GetAttackersToSquare(kingSquare, side.enemy)
	evasions := checkEvasions(checkers, kingSquare)

	writeIndex := 0
	for readIndex := range moves.Count {
		move := moves.Moves[readIndex]

		if isMoveLegal(b, move, kingSquare, pinnedPieces, checkers != 0, evasions, enemy) {
			moves.Moves[writeIndex] = move
			writeIndex++
		}
	}

	moves.Count = writeIndex
	moves.Moves = moves.Moves[:writeIndex]
}

func calculatePinnedPieces(b *board.Board, kingSquare int, enemy *sideInfo) board.Bitboard {
	var pinnedPieces board.Bitboard

	queens := b.GetPieceBitboard(enemy.queen)
	snipers := board.GetRookAttacks(kingSquare, 0)&(b.GetPieceBitboard(enemy.rook)|queens) |
		board.GetBishopAttacks(kingSquare, 0)&(b.GetPieceBitboard(enemy.bishop)|queens)

	for snipers != 0 {
		attackerSquare, newBitboard := snipers.PopLSB()
		snipers = newBitboard

		blockers := board.GetBetween(attackerSquare, kingSquare) & b.AllPieces
		if blockers.PopCount() == 1 {
			pinnedPieces |= blockers
		}
	}

	return pinnedPieces
}

func checkEvasions(checkers board.Bitboard, kingSquare int) board.Bitboard {
	switch checkers.PopCount() {
	case 0:
		return ^board.Bitboard(0)
	case 1:
		return checkers | board.GetBetween(checkers.LSB(), kingSquare)
	default:
		return 0
	}
}

func isMoveLegal(b *board.Board, move board.Move, kingSquare int, pinnedPieces board.Bitboard, inCheck bool, evasions board.Bitboard, enemy *sideInfo) bool {
	fromSquare := move.From.Index()
	toSquare := move.To.Index()

	if move.Piece == board.WhiteKing || move.Piece == board.BlackKing {
		return b.GetAttackersToSquareWithOccupancy(toSquare, enemy.color, b.AllPieces.ClearBit(kingSquare)) == 0
	}

	if !evasions.HasBit(toSquare) && !(move.IsEnPassant && evasions.HasBit(move.EnPassantCapturedSquare().Index())) {
		return false
	}

	if pinnedPieces.HasBit(fromSquare) {
		return board.GetLine(kingSquare, fromSquare).HasBit(toSquare) || board.GetLine(kingSquare, toSquare).HasBit(fromSquare)
	}

	if move.IsEnPassant && !inCheck {
		return isEnPassantLegal(b, move, kingSquare, enemy)
	}

	return true
}

func isEnPassantLegal(b *board.Board, move board.Move, kingSquare int, enemy *sideInfo) bool {
	if kingSquare/8 != int(move.From.Rank) {
		return true
	}

	rankAttackers := (b.GetPieceBitboard(enemy.rook) | b.GetPieceBitboard(enemy.queen)) & board.RankMask(int(move.From.Rank))
	if rankAttackers == 0 {
		return true
	}

	fromSquare := move.From.Index()
	toSquare := move.To.Index()
	occupancyAfterMove := b.AllPieces.ClearBit(fromSquare).ClearBit(move.EnPassantCapturedSquare().Index()).SetBit(toSquare)

	return board.GetRookAttacks(kingSquare, occupancyAfterMove)&rankAttackers == 0
}
