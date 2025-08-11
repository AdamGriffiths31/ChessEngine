package board

// Bitboard-based attack detection for high-performance chess move generation
// This file provides fast attack detection using bitboard operations

// IsSquareAttackedByColor checks if a square is attacked by pieces of a given color
// Uses bitboard operations for maximum performance
func (b *Board) IsSquareAttackedByColor(square int, color BitboardColor) bool {
	// Ensure tables are initialized
	
	if square < 0 || square > 63 {
		return false
	}
	
	// Check pawn attacks
	if b.isSquareAttackedByPawns(square, color) {
		return true
	}
	
	// Check knight attacks
	if b.isSquareAttackedByKnights(square, color) {
		return true
	}
	
	// Check sliding piece attacks (rooks, bishops, queens)
	if b.isSquareAttackedBySlidingPieces(square, color) {
		return true
	}
	
	// Check king attacks
	if b.isSquareAttackedByKing(square, color) {
		return true
	}
	
	return false
}

// isSquareAttackedByPawns checks if pawns of the given color attack the square
func (b *Board) isSquareAttackedByPawns(square int, color BitboardColor) bool {
	// Get pawn attack pattern for the opposite color (pawns attack forward)
	// If we're checking for white pawn attacks, we look at black pawn attack patterns from the target square
	oppositeColor := OppositeBitboardColor(color)
	pawnAttacks := GetPawnAttacks(square, oppositeColor)
	
	// Get the attacking color's pawn bitboard
	var pawnBitboard Bitboard
	if color == BitboardWhite {
		pawnBitboard = b.GetPieceBitboard(WhitePawn)
	} else {
		pawnBitboard = b.GetPieceBitboard(BlackPawn)
	}
	
	// Check if any pawns are on squares that can attack the target square
	return (pawnAttacks & pawnBitboard) != 0
}

// isSquareAttackedByKnights checks if knights of the given color attack the square
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

// isSquareAttackedBySlidingPieces checks if sliding pieces (rooks, bishops, queens) attack the square
func (b *Board) isSquareAttackedBySlidingPieces(square int, color BitboardColor) bool {
	occupancy := b.AllPieces
	
	// Use magic bitboards directly
	bishopAttacks := GetBishopAttacks(square, occupancy)
	rookAttacks := GetRookAttacks(square, occupancy)
	
	if color == BitboardWhite {
		// Check if any white bishops or queens are on bishop attack squares
		if (bishopAttacks & (b.GetPieceBitboard(WhiteBishop) | b.GetPieceBitboard(WhiteQueen))) != 0 {
			return true
		}
		// Check if any white rooks or queens are on rook attack squares
		if (rookAttacks & (b.GetPieceBitboard(WhiteRook) | b.GetPieceBitboard(WhiteQueen))) != 0 {
			return true
		}
	} else {
		// Check if any black bishops or queens are on bishop attack squares
		if (bishopAttacks & (b.GetPieceBitboard(BlackBishop) | b.GetPieceBitboard(BlackQueen))) != 0 {
			return true
		}
		// Check if any black rooks or queens are on rook attack squares
		if (rookAttacks & (b.GetPieceBitboard(BlackRook) | b.GetPieceBitboard(BlackQueen))) != 0 {
			return true
		}
	}
	
	return false
}

// isSquareAttackedByKing checks if the king of the given color attacks the square
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
	var attackers Bitboard
	
	if square < 0 || square > 63 {
		return attackers
	}
	
	// Ensure tables are initialized
	
	occupancy := b.AllPieces
	
	// Pawn attackers
	oppositeColor := OppositeBitboardColor(color)
	pawnAttacks := GetPawnAttacks(square, oppositeColor)
	var pawnBitboard Bitboard
	if color == BitboardWhite {
		pawnBitboard = b.GetPieceBitboard(WhitePawn)
	} else {
		pawnBitboard = b.GetPieceBitboard(BlackPawn)
	}
	attackers |= pawnAttacks & pawnBitboard
	
	// Knight attackers
	knightAttacks := GetKnightAttacks(square)
	var knightBitboard Bitboard
	if color == BitboardWhite {
		knightBitboard = b.GetPieceBitboard(WhiteKnight)
	} else {
		knightBitboard = b.GetPieceBitboard(BlackKnight)
	}
	attackers |= knightAttacks & knightBitboard
	
	// Sliding piece attackers - use magic bitboards directly
	rookAttacks := GetRookAttacks(square, occupancy)
	bishopAttacks := GetBishopAttacks(square, occupancy)
	
	if color == BitboardWhite {
		// Rook and queen attackers (horizontal/vertical)
		attackers |= rookAttacks & (b.GetPieceBitboard(WhiteRook) | b.GetPieceBitboard(WhiteQueen))
		// Bishop and queen attackers (diagonal)
		attackers |= bishopAttacks & (b.GetPieceBitboard(WhiteBishop) | b.GetPieceBitboard(WhiteQueen))
	} else {
		// Rook and queen attackers (horizontal/vertical)
		attackers |= rookAttacks & (b.GetPieceBitboard(BlackRook) | b.GetPieceBitboard(BlackQueen))
		// Bishop and queen attackers (diagonal)
		attackers |= bishopAttacks & (b.GetPieceBitboard(BlackBishop) | b.GetPieceBitboard(BlackQueen))
	}
	
	// King attackers
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
	// Find the king position
	var kingBitboard Bitboard
	if color == BitboardWhite {
		kingBitboard = b.GetPieceBitboard(WhiteKing)
	} else {
		kingBitboard = b.GetPieceBitboard(BlackKing)
	}
	
	if kingBitboard == 0 {
		return false // No king found
	}
	
	kingSquare := kingBitboard.LSB()
	if kingSquare == -1 {
		return false // Invalid king position
	}
	
	// Check if the king is attacked by the opposite color
	oppositeColor := OppositeBitboardColor(color)
	return b.IsSquareAttackedByColor(kingSquare, oppositeColor)
}

// GetPieceAttacks returns all squares attacked by a piece of given type on given square
func (b *Board) GetPieceAttacks(piece Piece, square int) Bitboard {
	if square < 0 || square > 63 {
		return 0
	}
	
	// Ensure tables are initialized
	
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
	
	// Ensure tables are initialized
	
	occupancy := b.AllPieces
	
	// Pawn attacks
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
	
	// Knight attacks
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
	
	// Sliding piece attacks
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
	
	// Rook attacks
	rookSquares := rookBitboard.BitList()
	for _, square := range rookSquares {
		attacks |= GetRookAttacks(square, occupancy)
	}
	
	// Bishop attacks
	bishopSquares := bishopBitboard.BitList()
	for _, square := range bishopSquares {
		attacks |= GetBishopAttacks(square, occupancy)
	}
	
	// Queen attacks
	queenSquares := queenBitboard.BitList()
	for _, square := range queenSquares {
		attacks |= GetQueenAttacks(square, occupancy)
	}
	
	// King attacks
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
	
	// Check each piece type
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