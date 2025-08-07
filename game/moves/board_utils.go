package moves

import "github.com/AdamGriffiths31/ChessEngine/board"

// MoveExecutor handles making and unmaking moves
type MoveExecutor struct{}

// MakeMove executes a move on the board and returns history for undoing
func (me *MoveExecutor) MakeMove(b *board.Board, move board.Move, updateBoardState func(*board.Board, board.Move)) *MoveHistory {
	// Create history to enable undoing
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
	
	// Handle en passant capture
	if move.IsEnPassant {
		// Remove the captured pawn
		captureRank := move.From.Rank
		history.CapturedPiece = b.GetPiece(captureRank, move.To.File)
		b.SetPiece(captureRank, move.To.File, board.Empty)
	} else if move.IsCapture {
		// Store captured piece for normal captures
		history.CapturedPiece = b.GetPiece(move.To.Rank, move.To.File)
	}
	
	// Handle castling
	if move.IsCastling {
		// Move the rook
		var rookFrom, rookTo board.Square
		if move.To.File == KingsideFile { // Kingside
			rookFrom = board.Square{File: KingsideRookFromFile, Rank: move.From.Rank}
			rookTo = board.Square{File: KingsideRookToFile, Rank: move.From.Rank}
		} else { // Queenside
			rookFrom = board.Square{File: QueensideRookFromFile, Rank: move.From.Rank}
			rookTo = board.Square{File: QueensideRookToFile, Rank: move.From.Rank}
		}
		rook := b.GetPiece(rookFrom.Rank, rookFrom.File)
		b.SetPiece(rookFrom.Rank, rookFrom.File, board.Empty)
		b.SetPiece(rookTo.Rank, rookTo.File, rook)
	}
	
	// Move the piece
	piece := b.GetPiece(move.From.Rank, move.From.File)
	b.SetPiece(move.From.Rank, move.From.File, board.Empty)
	
	// Handle promotion
	if move.Promotion != board.Empty {
		b.SetPiece(move.To.Rank, move.To.File, move.Promotion)
	} else {
		b.SetPiece(move.To.Rank, move.To.File, piece)
	}
	
	// Update board state (castling rights, en passant, etc.)
	updateBoardState(b, move)
	
	return history
}

// UnmakeMove undoes a move using the stored history
func (me *MoveExecutor) UnmakeMove(b *board.Board, history *MoveHistory) {
	move := history.Move
	
	// Restore the piece to its original position
	piece := b.GetPiece(move.To.Rank, move.To.File)
	if move.Promotion != board.Empty {
		// For promotion, restore the original pawn
		var pawnPiece board.Piece
		if move.To.Rank == 7 { // White promotion
			pawnPiece = board.WhitePawn
		} else { // Black promotion
			pawnPiece = board.BlackPawn
		}
		b.SetPiece(move.From.Rank, move.From.File, pawnPiece)
	} else {
		b.SetPiece(move.From.Rank, move.From.File, piece)
	}
	
	// Restore the target square
	if history.WasEnPassant {
		// For en passant, restore the captured pawn to its original position
		b.SetPiece(move.To.Rank, move.To.File, board.Empty)
		captureRank := move.From.Rank
		b.SetPiece(captureRank, move.To.File, history.CapturedPiece)
	} else if history.CapturedPiece != board.Empty {
		// Restore captured piece
		b.SetPiece(move.To.Rank, move.To.File, history.CapturedPiece)
	} else {
		// Empty the target square
		b.SetPiece(move.To.Rank, move.To.File, board.Empty)
	}
	
	// Undo castling
	if history.WasCastling {
		// Restore the rook
		var rookFrom, rookTo board.Square
		if move.To.File == KingsideFile { // Kingside
			rookFrom = board.Square{File: KingsideRookFromFile, Rank: move.From.Rank}
			rookTo = board.Square{File: KingsideRookToFile, Rank: move.From.Rank}
		} else { // Queenside
			rookFrom = board.Square{File: QueensideRookFromFile, Rank: move.From.Rank}
			rookTo = board.Square{File: QueensideRookToFile, Rank: move.From.Rank}
		}
		rook := b.GetPiece(rookTo.Rank, rookTo.File)
		b.SetPiece(rookTo.Rank, rookTo.File, board.Empty)
		b.SetPiece(rookFrom.Rank, rookFrom.File, rook)
	}
	
	// Restore board state
	b.SetCastlingRights(history.CastlingRights)
	b.SetEnPassantTarget(history.EnPassantTarget)
	b.SetHalfMoveClock(history.HalfMoveClock)
	b.SetFullMoveNumber(history.FullMoveNumber)
}


