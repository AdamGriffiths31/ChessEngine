package board

// MoveGivesCheck determines if a move puts the opponent's king in check
func MoveGivesCheck(b *Board, move Move) bool {
	piece := b.PieceAt(move.From)
	if piece == Empty {
		return false
	}

	var enemyKingBB Bitboard
	if IsWhitePiece(piece) {
		enemyKingBB = b.GetPieceBitboard(BlackKing)
	} else {
		enemyKingBB = b.GetPieceBitboard(WhiteKing)
	}

	if enemyKingBB == 0 {
		return false
	}

	kingSquare := enemyKingBB.LSB()
	toSquare := move.To.Index()
	fromSquare := move.From.Index()

	if isDirectCheck(b, piece, toSquare, kingSquare, fromSquare) {
		return true
	}

	if isDiscoveredCheck(b, fromSquare, toSquare, kingSquare, piece) {
		return true
	}

	if move.HasPromotion() {
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
	rooksQueens, bishopsQueens := sliders(b, IsWhitePiece(movingPiece))

	if (onSameRank || onSameFile) && rookSliderAttacks(rooksQueens.ClearBit(fromSquare), kingSquare, occupancy) {
		return true
	}

	return (onSameDiagonal || onSameAntiDiag) && bishopSliderAttacks(bishopsQueens.ClearBit(fromSquare), kingSquare, occupancy)
}

func isEnPassantCheck(b *Board, move Move, kingSquare int) bool {
	fromSquare := move.From.Index()
	toSquare := move.To.Index()
	capturedPawnSquare := move.EnPassantCapturedSquare().Index()

	occupancy := b.AllPieces.ClearBit(fromSquare).ClearBit(capturedPawnSquare).SetBit(toSquare)
	rooksQueens, bishopsQueens := sliders(b, IsWhitePiece(b.PieceAt(move.From)))

	return rookSliderAttacks(rooksQueens, kingSquare, occupancy) || bishopSliderAttacks(bishopsQueens, kingSquare, occupancy)
}

func sliders(b *Board, white bool) (rooksQueens, bishopsQueens Bitboard) {
	if white {
		queens := b.GetPieceBitboard(WhiteQueen)
		return b.GetPieceBitboard(WhiteRook) | queens, b.GetPieceBitboard(WhiteBishop) | queens
	}
	queens := b.GetPieceBitboard(BlackQueen)
	return b.GetPieceBitboard(BlackRook) | queens, b.GetPieceBitboard(BlackBishop) | queens
}

func rookSliderAttacks(attackers Bitboard, target int, occupancy Bitboard) bool {
	for attackers != 0 {
		attackerSq := attackers.LSB()
		attackers = attackers.ClearBit(attackerSq)

		if GetRookAttacks(attackerSq, occupancy).HasBit(target) {
			return true
		}
	}
	return false
}

func bishopSliderAttacks(attackers Bitboard, target int, occupancy Bitboard) bool {
	for attackers != 0 {
		attackerSq := attackers.LSB()
		attackers = attackers.ClearBit(attackerSq)

		if GetBishopAttacks(attackerSq, occupancy).HasBit(target) {
			return true
		}
	}
	return false
}
