// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/evaluation"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

type moveScore struct {
	score int
}

// getAbsPieceValue returns the absolute value of a piece
func getAbsPieceValue(piece board.Piece) int {
	value := evaluation.GetPieceValue(piece)
	if value < 0 {
		return -value
	}
	return value
}

// pickNextMove finds the highest-scored move from startIndex onwards and swaps it to startIndex
func (m *MinimaxEngine) pickNextMove(moveList *moves.MoveList, startIndex, ply int) {
	if startIndex >= moveList.Count-1 {
		return
	}

	if ply < 0 || ply >= MaxKillerDepth {
		return
	}

	buffer := m.searchState.moveOrderBuffers[ply]
	bestIdx := startIndex
	bestScore := buffer[startIndex].score

	for i := startIndex + 1; i < moveList.Count; i++ {
		if buffer[i].score > bestScore {
			bestScore = buffer[i].score
			bestIdx = i
		}
	}

	if bestIdx != startIndex {
		moveList.Moves[startIndex], moveList.Moves[bestIdx] =
			moveList.Moves[bestIdx], moveList.Moves[startIndex]
		buffer[startIndex], buffer[bestIdx] =
			buffer[bestIdx], buffer[startIndex]
	}
}

// scoreMoves assigns scores to all moves for later pick-best selection
func (m *MinimaxEngine) scoreMoves(b *board.Board, moveList *moves.MoveList, depth, ply int, ttMove board.Move) {
	if moveList.Count == 0 {
		return
	}

	if ply < 0 || ply >= MaxKillerDepth {
		return
	}

	if cap(m.searchState.moveOrderBuffers[ply]) < moveList.Count {
		m.searchState.moveOrderBuffers[ply] = make([]moveScore, moveList.Count)
	} else if len(m.searchState.moveOrderBuffers[ply]) < moveList.Count {
		m.searchState.moveOrderBuffers[ply] = m.searchState.moveOrderBuffers[ply][:moveList.Count]
	}

	buffer := m.searchState.moveOrderBuffers[ply]

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

		buffer[i] = moveScore{score: score}
	}
}

// orderMovesAtRoot scores and fully sorts moves for the root node
func (m *MinimaxEngine) orderMovesAtRoot(b *board.Board, moveList *moves.MoveList, ttMove board.Move) {
	m.scoreMoves(b, moveList, 0, 0, ttMove)
	for i := 0; i < moveList.Count-1; i++ {
		m.pickNextMove(moveList, i, 0)
	}
}

// scoreCaptures scores captures using MVV-LVA for later pick-best selection
func (m *MinimaxEngine) scoreCaptures(moveList *moves.MoveList, ply int) {
	if moveList.Count == 0 {
		return
	}

	if ply < 0 || ply >= MaxKillerDepth {
		return
	}

	if cap(m.searchState.moveOrderBuffers[ply]) < moveList.Count {
		m.searchState.moveOrderBuffers[ply] = make([]moveScore, moveList.Count)
	} else if len(m.searchState.moveOrderBuffers[ply]) < moveList.Count {
		m.searchState.moveOrderBuffers[ply] = m.searchState.moveOrderBuffers[ply][:moveList.Count]
	}

	buffer := m.searchState.moveOrderBuffers[ply]

	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]

		victimValue := getAbsPieceValue(move.Captured)
		attackerValue := getAbsPieceValue(move.Piece)

		score := (victimValue * 10) - attackerValue
		buffer[i] = moveScore{score: score}
	}
}

// getCaptureScore calculates the capture score using SEE for accurate evaluation.
// Higher scores indicate more valuable captures (better moves to try first).
//
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
// 10. Quiet moves with history: 0-10,000
// 11. Other quiet moves: 0
func (m *MinimaxEngine) getCaptureScore(b *board.Board, move board.Move) int {
	if !move.IsCapture || move.Captured == board.Empty {
		return 0
	}

	victimValue := getAbsPieceValue(move.Captured)
	attackerValue := getAbsPieceValue(move.Piece)

	mvvLvaScore := (victimValue * 10) - attackerValue

	if victimValue > attackerValue {
		return 1000000 + mvvLvaScore
	}

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
