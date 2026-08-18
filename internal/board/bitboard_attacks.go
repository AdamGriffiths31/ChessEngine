package board

// IsSquareAttackedByColor checks if a square is attacked by pieces of a given color
// Uses bitboard operations for maximum performance
func (b *Board) IsSquareAttackedByColor(square int, color BitboardColor) bool {
	if square < 0 || square > 63 {
		return false
	}

	if b.isSquareAttackedByPawns(square, color) {
		return true
	}

	if b.isSquareAttackedByKnights(square, color) {
		return true
	}

	if b.isSquareAttackedBySlidingPieces(square, color) {
		return true
	}

	if b.isSquareAttackedByKing(square, color) {
		return true
	}

	return false
}

func (b *Board) isSquareAttackedByPawns(square int, color BitboardColor) bool {
	// Get pawn attack pattern for the opposite color (pawns attack forward)
	// If we're checking for white pawn attacks, we look at black pawn attack patterns from the target square
	oppositeColor := OppositeBitboardColor(color)
	pawnAttacks := GetPawnAttacks(square, oppositeColor)

	var pawnBitboard Bitboard
	if color == BitboardWhite {
		pawnBitboard = b.GetPieceBitboard(WhitePawn)
	} else {
		pawnBitboard = b.GetPieceBitboard(BlackPawn)
	}

	return (pawnAttacks & pawnBitboard) != 0
}

func (b *Board) isSquareAttackedByKnights(square int, color BitboardColor) bool {
	knightAttacks := GetKnightAttacks(square)

	var knightBitboard Bitboard
	if color == BitboardWhite {
		knightBitboard = b.GetPieceBitboard(WhiteKnight)
	} else {
		knightBitboard = b.GetPieceBitboard(BlackKnight)
	}

	return (knightAttacks & knightBitboard) != 0
}

func (b *Board) isSquareAttackedBySlidingPieces(square int, color BitboardColor) bool {
	occupancy := b.AllPieces

	bishopAttacks := GetBishopAttacks(square, occupancy)
	rookAttacks := GetRookAttacks(square, occupancy)

	if color == BitboardWhite {
		if (bishopAttacks & (b.GetPieceBitboard(WhiteBishop) | b.GetPieceBitboard(WhiteQueen))) != 0 {
			return true
		}
		if (rookAttacks & (b.GetPieceBitboard(WhiteRook) | b.GetPieceBitboard(WhiteQueen))) != 0 {
			return true
		}
	} else {
		if (bishopAttacks & (b.GetPieceBitboard(BlackBishop) | b.GetPieceBitboard(BlackQueen))) != 0 {
			return true
		}
		if (rookAttacks & (b.GetPieceBitboard(BlackRook) | b.GetPieceBitboard(BlackQueen))) != 0 {
			return true
		}
	}

	return false
}

func (b *Board) isSquareAttackedByKing(square int, color BitboardColor) bool {
	kingAttacks := GetKingAttacks(square)

	var kingBitboard Bitboard
	if color == BitboardWhite {
		kingBitboard = b.GetPieceBitboard(WhiteKing)
	} else {
		kingBitboard = b.GetPieceBitboard(BlackKing)
	}

	return (kingAttacks & kingBitboard) != 0
}

// GetAttackersToSquare returns a bitboard of all pieces of the given color that attack the square
func (b *Board) GetAttackersToSquare(square int, color BitboardColor) Bitboard {
	return b.GetAttackersToSquareWithOccupancy(square, color, b.AllPieces)
}

// GetAttackersToSquareWithOccupancy returns a bitboard of all pieces of the given color that
// attack the square, computing sliding attacks against the supplied occupancy instead of the
// board's current occupancy. This lets callers (e.g. SEE) reveal X-ray attackers by removing
// pieces from the occupancy without mutating the board.
func (b *Board) GetAttackersToSquareWithOccupancy(square int, color BitboardColor, occupancy Bitboard) Bitboard {
	var attackers Bitboard

	if square < 0 || square > 63 {
		return attackers
	}

	oppositeColor := OppositeBitboardColor(color)
	pawnAttacks := GetPawnAttacks(square, oppositeColor)
	var pawnBitboard Bitboard
	if color == BitboardWhite {
		pawnBitboard = b.GetPieceBitboard(WhitePawn)
	} else {
		pawnBitboard = b.GetPieceBitboard(BlackPawn)
	}
	attackers |= pawnAttacks & pawnBitboard

	knightAttacks := GetKnightAttacks(square)
	var knightBitboard Bitboard
	if color == BitboardWhite {
		knightBitboard = b.GetPieceBitboard(WhiteKnight)
	} else {
		knightBitboard = b.GetPieceBitboard(BlackKnight)
	}
	attackers |= knightAttacks & knightBitboard

	rookAttacks := GetRookAttacks(square, occupancy)
	bishopAttacks := GetBishopAttacks(square, occupancy)

	if color == BitboardWhite {
		attackers |= rookAttacks & (b.GetPieceBitboard(WhiteRook) | b.GetPieceBitboard(WhiteQueen))
		attackers |= bishopAttacks & (b.GetPieceBitboard(WhiteBishop) | b.GetPieceBitboard(WhiteQueen))
	} else {
		attackers |= rookAttacks & (b.GetPieceBitboard(BlackRook) | b.GetPieceBitboard(BlackQueen))
		attackers |= bishopAttacks & (b.GetPieceBitboard(BlackBishop) | b.GetPieceBitboard(BlackQueen))
	}

	kingAttacks := GetKingAttacks(square)
	var kingBitboard Bitboard
	if color == BitboardWhite {
		kingBitboard = b.GetPieceBitboard(WhiteKing)
	} else {
		kingBitboard = b.GetPieceBitboard(BlackKing)
	}
	attackers |= kingAttacks & kingBitboard

	return attackers
}

