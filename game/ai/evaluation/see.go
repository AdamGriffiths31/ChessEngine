package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// SEECalculator implements Static Exchange Evaluation
type SEECalculator struct {
	gain [32]int // Reusable gain array to avoid allocations
}

// NewSEECalculator creates a new SEE calculator
func NewSEECalculator() *SEECalculator {
	return &SEECalculator{}
}

// SEE calculates the Static Exchange Evaluation for a move
// Returns the material balance after all exchanges on the target square
// Positive values favor the side making the initial capture
func (see *SEECalculator) SEE(b *board.Board, move board.Move) int {
	if b == nil {
		return 0
	}
	if !move.IsCapture {
		return 0
	}

	target := move.To
	depth := 0

	see.gain[depth] = see.getPieceValue(move.Captured)

	if move.IsEnPassant {
		see.gain[depth] = 100
	}

	occupied := b.AllPieces

	targetSquareIndex := board.FileRankToSquare(target.File, target.Rank)

	whiteAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardWhite)
	blackAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardBlack)

	fromSquare := board.FileRankToSquare(move.From.File, move.From.Rank)
	occupied = occupied.ClearBit(fromSquare)

	see.updateAttackersAfterMove(b, &whiteAttackers, &blackAttackers, target, fromSquare, occupied)

	sideToMove := see.getOppositeSide(move.Piece)
	attackingPiece := move.Piece

	for {
		depth++

		var attackers *board.Bitboard
		if sideToMove == "w" {
			attackers = &whiteAttackers
		} else {
			attackers = &blackAttackers
		}

		nextAttacker := see.getLeastValuableAttacker(b, attackers, sideToMove, occupied)

		if nextAttacker.piece == board.Empty {
			break
		}

		capturedValue := see.getPieceValue(attackingPiece)

		see.gain[depth] = capturedValue - see.gain[depth-1]

		occupied = occupied.ClearBit(nextAttacker.square)

		see.updateAttackersAfterMove(b, &whiteAttackers, &blackAttackers, target, nextAttacker.square, occupied)

		if sideToMove == "w" {
			sideToMove = "b"
		} else {
			sideToMove = "w"
		}
		attackingPiece = nextAttacker.piece
	}

	for depth--; depth > 0; depth-- {
		if see.gain[depth-1] > -see.gain[depth] {
			see.gain[depth-1] = -see.gain[depth]
		}
	}

	return see.gain[0]
}

// attacker represents a piece that can attack a square
type attacker struct {
	piece  board.Piece
	square int
}

// getLeastValuableAttacker finds the least valuable piece attacking the target square
func (see *SEECalculator) getLeastValuableAttacker(b *board.Board, attackers *board.Bitboard, side string, occupied board.Bitboard) attacker {
	if b == nil || attackers == nil {
		return attacker{piece: board.Empty, square: -1}
	}
	pieceOrder := see.getPieceOrderForSide(side)

	for _, pieceType := range pieceOrder {
		pieceBitboard := b.GetPieceBitboard(pieceType)
		attackingPieces := pieceBitboard & *attackers & occupied

		if attackingPieces != 0 {
			square, _ := attackingPieces.PopLSB()

			*attackers = (*attackers).ClearBit(square)

			return attacker{
				piece:  pieceType,
				square: square,
			}
		}
	}

	return attacker{piece: board.Empty, square: -1}
}

// updateAttackersAfterMove updates attacker bitboards after a piece moves (for X-ray attacks)
func (see *SEECalculator) updateAttackersAfterMove(b *board.Board, whiteAttackers, blackAttackers *board.Bitboard, target board.Square, _ int, occupied board.Bitboard) {
	targetSquareIndex := board.FileRankToSquare(target.File, target.Rank)

	newWhiteAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardWhite) & occupied
	newBlackAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardBlack) & occupied

	*whiteAttackers = newWhiteAttackers
	*blackAttackers = newBlackAttackers
}

// getPieceValue returns the value of a piece for SEE calculation
func (see *SEECalculator) getPieceValue(piece board.Piece) int {
	if piece == board.WhiteKing || piece == board.BlackKing {
		return 10000
	}

	if value, ok := PieceValues[piece]; ok {
		if value < 0 {
			return -value
		}
		return value
	}

	return 0
}

// getPieceOrderForSide returns pieces in order from least to most valuable
func (see *SEECalculator) getPieceOrderForSide(side string) []board.Piece {
	if side == "w" {
		return []board.Piece{
			board.WhitePawn, board.WhiteKnight, board.WhiteBishop,
			board.WhiteRook, board.WhiteQueen, board.WhiteKing,
		}
	}
	return []board.Piece{
		board.BlackPawn, board.BlackKnight, board.BlackBishop,
		board.BlackRook, board.BlackQueen, board.BlackKing,
	}
}

// getOppositeSide returns the opposite side color
func (see *SEECalculator) getOppositeSide(piece board.Piece) string {
	if see.isWhitePiece(piece) {
		return "b"
	}
	return "w"
}

// isWhitePiece checks if a piece is white
func (see *SEECalculator) isWhitePiece(piece board.Piece) bool {
	return board.IsWhitePiece(piece)
}
