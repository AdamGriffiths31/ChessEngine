// Package search provides chess move search algorithms and transposition table implementation.
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

	// HashPosition only XORs the en-passant key when the side to move has a
	// pawn adjacent to the en-passant file that could actually capture. XOR
	// out the old position's key (evaluated against the PRE-move board — the
	// move may have relocated the very pawn that made the capture possible)
	// and XOR in the new position's key (evaluated against the post-move
	// board, which b already is). If both terms apply to the same file the
	// XORs cancel, matching HashPosition.
	if oldState.HasEnPassant && hadAdjacentCapturingPawnBeforeMove(b, move, oldState.EnPassantSquare, oldState.SideToMove) {
		hashDelta ^= m.zobrist.GetEnPassantKey(oldState.EnPassantSquare.File)
	}
	if newHasEP && hasAdjacentCapturingPawn(b, newEP, b.GetSideToMove()) {
		hashDelta ^= m.zobrist.GetEnPassantKey(newEP.File)
	}

	return hashDelta
}

// GetNullMoveDelta returns the hash delta for a null move (flip side to move)
func (m *MinimaxEngine) GetNullMoveDelta() uint64 {
	return m.zobrist.GetSideKey()
}

// epCapturerRankAndPiece returns the rank index and pawn piece of the pawns
// belonging to sideToMove that could capture on an en-passant target square,
// mirroring HashPosition's convention: white to move captures with a white
// pawn on rank index 4; black to move with a black pawn on rank index 3.
func epCapturerRankAndPiece(sideToMove string) (int, board.Piece) {
	if sideToMove == "w" {
		return 4, board.WhitePawn
	}
	return 3, board.BlackPawn
}

// hasAdjacentCapturingPawn checks if sideToMove (the side to move in the
// position represented by b) has a pawn adjacent to the en passant target
// file that can capture. This implements the same logic as the full
// HashPosition function.
func hasAdjacentCapturingPawn(b *board.Board, epTarget board.Square, sideToMove string) bool {
	pawnRank, pawnPiece := epCapturerRankAndPiece(sideToMove)

	for _, df := range []int{-1, 1} {
		adjFile := epTarget.File + df
		if adjFile >= 0 && adjFile < 8 {
			if b.GetPiece(pawnRank, adjFile) == pawnPiece {
				return true
			}
		}
	}

	return false
}

// hadAdjacentCapturingPawnBeforeMove is hasAdjacentCapturingPawn evaluated
// against the position as it was BEFORE move was applied to b. GetHashDelta
// runs after MakeMove has mutated the board, but whether the old en-passant
// key was part of the old hash depends on the old board's pawn layout.
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

// pieceBeforeMove returns the piece that occupied (rank, file) before move
// was applied to b. Only the squares a move can change are reconstructed:
// move.From, move.To, and (for en passant) the captured pawn's square.
// Castling rook squares are not reconstructed: they lie on ranks 0 and 7,
// and this helper is only used for en-passant checks on ranks 3 and 4.
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
	if move.IsEnPassant {
		captureRank := 3
		if move.Piece == board.WhitePawn {
			captureRank = 4
		}
		if rank == captureRank && file == move.To.File {
			return move.Captured
		}
	}
	return b.GetPiece(rank, file)
}
