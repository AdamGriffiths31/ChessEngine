package movegen

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// filterLegalMovesInPlace filters out moves that would leave the king in check
// Optimized implementation using pin analysis instead of make/unmake
func (bmg *BitboardMoveGenerator) filterLegalMovesInPlace(b *board.Board, player Player, moves *MoveList) {
	var ourKingPiece board.Piece
	var opponentColor board.BitboardColor
	if player == White {
		ourKingPiece = board.WhiteKing
		opponentColor = board.BitboardBlack
	} else {
		ourKingPiece = board.BlackKing
		opponentColor = board.BitboardWhite
	}

	ourKingBitboard := b.GetPieceBitboard(ourKingPiece)
	if ourKingBitboard == 0 {
		moves.Count = 0
		return
	}
	kingSquare := ourKingBitboard.LSB()

	pinnedPieces := bmg.calculatePinnedPieces(b, kingSquare, opponentColor)

	inCheck := b.IsSquareAttackedByColor(kingSquare, opponentColor)

	writeIndex := 0
	for readIndex := 0; readIndex < moves.Count; readIndex++ {
		move := moves.Moves[readIndex]

		if bmg.isMoveLegal(b, move, kingSquare, pinnedPieces, inCheck, opponentColor) {
			moves.Moves[writeIndex] = move
			writeIndex++
		}
	}

	moves.Count = writeIndex
	moves.Moves = moves.Moves[:writeIndex]
}

func (bmg *BitboardMoveGenerator) calculatePinnedPieces(b *board.Board, kingSquare int, opponentColor board.BitboardColor) board.Bitboard {
	var pinnedPieces board.Bitboard

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

	rookAttackers := opponentRooks | opponentQueens
	for rookAttackers != 0 {
		attackerSquare, newBitboard := rookAttackers.PopLSB()
		rookAttackers = newBitboard

		// Check if attacker and king are on the same rank or file (rooks can't pin diagonally)
		if bmg.areOnSameRankOrFile(attackerSquare, kingSquare) {
			between := board.GetBetween(attackerSquare, kingSquare)
			blockers := between & b.AllPieces
			if blockers.PopCount() == 1 {
				// Exactly one piece between attacker and king - it's pinned
				pinnedPieces |= blockers
			}
		}
	}

	bishopAttackers := opponentBishops | opponentQueens
	for bishopAttackers != 0 {
		attackerSquare, newBitboard := bishopAttackers.PopLSB()
		bishopAttackers = newBitboard

		// Only consider diagonal lines for bishop pins (not rank/file lines)
		if bmg.areOnSameDiagonal(attackerSquare, kingSquare) {
			between := board.GetBetween(attackerSquare, kingSquare)
			blockers := between & b.AllPieces
			if blockers.PopCount() == 1 {
				// Exactly one piece between attacker and king - it's pinned
				pinnedPieces |= blockers
			}
		}
	}

	return pinnedPieces
}

func (bmg *BitboardMoveGenerator) enPassantCapturesAttacker(move board.Move, attackerSquare int) bool {
	toSquare := move.To.Rank*8 + move.To.File

	var capturedPawnSquare int
	if move.Piece == board.WhitePawn {
		capturedPawnSquare = toSquare - 8 // Black pawn is one rank below en passant destination
	} else {
		capturedPawnSquare = toSquare + 8 // White pawn is one rank above en passant destination
	}

	return capturedPawnSquare == attackerSquare
}

func (bmg *BitboardMoveGenerator) areOnSameRankOrFile(square1, square2 int) bool {
	file1 := square1 % 8
	rank1 := square1 / 8
	file2 := square2 % 8
	rank2 := square2 / 8

	return file1 == file2 || rank1 == rank2
}

func (bmg *BitboardMoveGenerator) areOnSameDiagonal(square1, square2 int) bool {
	file1 := square1 % 8
	rank1 := square1 / 8
	file2 := square2 % 8
	rank2 := square2 / 8

	// Check if the absolute difference in files equals the absolute difference in ranks
	fileDiff := file1 - file2
	rankDiff := rank1 - rank2

	if fileDiff < 0 {
		fileDiff = -fileDiff
	}
	if rankDiff < 0 {
		rankDiff = -rankDiff
	}

	return fileDiff == rankDiff && fileDiff != 0
}

