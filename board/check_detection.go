package board

// MoveGivesCheck determines if a move puts the opponent's king in check
func MoveGivesCheck(b *Board, move Move) bool {
	piece := b.GetPiece(move.From.Rank, move.From.File)
	if piece == Empty {
		return false
	}

	// Determine enemy king position
	var enemyKingBB Bitboard
	if piece < BlackPawn { // White piece
		enemyKingBB = b.GetPieceBitboard(BlackKing)
	} else { // Black piece
		enemyKingBB = b.GetPieceBitboard(WhiteKing)
	}

	if enemyKingBB == 0 {
		return false
	}

	kingSquare := enemyKingBB.LSB()
	toSquare := move.To.Rank*8 + move.To.File
	fromSquare := move.From.Rank*8 + move.From.File

	// 1. Direct check: Does the piece attack the king from its destination?
	if isDirectCheck(b, piece, toSquare, kingSquare) {
		return true
	}

	// 2. Discovered check: Moving this piece might uncover an attack
	if isDiscoveredCheck(b, fromSquare, toSquare, kingSquare, piece) {
		return true
	}

	// 3. Promotion check (special case)
	if move.Promotion != Empty {
		return isDirectCheck(b, move.Promotion, toSquare, kingSquare)
	}

	// 4. En passant discovered check (rare but possible)
	if move.IsEnPassant {
		return isEnPassantCheck(b, move, kingSquare)
	}

	return false
}

// isDirectCheck checks if a piece on a square directly attacks the king
func isDirectCheck(b *Board, piece Piece, fromSquare, kingSquare int) bool {
	switch piece {
	case WhitePawn:
		pawnAttacks := GetPawnAttacks(fromSquare, BitboardWhite)
		return pawnAttacks.HasBit(kingSquare)
	case BlackPawn:
		pawnAttacks := GetPawnAttacks(fromSquare, BitboardBlack)
		return pawnAttacks.HasBit(kingSquare)

	case WhiteKnight, BlackKnight:
		return GetKnightAttacks(fromSquare).HasBit(kingSquare)

	case WhiteBishop, BlackBishop:
		// Use magic bitboards with current board occupancy
		return GetBishopAttacks(fromSquare, b.AllPieces).HasBit(kingSquare)

	case WhiteRook, BlackRook:
		return GetRookAttacks(fromSquare, b.AllPieces).HasBit(kingSquare)

	case WhiteQueen, BlackQueen:
		return GetQueenAttacks(fromSquare, b.AllPieces).HasBit(kingSquare)

	case WhiteKing, BlackKing:
		return GetKingAttacks(fromSquare).HasBit(kingSquare)
	}

	return false
}

// isDiscoveredCheck detects if moving a piece discovers a check
func isDiscoveredCheck(b *Board, fromSquare, toSquare, kingSquare int, movingPiece Piece) bool {
	// Quick rejection: if piece isn't between any of our sliders and enemy king, no discovered check
	fromRank, fromFile := fromSquare/8, fromSquare%8
	kingRank, kingFile := kingSquare/8, kingSquare%8

	onSameRank := fromRank == kingRank
	onSameFile := fromFile == kingFile
	onSameDiagonal := (fromRank - kingRank) == (fromFile - kingFile)
	onSameAntiDiag := (fromRank - kingRank) == -(fromFile - kingFile)

	if !onSameRank && !onSameFile && !onSameDiagonal && !onSameAntiDiag {
		return false // Can't discover check
	}

	occupancy := b.AllPieces.ClearBit(fromSquare).SetBit(toSquare)

	// Check our sliding pieces that could give discovered check
	if movingPiece < BlackPawn { // White piece moving
		if onSameRank || onSameFile {
			rooksQueens := b.GetPieceBitboard(WhiteRook) | b.GetPieceBitboard(WhiteQueen)
			rooksQueens = rooksQueens.ClearBit(fromSquare) // Don't count moving piece

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
	} else { // Black piece moving
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

// isEnPassantCheck checks if en passant capture discovers check
func isEnPassantCheck(b *Board, move Move, kingSquare int) bool {
	// En passant can discover check on the rank
	fromSquare := move.From.Rank*8 + move.From.File
	toSquare := move.To.Rank*8 + move.To.File

	// Calculate captured pawn square
	capturedPawnSquare := toSquare
	if move.From.Rank > move.To.Rank { // White capturing
		capturedPawnSquare -= 8
	} else { // Black capturing
		capturedPawnSquare += 8
	}

	occupancy := b.AllPieces.ClearBit(fromSquare).ClearBit(capturedPawnSquare).SetBit(toSquare)

	// Check if any rook/queen can now attack the king
	piece := b.GetPiece(move.From.Rank, move.From.File)
	if piece < BlackPawn { // White piece
		rooksQueens := b.GetPieceBitboard(WhiteRook) | b.GetPieceBitboard(WhiteQueen)
		for rooksQueens != 0 {
			attackerSq := rooksQueens.LSB()
			rooksQueens = rooksQueens.ClearBit(attackerSq)

			if GetRookAttacks(attackerSq, occupancy).HasBit(kingSquare) {
				return true
			}
		}
	} else { // Black piece
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