// IsInCheck checks if the king of the given color is in check
func (b *Board) IsInCheck(color BitboardColor) bool {
	var kingBitboard Bitboard
	if color == BitboardWhite {
		kingBitboard = b.GetPieceBitboard(WhiteKing)
	} else {
		kingBitboard = b.GetPieceBitboard(BlackKing)
	}

	if kingBitboard == 0 {
		return false
	}

	kingSquare := kingBitboard.LSB()
	if kingSquare == -1 {
		return false
	}

	oppositeColor := OppositeBitboardColor(color)
	return b.IsSquareAttackedByColor(kingSquare, oppositeColor)
}

// GetPieceAttacks returns all squares attacked by a piece of given type on given square
func (b *Board) GetPieceAttacks(piece Piece, square int) Bitboard {
	if square < 0 || square > 63 {
		return 0
	}

	occupancy := b.AllPieces

	switch piece {
	case WhitePawn:
		return GetPawnAttacks(square, BitboardWhite)
	case BlackPawn:
		return GetPawnAttacks(square, BitboardBlack)
	case WhiteKnight, BlackKnight:
		return GetKnightAttacks(square)
	case WhiteRook, BlackRook:
		return GetRookAttacks(square, occupancy)
	case WhiteBishop, BlackBishop:
		return GetBishopAttacks(square, occupancy)
	case WhiteQueen, BlackQueen:
		return GetQueenAttacks(square, occupancy)
	case WhiteKing, BlackKing:
		return GetKingAttacks(square)
	default:
		return 0
	}
}

// GetAllAttackedSquares returns a bitboard of all squares attacked by pieces of the given color
func (b *Board) GetAllAttackedSquares(color BitboardColor) Bitboard {
	var attacks Bitboard

	occupancy := b.AllPieces

	var pawnBitboard Bitboard
	if color == BitboardWhite {
		pawnBitboard = b.GetPieceBitboard(WhitePawn)
		// White pawns attack northeast and northwest
		attacks |= pawnBitboard.ShiftNorthEast() | pawnBitboard.ShiftNorthWest()
	} else {
		pawnBitboard = b.GetPieceBitboard(BlackPawn)
		// Black pawns attack southeast and southwest
		attacks |= pawnBitboard.ShiftSouthEast() | pawnBitboard.ShiftSouthWest()
	}

	var knightBitboard Bitboard
	if color == BitboardWhite {
		knightBitboard = b.GetPieceBitboard(WhiteKnight)
	} else {
		knightBitboard = b.GetPieceBitboard(BlackKnight)
	}

	knightSquares := knightBitboard.BitList()
	for _, square := range knightSquares {
		attacks |= GetKnightAttacks(square)
	}

	var rookBitboard, bishopBitboard, queenBitboard Bitboard
	if color == BitboardWhite {
		rookBitboard = b.GetPieceBitboard(WhiteRook)
		bishopBitboard = b.GetPieceBitboard(WhiteBishop)
		queenBitboard = b.GetPieceBitboard(WhiteQueen)
	} else {
		rookBitboard = b.GetPieceBitboard(BlackRook)
		bishopBitboard = b.GetPieceBitboard(BlackBishop)
		queenBitboard = b.GetPieceBitboard(BlackQueen)
	}

	rookSquares := rookBitboard.BitList()
	for _, square := range rookSquares {
		attacks |= GetRookAttacks(square, occupancy)
	}

	bishopSquares := bishopBitboard.BitList()
	for _, square := range bishopSquares {
		attacks |= GetBishopAttacks(square, occupancy)
	}

	queenSquares := queenBitboard.BitList()
	for _, square := range queenSquares {
		attacks |= GetQueenAttacks(square, occupancy)
	}

	var kingBitboard Bitboard
	if color == BitboardWhite {
		kingBitboard = b.GetPieceBitboard(WhiteKing)
	} else {
		kingBitboard = b.GetPieceBitboard(BlackKing)
	}

	kingSquares := kingBitboard.BitList()
	for _, square := range kingSquares {
		attacks |= GetKingAttacks(square)
	}

	return attacks
}

// IsSquareEmptyBitboard checks if a square is empty using bitboards
func (b *Board) IsSquareEmptyBitboard(square int) bool {
	if square < 0 || square > 63 {
		return false
	}
	return !b.AllPieces.HasBit(square)
}

// GetPieceOnSquare returns the piece on a square using bitboards (faster than array lookup)
func (b *Board) GetPieceOnSquare(square int) Piece {
	if square < 0 || square > 63 {
		return Empty
	}

	if !b.AllPieces.HasBit(square) {
		return Empty
	}

	pieces := []Piece{
		WhitePawn, WhiteRook, WhiteKnight, WhiteBishop, WhiteQueen, WhiteKing,
		BlackPawn, BlackRook, BlackKnight, BlackBishop, BlackQueen, BlackKing,
	}

	for _, piece := range pieces {
		if b.GetPieceBitboard(piece).HasBit(square) {
			return piece
		}
	}

	return Empty
}
