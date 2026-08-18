// Package search provides chess move search algorithms and transposition table implementation.
package search

import "github.com/AdamGriffiths31/ChessEngine/internal/board"

// pvTable holds triangular principal-variation storage: one fixed-length row
// per ply, each row large enough to hold a full line from that ply. It
// replaces the hand-rolled [][]board.Move that negamax and the iterative
// deepening loop previously shared through State.pv. The store/clear/read
// semantics are byte-for-byte the ones the inline loops implemented, including
// the truncation-at-row-capacity behavior and the trailing zero-move
// terminator; nothing here "improves" edge behavior.
type pvTable struct {
	lines [][]board.Move
}

// newPVTable allocates a maxPly-by-maxPly triangular table, matching the old
// make([][]board.Move, maxPly) + per-row make([]board.Move, maxPly) allocation
// in runIterativeDeepening.
func newPVTable(maxPly int) *pvTable {
	lines := make([][]board.Move, maxPly)
	for i := range lines {
		lines[i] = make([]board.Move, maxPly)
	}
	return &pvTable{lines: lines}
}

// store records move as the first move of the line at ply, then copies the
// child line (ply+1) in behind it. This is exactly the inline triangular copy
// negamax and the root loop performed: the child line is copied up to its
// first zero move, the destination row is truncated at its capacity, and a
// trailing zero move terminates the line when room remains. Out-of-range ply
// is a no-op, matching negamax's former `ply < len(pv)` guard.
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

// clearPly resets the first slot of the line at ply, terminating it. This is
// the `pv[ply][0] = board.Move{}` reset negamax applied to the child line
// before searching. Out-of-range ply is a no-op, matching the former
// `ply+1 < len(pv)` guard at the call site.
func (p *pvTable) clearPly(ply int) {
	if ply < 0 || ply >= len(p.lines) {
		return
	}
	p.lines[ply][0] = board.Move{}
}

// reset terminates every line, matching the per-depth
// `for i := range pv { pv[i][0] = board.Move{} }` loop in runIterativeDeepening.
func (p *pvTable) reset() {
	for i := range p.lines {
		p.lines[i][0] = board.Move{}
	}
}

// lineAt returns a fresh slice of the moves in the line at ply, up to (but not
// including) the first zero move — the same walk the PrincipalVariation
// builder performed over pv[0]. Out-of-range ply returns nil.
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
