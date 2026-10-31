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

	depth := 0

	see.gain[depth] = see.getPieceValue(move.Captured)

	if move.IsEnPassant {
		see.gain[depth] = see.getPieceValue(board.WhitePawn)
	}

	if move.HasPromotion() {
		see.gain[depth] += see.getPieceValue(move.Promotion) - see.getPieceValue(move.Piece)
	}

	targetSquare := board.FileRankToSquare(move.To.File, move.To.Rank)

	occupied := b.AllPieces.ClearBit(board.FileRankToSquare(move.From.File, move.From.Rank))

	if move.IsEnPassant {
		capturedPawnSquare := board.FileRankToSquare(move.To.File, move.From.Rank)
		occupied = occupied.ClearBit(capturedPawnSquare)
	}

	sideToMove := board.BitboardWhite
	if board.IsWhitePiece(move.Piece) {
		sideToMove = board.BitboardBlack
	}
	attackingPiece := move.Piece
	if move.HasPromotion() {
		attackingPiece = move.Promotion
	}

	for {
		depth++

		attackers := b.GetAttackersToSquareWithOccupancy(targetSquare, sideToMove, occupied) & occupied

		nextAttacker := getLeastValuableAttacker(b, attackers, sideToMove)

		if nextAttacker.piece == board.Empty {
			break
		}

		if nextAttacker.piece == board.WhiteKing || nextAttacker.piece == board.BlackKing {
			occAfterKingCapture := occupied.ClearBit(nextAttacker.square)
			defenders := b.GetAttackersToSquareWithOccupancy(targetSquare, board.OppositeBitboardColor(sideToMove), occAfterKingCapture) & occAfterKingCapture
			if defenders != 0 {
				break
			}
		}

		capturedValue := see.getPieceValue(attackingPiece)

		see.gain[depth] = capturedValue - see.gain[depth-1]

		occupied = occupied.ClearBit(nextAttacker.square)

		sideToMove = board.OppositeBitboardColor(sideToMove)
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

type seePiece struct {
	piece board.Piece
	index int
}

var seePieceOrder = [2][6]seePiece{
	board.BitboardWhite: {
		{board.WhitePawn, board.WhitePawnIndex}, {board.WhiteKnight, board.WhiteKnightIndex}, {board.WhiteBishop, board.WhiteBishopIndex},
		{board.WhiteRook, board.WhiteRookIndex}, {board.WhiteQueen, board.WhiteQueenIndex}, {board.WhiteKing, board.WhiteKingIndex},
	},
	board.BitboardBlack: {
		{board.BlackPawn, board.BlackPawnIndex}, {board.BlackKnight, board.BlackKnightIndex}, {board.BlackBishop, board.BlackBishopIndex},
		{board.BlackRook, board.BlackRookIndex}, {board.BlackQueen, board.BlackQueenIndex}, {board.BlackKing, board.BlackKingIndex},
	},
}

func getLeastValuableAttacker(b *board.Board, attackers board.Bitboard, side board.BitboardColor) attacker {
	for _, candidate := range &seePieceOrder[side] {
		attackingPieces := b.PieceBitboards[candidate.index] & attackers

		if attackingPieces != 0 {
			return attacker{
				piece:  candidate.piece,
				square: attackingPieces.LSB(),
			}
		}
	}

	return attacker{piece: board.Empty, square: -1}
}

func (see *SEECalculator) getPieceValue(piece board.Piece) int {
	if piece == board.WhiteKing || piece == board.BlackKing {
		return 10000
	}

	return abs(GetPieceValue(piece))
}
