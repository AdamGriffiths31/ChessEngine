package eval

import (
	"math"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/engine"
	"github.com/AdamGriffiths31/ChessEngine/util"
)

type EvaluationService struct {
	Weights
	pieceCount [2][7]int

	mobilityAreas [2]uint64
	pawnAttacks   [2]uint64
}

func NewEvaluationService() *EvaluationService {
	var es = &EvaluationService{}
	es.Weights.init()
	return es
}

const (
	QueenSideBB = data.FileAMask | data.FileBMask | data.FileCMask | data.FileDMask
	KingSideBB  = data.FileEMask | data.FileFMask | data.FileGMask | data.FileHMask
)

func (e *EvaluationService) Evaluate(p *engine.Position) int {

	if p.Board.WhitePawn == 0 && p.Board.BlackPawn == 0 && e.IsMaterialDraw(p) {
		return 0
	}

	bothPawns := p.Board.WhitePawn | p.Board.BlackPawn

	e.SetupEvaluate(p)

	eval := e.calculateEvalPawns(p)
	eval += e.calculateEvalKnights(p)
	eval += e.calculateEvalBishop(p)
	eval += e.calculateEvalRook(p, bothPawns)
	eval += e.calculateEvalQueens(p, bothPawns)
	eval += e.calculateEvalKings(p)

	eval += e.evaluateThreats(p, data.White, bothPawns) - e.evaluateThreats(p, data.Black, bothPawns)
	eval += e.evaluateMobility(p)

	eval += e.PawnValue * Score(e.pieceCount[data.White][data.WP]-e.pieceCount[data.Black][data.WP])
	eval += e.KnightValue * Score(e.pieceCount[data.White][data.WN]-e.pieceCount[data.Black][data.WN])
	eval += e.BishopValue * Score(e.pieceCount[data.White][data.WB]-e.pieceCount[data.Black][data.WB])
	eval += e.RookValue * Score(e.pieceCount[data.White][data.WR]-e.pieceCount[data.Black][data.WR])
	eval += e.QueenValue * Score(e.pieceCount[data.White][data.WQ]-e.pieceCount[data.Black][data.WQ])

	factor := computeFactor(e, p, eval, bothPawns)

	var phase = 4*(e.pieceCount[data.White][data.WQ]+e.pieceCount[data.Black][data.WQ]) +
		2*(e.pieceCount[data.White][data.WR]+e.pieceCount[data.Black][data.WR]) +
		1*(e.pieceCount[data.White][data.WN]+e.pieceCount[data.Black][data.WN]+
			e.pieceCount[data.White][data.WB]+e.pieceCount[data.Black][data.WB])

	phase = (phase*256 + 12) / 24

	result := (eval.Middle()*phase +
		eval.End()*(256-phase)*factor/scaleFactorNormal) / 256

	if p.Side == data.White {
		return result
	} else {
		return -result
	}
}

func (e *EvaluationService) SetupEvaluate(p *engine.Position) {
	for pt := data.WP; pt <= data.WK; pt++ {
		e.pieceCount[data.White][pt] = 0
		e.pieceCount[data.Black][pt] = 0
	}

	bothPawns := p.Board.WhitePawn | p.Board.BlackPawn

	e.pawnAttacks[data.White] = p.Board.AllWhitePawnAttacks(bothPawns & p.Board.WhitePawn)
	e.pawnAttacks[data.Black] = p.Board.AllBlackPawnAttacks(bothPawns & p.Board.BlackPawn)

	e.mobilityAreas[data.White] = ^(e.pawnAttacks[data.Black] | bothPawns&p.Board.WhitePieces&(data.Rank2Mask|(p.Board.Pieces>>8)))
	e.mobilityAreas[data.Black] = ^(e.pawnAttacks[data.White] | bothPawns&p.Board.BlackPieces&(data.Rank7Mask|(p.Board.Pieces<<8)))
}

func isEndGame() int {
	return (1 * data.PieceVal[data.WR]) + (2 * data.PieceVal[data.WN]) + (2 * data.PieceVal[data.WP]) + data.PieceVal[data.WK]
}

