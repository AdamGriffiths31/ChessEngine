// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// GetHashDelta implements the board.HashUpdater interface to calculate incremental hash updates.
// Computes the Zobrist hash delta for a move by XORing keys for changed board state:
//   - Side to move (always flips)
//   - Moving piece (from square removed, to square added or promoted piece added)
//   - Captured piece (removed from capture square, en passant handled specially)
//   - Castling rook (for castling moves, rook moves from corner to beside king)
//   - Castling rights (if changed by move)
//   - En passant target (if changed, only included if adjacent capturing pawn exists)
//
// This delta can be XORed with the old hash to get the new hash, avoiding full board rehashing.
func (m *MinimaxEngine) GetHashDelta(b *board.Board, move board.Move, oldState board.State) uint64 {
	var hashDelta uint64

	hashDelta ^= m.zobrist.GetSideKey()

	fromSquare := move.From.Rank*8 + move.From.File
	toSquare := move.To.Rank*8 + move.To.File

	if move.Piece != board.Empty {
		pieceIndex := m.zobrist.GetPieceIndex(move.Piece)
		hashDelta ^= m.zobrist.GetPieceKey(fromSquare, pieceIndex)
	}

	var destPiece board.Piece
	if move.Promotion != board.Empty {
		destPiece = move.Promotion
	} else {
		destPiece = move.Piece
	}
	if destPiece != board.Empty {
		pieceIndex := m.zobrist.GetPieceIndex(destPiece)
		hashDelta ^= m.zobrist.GetPieceKey(toSquare, pieceIndex)
	}

	if move.IsCapture && move.Captured != board.Empty {
		capturedPieceIndex := m.zobrist.GetPieceIndex(move.Captured)
		if move.IsEnPassant {
			var captureRank int
			if move.Piece == board.WhitePawn {
				captureRank = 4
			} else {
				captureRank = 3
			}
			captureSquare := captureRank*8 + move.To.File
			hashDelta ^= m.zobrist.GetPieceKey(captureSquare, capturedPieceIndex)
		} else {
			hashDelta ^= m.zobrist.GetPieceKey(toSquare, capturedPieceIndex)
		}
	}

	if move.IsCastling {
		var rookFrom, rookTo int
		switch move.To.File {
		case 6:
			rookFrom = move.From.Rank*8 + 7
			rookTo = move.From.Rank*8 + 5
		case 2:
			rookFrom = move.From.Rank*8 + 0
			rookTo = move.From.Rank*8 + 3
		}

		var rook board.Piece
		if move.From.Rank == 0 {
			rook = board.WhiteRook
		} else {
			rook = board.BlackRook
		}
		rookIndex := m.zobrist.GetPieceIndex(rook)
		hashDelta ^= m.zobrist.GetPieceKey(rookFrom, rookIndex)
		hashDelta ^= m.zobrist.GetPieceKey(rookTo, rookIndex)
	}

	if oldState.CastlingRights != b.GetCastlingRights() {
		oldRights := m.zobrist.GetCastlingKey(oldState.CastlingRights)
		newRights := m.zobrist.GetCastlingKey(b.GetCastlingRights())
		hashDelta ^= oldRights ^ newRights
	}

	newEP, newHasEP := b.GetEnPassantTarget()

	if oldState.HasEnPassant != newHasEP ||
		(oldState.HasEnPassant && newHasEP && oldState.EnPassantSquare.File != newEP.File) {

		if oldState.HasEnPassant && hasAdjacentCapturingPawn(b, oldState.EnPassantSquare, getOppositeSide(oldState.SideToMove)) {
			hashDelta ^= m.zobrist.GetEnPassantKey(oldState.EnPassantSquare.File)
		}
		if newHasEP && hasAdjacentCapturingPawn(b, newEP, getOppositeSide(b.GetSideToMove())) {
			hashDelta ^= m.zobrist.GetEnPassantKey(newEP.File)
		}
	}

	return hashDelta
}

// GetNullMoveDelta returns the hash delta for a null move (flip side to move)
func (m *MinimaxEngine) GetNullMoveDelta() uint64 {
	return m.zobrist.GetSideKey()
}

// hasAdjacentCapturingPawn checks if there's a pawn adjacent to the en passant target that can capture
// This implements the same logic as the full HashPosition function
func hasAdjacentCapturingPawn(b *board.Board, epTarget board.Square, sideToMove string) bool {
	var pawnRank int
	var pawnPiece board.Piece

	if sideToMove == "b" {
		pawnRank = 4
		pawnPiece = board.BlackPawn
	} else if sideToMove == "w" {
		pawnRank = 3
		pawnPiece = board.WhitePawn
	} else {
		panic("invalid sideToMove: " + sideToMove)
	}

	epFile := epTarget.File

	for _, df := range []int{-1, 1} {
		adjFile := epFile + df
		if adjFile >= 0 && adjFile < 8 {
			if b.GetPiece(pawnRank, adjFile) == pawnPiece {
				return true
			}
		}
	}

	return false
}

// getOppositeSide returns the opposite side
func getOppositeSide(side string) string {
	if side == "w" {
		return "b"
	}
	return "w"
}
