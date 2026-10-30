package board

// MoveGivesCheck determines if a move puts the opponent's king in check
func MoveGivesCheck(b *Board, move Move) bool {
	piece := b.GetPiece(move.From.Rank, move.From.File)
	if piece == Empty {
		return false
	}

	var enemyKingBB Bitboard
	if piece < BlackPawn {
		enemyKingBB = b.GetPieceBitboard(BlackKing)
	} else {
		enemyKingBB = b.GetPieceBitboard(WhiteKing)
	}

	if enemyKingBB == 0 {
		return false
	}

	kingSquare := enemyKingBB.LSB()
	toSquare := move.To.Rank*8 + move.To.File
	fromSquare := move.From.Rank*8 + move.From.File

	if isDirectCheck(b, piece, toSquare, kingSquare, fromSquare) {
		return true
	}

	if isDiscoveredCheck(b, fromSquare, toSquare, kingSquare, piece) {
		return true
	}

	if move.Promotion != Empty {
		return isDirectCheck(b, move.Promotion, toSquare, kingSquare, fromSquare)
	}

	if move.IsEnPassant {
		return isEnPassantCheck(b, move, kingSquare)
	}

	return false
}

func isDirectCheck(b *Board, piece Piece, attackFrom, kingSquare, moverOrigin int) bool {
	occupancy := b.AllPieces.ClearBit(moverOrigin)

	switch piece {
	case WhitePawn:
		pawnAttacks := GetPawnAttacks(attackFrom, BitboardWhite)
		return pawnAttacks.HasBit(kingSquare)
	case BlackPawn:
		pawnAttacks := GetPawnAttacks(attackFrom, BitboardBlack)
		return pawnAttacks.HasBit(kingSquare)

	case WhiteKnight, BlackKnight:
		return GetKnightAttacks(attackFrom).HasBit(kingSquare)

	case WhiteBishop, BlackBishop:
		return GetBishopAttacks(attackFrom, occupancy).HasBit(kingSquare)

	case WhiteRook, BlackRook:
		return GetRookAttacks(attackFrom, occupancy).HasBit(kingSquare)

	case WhiteQueen, BlackQueen:
		return GetQueenAttacks(attackFrom, occupancy).HasBit(kingSquare)

	case WhiteKing, BlackKing:
		return GetKingAttacks(attackFrom).HasBit(kingSquare)
	}

	return false
}

func isDiscoveredCheck(b *Board, fromSquare, toSquare, kingSquare int, movingPiece Piece) bool {
	fromRank, fromFile := fromSquare/8, fromSquare%8
	kingRank, kingFile := kingSquare/8, kingSquare%8

	onSameRank := fromRank == kingRank
	onSameFile := fromFile == kingFile
	onSameDiagonal := (fromRank - kingRank) == (fromFile - kingFile)
	onSameAntiDiag := (fromRank - kingRank) == -(fromFile - kingFile)

	if !onSameRank && !onSameFile && !onSameDiagonal && !onSameAntiDiag {
		return false
	}

	occupancy := b.AllPieces.ClearBit(fromSquare).SetBit(toSquare)

	if movingPiece < BlackPawn {
		if onSameRank || onSameFile {
			rooksQueens := b.GetPieceBitboard(WhiteRook) | b.GetPieceBitboard(WhiteQueen)
			rooksQueens = rooksQueens.ClearBit(fromSquare)

			for rooksQueens != 0 {
				attackerSq := rooksQueens.LSB()
				rooksQueens = rooksQueens.ClearBit(attackerSq)

				if GetRookAttacks(attackerSq, occupancy).HasBit(kingSquare) {
					return true
				}
			}
		}

		if onSameDiagonal || onSameAntiDiag {
			bishopsQueens := b.GetPieceBitboard(WhiteBishop) | b.GetPieceBitboard(WhiteQueen)
			bishopsQueens = bishopsQueens.ClearBit(fromSquare)

			for bishopsQueens != 0 {
				attackerSq := bishopsQueens.LSB()
				bishopsQueens = bishopsQueens.ClearBit(attackerSq)

				if GetBishopAttacks(attackerSq, occupancy).HasBit(kingSquare) {
					return true
				}
			}
		}
	} else {
		if onSameRank || onSameFile {
			rooksQueens := b.GetPieceBitboard(BlackRook) | b.GetPieceBitboard(BlackQueen)
			rooksQueens = rooksQueens.ClearBit(fromSquare)

			for rooksQueens != 0 {
				attackerSq := rooksQueens.LSB()
				rooksQueens = rooksQueens.ClearBit(attackerSq)

				if GetRookAttacks(attackerSq, occupancy).HasBit(kingSquare) {
					return true
				}
			}
		}

		if onSameDiagonal || onSameAntiDiag {
			bishopsQueens := b.GetPieceBitboard(BlackBishop) | b.GetPieceBitboard(BlackQueen)
			bishopsQueens = bishopsQueens.ClearBit(fromSquare)

			for bishopsQueens != 0 {
				attackerSq := bishopsQueens.LSB()
				bishopsQueens = bishopsQueens.ClearBit(attackerSq)

				if GetBishopAttacks(attackerSq, occupancy).HasBit(kingSquare) {
					return true
				}
			}
		}
	}

	return false
}

func isEnPassantCheck(b *Board, move Move, kingSquare int) bool {
	fromSquare := move.From.Rank*8 + move.From.File
	toSquare := move.To.Rank*8 + move.To.File

	capturedPawnSquare := toSquare
	if move.From.Rank > move.To.Rank {
		capturedPawnSquare -= 8
	} else {
		capturedPawnSquare += 8
	}

	occupancy := b.AllPieces.ClearBit(fromSquare).ClearBit(capturedPawnSquare).SetBit(toSquare)

	piece := b.GetPiece(move.From.Rank, move.From.File)
	if piece < BlackPawn {
		rooksQueens := b.GetPieceBitboard(WhiteRook) | b.GetPieceBitboard(WhiteQueen)
		for rooksQueens != 0 {
			attackerSq := rooksQueens.LSB()
			rooksQueens = rooksQueens.ClearBit(attackerSq)

			if GetRookAttacks(attackerSq, occupancy).HasBit(kingSquare) {
				return true
			}
		}
	} else {
		rooksQueens := b.GetPieceBitboard(BlackRook) | b.GetPieceBitboard(BlackQueen)
		for rooksQueens != 0 {
			attackerSq := rooksQueens.LSB()
			rooksQueens = rooksQueens.ClearBit(attackerSq)

			if GetRookAttacks(attackerSq, occupancy).HasBit(kingSquare) {
				return true
			}
		}
	}

	return false
}
