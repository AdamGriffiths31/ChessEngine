package moves

import "github.com/AdamGriffiths31/ChessEngine/board"

// isMoveLegalIdiomatic implements the standard chess engine approach:
// Direct calculation when needed, no expensive pre-caching
func (bmg *BitboardMoveGenerator) isMoveLegalIdiomatic(b *board.Board, move board.Move, kingSquare int, pinnedPieces board.Bitboard, inCheck bool, opponentColor board.BitboardColor) bool {
	fromSquare := move.From.Rank*8 + move.From.File
	toSquare := move.To.Rank*8 + move.To.File
	
	// King moves - validate safety directly when needed
	if move.Piece == board.WhiteKing || move.Piece == board.BlackKing {
		return bmg.isKingMoveIntoSafetyIdiomatic(b, fromSquare, toSquare, opponentColor)
	}
	
	// Double check - only king moves are legal
	if inCheck {
		attackers := b.GetAttackersToSquare(kingSquare, opponentColor)
		if attackers.PopCount() > 1 {
			return false
		}
		
		// Single check - must capture attacker or block check
		attackerSquare := attackers.LSB()
		
		// Capturing the attacker (including en passant)
		if toSquare == attackerSquare || (move.IsEnPassant && bmg.enPassantCapturesAttacker(move, attackerSquare)) {
			// Still need to verify piece isn't pinned in a way that prevents this capture
			if pinnedPieces.HasBit(fromSquare) {
				return bmg.moveStaysOnPinRay(b, fromSquare, toSquare, kingSquare, opponentColor)
			}
			return true
		}
		
		// Blocking the check
		between := board.GetBetween(attackerSquare, kingSquare)
		if between.HasBit(toSquare) {
			// Verify piece isn't pinned in a way that prevents blocking
			if pinnedPieces.HasBit(fromSquare) {
				return bmg.moveStaysOnPinRay(b, fromSquare, toSquare, kingSquare, opponentColor)
			}
			return true
		}
		
		return false // Move doesn't address the check
	}
	
	// Pinned piece validation - must stay on pin ray
	if pinnedPieces.HasBit(fromSquare) {
		return bmg.moveStaysOnPinRay(b, fromSquare, toSquare, kingSquare, opponentColor)
	}
	
	// En passant special case - can expose king to rank attacks
	if move.IsEnPassant {
		return bmg.isEnPassantLegal(b, move, kingSquare, opponentColor)
	}
	
	return true
}

// isKingMoveIntoSafetyIdiomatic uses the standard chess engine approach:
// Direct calculation with modified occupancy when needed
func (bmg *BitboardMoveGenerator) isKingMoveIntoSafetyIdiomatic(b *board.Board, fromSquare, toSquare int, opponentColor board.BitboardColor) bool {
	// Special case for castling - use existing complex validation
	if bmg.isCastlingMove(fromSquare, toSquare) {
		return bmg.isKingMoveIntoSafety(b, board.Move{
			From: board.Square{File: fromSquare % 8, Rank: fromSquare / 8},
			To:   board.Square{File: toSquare % 8, Rank: toSquare / 8},
		}, fromSquare, toSquare, opponentColor)
	}
	
	// Regular king move - check if destination would be safe
	// Create modified occupancy with king removed from current square
	modifiedOccupancy := b.AllPieces.ClearBit(fromSquare)
	
	// Direct attack calculation - no caching overhead
	return !bmg.isSquareAttackedWithOccupancyIdiomatic(b, toSquare, opponentColor, modifiedOccupancy)
}

// isSquareAttackedWithOccupancyIdiomatic implements direct attack detection
// This is the standard approach used by strong chess engines
func (bmg *BitboardMoveGenerator) isSquareAttackedWithOccupancyIdiomatic(b *board.Board, square int, attackerColor board.BitboardColor, occupancy board.Bitboard) bool {
	// Check pawn attacks (occupancy-independent)
	pawnAttacks := board.GetPawnAttacks(square, board.OppositeBitboardColor(attackerColor))
	if attackerColor == board.BitboardWhite {
		if (pawnAttacks & b.GetPieceBitboard(board.WhitePawn)) != 0 {
			return true
		}
	} else {
		if (pawnAttacks & b.GetPieceBitboard(board.BlackPawn)) != 0 {
			return true
		}
	}
	
	// Check knight attacks (occupancy-independent)
	knightAttacks := board.GetKnightAttacks(square)
	if attackerColor == board.BitboardWhite {
		if (knightAttacks & b.GetPieceBitboard(board.WhiteKnight)) != 0 {
			return true
		}
	} else {
		if (knightAttacks & b.GetPieceBitboard(board.BlackKnight)) != 0 {
			return true
		}
	}
	
	// Check sliding piece attacks with the provided occupancy
	rookAttacks := board.GetRookAttacks(square, occupancy)
	bishopAttacks := board.GetBishopAttacks(square, occupancy)
	
	if attackerColor == board.BitboardWhite {
		// Check rook/queen attacks
		if (rookAttacks & (b.GetPieceBitboard(board.WhiteRook) | b.GetPieceBitboard(board.WhiteQueen))) != 0 {
			return true
		}
		// Check bishop/queen attacks
		if (bishopAttacks & (b.GetPieceBitboard(board.WhiteBishop) | b.GetPieceBitboard(board.WhiteQueen))) != 0 {
			return true
		}
		// Check king attacks
		if (board.GetKingAttacks(square) & b.GetPieceBitboard(board.WhiteKing)) != 0 {
			return true
		}
	} else {
		// Check rook/queen attacks  
		if (rookAttacks & (b.GetPieceBitboard(board.BlackRook) | b.GetPieceBitboard(board.BlackQueen))) != 0 {
			return true
		}
		// Check bishop/queen attacks
		if (bishopAttacks & (b.GetPieceBitboard(board.BlackBishop) | b.GetPieceBitboard(board.BlackQueen))) != 0 {
			return true
		}
		// Check king attacks
		if (board.GetKingAttacks(square) & b.GetPieceBitboard(board.BlackKing)) != 0 {
			return true
		}
	}
	
	return false
}