func (e *EvaluationService) calculateEvalPawns(p *engine.Position) Score {
	var eval Score
	for wp := p.Board.WhitePawn; wp != 0; wp &= wp - 1 {
		e.pieceCount[data.White][data.WP]++
		sq := engine.FirstSquare(wp)
		eval += e.PSQT[data.White][data.WP][sq]
		if data.IsolatedMask[sq]&p.Board.WhitePawn == 0 {
			eval += e.PawnIsolated
		}

		if data.WhitePassedMask[sq]&p.Board.BlackPawn == 0 {
			eval += e.PassedPawn[data.RanksBoard[data.Square64ToSquare120[sq]]]
		}
	}

	for bp := p.Board.BlackPawn; bp != 0; bp &= bp - 1 {
		e.pieceCount[data.Black][data.WP]++
		sq := engine.FirstSquare(bp)
		eval -= e.PSQT[data.Black][data.WP][sq]

		if data.IsolatedMask[sq]&p.Board.BlackPawn == 0 {
			eval -= e.PawnIsolated
		}

		if data.BlackPassedMask[sq]&p.Board.WhitePawn == 0 {
			eval -= e.PassedPawn[data.RanksBoard[data.Square64ToSquare120[flip(sq)]]]
		}
	}

	return eval
}

func (e *EvaluationService) calculateEvalKnights(p *engine.Position) Score {
	var eval Score
	bb := p.Board.WhiteKnight
	for bb != 0 {
		e.pieceCount[data.White][data.WN]++
		sq := engine.FirstSquare(bb)
		eval += e.PSQT[data.White][data.WN][sq]
		bb ^= (1 << sq)
	}
	bb = p.Board.BlackKnight
	for bb != 0 {
		e.pieceCount[data.Black][data.WN]++
		sq := engine.FirstSquare(bb)
		eval -= e.PSQT[data.Black][data.WN][sq]
		bb ^= (1 << sq)
	}

	return eval
}

func (e *EvaluationService) calculateEvalBishop(p *engine.Position) Score {
	var eval Score
	bb := p.Board.WhiteBishop
	countWB := 0
	countBB := 0
	for bb != 0 {
		e.pieceCount[data.White][data.WB]++
		countWB++
		sq := engine.FirstSquare(bb)
		eval += e.PSQT[data.White][data.WB][sq]
		bb ^= (1 << sq)
	}

	bb = p.Board.BlackBishop
	for bb != 0 {
		e.pieceCount[data.Black][data.WB]++
		countBB++
		sq := engine.FirstSquare(bb)
		eval -= e.PSQT[data.Black][data.WB][sq]
		bb ^= (1 << sq)
	}
	if countWB >= 2 {
		eval += e.BishopPair
	}
	if countBB >= 2 {
		eval -= e.BishopPair
	}
	return eval
}

func (e *EvaluationService) calculateEvalRook(p *engine.Position, bothPawns uint64) Score {
	var eval Score

	bb := p.Board.WhiteRook
	for bb != 0 {
		e.pieceCount[data.White][data.WR]++
		sq := engine.FirstSquare(bb)
		eval += e.PSQT[data.White][data.WR][sq]
		bb ^= (1 << sq)

		if bothPawns&data.FileBBMask[data.FilesBoard[data.Square64ToSquare120[sq]]] == 0 {
			eval += e.RookOpenFile
		} else if p.Board.WhitePawn&data.FileBBMask[data.FilesBoard[data.Square64ToSquare120[sq]]] == 0 {
			eval += e.RookSemiOpenFile
		}
	}

	bb = p.Board.BlackRook
	for bb != 0 {
		e.pieceCount[data.Black][data.WR]++
		sq := engine.FirstSquare(bb)
		eval -= e.PSQT[data.Black][data.WR][sq]
		bb ^= (1 << sq)

		if bothPawns&data.FileBBMask[data.FilesBoard[data.Square64ToSquare120[sq]]] == 0 {
			eval -= e.RookOpenFile
		} else if p.Board.BlackPawn&data.FileBBMask[data.FilesBoard[data.Square64ToSquare120[sq]]] == 0 {
			eval -= e.RookSemiOpenFile
		}
	}

	return eval
}

func (e *EvaluationService) calculateEvalQueens(p *engine.Position, bothPawns uint64) Score {
	var eval Score

	bb := p.Board.WhiteQueen
	for bb != 0 {
		e.pieceCount[data.White][data.WQ]++
		sq := engine.FirstSquare(bb)
		eval += e.PSQT[data.White][data.WQ][sq]
		bb ^= (1 << sq)

		if bothPawns&data.FileBBMask[data.FilesBoard[data.Square64ToSquare120[sq]]] == 0 {
			eval += e.QueenOpenFile
		} else if p.Board.WhitePawn&data.FileBBMask[data.FilesBoard[data.Square64ToSquare120[sq]]] == 0 {
			eval += e.QueenSemiOpenFile
		}
	}
	bb = p.Board.BlackQueen
	for bb != 0 {
		e.pieceCount[data.Black][data.WQ]++
		sq := engine.FirstSquare(bb)
		eval -= e.PSQT[data.Black][data.WQ][sq]
		bb ^= (1 << sq)

		if bothPawns&data.FileBBMask[data.FilesBoard[data.Square64ToSquare120[sq]]] == 0 {
			eval -= e.QueenOpenFile
		} else if p.Board.BlackPawn&data.FileBBMask[data.FilesBoard[data.Square64ToSquare120[sq]]] == 0 {
			eval -= e.QueenSemiOpenFile
		}
	}

	return eval
}

