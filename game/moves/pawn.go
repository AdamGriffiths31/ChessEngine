package moves

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// GeneratePawnMoves generates all legal pawn moves for a given player
func (g *Generator) GeneratePawnMoves(b *board.Board, player Player) *MoveList {
	moveList := NewMoveList()

	// Determine pawn piece and direction based on player
	var pawnPiece board.Piece
	var direction int
	var startRank, promotionRank int

	if player == White {
		pawnPiece = board.WhitePawn
		direction = 1     // White pawns move up (increasing rank)
		startRank = 1     // 2nd rank (index 1)
		promotionRank = 7 // 8th rank (index 7)
	} else {
		pawnPiece = board.BlackPawn
		direction = -1    // Black pawns move down (decreasing rank)
		startRank = 6     // 7th rank (index 6)
		promotionRank = 0 // 1st rank (index 0)
	}

	// Scan the board for pawns of the current player
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece == pawnPiece {
				// Generate moves for this pawn
				g.generatePawnMovesFromSquare(b, player, rank, file, direction, startRank, promotionRank, moveList)
			}
		}
	}

	return moveList
}

// generatePawnMovesFromSquare generates all moves for a pawn at a specific square
func (g *Generator) generatePawnMovesFromSquare(b *board.Board, player Player, rank, file, direction, startRank, promotionRank int, moveList *MoveList) {
	fromSquare := board.Square{File: file, Rank: rank}

	// Determine pawn piece
	var pawnPiece board.Piece
	if player == White {
		pawnPiece = board.WhitePawn
	} else {
		pawnPiece = board.BlackPawn
	}

	// Forward moves
	g.generatePawnForwardMoves(b, player, fromSquare, direction, startRank, promotionRank, pawnPiece, moveList)

	// Capture moves
	g.generatePawnCaptures(b, player, fromSquare, direction, promotionRank, pawnPiece, moveList)

	// En passant moves
	g.generateEnPassantMoves(b, player, fromSquare, direction, pawnPiece, moveList)
}

// generatePawnForwardMoves generates forward pawn moves (1 and 2 squares)
func (g *Generator) generatePawnForwardMoves(b *board.Board, player Player, from board.Square, direction, startRank, promotionRank int, pawnPiece board.Piece, moveList *MoveList) {
	// One square forward
	newRank := from.Rank + direction
	if newRank >= 0 && newRank <= 7 {
		if b.GetPiece(newRank, from.File) == board.Empty {
			to := board.Square{File: from.File, Rank: newRank}

			// Check for promotion
			if newRank == promotionRank {
				createMove := func(f, t board.Square, isCapture bool, captured, promotion board.Piece) board.Move {
					return g.createMove(b, f, t, isCapture, captured, promotion)
				}
				g.promotionHandler.AddPromotionMoves(b, from, to, player, moveList, createMove)
			} else {
				move := g.createMove(b, from, to, false, board.Empty, board.Empty)
				moveList.AddMove(move)
			}

			// Two squares forward (only from starting position)
			if from.Rank == startRank {
				newRank2 := newRank + direction
				if newRank2 >= 0 && newRank2 <= 7 && b.GetPiece(newRank2, from.File) == board.Empty {
					to2 := board.Square{File: from.File, Rank: newRank2}
					move := g.createMove(b, from, to2, false, board.Empty, board.Empty)
					moveList.AddMove(move)
				}
			}
		}
	}
}

// generatePawnCaptures generates diagonal capture moves
func (g *Generator) generatePawnCaptures(b *board.Board, player Player, from board.Square, direction, promotionRank int, pawnPiece board.Piece, moveList *MoveList) {
	newRank := from.Rank + direction
	if newRank < 0 || newRank > 7 {
		return
	}

	// Left capture
	if from.File > 0 {
		piece := b.GetPiece(newRank, from.File-1)
		if piece != board.Empty && g.isEnemyPiece(piece, player) {
			to := board.Square{File: from.File - 1, Rank: newRank}

			if newRank == promotionRank {
				createMove := func(f, t board.Square, isCapture bool, captured, promotion board.Piece) board.Move {
					return g.createMove(b, f, t, isCapture, captured, promotion)
				}
				g.promotionHandler.AddPromotionMoves(b, from, to, player, moveList, createMove)
			} else {
				move := g.createMove(b, from, to, true, piece, board.Empty)
				moveList.AddMove(move)
			}
		}
	}

	// Right capture
	if from.File < 7 {
		piece := b.GetPiece(newRank, from.File+1)
		if piece != board.Empty && g.isEnemyPiece(piece, player) {
			to := board.Square{File: from.File + 1, Rank: newRank}

			if newRank == promotionRank {
				createMove := func(f, t board.Square, isCapture bool, captured, promotion board.Piece) board.Move {
					return g.createMove(b, f, t, isCapture, captured, promotion)
				}
				g.promotionHandler.AddPromotionMoves(b, from, to, player, moveList, createMove)
			} else {
				move := g.createMove(b, from, to, true, piece, board.Empty)
				moveList.AddMove(move)
			}
		}
	}
}

// generateEnPassantMoves generates en passant capture moves
func (g *Generator) generateEnPassantMoves(b *board.Board, player Player, from board.Square, direction int, pawnPiece board.Piece, moveList *MoveList) {
	enPassantTarget := b.GetEnPassantTarget()
	if enPassantTarget == nil {
		return
	}

	// Check if we can capture en passant
	newRank := from.Rank + direction
	if newRank != enPassantTarget.Rank {
		return
	}

	// Check if the en passant target is diagonally adjacent
	if (from.File == enPassantTarget.File-1 || from.File == enPassantTarget.File+1) {
		// Verify there's an enemy pawn next to us (the one to be captured)
		captureRank := from.Rank
		captureFile := enPassantTarget.File
		piece := b.GetPiece(captureRank, captureFile)
		
		if piece != board.Empty && g.isEnemyPiece(piece, player) {
			// Check if it's actually a pawn
			if (player == White && piece == board.BlackPawn) || (player == Black && piece == board.WhitePawn) {
				move := g.createMove(b, from, *enPassantTarget, true, piece, board.Empty)
				move.IsEnPassant = true
				moveList.AddMove(move)
			}
		}
	}
}

// ValidatePawnMove validates if a pawn move is legal
func (v *Validator) ValidatePawnMove(b *board.Board, move board.Move, player Player) bool {
	legalMoves := v.generator.GeneratePawnMoves(b, player)
	return legalMoves.Contains(move)
}

// GetLegalPawnMoves returns all legal pawn moves for the current position
func GetLegalPawnMoves(b *board.Board, player Player) *MoveList {
	generator := NewGenerator()
	return generator.GeneratePawnMoves(b, player)
}