package movegen

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// GenerateKnightMovesBitboard generates knight moves using precomputed attack patterns (exported for testing)
func (bmg *BitboardMoveGenerator) GenerateKnightMovesBitboard(b *board.Board, player Player, moveList *MoveList) {
	var knightPiece board.Piece
	var bitboardColor board.BitboardColor

	if player == White {
		knightPiece = board.WhiteKnight
		bitboardColor = board.BitboardWhite
	} else {
		knightPiece = board.BlackKnight
		bitboardColor = board.BitboardBlack
	}

	knights := b.GetPieceBitboard(knightPiece)
	if knights == 0 {
		return
	}

	friendlyPieces := b.GetColorBitboard(bitboardColor)

	for knights != 0 {
		fromSquare, newBitboard := knights.PopLSB()
		knights = newBitboard

		attacks := board.GetKnightAttacks(fromSquare)
		// Remove squares occupied by friendly pieces
		validMoves := attacks &^ friendlyPieces

		for validMoves != 0 {
			toSquare, newValidMoves := validMoves.PopLSB()
			validMoves = newValidMoves

			capturedPiece := b.GetPieceOnSquare(toSquare)
			isCapture := capturedPiece != board.Empty

			toFile, toRank := board.SquareToFileRank(toSquare)
			fromFile, fromRank := board.SquareToFileRank(fromSquare)

			move := board.Move{
				From:        board.Square{File: fromFile, Rank: fromRank},
				To:          board.Square{File: toFile, Rank: toRank},
				Piece:       knightPiece,
				Captured:    capturedPiece,
				Promotion:   board.Empty,
				IsCapture:   isCapture,
				IsCastling:  false,
				IsEnPassant: false,
			}
			moveList.AddMove(move)
		}
	}
}

// generateKingMovesBitboard generates king moves including castling
func (bmg *BitboardMoveGenerator) generateKingMovesBitboard(b *board.Board, player Player, moveList *MoveList) {
	var kingPiece board.Piece
	var bitboardColor board.BitboardColor

	if player == White {
		kingPiece = board.WhiteKing
		bitboardColor = board.BitboardWhite
	} else {
		kingPiece = board.BlackKing
		bitboardColor = board.BitboardBlack
	}

	kings := b.GetPieceBitboard(kingPiece)
	if kings == 0 {
		return
	}

	friendlyPieces := b.GetColorBitboard(bitboardColor)

	// Should only be one king
	kingSquare := kings.LSB()
	if kingSquare == -1 {
		return
	}

	attacks := board.GetKingAttacks(kingSquare)
	// Remove squares occupied by friendly pieces
	validMoves := attacks &^ friendlyPieces

	fromFile, fromRank := board.SquareToFileRank(kingSquare)

	for validMoves != 0 {
		toSquare, newValidMoves := validMoves.PopLSB()
		validMoves = newValidMoves

		capturedPiece := b.GetPieceOnSquare(toSquare)
		isCapture := capturedPiece != board.Empty

		toFile, toRank := board.SquareToFileRank(toSquare)

		move := board.Move{
			From:        board.Square{File: fromFile, Rank: fromRank},
			To:          board.Square{File: toFile, Rank: toRank},
			Piece:       kingPiece,
			Captured:    capturedPiece,
			Promotion:   board.Empty,
			IsCapture:   isCapture,
			IsCastling:  false,
			IsEnPassant: false,
		}
		moveList.AddMove(move)
	}

	bmg.generateCastlingMovesBitboard(b, player, kingSquare, moveList)
}