func (e *EvaluationService) calculateEvalKings(p *engine.Position) Score {
	var eval Score

	bb := p.Board.WhiteKing
	for bb != 0 {
		sq := engine.FirstSquare(bb)
		eval += e.PSQT[data.White][data.WK][sq]
		bb ^= (1 << sq)
	}

	bb = p.Board.BlackKing
	for bb != 0 {
		sq := engine.FirstSquare(bb)
		eval -= e.PSQT[data.Black][data.WK][sq]
		bb ^= (1 << sq)
	}

	return eval
}

func (e *EvaluationService) evaluateMobility(p *engine.Position) Score {
	eval := e.EvaluateMobilityKnights(p, data.White) - e.EvaluateMobilityKnights(p, data.Black)
	eval += e.EvaluateMobilityBishops(p, data.White) - e.EvaluateMobilityBishops(p, data.Black)
	eval += e.EvaluateMobilityRooks(p, data.White) - e.EvaluateMobilityRooks(p, data.Black)
	eval += e.EvaluateMobilityQueens(p, data.White) - e.EvaluateMobilityQueens(p, data.Black)

	return eval
}

func (e *EvaluationService) EvaluateMobilityKnights(p *engine.Position, colour int) Score {
	friendly := p.Board.GetPieces(colour, data.WN)
	var eval Score

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
	var eval Score

	for friendly != 0 {
		sq := engine.FirstSquare(friendly)
		attack := data.GetBishopAttacks(p.Board.Pieces, sq)
		eval += e.BishopMobility[p.Board.CountBits(e.mobilityAreas[colour]&attack)]
		friendly &= friendly - 1
	}

	return eval
}

func (e *EvaluationService) EvaluateMobilityQueens(p *engine.Position, colour int) Score {
	friendly := p.Board.GetPieces(colour, data.WQ)
	var eval Score

	for friendly != 0 {
		sq := engine.FirstSquare(friendly)
		bishopAttacks := data.GetBishopAttacks(p.Board.Pieces, sq)
		rookAttacks := data.GetRookAttacks(p.Board.Pieces, sq)
		eval += e.QueenMobility[p.Board.CountBits(e.mobilityAreas[colour]&(bishopAttacks|rookAttacks))]
		friendly &= friendly - 1
	}

	return eval
}

func (e *EvaluationService) EvaluateMobilityRooks(p *engine.Position, colour int) Score {
	friendly := p.Board.GetPieces(colour, data.WR)
	var eval Score

	for friendly != 0 {
		sq := engine.FirstSquare(friendly)
		attack := data.GetRookAttacks(p.Board.Pieces, sq)
		eval += e.RookMobility[p.Board.CountBits(e.mobilityAreas[colour]&attack)]
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

func computeFactor(e *EvaluationService, p *engine.Position, eval Score, bothPawns uint64) int {
	var strongSide int
	var strong uint64
	if eval.End() > 0 {
		strongSide = data.White
		strong = p.Board.WhitePieces
	} else {
		strongSide = data.Black
		strong = p.Board.Pieces
	}

	var strongPawnCount = e.pieceCount[strongSide][data.WP]
	var x = 8 - strongPawnCount
	var pawnScale = 128 - x*x

	if strong&bothPawns&QueenSideBB == 0 ||
		strong&bothPawns&KingSideBB == 0 {
		pawnScale -= 20
	}

	bishops := p.Board.WhiteBishop | p.Board.BlackBishop
	kings := p.Board.WhiteKing | p.Board.BlackKing

	if e.pieceCount[data.White][data.WB] == 1 &&
		e.pieceCount[data.Black][data.WB] == 1 &&
		OnlyOne(bishops&darkSquares) {

		var whiteNonPawnCount = p.Board.CountBits(p.Board.WhitePieces &^ (bothPawns | kings))
		var blackNonPawnCount = p.Board.CountBits(p.Board.BlackPieces &^ (bothPawns | kings))
		if whiteNonPawnCount == blackNonPawnCount &&
			whiteNonPawnCount <= 2 &&
			blackNonPawnCount <= 2 {

			if whiteNonPawnCount == 1 {
				pawnScale = util.Min(pawnScale, 64)
			} else {
				pawnScale = util.Min(pawnScale, 96)
			}
		}
	}

	return pawnScale
}
