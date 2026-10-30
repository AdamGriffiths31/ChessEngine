package eval

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// SEECalculator implements Static Exchange Evaluation
type SEECalculator struct {
	gain [32]int
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

	if move.HasPromotion() {
		see.gain[depth] += see.getPieceValue(move.Promotion) - see.getPieceValue(move.Piece)
	}

	occupied := b.AllPieces

	targetSquareIndex := board.FileRankToSquare(target.File, target.Rank)

	whiteAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardWhite)
	blackAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardBlack)

	fromSquare := board.FileRankToSquare(move.From.File, move.From.Rank)
	occupied = occupied.ClearBit(fromSquare)

	if move.IsEnPassant {
		capturedPawnSquare := board.FileRankToSquare(move.To.File, move.From.Rank)
		occupied = occupied.ClearBit(capturedPawnSquare)
	}

	see.updateAttackersAfterMove(b, &whiteAttackers, &blackAttackers, target, occupied)

	sideToMove := see.getOppositeSide(move.Piece)
	attackingPiece := move.Piece
	if move.HasPromotion() {
		attackingPiece = move.Promotion
	}

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

		if nextAttacker.piece == board.WhiteKing || nextAttacker.piece == board.BlackKing {
			occAfterKingCapture := occupied.ClearBit(nextAttacker.square)
			var defenders board.Bitboard
			if sideToMove == "w" {
				defenders = b.GetAttackersToSquareWithOccupancy(targetSquareIndex, board.BitboardBlack, occAfterKingCapture) & occAfterKingCapture
			} else {
				defenders = b.GetAttackersToSquareWithOccupancy(targetSquareIndex, board.BitboardWhite, occAfterKingCapture) & occAfterKingCapture
			}
			if defenders != 0 {
				break
			}
		}

		capturedValue := see.getPieceValue(attackingPiece)

		see.gain[depth] = capturedValue - see.gain[depth-1]

		occupied = occupied.ClearBit(nextAttacker.square)

		see.updateAttackersAfterMove(b, &whiteAttackers, &blackAttackers, target, occupied)

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

type attacker struct {
	piece  board.Piece
	square int
}

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

			*attackers = attackers.ClearBit(square)

			return attacker{
				piece:  pieceType,
				square: square,
			}
		}
	}

	return attacker{piece: board.Empty, square: -1}
}

func (see *SEECalculator) updateAttackersAfterMove(b *board.Board, whiteAttackers, blackAttackers *board.Bitboard, target board.Square, occupied board.Bitboard) {
	targetSquareIndex := board.FileRankToSquare(target.File, target.Rank)

	newWhiteAttackers := b.GetAttackersToSquareWithOccupancy(targetSquareIndex, board.BitboardWhite, occupied) & occupied
	newBlackAttackers := b.GetAttackersToSquareWithOccupancy(targetSquareIndex, board.BitboardBlack, occupied) & occupied

	*whiteAttackers = newWhiteAttackers
	*blackAttackers = newBlackAttackers
}

func (see *SEECalculator) getPieceValue(piece board.Piece) int {
	if piece == board.WhiteKing || piece == board.BlackKing {
		return 10000
	}

	value := GetPieceValue(piece)
	if value < 0 {
		return -value
	}
	return value
}

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

func (see *SEECalculator) getOppositeSide(piece board.Piece) string {
	if see.isWhitePiece(piece) {
		return "b"
	}
	return "w"
}

func (see *SEECalculator) isWhitePiece(piece board.Piece) bool {
	return board.IsWhitePiece(piece)
}
