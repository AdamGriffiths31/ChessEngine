package search

import "github.com/AdamGriffiths31/ChessEngine/internal/board"

type pvTable struct {
	lines [][]board.Move
}

func newPVTable(maxPly int) *pvTable {
	lines := make([][]board.Move, maxPly)
	for i := range lines {
		lines[i] = make([]board.Move, maxPly)
	}
	return &pvTable{lines: lines}
}

func (p *pvTable) store(ply int, move board.Move) {
	if ply < 0 || ply >= len(p.lines) {
		return
	}
	line := p.lines[ply]
	line[0] = move
	pvLen := 1
	if ply+1 < len(p.lines) {
		child := p.lines[ply+1]
		for i := 0; i < len(child) && child[i] != (board.Move{}); i++ {
			line[pvLen] = child[i]
			pvLen++
			if pvLen >= len(line) {
				break
			}
		}
	}
	if pvLen < len(line) {
		line[pvLen] = board.Move{}
	}
}

func (p *pvTable) clearPly(ply int) {
	if ply < 0 || ply >= len(p.lines) {
		return
	}
	p.lines[ply][0] = board.Move{}
}

func (p *pvTable) reset() {
	for i := range p.lines {
		p.lines[i][0] = board.Move{}
	}
}

func (p *pvTable) lineAt(ply int) []board.Move {
	if ply < 0 || ply >= len(p.lines) {
		return nil
	}
	line := p.lines[ply]
	out := make([]board.Move, 0, len(line))
	for i := 0; i < len(line) && line[i] != (board.Move{}); i++ {
		out = append(out, line[i])
	}
	return out
}