// isMoveLegal checks if a move is legal without making/unmaking it
func (bmg *BitboardMoveGenerator) isMoveLegal(b *board.Board, move board.Move, kingSquare int, pinnedPieces board.Bitboard, inCheck bool, opponentColor board.BitboardColor) bool {
	fromSquare := move.From.Rank*8 + move.From.File
	toSquare := move.To.Rank*8 + move.To.File

	if move.Piece == board.WhiteKing || move.Piece == board.BlackKing {
		// For king moves, we need to check if the destination would be attacked
		// with the king on the new square (not the old one)
		return bmg.isKingMoveIntoSafety(b, move, kingSquare, toSquare, opponentColor)
	}

	// If we're in double check, only king moves are legal (already handled above)
	if inCheck {
		attackers := bmg.getAttackersToSquare(b, kingSquare, opponentColor)
		if attackers.PopCount() > 1 {
			return false
		}

		// Single check - move must block check or capture attacking piece
		attackerSquare := attackers.LSB()

		// Capturing the attacker (including en passant captures)
		if toSquare == attackerSquare || (move.IsEnPassant && bmg.enPassantCapturesAttacker(move, attackerSquare)) {
			// Still need to check if moving piece is pinned
			if pinnedPieces.HasBit(fromSquare) {
				pinRay := board.GetLine(kingSquare, bmg.findPinningPiece(b, fromSquare, kingSquare, opponentColor))
				return pinRay.HasBit(toSquare)
			}
			return true
		}

		// Blocking the check (only works against sliding pieces)
		between := board.GetBetween(attackerSquare, kingSquare)
		if between.HasBit(toSquare) {
			// Still need to check if moving piece is pinned
			if pinnedPieces.HasBit(fromSquare) {
				pinRay := board.GetLine(kingSquare, bmg.findPinningPiece(b, fromSquare, kingSquare, opponentColor))
				return pinRay.HasBit(toSquare)
			}
			return true
		}

		return false
	}

	// Handle pinned pieces - they can only move along the pin ray
	if pinnedPieces.HasBit(fromSquare) {
		pinningPieceSquare := bmg.findPinningPiece(b, fromSquare, kingSquare, opponentColor)
		if pinningPieceSquare == -1 {
			return false
		}

		pinRay := board.GetLine(kingSquare, pinningPieceSquare)
		return pinRay.HasBit(toSquare)
	}

	// Special case: En passant can expose king to rank attacks
	if move.IsEnPassant {
		return bmg.isEnPassantLegal(b, move, kingSquare, opponentColor)
	}

	return true
}

func (bmg *BitboardMoveGenerator) getAttackersToSquare(b *board.Board, square int, attackerColor board.BitboardColor) board.Bitboard {
	var attackers board.Bitboard

	if attackerColor == board.BitboardWhite {
		whitePawnAttacks := board.GetPawnAttacks(square, board.BitboardBlack) // Reverse direction
		attackers |= whitePawnAttacks & b.GetPieceBitboard(board.WhitePawn)

		knightAttacks := board.GetKnightAttacks(square)
		attackers |= knightAttacks & b.GetPieceBitboard(board.WhiteKnight)

		rookAttacks := board.GetRookAttacks(square, b.AllPieces)
		attackers |= rookAttacks & (b.GetPieceBitboard(board.WhiteRook) | b.GetPieceBitboard(board.WhiteQueen))

		bishopAttacks := board.GetBishopAttacks(square, b.AllPieces)
		attackers |= bishopAttacks & (b.GetPieceBitboard(board.WhiteBishop) | b.GetPieceBitboard(board.WhiteQueen))

		kingAttacks := board.GetKingAttacks(square)
		attackers |= kingAttacks & b.GetPieceBitboard(board.WhiteKing)
	} else {
		blackPawnAttacks := board.GetPawnAttacks(square, board.BitboardWhite) // Reverse direction
		attackers |= blackPawnAttacks & b.GetPieceBitboard(board.BlackPawn)

		knightAttacks := board.GetKnightAttacks(square)
		attackers |= knightAttacks & b.GetPieceBitboard(board.BlackKnight)

		rookAttacks := board.GetRookAttacks(square, b.AllPieces)
		attackers |= rookAttacks & (b.GetPieceBitboard(board.BlackRook) | b.GetPieceBitboard(board.BlackQueen))

		bishopAttacks := board.GetBishopAttacks(square, b.AllPieces)
		attackers |= bishopAttacks & (b.GetPieceBitboard(board.BlackBishop) | b.GetPieceBitboard(board.BlackQueen))

		kingAttacks := board.GetKingAttacks(square)
		attackers |= kingAttacks & b.GetPieceBitboard(board.BlackKing)
	}

	return attackers
}

func (bmg *BitboardMoveGenerator) findPinningPiece(b *board.Board, pinnedSquare, kingSquare int, opponentColor board.BitboardColor) int {
	var opponentSliders board.Bitboard
	if opponentColor == board.BitboardWhite {
		opponentSliders = b.GetPieceBitboard(board.WhiteRook) | b.GetPieceBitboard(board.WhiteBishop) | b.GetPieceBitboard(board.WhiteQueen)
	} else {
		opponentSliders = b.GetPieceBitboard(board.BlackRook) | b.GetPieceBitboard(board.BlackBishop) | b.GetPieceBitboard(board.BlackQueen)
	}

	for opponentSliders != 0 {
		attackerSquare, newBitboard := opponentSliders.PopLSB()
		opponentSliders = newBitboard

		line := board.GetLine(attackerSquare, kingSquare)
		if line != 0 {
			between := board.GetBetween(attackerSquare, kingSquare)
			blockers := between & b.AllPieces

			// If exactly one piece between attacker and king, and it's our pinned piece
			if blockers.PopCount() == 1 && blockers.HasBit(pinnedSquare) {
				return attackerSquare
			}
		}
	}

	return -1
}

