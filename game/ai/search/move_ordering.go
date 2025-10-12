// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/evaluation"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// moveScore is a temporary structure used during move ordering to pair each move
// with its calculated score, allowing moves to be sorted by priority before search.
type moveScore struct {
	index int // Original index in the move list
	score int // Calculated ordering score (higher = better)
}

// orderMoves orders moves for search
func (m *MinimaxEngine) orderMoves(b *board.Board, moveList *moves.MoveList, depth int, ttMove board.Move) {
	if moveList.Count <= 1 {
		return
	}

	if cap(m.searchState.moveOrderBuffer) < moveList.Count {
		m.searchState.moveOrderBuffer = make([]moveScore, moveList.Count)
	} else {
		m.searchState.moveOrderBuffer = m.searchState.moveOrderBuffer[:moveList.Count]
	}

	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		score := 0

		if move.From == ttMove.From && move.To == ttMove.To && move.Promotion == ttMove.Promotion {
			score = 3000000
		} else {
			if move.IsCapture {
				score = m.getCaptureScore(b, move)
			}

			if move.Promotion != board.Empty {
				switch move.Promotion {
				case board.WhiteQueen, board.BlackQueen:
					score += 9000
				case board.WhiteRook, board.BlackRook:
					score += 5000
				case board.WhiteBishop, board.BlackBishop, board.WhiteKnight, board.BlackKnight:
					score += 3000
				}
			}

			if !move.IsCapture && m.isKillerMove(move, depth) {
				score = 500000
			}

			if !move.IsCapture && move.Promotion == board.Empty {
				score += int(m.getHistoryScore(move))
			}
		}

		m.searchState.moveOrderBuffer[i] = moveScore{index: i, score: score}
	}

	// Insertion sort - O(n²) worst case but O(n) best case, better cache locality
	for i := 1; i < moveList.Count; i++ {
		key := m.searchState.moveOrderBuffer[i]
		j := i - 1
		for j >= 0 && m.searchState.moveOrderBuffer[j].score < key.score {
			m.searchState.moveOrderBuffer[j+1] = m.searchState.moveOrderBuffer[j]
			j--
		}
		m.searchState.moveOrderBuffer[j+1] = key
	}

	if cap(m.searchState.reorderBuffer) < moveList.Count {
		m.searchState.reorderBuffer = make([]board.Move, moveList.Count)
	} else {
		m.searchState.reorderBuffer = m.searchState.reorderBuffer[:moveList.Count]
	}

	for i := 0; i < moveList.Count; i++ {
		origIndex := m.searchState.moveOrderBuffer[i].index
		m.searchState.reorderBuffer[i] = moveList.Moves[origIndex]
	}

	copy(moveList.Moves[:moveList.Count], m.searchState.reorderBuffer)
}

// orderCaptures orders captures using SEE scores
func (m *MinimaxEngine) orderCaptures(b *board.Board, moveList *moves.MoveList) {
	if moveList.Count <= 1 {
		return
	}

	if cap(m.searchState.moveOrderBuffer) < moveList.Count {
		m.searchState.moveOrderBuffer = make([]moveScore, moveList.Count)
	} else {
		m.searchState.moveOrderBuffer = m.searchState.moveOrderBuffer[:moveList.Count]
	}

	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		score := m.seeCalculator.SEE(b, move)
		m.searchState.moveOrderBuffer[i] = moveScore{index: i, score: score}
	}

	// Insertion sort - O(n²) worst case but O(n) best case, better cache locality
	for i := 1; i < moveList.Count; i++ {
		key := m.searchState.moveOrderBuffer[i]
		j := i - 1
		for j >= 0 && m.searchState.moveOrderBuffer[j].score < key.score {
			m.searchState.moveOrderBuffer[j+1] = m.searchState.moveOrderBuffer[j]
			j--
		}
		m.searchState.moveOrderBuffer[j+1] = key
	}

	if cap(m.searchState.reorderBuffer) < moveList.Count {
		m.searchState.reorderBuffer = make([]board.Move, moveList.Count)
	} else {
		m.searchState.reorderBuffer = m.searchState.reorderBuffer[:moveList.Count]
	}

	for i := 0; i < moveList.Count; i++ {
		origIndex := m.searchState.moveOrderBuffer[i].index
		m.searchState.reorderBuffer[i] = moveList.Moves[origIndex]
	}

	copy(moveList.Moves[:moveList.Count], m.searchState.reorderBuffer)
}

// getCaptureScore calculates the capture score using SEE for accurate evaluation
// Higher scores indicate more valuable captures (better moves to try first)
// Move ordering priorities:
//  1. TT moves: 3,000,000+
//  2. Good captures (SEE > 0): 1,000,000+
//  3. Equal exchanges (SEE = 0): 900,000
//  4. Killer moves: 500,000
//  5. Good history moves: up to ~50,000
//  6. Slightly bad captures (SEE >= -100): 50,000+
//  7. Terrible captures (SEE < -100): 25,000+
//  8. Quiet moves: 0
func (m *MinimaxEngine) getCaptureScore(b *board.Board, move board.Move) int {
	if !move.IsCapture || move.Captured == board.Empty {
		return 0
	}

	seeValue := m.seeCalculator.SEE(b, move)

	victimValue := evaluation.GetPieceValue(move.Captured)
	if victimValue < 0 {
		victimValue = -victimValue
	}

	attackerValue := evaluation.GetPieceValue(move.Piece)
	if attackerValue < 0 {
		attackerValue = -attackerValue
	}

	mvvLvaScore := (victimValue * 10) - attackerValue

	if seeValue > 0 {
		return 1000000 + seeValue + mvvLvaScore
	} else if seeValue == 0 {
		return 900000 + mvvLvaScore
	} else if seeValue >= -100 {
		return 50000 + seeValue + 100 + mvvLvaScore
	}
	return 25000 + seeValue + 1000 + mvvLvaScore
}

// getHistoryScore returns the history score for a move
func (m *MinimaxEngine) getHistoryScore(move board.Move) int32 {
	if m.historyTable == nil {
		return 0
	}
	return m.historyTable.GetHistoryScore(move)
}
