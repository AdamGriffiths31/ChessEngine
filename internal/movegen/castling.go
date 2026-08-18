package movegen

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"strings"
)

// Square constants for castling
const (
	E1 = 4
	G1 = 6
	C1 = 2
	H1 = 7
	A1 = 0
	F1 = 5
	D1 = 3
	B1 = 1
	E8 = 60
	G8 = 62
	C8 = 58
	H8 = 63
	A8 = 56
	F8 = 61
	D8 = 59
	B8 = 57
)

// generateCastlingMovesBitboard generates castling moves using bitboard path checking
func (bmg *BitboardMoveGenerator) generateCastlingMovesBitboard(b *board.Board, player Player, kingSquare int, moveList *MoveList) {
	castlingRights := b.GetCastlingRights()
	if castlingRights == "-" {
		return
	}

	var kingPiece board.Piece
	var kingside, queenside byte
	var kingStartSquare int
	var kingsideKingTarget, kingsideRookTarget int
	var queensideKingTarget, queensideRookTarget int

	if player == White {
		kingPiece = board.WhiteKing
		kingside = 'K'
		queenside = 'Q'
		kingStartSquare = E1
		kingsideKingTarget = G1
		kingsideRookTarget = F1
		queensideKingTarget = C1
		queensideRookTarget = D1
	} else {
		kingPiece = board.BlackKing
		kingside = 'k'
		queenside = 'q'
		kingStartSquare = E8
		kingsideKingTarget = G8
		kingsideRookTarget = F8
		queensideKingTarget = C8
		queensideRookTarget = D8
	}

	if kingSquare != kingStartSquare {
		return
	}

	var oppositeColor board.BitboardColor
	if player == White {
		oppositeColor = board.BitboardBlack
	} else {
		oppositeColor = board.BitboardWhite
	}

	if strings.ContainsRune(castlingRights, rune(kingside)) {
		pathMask := board.Bitboard(0)
		pathMask = pathMask.SetBit(kingsideRookTarget).SetBit(kingsideKingTarget)

		if (b.AllPieces & pathMask) == 0 {
			if !b.IsSquareAttackedByColor(kingSquare, oppositeColor) &&
				!b.IsSquareAttackedByColor(kingsideRookTarget, oppositeColor) &&
				!b.IsSquareAttackedByColor(kingsideKingTarget, oppositeColor) {

				kingFromFile, kingFromRank := board.SquareToFileRank(kingSquare)
				kingToFile, kingToRank := board.SquareToFileRank(kingsideKingTarget)

				move := board.Move{
					From:        board.Square{File: kingFromFile, Rank: kingFromRank},
					To:          board.Square{File: kingToFile, Rank: kingToRank},
					Piece:       kingPiece,
					Captured:    board.Empty,
					Promotion:   board.Empty,
					IsCapture:   false,
					IsCastling:  true,
					IsEnPassant: false,
				}
				moveList.AddMove(move)
			}
		}
	}

	if strings.ContainsRune(castlingRights, rune(queenside)) {
		// Check if path is clear (queenside includes an extra square)
		pathMask := board.Bitboard(0)
		pathMask = pathMask.SetBit(queensideRookTarget).SetBit(queensideKingTarget).SetBit(B1)
		if player == Black {
			pathMask = pathMask.ClearBit(B1).SetBit(B8)
		}

		if (b.AllPieces & pathMask) == 0 {
			// Check if king or key path squares are under attack
			if !b.IsSquareAttackedByColor(kingSquare, oppositeColor) &&
				!b.IsSquareAttackedByColor(queensideRookTarget, oppositeColor) &&
				!b.IsSquareAttackedByColor(queensideKingTarget, oppositeColor) {

				kingFromFile, kingFromRank := board.SquareToFileRank(kingSquare)
				kingToFile, kingToRank := board.SquareToFileRank(queensideKingTarget)

				move := board.Move{
					From:        board.Square{File: kingFromFile, Rank: kingFromRank},
					To:          board.Square{File: kingToFile, Rank: kingToRank},
					Piece:       kingPiece,
					Captured:    board.Empty,
					Promotion:   board.Empty,
					IsCapture:   false,
					IsCastling:  true,
					IsEnPassant: false,
				}
				moveList.AddMove(move)
			}
		}
	}
}
