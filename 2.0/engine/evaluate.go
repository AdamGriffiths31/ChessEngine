package engine

import (
	"math/bits"

	"github.com/AdamGriffiths31/ChessEngine/data"
)

var pawnTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	10, 10, 0, -10, -10, 0, 10, 10,
	5, 0, 0, 5, 5, 0, 0, 5,
	0, 0, 10, 20, 20, 10, 0, 0,
	5, 5, 5, 10, 10, 5, 5, 5,
	10, 10, 10, 20, 20, 10, 10, 10,
	20, 20, 20, 30, 30, 20, 20, 20,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightTable = [64]int{
	0, -10, 0, 0, 0, 0, -10, 0,
	0, 0, 0, 5, 5, 0, 0, 0,
	0, 0, 10, 10, 10, 10, 0, 0,
	0, 0, 10, 20, 20, 10, 5, 0,
	5, 10, 15, 20, 20, 15, 10, 5,
	5, 10, 10, 20, 20, 10, 10, 5,
	0, 0, 5, 10, 10, 5, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var bishopTable = [64]int{
	0, 0, -10, 0, 0, -10, 0, 0,
	0, 0, 0, 10, 10, 0, 0, 0,
	0, 0, 10, 15, 15, 10, 0, 0,
	0, 10, 15, 20, 20, 15, 10, 0,
	0, 10, 15, 20, 20, 15, 10, 0,
	0, 0, 10, 15, 15, 10, 0, 0,
	0, 0, 0, 10, 10, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var rookTable = [64]int{
	0, 0, 5, 10, 10, 5, 0, 0,
	0, 0, 5, 10, 10, 5, 0, 0,
	0, 0, 5, 10, 10, 5, 0, 0,
	0, 0, 5, 10, 10, 5, 0, 0,
	0, 0, 5, 10, 10, 5, 0, 0,
	0, 0, 5, 10, 10, 5, 0, 0,
	25, 25, 25, 25, 25, 25, 25, 25,
	0, 0, 5, 10, 10, 5, 0, 0,
}

var kingE = [64]int{
	-50, -10, 0, 0, 0, 0, -10, -50,
	-10, 0, 10, 10, 10, 10, 0, -10,
	0, 10, 20, 20, 20, 20, 10, 0,
	0, 10, 20, 40, 40, 20, 10, 0,
	0, 10, 20, 40, 40, 20, 10, 0,
	0, 10, 20, 20, 20, 20, 10, 0,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-50, -10, 0, 0, 0, 0, -10, -50,
}

var kingO = [64]int{
	0, 5, 5, -10, -10, 0, 10, 5,
	-30, -30, -30, -30, -30, -30, -30, -30,
	-50, -50, -50, -50, -50, -50, -50, -50,
	-70, -70, -70, -70, -70, -70, -70, -70,
	-70, -70, -70, -70, -70, -70, -70, -70,
	-70, -70, -70, -70, -70, -70, -70, -70,
	-70, -70, -70, -70, -70, -70, -70, -70,
	-70, -70, -70, -70, -70, -70, -70, -70,
}

var pawnIsolated = -10
var pawnPassed = [8]int{0, 5, 10, 20, 35, 60, 100, 200}
var rookOpenFile = 10
var rookSemiOpenFile = 5
var queenOpenFile = 5
var queenSemiOpenFile = 3
var bishopPair = 30

// var score int

func (p *Position) Evaluate() int {

	if p.isMaterialDraw() {
		return 0
	}
	p.CurrentScore = p.calculateWhiteMaterial() - p.calculateBlackMaterial()

	bothPawns := p.Board.WhitePawn | p.Board.BlackPawn

	p.calculateEvalPawns()
	p.calculateEvalKnights()
	p.calculateEvalBishop()
	p.calculateEvalRook(bothPawns)
	p.calculateEvalQueens(bothPawns)
	p.calculateEvalKings()

	if p.Side == data.White {
		return p.CurrentScore
	} else {
		return -p.CurrentScore
	}
}

func (p *Position) calculateWhiteMaterial() int {
	score := 0
	score += p.Board.CountBits(p.Board.WhitePawn) * data.PieceVal[WP]
	score += p.Board.CountBits(p.Board.WhiteKnight) * data.PieceVal[WN]
	score += p.Board.CountBits(p.Board.WhiteBishop) * data.PieceVal[WB]
	score += p.Board.CountBits(p.Board.WhiteRook) * data.PieceVal[WR]
	score += p.Board.CountBits(p.Board.WhiteQueen) * data.PieceVal[WQ]
	score += 1 * data.PieceVal[WK]
	return score
}

func (p *Position) calculateBlackMaterial() int {
	score := 0
	score += p.Board.CountBits(p.Board.BlackPawn) * data.PieceVal[WP]
	score += p.Board.CountBits(p.Board.BlackKnight) * data.PieceVal[WN]
	score += p.Board.CountBits(p.Board.BlackBishop) * data.PieceVal[WB]
	score += p.Board.CountBits(p.Board.BlackRook) * data.PieceVal[WR]
	score += p.Board.CountBits(p.Board.BlackQueen) * data.PieceVal[WQ]
	score += 1 * data.PieceVal[WK]
	return score
}

func isEndGame() int {
	return (1 * data.PieceVal[data.WR]) + (2 * data.PieceVal[data.WN]) + (2 * data.PieceVal[data.WP]) + data.PieceVal[data.WK]
}

func (p *Position) calculateEvalPawns() {
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
func (p *Position) calculateEvalKnights() {
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

func (p *Position) calculateEvalBishop() {
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
func (p *Position) calculateEvalRook(bothPawns uint64) {
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

func (p *Position) calculateEvalQueens(bothPawns uint64) {
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

func (p *Position) calculateEvalKings() {
	bb := p.Board.WhiteKing
	for bb != 0 {
		sq := bits.TrailingZeros64(bb)
		bb ^= (1 << sq)
		if p.calculateBlackMaterial() <= isEndGame() {
			p.CurrentScore += kingE[sq]
		} else {
			p.CurrentScore += kingO[sq]
		}
	}
	bb = p.Board.BlackKing
	for bb != 0 {
		sq := bits.TrailingZeros64(bb)
		bb ^= (1 << sq)
		if p.calculateWhiteMaterial() <= isEndGame() {
			p.CurrentScore -= kingE[data.Mirror64[sq]]
		} else {
			p.CurrentScore -= kingO[data.Mirror64[sq]]
		}
	}
}

func (p *Position) isMaterialDraw() bool {
	if p.Board.WhitePawn != 0 || p.Board.BlackPawn != 0 ||
		p.Board.WhiteQueen != 0 || p.Board.BlackQueen != 0 ||
		p.Board.WhiteRook != 0 || p.Board.BlackRook != 0 {
		return false
	}
	wKnightsNum := p.Board.CountBits(p.Board.WhiteKnight)
	bKnightsNum := p.Board.CountBits(p.Board.BlackKnight)
	wBishopsNum := p.Board.CountBits(p.Board.WhiteBishop)
	bBishopsNum := p.Board.CountBits(p.Board.BlackBishop)
	all := wKnightsNum + bKnightsNum + wBishopsNum + bBishopsNum

	if all <= 1 {
		if wKnightsNum != 0 {
			return wKnightsNum == 1
		} else if bKnightsNum != 0 {
			return bKnightsNum == 1
		} else if wBishopsNum != 0 {
			return wBishopsNum == 1
		} else if bBishopsNum != 0 {
			return wBishopsNum == 1
		}
	}

	return false
}
