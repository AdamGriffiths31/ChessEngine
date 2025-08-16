// Package moves provides chess move execution and board state management utilities.
package moves

import "github.com/AdamGriffiths31/ChessEngine/board"

// MoveExecutor handles making and unmaking moves
type MoveExecutor struct{}

// MakeMove executes a move on the board and returns history for undoing
func (me *MoveExecutor) MakeMove(b *board.Board, move board.Move, updateBoardState func(*board.Board, board.Move)) *MoveHistory {
	history := &MoveHistory{
		Move:            move,
		CapturedPiece:   board.Empty,
		CastlingRights:  b.GetCastlingRights(),
		EnPassantTarget: b.GetEnPassantTarget(),
		HalfMoveClock:   b.GetHalfMoveClock(),
		FullMoveNumber:  b.GetFullMoveNumber(),
		WasEnPassant:    move.IsEnPassant,
		WasCastling:     move.IsCastling,
	}

	if move.IsEnPassant {
		captureRank := move.From.Rank
		history.CapturedPiece = b.GetPiece(captureRank, move.To.File)
		b.SetPiece(captureRank, move.To.File, board.Empty)
	} else if move.IsCapture {
		history.CapturedPiece = b.GetPiece(move.To.Rank, move.To.File)
	}

	if move.IsCastling {
		var rookFrom, rookTo board.Square
		if move.To.File == KingsideFile {
			rookFrom = board.Square{File: KingsideRookFromFile, Rank: move.From.Rank}
			rookTo = board.Square{File: KingsideRookToFile, Rank: move.From.Rank}
		} else {
			rookFrom = board.Square{File: QueensideRookFromFile, Rank: move.From.Rank}
			rookTo = board.Square{File: QueensideRookToFile, Rank: move.From.Rank}
		}
		rook := b.GetPiece(rookFrom.Rank, rookFrom.File)
		b.SetPiece(rookFrom.Rank, rookFrom.File, board.Empty)
		b.SetPiece(rookTo.Rank, rookTo.File, rook)
	}

	piece := b.GetPiece(move.From.Rank, move.From.File)
	b.SetPiece(move.From.Rank, move.From.File, board.Empty)

	if move.Promotion != board.Empty {
		b.SetPiece(move.To.Rank, move.To.File, move.Promotion)
	} else {
		b.SetPiece(move.To.Rank, move.To.File, piece)
	}

	updateBoardState(b, move)

	return history
}

// UnmakeMove undoes a move using the stored history
func (me *MoveExecutor) UnmakeMove(b *board.Board, history *MoveHistory) {
	move := history.Move

	piece := b.GetPiece(move.To.Rank, move.To.File)
	if move.Promotion != board.Empty {
		var pawnPiece board.Piece
		if move.To.Rank == 7 {
			pawnPiece = board.WhitePawn
		} else {
			pawnPiece = board.BlackPawn
		}
		b.SetPiece(move.From.Rank, move.From.File, pawnPiece)
	} else {
		b.SetPiece(move.From.Rank, move.From.File, piece)
	}

	if history.WasEnPassant {
		b.SetPiece(move.To.Rank, move.To.File, board.Empty)
		captureRank := move.From.Rank
		b.SetPiece(captureRank, move.To.File, history.CapturedPiece)
	} else if history.CapturedPiece != board.Empty {
		b.SetPiece(move.To.Rank, move.To.File, history.CapturedPiece)
	} else {
		b.SetPiece(move.To.Rank, move.To.File, board.Empty)
	}

	if history.WasCastling {
		var rookFrom, rookTo board.Square
		if move.To.File == KingsideFile {
			rookFrom = board.Square{File: KingsideRookFromFile, Rank: move.From.Rank}
			rookTo = board.Square{File: KingsideRookToFile, Rank: move.From.Rank}
		} else {
			rookFrom = board.Square{File: QueensideRookFromFile, Rank: move.From.Rank}
			rookTo = board.Square{File: QueensideRookToFile, Rank: move.From.Rank}
		}
		rook := b.GetPiece(rookTo.Rank, rookTo.File)
		b.SetPiece(rookTo.Rank, rookTo.File, board.Empty)
		b.SetPiece(rookFrom.Rank, rookFrom.File, rook)
	}

	b.SetCastlingRights(history.CastlingRights)
	b.SetEnPassantTarget(history.EnPassantTarget)
	b.SetHalfMoveClock(history.HalfMoveClock)
	b.SetFullMoveNumber(history.FullMoveNumber)
}
