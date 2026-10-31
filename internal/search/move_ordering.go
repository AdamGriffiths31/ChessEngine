package search

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

const (
	scoreTTMove        = 3000000
	scoreGoodCapture   = 1000000
	scoreEqualCapture  = 900000
	scoreKiller        = 500000
	scoreBadCapture    = 50000
	scoreWorstCapture  = 25000
	scoreQueenPromo    = 9000
	scoreRookPromo     = 5000
	scoreMinorPromo    = 3000
	historyScoreScale  = 16
	badCaptureSEEFloor = 100
	worstCaptureOffset = 1000
)

func (m *MinimaxEngine) orderBuffer(ply, n int) []int {
	buf := m.searchState.moveOrderBuffers[ply]
	if cap(buf) < n {
		buf = make([]int, n)
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

func (m *MinimaxEngine) pickNextMove(moveList *movegen.MoveList, startIndex, ply int) {
	if startIndex >= moveList.Count-1 {
		return
	}

	buffer := m.searchState.moveOrderBuffers[ply]
	bestIdx := startIndex
	bestScore := buffer[startIndex]

	for i := startIndex + 1; i < moveList.Count; i++ {
		if buffer[i] > bestScore {
			bestScore = buffer[i]
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

func (m *MinimaxEngine) scoreMoves(b *board.Board, moveList *movegen.MoveList, ply int, ttMove board.Move) {
	buffer := m.orderBuffer(ply, moveList.Count)

	for i := range moveList.Count {
		move := moveList.Moves[i]
		score := 0

		if sameMove(&move, &ttMove) {
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
				score += int(m.getHistoryScore(move)) * historyScoreScale
			}
		}

		buffer[i] = score
	}
}

func (m *MinimaxEngine) orderMovesAtRoot(b *board.Board, moveList *movegen.MoveList, ttMove board.Move) {
	m.scoreMoves(b, moveList, 0, ttMove)
	for i := range moveList.Count - 1 {
		m.pickNextMove(moveList, i, 0)
	}
}

func (m *MinimaxEngine) scoreCaptures(moveList *movegen.MoveList, ply int) {
	buffer := m.orderBuffer(ply, moveList.Count)

	for i := range moveList.Count {
		move := &moveList.Moves[i]
		buffer[i] = mvvLva(getAbsPieceValue(move.Captured), getAbsPieceValue(move.Piece))
	}
}

func mvvLva(victimValue, attackerValue int) int {
	return (victimValue * 10) - attackerValue
}

func (m *MinimaxEngine) getCaptureScore(b *board.Board, move board.Move) int {
	if move.Captured == board.Empty {
		return 0
	}

	victimValue := getAbsPieceValue(move.Captured)
	attackerValue := getAbsPieceValue(move.Piece)

	mvvLvaScore := mvvLva(victimValue, attackerValue)

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
	return m.historyTable.GetHistoryScore(move)
}
