// Package ui provides user interface components for displaying game state and moves.
package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// MovesDisplayer handles formatting and displaying move lists
type MovesDisplayer struct{}

// NewMovesDisplayer creates a new moves displayer
func NewMovesDisplayer() *MovesDisplayer {
	return &MovesDisplayer{}
}

// FormatMoveList formats a move list for display
func (md *MovesDisplayer) FormatMoveList(moveList *moves.MoveList, playerName string) string {
	if moveList.Count == 0 {
		return fmt.Sprintf("No legal moves available for %s", playerName)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Available moves for %s:\n", playerName))

	// Group moves by type
	pawnMoves := md.filterPawnMoves(moveList)

	if len(pawnMoves) > 0 {
		result.WriteString("Pawn moves:\n")
		result.WriteString(md.formatMoveGroup(pawnMoves))
		result.WriteString("\n")
	}

	// Future: Add other piece types here
	// knightMoves := md.filterKnightMoves(moveList)
	// bishopMoves := md.filterBishopMoves(moveList)
	// etc.

	return result.String()
}

// FormatMoveListCompact formats moves in a compact single-line format
func (md *MovesDisplayer) FormatMoveListCompact(moveList *moves.MoveList) string {
	if moveList.Count == 0 {
		return "No moves"
	}

	moveStrings := make([]string, 0, moveList.Count)
	for _, move := range moveList.Moves {
		moveStrings = append(moveStrings, md.formatMove(move))
	}

	sort.Strings(moveStrings)
	return strings.Join(moveStrings, ", ")
}

// filterPawnMoves extracts pawn moves from the move list
func (md *MovesDisplayer) filterPawnMoves(moveList *moves.MoveList) []board.Move {
	// For Phase 3a, all moves are pawn moves since we only generate pawn moves
	pawnMoves := append([]board.Move(nil), moveList.Moves...)

	return pawnMoves
}

// formatMoveGroup formats a group of moves with proper line wrapping
func (md *MovesDisplayer) formatMoveGroup(moves []board.Move) string {
	if len(moves) == 0 {
		return "  (none)\n"
	}

	// Convert moves to strings and sort them
	moveStrings := make([]string, len(moves))
	for i, move := range moves {
		moveStrings[i] = md.formatMove(move)
	}
	sort.Strings(moveStrings)

	// Format with line wrapping at ~70 characters
	var result strings.Builder
	result.WriteString("  ")

	currentLineLength := 2
	for i, moveStr := range moveStrings {
		if i > 0 {
			if currentLineLength+len(moveStr)+2 > 70 {
				result.WriteString("\n  ")
				currentLineLength = 2
			} else {
				result.WriteString(", ")
				currentLineLength += 2
			}
		}

		result.WriteString(moveStr)
		currentLineLength += len(moveStr)
	}

	result.WriteString("\n")
	return result.String()
}

// formatMove formats a single move for display
func (md *MovesDisplayer) formatMove(move board.Move) string {
	moveStr := fmt.Sprintf("%s%s", move.From.String(), move.To.String())

	// Add promotion notation
	if move.Promotion != board.Empty {
		promotionChar := md.getPromotionChar(move.Promotion)
		moveStr += string(promotionChar)
	}

	return moveStr
}

// getPromotionChar returns the character representation of a promotion piece
func (md *MovesDisplayer) getPromotionChar(piece board.Piece) rune {
	switch piece {
	case board.WhiteQueen, board.BlackQueen:
		return 'Q'
	case board.WhiteRook, board.BlackRook:
		return 'R'
	case board.WhiteBishop, board.BlackBishop:
		return 'B'
	case board.WhiteKnight, board.BlackKnight:
		return 'N'
	default:
		return '?'
	}
}

// ShowMoves displays the move list to the user
func (md *MovesDisplayer) ShowMoves(moveList *moves.MoveList, playerName string) {
	fmt.Print(md.FormatMoveList(moveList, playerName))
	fmt.Println()
}

// ShowMovesCompact displays moves in compact format
func (md *MovesDisplayer) ShowMovesCompact(moveList *moves.MoveList) {
	fmt.Println(md.FormatMoveListCompact(moveList))
}

// CountMovesByType returns counts of different move types for summary
func (md *MovesDisplayer) CountMovesByType(moveList *moves.MoveList) map[string]int {
	counts := make(map[string]int)

	for _, move := range moveList.Moves {
		if move.Promotion != board.Empty {
			counts["promotions"]++
		} else if move.IsCapture {
			counts["captures"]++
		} else {
			counts["quiet"]++
		}
	}

	return counts
}

// ShowMoveSummary displays a summary of available moves
func (md *MovesDisplayer) ShowMoveSummary(moveList *moves.MoveList, playerName string) {
	if moveList.Count == 0 {
		fmt.Printf("No legal moves available for %s\n", playerName)
		return
	}

	counts := md.CountMovesByType(moveList)

	fmt.Printf("%s has %d legal moves: ", playerName, moveList.Count)

	var parts []string
	if quiet, ok := counts["quiet"]; ok && quiet > 0 {
		parts = append(parts, fmt.Sprintf("%d quiet", quiet))
	}
	if captures, ok := counts["captures"]; ok && captures > 0 {
		parts = append(parts, fmt.Sprintf("%d captures", captures))
	}
	if promotions, ok := counts["promotions"]; ok && promotions > 0 {
		parts = append(parts, fmt.Sprintf("%d promotions", promotions))
	}

	fmt.Printf("(%s)\n", strings.Join(parts, ", "))
}
