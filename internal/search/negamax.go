package search

import (
	"context"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

type nodeInfo struct {
	player      movegen.Player
	depth       int
	ply         int
	beta        eval.EvaluationScore
	collectPV   bool
	inCheck     bool
	isPV        bool
	sparseBoard bool
	stmEval     eval.EvaluationScore
	lmpLimit    int
}

func sideRelative(score eval.EvaluationScore, player movegen.Player) eval.EvaluationScore {
	if player == movegen.Black {
		return -score
	}
	return score
}

func (m *MinimaxEngine) pollCancel(ctx context.Context) bool {
	if m.searchState.searchStats.NodesSearched&1023 != 0 {
		return false
	}
	select {
	case <-ctx.Done():
		m.searchState.searchCancelled = true
		return true
	default:
		return false
	}
}

func (m *MinimaxEngine) negamax(ctx context.Context, player movegen.Player, depth int, alpha, beta eval.EvaluationScore, ply int, collectPV bool) eval.EvaluationScore {
	b := m.searchState.board

	m.searchState.searchStats.NodesSearched++
	if ply < len(m.searchState.searchStats.NodesByDepth) {
		m.searchState.searchStats.NodesByDepth[ply]++
	}

	if m.pollCancel(ctx) {
		return alpha
	}

	inCheck := m.generator.IsKingInCheck(b, player)
	if inCheck {
		depth++
	}

	hash := b.GetHash()
	if m.isDrawNode(b, hash, inCheck) {
		return eval.DrawScore
	}

	alpha = max(alpha, -eval.MateScore+eval.EvaluationScore(ply))
	beta = min(beta, eval.MateScore-eval.EvaluationScore(ply+1))
	if alpha >= beta {
		return alpha
	}

	ttMove, ttScore, ttAlpha, ttBeta, ttDone := m.probeTT(hash, depth, ply, alpha, beta)
	if ttDone {
		return ttScore
	}
	alpha, beta = ttAlpha, ttBeta

	if depth <= 0 {
		if ply == 1 {
			return m.quiescence(ctx, player, alpha, beta, ply)
		}
		if m.enterQuiescenceNode(ctx) {
			return alpha
		}
		return m.quiescenceBody(ctx, player, alpha, beta, ply, hash, inCheck)
	}

	var stmEval eval.EvaluationScore
	if !inCheck {
		stmEval = sideRelative(m.evaluator.Evaluate(b), player)
	}
	improving := m.updateImproving(ply, inCheck, stmEval)

	if score, done := m.tryReverseFutility(b, player, depth, alpha, beta, stmEval, inCheck); done {
		return score
	}

	if score, done := m.tryNullMove(ctx, b, player, depth, beta, ply, stmEval, inCheck); done {
		return score
	}

	if score, done := m.tryRazoring(ctx, player, depth, alpha, beta, ply, stmEval, inCheck); done {
		return score
	}

	if ply >= MaxKillerDepth-1 {
		return m.quiescence(ctx, player, alpha, beta, ply)
	}

	pseudoMoves := m.generator.GeneratePseudoLegalMoves(b, player)
	defer movegen.ReleaseMoveList(pseudoMoves)

	m.scoreMoves(b, pseudoMoves, ply, ttMove)

	lmpLimit := 3 + depth*depth
	if improving {
		lmpLimit *= 2
	}

	n := nodeInfo{
		player:      player,
		depth:       depth,
		ply:         ply,
		beta:        beta,
		collectPV:   collectPV,
		inCheck:     inCheck,
		isPV:        beta-alpha > 1,
		sparseBoard: b.AllPieces.PopCount() <= NullMoveMaxPieces,
		stmEval:     stmEval,
		lmpLimit:    lmpLimit,
	}
	return m.searchMoves(ctx, &n, hash, pseudoMoves, alpha)
}

func (m *MinimaxEngine) isDrawNode(b *board.Board, hash uint64, inCheck bool) bool {
	if m.isDrawByRepetition(hash, b.GetHalfMoveClock()) || isInsufficientMaterial(b) {
		return true
	}
	return !inCheck && b.GetHalfMoveClock() >= FiftyMoveHalfMoves
}

func (m *MinimaxEngine) updateImproving(ply int, inCheck bool, stmEval eval.EvaluationScore) bool {
	stack := &m.searchState.staticEvalStack
	if ply == 1 {
		stack[0] = noStaticEval
	}
	if ply >= MaxKillerDepth {
		return false
	}
	if inCheck {
		stack[ply] = noStaticEval
		return false
	}
	stack[ply] = stmEval
	if ply < 2 {
		return false
	}
	prev := stack[ply-2]
	return prev != noStaticEval && stmEval > prev
}

func (m *MinimaxEngine) searchMoves(ctx context.Context, n *nodeInfo, hash uint64, moves *movegen.MoveList, alpha eval.EvaluationScore) eval.EvaluationScore {
	b := m.searchState.board
	pv := m.searchState.pv

	bestScore := -scoreInfinity
	bestMove := board.Move{}
	legalMoveCount := 0

	quietsSearched := 0

	quietMovesPruned := false

	alphaImproved := false

	evasions := allSquares
	if n.inCheck {
		evasions = m.generator.CheckEvasionMask(b, n.player)
	}

	for i := range moves.Count {
		m.pickNextMove(moves, i, n.ply)
		move := moves.Moves[i]

		if m.searchState.searchCancelled {
			break
		}

		if cannotEvade(&move, evasions) {
			continue
		}

		givesCheck := move.IsQuiet() && !n.inCheck && board.MoveGivesCheck(b, move)

		if legalMoveCount > 0 && m.shouldPruneQuiet(n, move, givesCheck, alpha, quietsSearched) {
			continue
		}

		undo, ok := m.tryMove(move, n.player)
		if !ok {
			continue
		}

		if legalMoveCount == 0 && m.shouldPruneQuiet(n, move, givesCheck, alpha, quietsSearched) {
			b.UnmakeMove(undo)
			quietMovesPruned = true
			continue
		}

		m.addHistory(b.GetHash())

		legalMoveCount++
		if n.collectPV {
			pv.clearPly(n.ply + 1)
		}

		score := m.searchChild(ctx, n, move, givesCheck, legalMoveCount, alpha)

		b.UnmakeMove(undo)

		m.removeHistory()

		if move.IsQuiet() {
			quietsSearched++
		}

		if score > bestScore {
			bestScore = score
			bestMove = move
		}

		if score > alpha {
			alpha = score
			alphaImproved = true

			if n.collectPV {
				pv.store(n.ply, move)
			}

			if alpha >= n.beta {
				m.recordCutoff(move, n.depth, n.ply, legalMoveCount, hash, bestScore, moves, i)
				return n.beta
			}
		}
	}

	if legalMoveCount == 0 {
		if !quietMovesPruned {
			return m.handleNoLegalMoves(n.inCheck, n.depth, n.ply, hash)
		}
		return alpha
	}

	m.storeNodeResult(hash, n.depth, n.ply, bestScore, bestMove, alphaImproved)
	return bestScore
}

func cannotEvade(move *board.Move, evasions board.Bitboard) bool {
	if evasions == allSquares || move.IsEnPassant || move.Piece == board.WhiteKing || move.Piece == board.BlackKing {
		return false
	}
	return !evasions.HasBit(board.FileRankToSquare(move.To.File, move.To.Rank))
}

func (m *MinimaxEngine) shouldPruneQuiet(n *nodeInfo, move board.Move, givesCheck bool, alpha eval.EvaluationScore, quietsSearched int) bool {
	if n.sparseBoard || n.inCheck || !move.IsQuiet() ||
		givesCheck || m.isKillerMove(move, n.ply) ||
		isMateScore(alpha) {
		return false
	}

	p := &m.searchState.searchParams
	switch {
	case n.depth <= p.FutilityMaxDepth && n.stmEval+p.FutilityMargins[n.depth] <= alpha:
		m.searchState.searchStats.FutilityPrunes++
		return true
	case !n.isPV && n.depth <= p.LMPMaxDepth && quietsSearched >= n.lmpLimit:
		m.searchState.searchStats.LMPPrunes++
		return true
	}
	return false
}

func (m *MinimaxEngine) searchChild(ctx context.Context, n *nodeInfo, move board.Move, givesCheck bool, legalMoveCount int, alpha eval.EvaluationScore) eval.EvaluationScore {
	opponent := oppositePlayer(n.player)

	reduction := m.calculateLMRReduction(n.depth, legalMoveCount, n.inCheck, givesCheck, move, n.ply)
	if reduction > 0 {
		m.searchState.searchStats.LMRReductions++

		score := -m.negamax(ctx, opponent, n.depth-1-reduction, -alpha-1, -alpha, n.ply+1, false)
		if score <= alpha {
			return score
		}
		m.searchState.searchStats.LMRReSearches++
	}

	return -m.negamax(ctx, opponent, n.depth-1, -n.beta, -alpha, n.ply+1, n.collectPV)
}

func (m *MinimaxEngine) storeNodeResult(hash uint64, depth, ply int, bestScore eval.EvaluationScore, bestMove board.Move, alphaImproved bool) {
	entryType := EntryUpperBound
	if alphaImproved {
		entryType = EntryExact
		m.searchState.searchStats.PVNodes++
	} else {
		m.searchState.searchStats.AllNodes++
	}

	if m.transpositionTable != nil && !m.searchState.searchCancelled {
		m.transpositionTable.Store(hash, depth, scoreToTT(bestScore, ply), entryType, bestMove)
	}
}

func (m *MinimaxEngine) probeTT(hash uint64, depth, ply int, alpha, beta eval.EvaluationScore) (ttMove board.Move, score, alphaOut, betaOut eval.EvaluationScore, done bool) {
	if m.transpositionTable == nil {
		return ttMove, 0, alpha, beta, false
	}

	m.searchState.searchStats.TTProbes++
	entry, found := m.transpositionTable.Probe(hash)
	if !found {
		return ttMove, 0, alpha, beta, false
	}

	m.searchState.searchStats.TTHits++
	ttMove = entry.GetMove()
	if !ttMove.IsValid() {
		ttMove = board.Move{}
	}

	if entry.GetDepth() < depth {
		return ttMove, 0, alpha, beta, false
	}

	rootChild := ply == 1
	ttScore := scoreFromTT(entry.Score, ply)
	switch entry.GetType() {
	case EntryExact:
		if !rootChild {
			m.searchState.searchStats.TTCutoffs++
			m.searchState.searchStats.PVNodes++
			return ttMove, ttScore, alpha, beta, true
		}
	case EntryLowerBound:
		switch {
		case ttScore < beta:
			alpha = max(alpha, ttScore)
		case !rootChild:
			m.searchState.searchStats.TTCutoffs++
			m.searchState.searchStats.CutNodes++
			return ttMove, ttScore, alpha, beta, true
		}
	case EntryUpperBound:
		switch {
		case ttScore > alpha:
			beta = min(beta, ttScore)
		case !rootChild:
			m.searchState.searchStats.AllNodes++
			return ttMove, ttScore, alpha, beta, true
		}
	}
	return ttMove, 0, alpha, beta, false
}

func (m *MinimaxEngine) tryReverseFutility(b *board.Board, player movegen.Player, depth int, alpha, beta, stmEval eval.EvaluationScore, inCheck bool) (score eval.EvaluationScore, done bool) {
	p := &m.searchState.searchParams
	if !p.RFPEnabled || inCheck || depth <= 0 || depth > p.RFPMaxDepth ||
		beta-alpha != 1 ||
		isMateScore(beta) ||
		!hasNonPawnMaterial(b, player) {
		return 0, false
	}
	margin := eval.EvaluationScore(depth) * p.RFPMargin
	if stmEval-margin >= beta {
		m.searchState.searchStats.RFPCutoffs++
		m.searchState.searchStats.CutNodes++
		return stmEval - margin, true
	}
	return 0, false
}

func (m *MinimaxEngine) tryNullMove(ctx context.Context, b *board.Board, player movegen.Player, depth int, beta eval.EvaluationScore, ply int, stmEval eval.EvaluationScore, inCheck bool) (score eval.EvaluationScore, done bool) {
	if inCheck || depth < NullMoveMinDepth || stmEval < beta || isMateScore(beta) {
		return 0, false
	}

	if !hasNonPawnMaterial(b, player) {
		m.searchState.searchStats.NullMoveZugzwangSkipped++
		return 0, false
	}
	if b.AllPieces.PopCount() <= NullMoveMaxPieces {
		m.searchState.searchStats.NullMoveSparseSkipped++
		return 0, false
	}

	nullReduction := m.searchState.searchParams.NullMoveReduction
	if depth >= NullMoveDeepDepth {
		nullReduction++
	}

	m.searchState.searchStats.NullMoves++

	nullUndo := b.MakeNullMove()

	nullScore := -m.negamax(ctx, oppositePlayer(player), depth-1-nullReduction, -beta, -beta+1, ply+1, false)

	b.UnmakeNullMove(nullUndo)

	if nullScore >= beta && nullScore < eval.MateScore-MateDistanceThreshold {
		m.searchState.searchStats.NullCutoffs++
		m.searchState.searchStats.CutNodes++
		return beta, true
	}
	return 0, false
}

func isInsufficientMaterial(b *board.Board) bool {
	if b.AllPieces.PopCount() > 3 {
		return false
	}
	return b.GetPieceBitboard(board.WhitePawn)|b.GetPieceBitboard(board.BlackPawn)|
		b.GetPieceBitboard(board.WhiteRook)|b.GetPieceBitboard(board.BlackRook)|
		b.GetPieceBitboard(board.WhiteQueen)|b.GetPieceBitboard(board.BlackQueen) == 0
}

func hasNonPawnMaterial(b *board.Board, player movegen.Player) bool {
	if player == movegen.White {
		return b.GetPieceBitboard(board.WhiteKnight) != 0 ||
			b.GetPieceBitboard(board.WhiteBishop) != 0 ||
			b.GetPieceBitboard(board.WhiteRook) != 0 ||
			b.GetPieceBitboard(board.WhiteQueen) != 0
	}
	return b.GetPieceBitboard(board.BlackKnight) != 0 ||
		b.GetPieceBitboard(board.BlackBishop) != 0 ||
		b.GetPieceBitboard(board.BlackRook) != 0 ||
		b.GetPieceBitboard(board.BlackQueen) != 0
}

func (m *MinimaxEngine) tryRazoring(ctx context.Context, player movegen.Player, depth int, alpha, beta eval.EvaluationScore, ply int, stmEval eval.EvaluationScore, inCheck bool) (score eval.EvaluationScore, done bool) {
	p := &m.searchState.searchParams
	if !p.RazoringEnabled || inCheck || depth > p.RazoringMaxDepth || depth <= 0 ||
		isMateScore(alpha) || stmEval+p.RazoringMargins[depth] >= alpha {
		return 0, false
	}

	m.searchState.searchStats.RazoringAttempts++

	qScore := m.quiescence(ctx, player, alpha, beta, ply)

	if qScore <= alpha {
		m.searchState.searchStats.RazoringCutoffs++
		m.searchState.searchStats.AllNodes++
		return qScore, true
	}
	m.searchState.searchStats.RazoringFailed++
	return 0, false
}

func (m *MinimaxEngine) recordCutoff(move board.Move, depth, ply, legalMoveCount int, hash uint64, bestScore eval.EvaluationScore, moveList *movegen.MoveList, triedIndex int) {
	m.searchState.searchStats.TotalCutoffs++
	if legalMoveCount == 1 {
		m.searchState.searchStats.FirstMoveCutoffs++
	}
	if legalMoveCount-1 < len(m.searchState.searchStats.CutoffsByMoveIndex) {
		m.searchState.searchStats.CutoffsByMoveIndex[legalMoveCount-1]++
	}
	m.searchState.searchStats.CutNodes++

	if !move.IsCapture {
		m.storeKiller(move, ply)
	}

	if !move.IsCapture {
		m.historyTable.Bonus(move, depth)
	}
	for j := range triedIndex {
		earlier := &moveList.Moves[j]
		if !earlier.IsQuiet() || sameMove(earlier, &move) {
			continue
		}
		m.historyTable.Malus(*earlier, depth)
	}

	if m.transpositionTable != nil && !m.searchState.searchCancelled {
		m.transpositionTable.Store(hash, depth, scoreToTT(bestScore, ply), EntryLowerBound, move)
	}
}

func (m *MinimaxEngine) handleNoLegalMoves(inCheck bool, depth, ply int, hash uint64) eval.EvaluationScore {
	score := eval.DrawScore
	if inCheck {
		score = -eval.MateScore + eval.EvaluationScore(ply)
	}
	if m.transpositionTable != nil {
		m.transpositionTable.Store(hash, depth, scoreToTT(score, ply), EntryExact, board.Move{})
	}
	return score
}

func (m *MinimaxEngine) calculateLMRReduction(depth, legalMoveCount int, inCheck, givesCheck bool, move board.Move, ply int) int {
	if depth < m.searchState.searchParams.LMRMinDepth ||
		legalMoveCount <= m.searchState.searchParams.LMRMinMoves ||
		inCheck ||
		!move.IsQuiet() ||
		m.isKillerMove(move, ply) {
		return 0
	}

	reduction := LMRTable[min(depth, lmrTableDepths-1)][min(legalMoveCount, lmrTableMoves-1)]

	historyScore := m.getHistoryScore(move)
	switch {
	case historyScore > m.searchState.searchParams.HistoryHighThreshold:
		return 0
	case historyScore > m.searchState.searchParams.HistoryMedThreshold:
		reduction = reduction * 2 / 3
	case historyScore < m.searchState.searchParams.HistoryLowThreshold:
		reduction = reduction * 4 / 3
	}

	reduction = min(reduction, depth-1)

	if givesCheck {
		reduction = min(reduction, 1)
	}

	return reduction
}