// updateBoardState updates castling rights, en passant, and move counters
func (g *Generator) updateBoardState(b *board.Board, move board.Move) {
	// Update castling rights based on the move
	castlingRights := b.GetCastlingRights()
	piece := b.GetPiece(move.To.Rank, move.To.File)
	
	// King moves remove all castling rights for that side
	if piece == board.WhiteKing {
		castlingRights = g.removeCastlingRights(castlingRights, "KQ")
	} else if piece == board.BlackKing {
		castlingRights = g.removeCastlingRights(castlingRights, "kq")
	}
	
	// Rook moves remove castling rights for that side
	if piece == board.WhiteRook {
		if move.From.File == QueensideRookFromFile && move.From.Rank == 0 { // Queenside rook
			castlingRights = g.removeCastlingRights(castlingRights, "Q")
		} else if move.From.File == KingsideRookFromFile && move.From.Rank == 0 { // Kingside rook
			castlingRights = g.removeCastlingRights(castlingRights, "K")
		}
	} else if piece == board.BlackRook {
		if move.From.File == QueensideRookFromFile && move.From.Rank == 7 { // Queenside rook
			castlingRights = g.removeCastlingRights(castlingRights, "q")
		} else if move.From.File == KingsideRookFromFile && move.From.Rank == 7 { // Kingside rook
			castlingRights = g.removeCastlingRights(castlingRights, "k")
		}
	}
	
	// Captured rook removes castling rights
	if move.IsCapture {
		if move.To.File == QueensideRookFromFile && move.To.Rank == 0 { // White queenside rook captured
			castlingRights = g.removeCastlingRights(castlingRights, "Q")
		} else if move.To.File == KingsideRookFromFile && move.To.Rank == 0 { // White kingside rook captured
			castlingRights = g.removeCastlingRights(castlingRights, "K")
		} else if move.To.File == QueensideRookFromFile && move.To.Rank == 7 { // Black queenside rook captured
			castlingRights = g.removeCastlingRights(castlingRights, "q")
		} else if move.To.File == KingsideRookFromFile && move.To.Rank == 7 { // Black kingside rook captured
			castlingRights = g.removeCastlingRights(castlingRights, "k")
		}
	}
	
	b.SetCastlingRights(castlingRights)
	
	// Set en passant target for pawn two-square moves
	if piece == board.WhitePawn || piece == board.BlackPawn {
		if abs(move.To.Rank - move.From.Rank) == 2 {
			// Two-square pawn move - set en passant target
			targetRank := (move.From.Rank + move.To.Rank) / 2
			enPassantTarget := &board.Square{File: move.From.File, Rank: targetRank}
			b.SetEnPassantTarget(enPassantTarget)
		} else {
			b.SetEnPassantTarget(nil)
		}
	} else {
		b.SetEnPassantTarget(nil)
	}
	
	// Update move counters
	halfMoveClock := b.GetHalfMoveClock()
	if move.IsCapture || piece == board.WhitePawn || piece == board.BlackPawn {
		halfMoveClock = 0
	} else {
		halfMoveClock++
	}
	b.SetHalfMoveClock(halfMoveClock)
	
	// Update full move number (increments after black's move)
	if b.GetSideToMove() == "b" {
		b.SetFullMoveNumber(b.GetFullMoveNumber() + 1)
	}
	
}

// removeCastlingRights removes specific castling rights from the string
func (g *Generator) removeCastlingRights(rights, toRemove string) string {
	result := ""
	for _, r := range rights {
		remove := false
		for _, remove_r := range toRemove {
			if r == remove_r {
				remove = true
				break
			}
		}
		if !remove {
			result += string(r)
		}
	}
	if result == "" {
		return "-"
	}
	return result
}