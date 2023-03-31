package eval

import (
	"math"
	"math/bits"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/engine"
)

type EvaluationService struct {
	Weights

	mobilityAreas [2]uint64
	pawnAttacks   [2]uint64
}

func NewEvaluationService() *EvaluationService {
	var es = &EvaluationService{}
	es.Weights.init()
	return es
}

func (e *EvaluationService) Evaluate(p *engine.Position) int {

	if p.Board.WhitePawn == 0 && p.Board.BlackPawn == 0 && e.IsMaterialDraw(p) {
		return 0
	}
	p.CurrentScore = e.calculateWhiteMaterial(p) - e.calculateBlackMaterial(p)

	bothPawns := p.Board.WhitePawn | p.Board.BlackPawn

	e.SetupEvaluate(p)

	e.calculateEvalPawns(p)
	e.calculateEvalKnights(p)
	e.calculateEvalBishop(p)
	e.calculateEvalRook(p, bothPawns)
	e.calculateEvalQueens(p, bothPawns)
	e.calculateEvalKings(p)

	eval := e.evaluateThreats(p, data.White, bothPawns) - e.evaluateThreats(p, data.Black, bothPawns)
	eval += e.evaluateMobility(p)

	p.CurrentScore += eval.Middle()

	if p.Side == data.White {
		return p.CurrentScore
	} else {
		return -p.CurrentScore
	}
}

func (e *EvaluationService) SetupEvaluate(p *engine.Position) {
	bothPawns := p.Board.WhitePawn | p.Board.BlackPawn

	e.pawnAttacks[data.White] = p.Board.AllWhitePawnAttacks(bothPawns & p.Board.WhitePawn)
	e.pawnAttacks[data.Black] = p.Board.AllBlackPawnAttacks(bothPawns & p.Board.BlackPawn)

	e.mobilityAreas[data.White] = ^(e.pawnAttacks[data.Black] | bothPawns&p.Board.WhitePieces&(data.Rank2Mask|(p.Board.Pieces>>8)))
	e.mobilityAreas[data.Black] = ^(e.pawnAttacks[data.White] | bothPawns&p.Board.BlackPieces&(data.Rank7Mask|(p.Board.Pieces<<8)))
}

func (e *EvaluationService) calculateWhiteMaterial(p *engine.Position) int {
	score := 0
	score += p.Board.CountBits(p.Board.WhitePawn) * data.PieceVal[data.WP]
	score += p.Board.CountBits(p.Board.WhiteKnight) * data.PieceVal[data.WN]
	score += p.Board.CountBits(p.Board.WhiteBishop) * data.PieceVal[data.WB]
	score += p.Board.CountBits(p.Board.WhiteRook) * data.PieceVal[data.WR]
	score += p.Board.CountBits(p.Board.WhiteQueen) * data.PieceVal[data.WQ]
	score += 1 * data.PieceVal[data.WK]
	return score
}

func (e *EvaluationService) calculateBlackMaterial(p *engine.Position) int {
	score := 0
	score += p.Board.CountBits(p.Board.BlackPawn) * data.PieceVal[data.WP]
	score += p.Board.CountBits(p.Board.BlackKnight) * data.PieceVal[data.WN]
	score += p.Board.CountBits(p.Board.BlackBishop) * data.PieceVal[data.WB]
	score += p.Board.CountBits(p.Board.BlackRook) * data.PieceVal[data.WR]
	score += p.Board.CountBits(p.Board.BlackQueen) * data.PieceVal[data.WQ]
	score += 1 * data.PieceVal[data.WK]
	return score
}

func isEndGame() int {
	return (1 * data.PieceVal[data.WR]) + (2 * data.PieceVal[data.WN]) + (2 * data.PieceVal[data.WP]) + data.PieceVal[data.WK]
}

func (e *EvaluationService) calculateEvalPawns(p *engine.Position) {
	for wp := p.Board.WhitePawn; wp != 0; wp &= wp - 1 {
		sq := bits.TrailingZeros64(wp)
		p.CurrentScore += pawnTable[sq]
		if data.IsolatedMask[sq]&p.Board.WhitePawn == 0 {
			p.CurrentScore += pawnIsolated
		}

		if data.WhitePassedMask[sq]&p.Board.BlackPawn == 0 {
			p.CurrentScore += pawnPassed[data.RanksBoard[data.Square64ToSquare120[sq]]]
		}
	}

	for bp := p.Board.BlackPawn; bp != 0; bp &= bp - 1 {
		sq := bits.TrailingZeros64(bp)
		p.CurrentScore -= pawnTable[data.Mirror64[sq]]

		if data.IsolatedMask[sq]&p.Board.BlackPawn == 0 {
			p.CurrentScore -= pawnIsolated
		}

		if data.BlackPassedMask[sq]&p.Board.WhitePawn == 0 {
			p.CurrentScore -= pawnPassed[data.Rank8-data.RanksBoard[data.Square64ToSquare120[sq]]]
		}
	}
}
func (e *EvaluationService) calculateEvalKnights(p *engine.Position) {
	bb := p.Board.WhiteKnight
	for bb != 0 {
		sq := bits.TrailingZeros64(bb)
		bb ^= (1 << sq)
		p.CurrentScore += knightTable[sq]
	}
	bb = p.Board.BlackKnight
	for bb != 0 {
		sq := bits.TrailingZeros64(bb)
		bb ^= (1 << sq)
		p.CurrentScore -= knightTable[data.Mirror64[sq]]
	}
}

