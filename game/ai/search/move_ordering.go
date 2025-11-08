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

// sortMoveScores sorts move scores in descending order (highest score first)
func sortMoveScores(moves []moveScore) {
	for i := 1; i < len(moves); i++ {
		key := moves[i]
		j := i - 1
		for j >= 0 && moves[j].score < key.score {
			moves[j+1] = moves[j]
			j--
		}
		moves[j+1] = key
	}
}

// getAbsPieceValue returns the absolute value of a piece
func getAbsPieceValue(piece board.Piece) int {
	value := evaluation.GetPieceValue(piece)
	if value < 0 {
		return -value
	}
	return value
}

// applyMoveOrdering reorders the move list based on sorted scores in moveOrderBuffer
func (m *MinimaxEngine) applyMoveOrdering(moveList *moves.MoveList) {
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

				tacticalBonus := m.getTacticalBonus(b, move)
				score += tacticalBonus
			}
		}

		m.searchState.moveOrderBuffer[i] = moveScore{index: i, score: score}
	}

	// Sort using optimized insertion sort (avoids reflection overhead)
	sortMoveScores(m.searchState.moveOrderBuffer)

	m.applyMoveOrdering(moveList)
}

// orderCaptures orders captures using MVV-LVA
func (m *MinimaxEngine) orderCaptures(moveList *moves.MoveList) {
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

		victimValue := getAbsPieceValue(move.Captured)
		attackerValue := getAbsPieceValue(move.Piece)

		score := (victimValue * 10) - attackerValue
		m.searchState.moveOrderBuffer[i] = moveScore{index: i, score: score}
	}

	sortMoveScores(m.searchState.moveOrderBuffer)

	m.applyMoveOrdering(moveList)
}

// getCaptureScore calculates the capture score using SEE for accurate evaluation
// Higher scores indicate more valuable captures (better moves to try first)
// Move ordering priorities:
//  1. TT moves: 3,000,000+
//  2. Good captures (SEE > 0): 1,000,000+
//  3. Equal exchanges (SEE = 0): 900,000
//  4. Killer moves: 500,000
//  5. Tactical quiet moves (attacks piece + king zone): 150,000
//  6. Tactical quiet moves (attacks piece): 100,000
//  7. Tactical quiet moves (attacks king zone): 50,000
//  8. Slightly bad captures (SEE >= -100): 50,000+
//  9. Terrible captures (SEE < -100): 25,000+
//
// 10. Quiet moves with history: 0-10,000
// 11. Other quiet moves: 0
func (m *MinimaxEngine) getCaptureScore(b *board.Board, move board.Move) int {
	if !move.IsCapture || move.Captured == board.Empty {
		return 0
	}

	victimValue := getAbsPieceValue(move.Captured)
	attackerValue := getAbsPieceValue(move.Piece)

	mvvLvaScore := (victimValue * 10) - attackerValue

	// Skip SEE for obviously good captures (capture higher value piece)
	if victimValue > attackerValue {
		return 1000000 + mvvLvaScore
	}

	// For equal/lower value captures, use SEE
	seeValue := m.seeCalculator.SEE(b, move)

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

// getTacticalBonus calculates bonus score for moves that attack enemy pieces or king zone
// This combines both checks into one to avoid duplicate GetPieceAttacks calls
func (m *MinimaxEngine) getTacticalBonus(b *board.Board, move board.Move) int {
	toSquare := move.To.Rank*8 + move.To.File
	attacks := b.GetPieceAttacks(move.Piece, toSquare)

	var enemyPieces board.Bitboard
	var enemyKing board.Bitboard
	if move.Piece >= board.WhitePawn && move.Piece <= board.WhiteKing {
		enemyPieces = b.BlackPieces
		enemyKing = b.GetPieceBitboard(board.BlackKing)
	} else {
		enemyPieces = b.WhitePieces
		enemyKing = b.GetPieceBitboard(board.WhiteKing)
	}

	bonus := 0

	if (attacks & enemyPieces) != 0 {
		bonus += 100000
	}

	if enemyKing != 0 {
		kingSquare, _ := enemyKing.PopLSB()
		kingZone := board.GetKingAttacks(kingSquare)
		if (attacks & kingZone) != 0 {
			bonus += 50000
		}
	}

	return bonus
}