// isEnPassantLegal checks if an en passant move is legal (doesn't expose king to rank attacks)
func (bmg *BitboardMoveGenerator) isEnPassantLegal(b *board.Board, move board.Move, kingSquare int, opponentColor board.BitboardColor) bool {
	// En passant captures remove a pawn from a different square than the destination
	// This can potentially expose the king to rank attacks

	fromSquare := move.From.Rank*8 + move.From.File
	toSquare := move.To.Rank*8 + move.To.File

	var capturedPawnSquare int
	if move.Piece == board.WhitePawn {
		capturedPawnSquare = toSquare - 8 // Black pawn is one rank below
	} else {
		capturedPawnSquare = toSquare + 8 // White pawn is one rank above
	}

	kingRank := kingSquare / 8
	moveRank := fromSquare / 8
	if kingRank != moveRank {
		return true // No rank attack possible
	}

	var opponentRooksQueens board.Bitboard
	if opponentColor == board.BitboardWhite {
		opponentRooksQueens = b.GetPieceBitboard(board.WhiteRook) | b.GetPieceBitboard(board.WhiteQueen)
	} else {
		opponentRooksQueens = b.GetPieceBitboard(board.BlackRook) | b.GetPieceBitboard(board.BlackQueen)
	}

	rankMask := board.RankMask(kingRank)
	rankAttackers := opponentRooksQueens & rankMask

	if rankAttackers == 0 {
		return true // No rank attackers
	}

	// Simulate the en passant capture by temporarily removing both pawns
	occupancyAfterMove := b.AllPieces
	occupancyAfterMove = occupancyAfterMove.ClearBit(fromSquare)
	occupancyAfterMove = occupancyAfterMove.ClearBit(capturedPawnSquare)
	occupancyAfterMove = occupancyAfterMove.SetBit(toSquare)

	for rankAttackers != 0 {
		attackerSquare, newBitboard := rankAttackers.PopLSB()
		rankAttackers = newBitboard

		attackRay := board.GetRookAttacks(attackerSquare, occupancyAfterMove)
		if attackRay.HasBit(kingSquare) {
			return false // En passant would expose king
		}
	}

	return true
}

func (bmg *BitboardMoveGenerator) isKingMoveIntoSafety(b *board.Board, _ board.Move, fromSquare, toSquare int, opponentColor board.BitboardColor) bool {
	// For king moves, we need to temporarily remove the king from its current square
	// and check if the destination square would be attacked
	modifiedOccupancy := b.AllPieces
	modifiedOccupancy = modifiedOccupancy.ClearBit(fromSquare)

	return !bmg.isSquareAttackedByColorWithOccupancy(b, toSquare, opponentColor, modifiedOccupancy)
}

func (bmg *BitboardMoveGenerator) isSquareAttackedByColorWithOccupancy(b *board.Board, square int, attackerColor board.BitboardColor, occupancy board.Bitboard) bool {
	if square < 0 || square > 63 {
		return false
	}

	if attackerColor == board.BitboardWhite {
		whitePawnAttacks := board.GetPawnAttacks(square, board.BitboardBlack) // Reverse direction
		if (whitePawnAttacks & b.GetPieceBitboard(board.WhitePawn)) != 0 {
			return true
		}

		knightAttacks := board.GetKnightAttacks(square)
		if (knightAttacks & b.GetPieceBitboard(board.WhiteKnight)) != 0 {
			return true
		}

		rookAttacks := board.GetRookAttacks(square, occupancy)
		if (rookAttacks & (b.GetPieceBitboard(board.WhiteRook) | b.GetPieceBitboard(board.WhiteQueen))) != 0 {
			return true
		}

		bishopAttacks := board.GetBishopAttacks(square, occupancy)
		if (bishopAttacks & (b.GetPieceBitboard(board.WhiteBishop) | b.GetPieceBitboard(board.WhiteQueen))) != 0 {
			return true
		}

		// Check king attacks (but not the moving king)
		kingAttacks := board.GetKingAttacks(square)
		if (kingAttacks & b.GetPieceBitboard(board.WhiteKing)) != 0 {
			return true
		}
	} else {
		blackPawnAttacks := board.GetPawnAttacks(square, board.BitboardWhite) // Reverse direction
		if (blackPawnAttacks & b.GetPieceBitboard(board.BlackPawn)) != 0 {
			return true
		}

		knightAttacks := board.GetKnightAttacks(square)
		if (knightAttacks & b.GetPieceBitboard(board.BlackKnight)) != 0 {
			return true
		}

		rookAttacks := board.GetRookAttacks(square, occupancy)
		if (rookAttacks & (b.GetPieceBitboard(board.BlackRook) | b.GetPieceBitboard(board.BlackQueen))) != 0 {
			return true
		}

		bishopAttacks := board.GetBishopAttacks(square, occupancy)
		if (bishopAttacks & (b.GetPieceBitboard(board.BlackBishop) | b.GetPieceBitboard(board.BlackQueen))) != 0 {
			return true
		}

		// Check king attacks (but not the moving king)
		kingAttacks := board.GetKingAttacks(square)
		if (kingAttacks & b.GetPieceBitboard(board.BlackKing)) != 0 {
			return true
		}
	}

	return false
}