func (e *EvaluationService) calculateEvalBishop(p *engine.Position) {
	bb := p.Board.WhiteBishop
	countWB := 0
	countBB := 0
	for bb != 0 {
		countWB++
		sq := bits.TrailingZeros64(bb)
		bb ^= (1 << sq)
		p.CurrentScore += bishopTable[sq]
	}
	bb = p.Board.BlackBishop
	for bb != 0 {
		countBB++
		sq := bits.TrailingZeros64(bb)
		bb ^= (1 << sq)
		p.CurrentScore -= bishopTable[data.Mirror64[sq]]
	}
	if countWB >= 2 {
		p.CurrentScore += bishopPair
	}
	if countBB >= 2 {
		p.CurrentScore -= bishopPair
	}
}
func (e *EvaluationService) calculateEvalRook(p *engine.Position, bothPawns uint64) {
	bb := p.Board.WhiteRook
	for bb != 0 {
		sq := bits.TrailingZeros64(bb)
		bb ^= (1 << sq)
		p.CurrentScore += rookTable[sq]

		if bothPawns&data.FileBBMask[data.FilesBoard[data.Square64ToSquare120[sq]]] == 0 {
			p.CurrentScore += rookOpenFile
		} else if p.Board.WhitePawn&data.FileBBMask[data.FilesBoard[data.Square64ToSquare120[sq]]] == 0 {
			p.CurrentScore += rookSemiOpenFile
		}
	}
	bb = p.Board.BlackRook
	for bb != 0 {
		sq := bits.TrailingZeros64(bb)
		bb ^= (1 << sq)
		p.CurrentScore -= rookTable[data.Mirror64[sq]]

		if bothPawns&data.FileBBMask[data.FilesBoard[data.Square64ToSquare120[sq]]] == 0 {
			p.CurrentScore -= rookOpenFile
		} else if p.Board.BlackPawn&data.FileBBMask[data.FilesBoard[data.Square64ToSquare120[sq]]] == 0 {
			p.CurrentScore -= rookSemiOpenFile
		}
	}
}

func (e *EvaluationService) calculateEvalQueens(p *engine.Position, bothPawns uint64) {
	bb := p.Board.WhiteQueen
	for bb != 0 {
		sq := bits.TrailingZeros64(bb)
		bb ^= (1 << sq)
		if bothPawns&data.FileBBMask[data.FilesBoard[data.Square64ToSquare120[sq]]] == 0 {
			p.CurrentScore += queenOpenFile
		} else if p.Board.WhitePawn&data.FileBBMask[data.FilesBoard[data.Square64ToSquare120[sq]]] == 0 {
			p.CurrentScore += queenSemiOpenFile
		}
	}
	bb = p.Board.BlackQueen
	for bb != 0 {
		sq := bits.TrailingZeros64(bb)
		bb ^= (1 << sq)
		if bothPawns&data.FileBBMask[data.FilesBoard[data.Square64ToSquare120[sq]]] == 0 {
			p.CurrentScore -= queenOpenFile
		} else if p.Board.BlackPawn&data.FileBBMask[data.FilesBoard[data.Square64ToSquare120[sq]]] == 0 {
			p.CurrentScore -= queenSemiOpenFile
		}
	}
}

func (e *EvaluationService) calculateEvalKings(p *engine.Position) {
	bb := p.Board.WhiteKing
	for bb != 0 {
		sq := bits.TrailingZeros64(bb)
		bb ^= (1 << sq)
		if e.calculateBlackMaterial(p) <= isEndGame() {
			p.CurrentScore += kingE[sq]
		} else {
			p.CurrentScore += kingO[sq]
		}
	}
	bb = p.Board.BlackKing
	for bb != 0 {
		sq := bits.TrailingZeros64(bb)
		bb ^= (1 << sq)
		if e.calculateWhiteMaterial(p) <= isEndGame() {
			p.CurrentScore -= kingE[data.Mirror64[sq]]
		} else {
			p.CurrentScore -= kingO[data.Mirror64[sq]]
		}
	}
}

