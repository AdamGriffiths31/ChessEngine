// Package movegen provides high-performance chess move generation using bitboard operations.
package movegen

import "github.com/AdamGriffiths31/ChessEngine/internal/board"

// Generator generates chess moves using bitboard operations.
type Generator struct{}

// NewGenerator creates a new move generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateAllMoves generates all legal moves for the given player.
// Returns an empty list if board is nil.
func (g *Generator) GenerateAllMoves(b *board.Board, player Player) *MoveList {
	moveList := g.GeneratePseudoLegalMoves(b, player)
	if b != nil {
		filterLegalMovesInPlace(b, &sides[player], moveList)
	}
	return moveList
}

// GeneratePseudoLegalMoves generates pseudo-legal moves without king safety validation.
// Callers must verify moves don't leave their own king in check before execution.
// Returns an empty list if board is nil.
func (g *Generator) GeneratePseudoLegalMoves(b *board.Board, player Player) *MoveList {
	moveList := GetMoveList()
	if b == nil {
		return moveList
	}

	side := &sides[player]
	targets := ^b.GetColorBitboard(side.color)

	generatePawnMoves(b, side, moveList)
	generateKnightMoves(b, side, targets, moveList)
	generateBishopMoves(b, side, targets, moveList)
	generateRookMoves(b, side, targets, moveList)
	generateQueenMoves(b, side, targets, moveList)
	generateKingMoves(b, side, moveList)

	return moveList
}

// GenerateCaptures generates pseudo-legal captures and promotions only.
// Returns an empty list if board is nil.
func (g *Generator) GenerateCaptures(b *board.Board, player Player) *MoveList {
	moveList := GetMoveList()
	if b == nil {
		return moveList
	}

	side := &sides[player]
	enemyPieces := b.GetColorBitboard(side.enemy)
	occupancy := b.AllPieces

	pawns := b.GetPieceBitboard(side.pawn)
	generatePawnCaptures(b, side, pawns, enemyPieces, moveList)
	generateEnPassantCaptures(b, side, pawns, moveList)

	promotionPushes := side.pushPawns(pawns) &^ occupancy & board.RankMask(side.promotionRank)
	for promotionPushes != 0 {
		toSquare, newBitboard := promotionPushes.PopLSB()
		promotionPushes = newBitboard

		addPromotionMoves(moveList, toSquare-side.push, toSquare, side, board.Empty)
	}

	generateKnightMoves(b, side, enemyPieces, moveList)
	generateBishopMoves(b, side, enemyPieces, moveList)
	generateRookMoves(b, side, enemyPieces, moveList)

	queens := b.GetPieceBitboard(side.queen)
	for queens != 0 {
		fromSquare, newBitboard := queens.PopLSB()
		queens = newBitboard

		captures := board.GetQueenAttacks(fromSquare, occupancy) & enemyPieces
		addSlidingPieceMoves(b, moveList, fromSquare, captures, enemyPieces, side.queen)
	}

	kings := b.GetPieceBitboard(side.king)
	if kings != 0 {
		fromSquare := kings.LSB()
		addStepMoves(b, moveList, fromSquare, board.GetKingAttacks(fromSquare)&enemyPieces, side.king)
	}

	return moveList
}

// CheckEvasionMask returns the squares a non-king move must reach to answer a
// check on the given player: the checker and the squares between it and the
// king, or none on a double check. It returns every square when not in check.
func (g *Generator) CheckEvasionMask(b *board.Board, player Player) board.Bitboard {
	side := &sides[player]
	kingSquare := b.GetPieceBitboard(side.king).LSB()
	return checkEvasions(b.GetAttackersToSquare(kingSquare, side.enemy), kingSquare)
}

// IsKingInCheck checks if the king of the given player is currently in check.
// Returns false if board is nil or king is not found.
func (g *Generator) IsKingInCheck(b *board.Board, player Player) bool {
	if b == nil {
		return false
	}

	return b.IsInCheck(sides[player].color)
}