// moveStaysOnPinRay checks if a move by a pinned piece stays on the pin ray
// Uses direct calculation instead of pre-cached pin rays
func (bmg *BitboardMoveGenerator) moveStaysOnPinRay(b *board.Board, fromSquare, toSquare, kingSquare int, opponentColor board.BitboardColor) bool {
	// Find the pinning piece by checking lines from king through the piece
	pinningPiece := bmg.findPinningPieceIdiomatic(b, fromSquare, kingSquare, opponentColor)
	if pinningPiece == -1 {
		return true // Not actually pinned
	}
	
	// Check if both source and destination are on the line between king and pinner
	pinRay := board.GetLine(kingSquare, pinningPiece)
	return pinRay.HasBit(toSquare)
}

// findPinningPieceIdiomatic finds which opponent piece is pinning the given piece
// Direct calculation - only done when needed for pinned pieces
func (bmg *BitboardMoveGenerator) findPinningPieceIdiomatic(b *board.Board, pinnedSquare, kingSquare int, opponentColor board.BitboardColor) int {
	// Get opponent sliding pieces that could potentially pin
	var opponentSliders board.Bitboard
	if opponentColor == board.BitboardWhite {
		opponentSliders = b.GetPieceBitboard(board.WhiteRook) | b.GetPieceBitboard(board.WhiteBishop) | b.GetPieceBitboard(board.WhiteQueen)
	} else {
		opponentSliders = b.GetPieceBitboard(board.BlackRook) | b.GetPieceBitboard(board.BlackBishop) | b.GetPieceBitboard(board.BlackQueen)
	}
	
	// Check each sliding piece to see if it pins our piece to the king
	for opponentSliders != 0 {
		attackerSquare, newBitboard := opponentSliders.PopLSB()
		opponentSliders = newBitboard
		
		// Check if attacker, pinned piece, and king are on the same line
		line := board.GetLine(attackerSquare, kingSquare)
		if line != 0 && line.HasBit(pinnedSquare) {
			// Verify exactly one piece (the pinned piece) is between attacker and king
			between := board.GetBetween(attackerSquare, kingSquare)
			blockers := between & b.AllPieces
			
			if blockers.PopCount() == 1 && blockers.HasBit(pinnedSquare) {
				return attackerSquare
			}
		}
	}
	
	return -1
}

// calculatePinnedPiecesIdiomatic calculates pinned pieces using direct method
// This is lightweight compared to the attack cache approach
func (bmg *BitboardMoveGenerator) calculatePinnedPiecesIdiomatic(b *board.Board, kingSquare int, opponentColor board.BitboardColor) board.Bitboard {
	var pinnedPieces board.Bitboard
	
	// Get opponent sliding pieces
	var opponentRooks, opponentBishops, opponentQueens board.Bitboard
	if opponentColor == board.BitboardWhite {
		opponentRooks = b.GetPieceBitboard(board.WhiteRook)
		opponentBishops = b.GetPieceBitboard(board.WhiteBishop)  
		opponentQueens = b.GetPieceBitboard(board.WhiteQueen)
	} else {
		opponentRooks = b.GetPieceBitboard(board.BlackRook)
		opponentBishops = b.GetPieceBitboard(board.BlackBishop)
		opponentQueens = b.GetPieceBitboard(board.BlackQueen)
	}
	
	// Check for pins by rooks/queens (rank and file attacks)
	rookAttackers := opponentRooks | opponentQueens
	for rookAttackers != 0 {
		attackerSquare, newBitboard := rookAttackers.PopLSB()
		rookAttackers = newBitboard
		
		// Check if attacker and king are on same rank or file
		if bmg.areOnSameRankOrFile(attackerSquare, kingSquare) {
			between := board.GetBetween(attackerSquare, kingSquare)
			blockers := between & b.AllPieces
			if blockers.PopCount() == 1 {
				pinnedPieces |= blockers
			}
		}
	}
	
	// Check for pins by bishops/queens (diagonal attacks)
	bishopAttackers := opponentBishops | opponentQueens
	for bishopAttackers != 0 {
		attackerSquare, newBitboard := bishopAttackers.PopLSB()
		bishopAttackers = newBitboard
		
		// Check if attacker and king are on same diagonal
		if bmg.areOnSameDiagonal(attackerSquare, kingSquare) {
			between := board.GetBetween(attackerSquare, kingSquare)
			blockers := between & b.AllPieces
			if blockers.PopCount() == 1 {
				pinnedPieces |= blockers
			}
		}
	}
	
	return pinnedPieces
}