func (e *EvaluationService) evaluateMobility(p *engine.Position) Score {
	eval := e.EvaluateMobilityKnights(p, data.White) - e.EvaluateMobilityKnights(p, data.Black)
	eval += e.EvaluateMobilityBishops(p, data.White) - e.EvaluateMobilityBishops(p, data.Black)

	return eval
}

func (e *EvaluationService) EvaluateMobilityKnights(p *engine.Position, colour int) Score {
	friendly := p.Board.GetPieces(colour, data.WN)
	eval := Score(0)

	for friendly != 0 {
		sq := engine.FirstSquare(friendly)
		eval += e.KnightMobility[p.Board.CountBits(e.mobilityAreas[colour]&engine.PreCalculatedKnightMoves[sq])]
		friendly &= friendly - 1
	}

	return eval
}

// TODO Add x-ray attacks
func (e *EvaluationService) EvaluateMobilityBishops(p *engine.Position, colour int) Score {
	friendly := p.Board.GetPieces(colour, data.WB)
	eval := Score(0)

	for friendly != 0 {
		sq := engine.FirstSquare(friendly)
		attack := data.GetBishopAttacks(p.Board.Pieces, sq)
		eval += e.BishopMobility[p.Board.CountBits(e.mobilityAreas[colour]&attack)]
		friendly &= friendly - 1
	}

	return eval
}

func (e *EvaluationService) evaluateThreats(p *engine.Position, colour int, bothPawns uint64) Score {
	var currentSide, enemySide = colour, colour ^ 1
	var friendly uint64
	var eval Score

	if currentSide == data.White {
		friendly = p.Board.WhitePieces
	} else {
		friendly = p.Board.BlackPieces
	}

	count := p.Board.CountBits(bothPawns & friendly & e.pawnAttacks[enemySide])
	eval += Score(count) * e.ThreatByPawn

	var pawnPushAttacks uint64
	if currentSide == data.White {
		pawnPushAttacks = (bothPawns << 8) & ^p.Board.Pieces
	} else {
		pawnPushAttacks = (bothPawns >> 8) & ^p.Board.Pieces
	}
	count = p.Board.CountBits(^bothPawns & friendly & pawnPushAttacks)
	eval += Score(count) * e.ThreatByPawnPush

	return eval
}

func (e *EvaluationService) IsMaterialDraw(p *engine.Position) bool {
	if p.Board.WhiteQueen == 0 && p.Board.BlackQueen == 0 ||
		p.Board.WhiteQueen == 0 && p.Board.BlackQueen == 0 && p.Board.WhiteRook == 0 && p.Board.BlackRook == 0 {
		wKnightsNum := p.Board.CountBits(p.Board.WhiteKnight)
		bKnightsNum := p.Board.CountBits(p.Board.BlackKnight)
		wBishopsNum := p.Board.CountBits(p.Board.WhiteBishop)
		bBishopsNum := p.Board.CountBits(p.Board.BlackBishop)
		wRooksNum := p.Board.CountBits(p.Board.WhiteRook)
		bRooksNum := p.Board.CountBits(p.Board.BlackRook)
		wQueenNum := p.Board.CountBits(p.Board.WhiteQueen)
		bQueenNum := p.Board.CountBits(p.Board.BlackQueen)

		if wRooksNum == 0 && bRooksNum == 0 && wQueenNum == 0 && bQueenNum == 0 {
			if wBishopsNum == 0 && bBishopsNum == 0 {
				if wKnightsNum < 3 && bKnightsNum < 3 {
					return true
				}
			} else if wKnightsNum == 0 && bKnightsNum == 0 {
				if math.Abs(float64(wBishopsNum-bBishopsNum)) < 2 {
					return true
				}
			} else if (wKnightsNum < 3 && wBishopsNum == 0) || (wBishopsNum == 1 && wKnightsNum == 0) {
				if (bKnightsNum < 3 && bBishopsNum == 0) || (bBishopsNum == 1 && bKnightsNum == 0) {
					return true
				}
			}
		} else if wQueenNum == 0 && bQueenNum == 0 {
			if wRooksNum == 1 && bRooksNum == 1 {
				if (wKnightsNum+wBishopsNum) < 2 && (bKnightsNum+bBishopsNum) < 2 {
					return true
				}
			} else if wRooksNum == 1 && bRooksNum == 0 {
				if (wKnightsNum+wBishopsNum == 0) && (((bKnightsNum + bBishopsNum) == 1) || ((bKnightsNum + bBishopsNum) == 2)) {
					return true
				}
			} else if bRooksNum == 1 && wRooksNum == 0 {
				if (bKnightsNum+bBishopsNum == 0) && (((wKnightsNum + wBishopsNum) == 1) || ((wKnightsNum + wBishopsNum) == 2)) {
					return true
				}
			}
		}
	}

	return false
}
