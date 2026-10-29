// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

type moveScore struct {
	score int
}

// Move-ordering score tiers, highest first. Quiet moves score by history alone
// (+-historyScoreScale x raw), which spans roughly +-260k: below killers, above
// the losing-capture tiers once well learned. See getCaptureScore.
const (
	scoreTTMove        = 3000000
	scoreGoodCapture   = 1000000 // SEE > 0, or victim worth more than attacker
	scoreEqualCapture  = 900000  // SEE == 0
	scoreKiller        = 500000
	scoreBadCapture    = 50000 // SEE in [-100, 0)
	scoreWorstCapture  = 25000 // SEE < -100
	scoreQueenPromo    = 9000
	scoreRookPromo     = 5000
	scoreMinorPromo    = 3000
	historyScoreScale  = 16
	badCaptureSEEFloor = 100  // lifts SEE in [-100, 0) to a non-negative offset
	worstCaptureOffset = 1000 // keeps scoreWorstCapture above 0 for very negative SEE
)

// orderBuffer returns the ply's move-ordering buffer sized for n moves,
// growing it only when it is too small.
func (m *MinimaxEngine) orderBuffer(ply, n int) []moveScore {
	buf := m.searchState.moveOrderBuffers[ply]
	if cap(buf) < n {
		buf = make([]moveScore, n)
	} else {
		buf = buf[:n]
	}
	m.searchState.moveOrderBuffers[ply] = buf
	return buf
}

func getAbsPieceValue(piece board.Piece) int {
	value := eval.GetPieceValue(piece)
	if value < 0 {
		return -value
	}
	return value
}

// pickNextMove finds the highest-scored move from startIndex onwards and swaps it to startIndex
func (m *MinimaxEngine) pickNextMove(moveList *movegen.MoveList, startIndex, ply int) {
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

// scoreMoves assigns scores to all moves for later pick-best selection.
// ply is the distance from the search root; it selects both the ordering
// buffer and the killer-move slots.
func (m *MinimaxEngine) scoreMoves(b *board.Board, moveList *movegen.MoveList, ply int, ttMove board.Move) {
	if moveList.Count == 0 {
		return
	}

	if ply < 0 || ply >= MaxKillerDepth {
		return
	}

	buffer := m.orderBuffer(ply, moveList.Count)

	for i := range moveList.Count {
		move := moveList.Moves[i]
		score := 0

		if move.From == ttMove.From && move.To == ttMove.To && move.Promotion == ttMove.Promotion {
			score = scoreTTMove
		} else {
			if move.IsCapture {
				score = m.getCaptureScore(b, move)
			}

			if move.HasPromotion() {
				switch move.Promotion {
				case board.WhiteQueen, board.BlackQueen:
					score += scoreQueenPromo
				case board.WhiteRook, board.BlackRook:
					score += scoreRookPromo
				case board.WhiteBishop, board.BlackBishop, board.WhiteKnight, board.BlackKnight:
					score += scoreMinorPromo
				}
			}

			if !move.IsCapture && m.isKillerMove(move, ply) {
				score = scoreKiller
			}

			if move.IsQuiet() {
				// Quiet moves are ordered by history, scaled x16 (+-262k): below
				// killers (500k), above losing captures once well learned.
				score += int(m.getHistoryScore(move)) * historyScoreScale
			}
		}

		buffer[i] = moveScore{score: score}
	}
}

// orderMovesAtRoot scores and fully sorts moves for the root node
func (m *MinimaxEngine) orderMovesAtRoot(b *board.Board, moveList *movegen.MoveList, ttMove board.Move) {
	m.scoreMoves(b, moveList, 0, ttMove)
	for i := range moveList.Count - 1 {
		m.pickNextMove(moveList, i, 0)
	}
}

// scoreCaptures scores captures using MVV-LVA for later pick-best selection
func (m *MinimaxEngine) scoreCaptures(moveList *movegen.MoveList, ply int) {
	if moveList.Count == 0 {
		return
	}

	if ply < 0 || ply >= MaxKillerDepth {
		return
	}

	buffer := m.orderBuffer(ply, moveList.Count)

	for i := range moveList.Count {
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
//
//  1. TT moves: 3,000,000+
//
//  2. Good captures (SEE > 0): 1,000,000+
//
//  3. Equal exchanges (SEE = 0): 900,000
//
//  4. Killer moves: 500,000
//
//  5. Slightly bad captures (SEE >= -100): 50,000+
//
//  6. Terrible captures (SEE < -100): 25,000+
//
//  7. Quiet moves, ordered by history: roughly -260k..+260k (raw +-16k x16),
//     so a well-learned quiet outranks slightly and very bad captures
//
//  8. Quiet moves with no history: 0
func (m *MinimaxEngine) getCaptureScore(b *board.Board, move board.Move) int {
	if !move.IsCapture || move.Captured == board.Empty {
		return 0
	}

	victimValue := getAbsPieceValue(move.Captured)
	attackerValue := getAbsPieceValue(move.Piece)

	mvvLvaScore := (victimValue * 10) - attackerValue

	if victimValue > attackerValue {
		return scoreGoodCapture + mvvLvaScore
	}

	seeValue := m.seeCalculator.SEE(b, move)

	switch {
	case seeValue > 0:
		return scoreGoodCapture + seeValue + mvvLvaScore
	case seeValue == 0:
		return scoreEqualCapture + mvvLvaScore
	case seeValue >= -100:
		return scoreBadCapture + seeValue + badCaptureSEEFloor + mvvLvaScore
	}
	return scoreWorstCapture + seeValue + worstCaptureOffset + mvvLvaScore
}

func (m *MinimaxEngine) getHistoryScore(move board.Move) int32 {
	if m.historyTable == nil {
		return 0
	}
	return m.historyTable.GetHistoryScore(move)
}
