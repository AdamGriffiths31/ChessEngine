package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// SEECalculator implements Static Exchange Evaluation
type SEECalculator struct {
}

// NewSEECalculator creates a new SEE calculator
func NewSEECalculator() *SEECalculator {
	return &SEECalculator{}
}

// SEE calculates the Static Exchange Evaluation for a move
// Returns the material balance after all exchanges on the target square
// Positive values favor the side making the initial capture
func (see *SEECalculator) SEE(b *board.Board, move board.Move) int {
	if !move.IsCapture {
		return 0
	}

	target := move.To
	gain := make([]int, 32) // Max possible depth of exchanges
	depth := 0

	// Initial capture value - use existing PieceValues
	gain[depth] = see.getPieceValue(move.Captured)
	
	// Handle en passant special case
	if move.IsEnPassant {
		gain[depth] = 100 // Pawn value
	}

	// Get occupied squares bitboard using existing functionality
	occupied := b.AllPieces
	
	// Get all attackers to the target square using existing functionality  
	targetSquareIndex := board.FileRankToSquare(target.File, target.Rank)
	
	whiteAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardWhite)
	blackAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardBlack)
	
	// Remove the piece that made the initial move
	fromSquare := board.FileRankToSquare(move.From.File, move.From.Rank)
	occupied = occupied.ClearBit(fromSquare)
	
	// Update attackers after the initial move (for X-ray attacks)
	see.updateAttackersAfterMove(b, &whiteAttackers, &blackAttackers, target, fromSquare, occupied)

	// Current side to move (opposite of the side that made the initial capture)
	sideToMove := see.getOppositeSide(move.Piece)
	attackingPiece := move.Piece

	for {
		depth++
		
		// Find the least valuable attacker for the current side
		var attackers *board.Bitboard
		if sideToMove == "w" {
			attackers = &whiteAttackers
		} else {
			attackers = &blackAttackers
		}
		
		nextAttacker := see.getLeastValuableAttacker(b, attackers, sideToMove, occupied)
		
		if nextAttacker.piece == board.Empty {
			break // No more attackers for this side
		}
		
		// Calculate gain from white's perspective
		// The "capturedValue" is what the current attacker will capture
		capturedValue := see.getPieceValue(attackingPiece)
		
		// Standard SEE: gain[depth] = value_of_captured_piece - gain[depth-1]
		// This works because we're alternating perspectives with the minimax
		gain[depth] = capturedValue - gain[depth-1]
		
		// Remove the attacking piece
		occupied = occupied.ClearBit(nextAttacker.square)
		
		// Update attackers for X-ray attacks
		see.updateAttackersAfterMove(b, &whiteAttackers, &blackAttackers, target, nextAttacker.square, occupied)
		
		// Switch sides and set up for next iteration
		if sideToMove == "w" {
			sideToMove = "b"
		} else {
			sideToMove = "w"
		}
		attackingPiece = nextAttacker.piece
	}

	// Minimax through the gain array
	for depth--; depth > 0; depth-- {
		// SEE minimax: each player will choose the move that's best for them
		// But this means the opponent will choose the move that's worst for us
		// If -gain[depth] is worse for the initial side than gain[depth-1], the opponent will force it
		if gain[depth-1] > -gain[depth] {
			gain[depth-1] = -gain[depth]
		}
	}

	return gain[0]
}

// attacker represents a piece that can attack a square
type attacker struct {
	piece  board.Piece
	square int
}

// getLeastValuableAttacker finds the least valuable piece attacking the target square
func (see *SEECalculator) getLeastValuableAttacker(b *board.Board, attackers *board.Bitboard, side string, occupied board.Bitboard) attacker {
	// Check each piece type in order of value (least to most valuable)
	pieceOrder := see.getPieceOrderForSide(side)
	
	for _, pieceType := range pieceOrder {
		// Get pieces of this type that are attacking and still on the board
		pieceBitboard := b.GetPieceBitboard(pieceType)
		attackingPieces := pieceBitboard & *attackers & occupied
		
		if attackingPieces != 0 {
			// Get the first (any) piece of this type
			square, _ := attackingPieces.PopLSB()
			
			// Remove this attacker from the attackers bitboard
			*attackers = attackers.ClearBit(square)
			
			return attacker{
				piece:  pieceType,
				square: square,
			}
		}
	}
	
	return attacker{piece: board.Empty, square: -1}
}

// updateAttackersAfterMove updates attacker bitboards after a piece moves (for X-ray attacks)
func (see *SEECalculator) updateAttackersAfterMove(b *board.Board, whiteAttackers, blackAttackers *board.Bitboard, target board.Square, removedSquare int, occupied board.Bitboard) {
	// Recalculate attackers to handle X-ray attacks
	// This is simpler and more reliable than trying to update incrementally
	targetSquareIndex := board.FileRankToSquare(target.File, target.Rank)
	
	newWhiteAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardWhite) & occupied
	newBlackAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardBlack) & occupied
	
	*whiteAttackers = newWhiteAttackers
	*blackAttackers = newBlackAttackers
}

// Helper functions


// getPieceValue returns the value of a piece for SEE calculation using existing PieceValues
func (see *SEECalculator) getPieceValue(piece board.Piece) int {
	// Kings are extremely valuable for SEE (they have evaluation value 0 in PieceValues)
	if piece == board.WhiteKing || piece == board.BlackKing {
		return 10000
	}
	
	if value, ok := PieceValues[piece]; ok {
		if value < 0 {
			return -value // Return absolute value
		}
		return value
	}
	
	return 0
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

// getOppositeSide returns the opposite side color
func (see *SEECalculator) getOppositeSide(piece board.Piece) string {
	if see.isWhitePiece(piece) {
		return "b"
	}
	return "w"
}

// isWhitePiece uses existing functionality to check if a piece is white
func (see *SEECalculator) isWhitePiece(piece board.Piece) bool {
	return board.IsWhitePiece(piece)
}