package moves

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Validator handles move validation against legal moves
type Validator struct {
	generator *Generator
}

// NewValidator creates a new move validator
func NewValidator() *Validator {
	return &Validator{
		generator: NewGenerator(),
	}
}

// ValidateMove checks if a move is legal for the current position
func (v *Validator) ValidateMove(b *board.Board, move board.Move, player Player) bool {
	legalMoves := v.generator.GenerateAllMoves(b, player)
	return legalMoves.Contains(move)
}

// ValidatePawnMove specifically validates pawn moves
func (v *Validator) ValidatePawnMove(b *board.Board, move board.Move, player Player) bool {
	legalMoves := v.generator.GeneratePawnMoves(b, player)
	return legalMoves.Contains(move)
}

// IsMoveLegal is a convenience function that checks if a move is legal
func IsMoveLegal(b *board.Board, move board.Move, player Player) bool {
	validator := NewValidator()
	return validator.ValidateMove(b, move, player)
}

// GetLegalMoves returns all legal moves for the current position
func GetLegalMoves(b *board.Board, player Player) *MoveList {
	generator := NewGenerator()
	return generator.GenerateAllMoves(b, player)
}

// GetLegalPawnMoves returns all legal pawn moves for the current position
func GetLegalPawnMoves(b *board.Board, player Player) *MoveList {
	generator := NewGenerator()
	return generator.GeneratePawnMoves(b, player)
}

// MoveMatchesInput checks if a move matches user input (considering promotion)
func MoveMatchesInput(move board.Move, inputMove board.Move) bool {
	// Basic move comparison
	if move.From.File != inputMove.From.File ||
		move.From.Rank != inputMove.From.Rank ||
		move.To.File != inputMove.To.File ||
		move.To.Rank != inputMove.To.Rank {
		return false
	}
	
	// If input has no promotion specified, match any promotion
	if inputMove.Promotion == board.Empty {
		return true
	}
	
	// If input has promotion specified, it must match exactly
	return move.Promotion == inputMove.Promotion
}

// FindMatchingMove finds a legal move that matches the input move
func FindMatchingMove(legalMoves *MoveList, inputMove board.Move) (board.Move, bool) {
	for _, move := range legalMoves.Moves {
		if MoveMatchesInput(move, inputMove) {
			return move, true
		}
	}
	return board.Move{}, false
}

// ValidateAndFindMove validates input and returns the matching legal move
func ValidateAndFindMove(b *board.Board, inputMove board.Move, player Player) (board.Move, bool) {
	legalMoves := GetLegalMoves(b, player)
	return FindMatchingMove(legalMoves, inputMove)
}