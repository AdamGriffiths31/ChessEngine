package search

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
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

	fromSquare := board.FileRankToSquare(move.From.File, move.From.Rank)
	toSquare := board.FileRankToSquare(move.To.File, move.To.Rank)

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
			captureSquare := board.FileRankToSquare(move.To.File, move.From.Rank)
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

	if oldState.HasEnPassant && hadAdjacentCapturingPawnBeforeMove(b, move, oldState.EnPassantSquare, oldState.SideToMove) {
		hashDelta ^= m.zobrist.GetEnPassantKey(oldState.EnPassantSquare.File)
	}
	if newHasEP && hasAdjacentCapturingPawn(b, newEP, b.GetSideToMove()) {
		hashDelta ^= m.zobrist.GetEnPassantKey(newEP.File)
	}

	return hashDelta
}

// GetNullMoveDelta returns the hash delta for a null move from position b.
// Called before MakeNullMove mutates anything: b still has the mover's
// side-to-move and a live en-passant target. The delta flips the
// side-to-move key and, if the EP target was actually hashed (a capturing
// pawn existed - the non-perturbation rule), removes its file key, since the
// null move clears that target.
func (m *MinimaxEngine) GetNullMoveDelta(b *board.Board) uint64 {
	delta := m.zobrist.GetSideKey()
	if ep, hasEP := b.GetEnPassantTarget(); hasEP && hasAdjacentCapturingPawn(b, ep, b.GetSideToMove()) {
		delta ^= m.zobrist.GetEnPassantKey(ep.File)
	}
	return delta
}

func epCapturerRankAndPiece(sideToMove string) (int, board.Piece) {
	if sideToMove == "w" {
		return 4, board.WhitePawn
	}
	return 3, board.BlackPawn
}

func hasAdjacentCapturingPawn(b *board.Board, epTarget board.Square, sideToMove string) bool {
	return hadAdjacentCapturingPawnBeforeMove(b, board.Move{}, epTarget, sideToMove)
}

func hadAdjacentCapturingPawnBeforeMove(b *board.Board, move board.Move, epTarget board.Square, sideToMove string) bool {
	pawnRank, pawnPiece := epCapturerRankAndPiece(sideToMove)

	for _, df := range []int{-1, 1} {
		adjFile := epTarget.File + df
		if adjFile >= 0 && adjFile < 8 {
			if pieceBeforeMove(b, move, pawnRank, adjFile) == pawnPiece {
				return true
			}
		}
	}

	return false
}

func pieceBeforeMove(b *board.Board, move board.Move, rank, file int) board.Piece {
	if move.From.Rank == rank && move.From.File == file {
		return move.Piece
	}
	if move.To.Rank == rank && move.To.File == file {
		if move.IsCapture && !move.IsEnPassant {
			return move.Captured
		}
		return board.Empty
	}
	if move.IsEnPassant && rank == move.From.Rank && file == move.To.File {
		return move.Captured
	}
	return b.GetPiece(rank, file)
}
