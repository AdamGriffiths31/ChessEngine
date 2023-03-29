package eval

import (
	"math/bits"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/engine"
)

type EvaluationService struct {
	Weights
	pieceCount [2][7]int
	force      [2]int
}

func NewEvaluationService() *EvaluationService {
	var es = &EvaluationService{}
	es.Weights.init()
	return es
}

const (
	darkSquares = uint64(0xAA55AA55AA55AA55)
	minorPhase  = 4
	rookPhase   = 6
	queenPhase  = 12
	totalPhase  = 2 * (4*minorPhase + 2*rookPhase + queenPhase)
)

func flip(sq int) int {
	return data.Mirror64[sq]
}

func (e *EvaluationService) Clear() {
	e.force[data.White] = 0
	e.force[data.Black] = 0

	for i := 0; i < 2; i++ {
		for j := 0; j < 6; j++ {
			e.pieceCount[i][j] = 0
		}
	}
}

func (e *EvaluationService) Evaluate(p *engine.Position) int {
	e.Clear()
	var score Score

	for wp := p.Board.WhitePieces; wp != 0; wp &= wp - 1 {
		sq := bits.TrailingZeros64(wp)
		pc := p.Board.PieceAt(sq)
		score += e.PST[data.White][pc-1][sq]
		e.pieceCount[data.White][pc-1]++
	}

	for bp := p.Board.BlackPieces; bp != 0; bp &= bp - 1 {
		sq := bits.TrailingZeros64(bp)
		pc := p.Board.PieceAt(sq)
		score += e.PST[data.Black][pc-7][sq]
		e.pieceCount[data.Black][pc-7]++
	}

	e.force[data.White] = minorPhase*(e.pieceCount[data.White][data.WN-1]+e.pieceCount[data.White][data.WB-1]) +
		rookPhase*e.pieceCount[data.White][data.WR-1] + queenPhase*e.pieceCount[data.White][data.WQ-1]

	e.force[data.Black] = minorPhase*(e.pieceCount[data.Black][data.WN-1]+e.pieceCount[data.Black][data.WB-1]) +
		rookPhase*e.pieceCount[data.Black][data.WR-1] + queenPhase*e.pieceCount[data.Black][data.WQ-1]

	if e.pieceCount[data.White][data.WB-1] >= 2 {
		score += 20
	}
	if e.pieceCount[data.Black][data.WB-1] >= 2 {
		score -= 20
	}

	phase := e.force[data.White] + e.force[data.Black]
	if phase > totalPhase {
		phase = totalPhase
	}

	result := (int(score.Middle())*phase + int(score.End())*(totalPhase-phase)) / totalPhase

	combinedPosition := combineBishops(p.Board.BlackBishop, p.Board.WhiteBishop)

	ocb := e.force[data.White] == minorPhase &&
		e.force[data.Black] == minorPhase &&
		(combinedPosition&darkSquares) != 0 &&
		(combinedPosition & ^darkSquares) != 0

	if result > 0 {
		result = result * computeFactor(e, data.White, ocb) / scaleNormal
	} else {
		result = result * computeFactor(e, data.Black, ocb) / scaleNormal
	}

	if p.Side == data.Black {
		result = -result
	}

	return result
}

const (
	scaleDraw   = 0
	scaleHard   = 1
	scaleNormal = 2
)

func combineBishops(white, black uint64) uint64 {
	return white | black
}

func computeFactor(e *EvaluationService, side int, ocb bool) int {
	if e.force[side] >= queenPhase+rookPhase {
		return scaleNormal
	}
	if e.pieceCount[side][data.WP-1] == 0 {
		if e.force[side] <= minorPhase {
			return scaleHard
		}
		if e.force[side] == 2*minorPhase && e.pieceCount[side][data.WN-1] == 2 && e.pieceCount[side^1][data.WP-1] == 0 {
			return scaleHard
		}
		if e.force[side]-e.force[side^1] <= minorPhase {
			return scaleHard
		}
	} else if e.pieceCount[side][data.WP-1] == 1 {
		if e.force[side] <= minorPhase && e.pieceCount[side^1][data.WN-1]+e.pieceCount[side^1][data.WB-1] != 0 {
			return scaleHard
		}
		if e.force[side] == e.force[side^1] && e.pieceCount[side^1][data.WN-1]+e.pieceCount[side^1][data.WB-1] != 0 {
			return scaleHard
		}
	} else if ocb && e.pieceCount[side][data.WP-1]-e.pieceCount[side^1][data.WP-1] <= 2 {
		return scaleHard
	}
	return scaleNormal
}
