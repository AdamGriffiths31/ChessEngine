package eval

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
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

	// Promotion-capture: the promoted piece is part of the exchange. Its
	// arrival adds net material up front (promoted value minus the pawn), and
	// it is the piece exposed to recapture on the target square - without
	// this, exd8=Q met by Rxd8 was priced as winning a rook for a pawn
	// instead of losing a queen. Both Empty ('.') and the zero value count
	// as "no promotion": hand-built Move literals leave Promotion unset,
	// which is not Empty.
	promoted := move.Promotion
	if promoted != board.Empty && promoted != 0 {
		see.gain[depth] += see.getPieceValue(promoted) - see.getPieceValue(move.Piece)
	}

	occupied := b.AllPieces

	targetSquareIndex := board.FileRankToSquare(target.File, target.Rank)

	whiteAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardWhite)
	blackAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardBlack)

	fromSquare := board.FileRankToSquare(move.From.File, move.From.Rank)
	occupied = occupied.ClearBit(fromSquare)

	if move.IsEnPassant {
		// The captured pawn sits beside the target square, not on it
		capturedPawnSquare := board.FileRankToSquare(move.To.File, move.From.Rank)
		occupied = occupied.ClearBit(capturedPawnSquare)
	}

	see.updateAttackersAfterMove(b, &whiteAttackers, &blackAttackers, target, fromSquare, occupied)

	sideToMove := see.getOppositeSide(move.Piece)
	attackingPiece := move.Piece
	if promoted != board.Empty && promoted != 0 {
		// A recapturer takes the promoted piece standing on the target
		// square, not the pawn that moved there.
		attackingPiece = promoted
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

		// King handling: a king recapture only happens when the target square
		// is undefended by the opposing side; otherwise the king may not take
		// and the exchange genuinely ends. Blindly stopping here overvalued
		// captures like QxP where the enemy king was the pawn's sole guard:
		// the code returned +100 while Kxd4 wins the queen (-800).
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
			// Legal king capture: fall through and apply it like any other;
			// the opponent has no reply, so the next iteration ends cleanly.
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

			*attackers = (*attackers).ClearBit(square)

			return attacker{
				piece:  pieceType,
				square: square,
			}
		}
	}

	return attacker{piece: board.Empty, square: -1}
}

func (see *SEECalculator) updateAttackersAfterMove(b *board.Board, whiteAttackers, blackAttackers *board.Bitboard, target board.Square, _ int, occupied board.Bitboard) {
	targetSquareIndex := board.FileRankToSquare(target.File, target.Rank)

	// Recompute sliding attacks against the reduced occupancy so pieces standing
	// behind a removed attacker (X-rays) are discovered as new attackers.
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

func (see *SEECalculator) getOppositeSide(piece board.Piece) string {
	if see.isWhitePiece(piece) {
		return "b"
	}
	return "w"
}

func (see *SEECalculator) isWhitePiece(piece board.Piece) bool {
	return board.IsWhitePiece(piece)
}
