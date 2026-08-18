// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"context"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

// negamax performs negamax search with alpha-beta pruning and optimizations
//
//nolint:gocyclo // refactored in Phase 4
func (m *MinimaxEngine) negamax(ctx context.Context, depth int, alpha, beta eval.EvaluationScore, ply int) eval.EvaluationScore {
	b := m.searchState.board
	player := m.searchState.player
	pv := m.searchState.pv
	collectPV := m.searchState.collectPV

	m.searchState.searchStats.NodesSearched++

	// ply indexes this node by originalMaxDepth − depth: its distance from the root along the unreduced, unextended spine. Reductions inflate it and extensions hold it constant, so it is NOT true recursion depth. Callers thread it so that on
	// entry ply == originalMaxDepth - depth (the quantity this code previously
	// recomputed from the State field). Because a child is searched at
	// depth' = depth - K (K = 1 plus any reduction, minus any extension applied
	// here), its ply is ply + (depth - depth'); see the recursive calls below.

	if ply >= 0 && ply < len(m.searchState.searchStats.NodesByDepth) {
		m.searchState.searchStats.NodesByDepth[ply]++
	}

	// Cancellation cadence: poll ctx only once per ~1024 node entries instead
	// of on every node. NodesSearched was just incremented at node entry, so
	// the mask fires when the (post-increment) counter is a multiple of 1024
	// (1024, 2048, ...) — never at the very first node, and roughly every 1024
	// nodes thereafter. The counter is shared with quiescence, so the combined
	// negamax+quiescence node stream is what gets sampled. On cancellation we
	// set searchCancelled (whose existing propagation via the move-loop breaks
	// is untouched) and return alpha exactly as the former per-node check did.
	if m.searchState.searchStats.NodesSearched&1023 == 0 {
		select {
		case <-ctx.Done():
			m.searchState.searchCancelled = true
			return alpha
		default:
		}
	}

	inCheck := m.generator.IsKingInCheck(b, player)
	// Check extension. The former guard `depth < originalMaxDepth` is exactly
	// originalMaxDepth-depth > 0, and on entry ply == originalMaxDepth-depth, so
	// it re-expresses identically as ply > 0. `extended` (0 or 1) records whether
	// this node spent one ply of extension budget: after depth++ the invariant
	// becomes originalMaxDepth-depth == ply-extended, which the quiescence and
	// mate paths below consume directly.
	extended := 0
	if inCheck && ply > 0 {
		depth++
		extended = 1
	}

	hash := b.GetHash()

	if m.isDrawByRepetition(hash) {
		return eval.DrawScore
	}

	ttMove, ttScore, ttAlpha, ttBeta, ttDone := m.probeTT(hash, depth, ply, alpha, beta)
	if ttDone {
		return ttScore
	}
	alpha, beta = ttAlpha, ttBeta

	staticEval := m.evaluator.Evaluate(b)

	// Null move pruning: If our position is so good we can give opponent a free
	// move and still achieve a beta cutoff, we can prune this branch.
	if score, done := m.tryNullMove(ctx, b, player, depth, beta, ply, extended, staticEval, inCheck); done {
		return score
	}

	// Razoring: If static eval + margin is below alpha at low depths, verify
	// with quiescence search and potentially prune.
	if score, done := m.tryRazoring(ctx, player, depth, alpha, beta, ply, extended, staticEval, inCheck); done {
		return score
	}

	if depth <= 0 {
		m.searchState.player = player
		// ply-extended == originalMaxDepth-depth (extended depth). Reaching here
		// needs depth<=0, and since child entry depths are >= 0 this node cannot
		// have extended (extension needs ply>0 and leaves depth>=1), so extended==0.
		// (the passed value ply−extended is exact in all cases; extended==0 here only explains why no extension budget is lost at a leaf)
		return m.quiescence(ctx, alpha, beta, ply-extended)
	}

	pseudoMoves := m.generator.GeneratePseudoLegalMoves(b, player)
	defer movegen.ReleaseMoveList(pseudoMoves)

	if pseudoMoves.Count == 0 {
		return m.handleNoLegalMoves(b, player, depth, ply, ply-extended, hash)
	}

	m.scoreMoves(b, pseudoMoves, ply, ttMove)

	bestScore := -eval.MateScore - 1
	bestMove := board.Move{}
	legalMoveCount := 0

	// Track if we improved alpha to determine correct entry type
	alphaImproved := false

	for i := 0; i < pseudoMoves.Count; i++ {
		m.pickNextMove(pseudoMoves, i, ply)
		move := pseudoMoves.Moves[i]

		if m.searchState.searchCancelled {
			break
		}

		undo, ok := m.tryMove(move, player)
		if !ok {
			continue
		}

		m.addHistory(b.GetHash())

		legalMoveCount++
		var score eval.EvaluationScore

		if collectPV {
			pv.clearPly(ply + 1)
		}

		// Late Move Reductions: Search later moves at reduced depth since
		// move ordering should place best moves first
		reduction := m.calculateLMRReduction(b, depth, legalMoveCount, inCheck, move, ply)

		if reduction > 0 {
			m.searchState.searchStats.LMRReductions++

			m.searchState.player = oppositePlayer(player)
			m.searchState.collectPV = false
			// Reduced scout child ply = originalMaxDepth - (depth-1-reduction) =
			// ply+1-extended+reduction (searched shallower, so a larger ply index).
			score = -m.negamax(ctx, depth-1-reduction, -alpha-1, -alpha, ply+1-extended+reduction)

			if score > alpha {
				m.searchState.searchStats.LMRReSearches++

				m.searchState.player = oppositePlayer(player)
				m.searchState.collectPV = collectPV
				score = -m.negamax(ctx, depth-1, -beta, -alpha, ply+1-extended)
			}
		} else {
			m.searchState.player = oppositePlayer(player)
			m.searchState.collectPV = collectPV
			// Child ply = originalMaxDepth - (depth-1) = ply+1-extended.
			score = -m.negamax(ctx, depth-1, -beta, -alpha, ply+1-extended)
		}

		b.UnmakeMove(undo)

		m.removeHistory()

		if score > bestScore {
			bestScore = score
			bestMove = move
		}

		if score > alpha {
			alpha = score
			alphaImproved = true

			if collectPV {
				pv.store(ply, move)
			}

			if alpha >= beta {
				m.recordCutoff(move, depth, ply, legalMoveCount, hash, bestScore)
				return beta
			}
		}
	}

	if legalMoveCount == 0 {
		return m.handleNoLegalMoves(b, player, depth, ply, ply-extended, hash)
	}

	// Track node types for search statistics
	if alphaImproved {
		m.searchState.searchStats.PVNodes++
	} else {
		m.searchState.searchStats.AllNodes++
	}

	if m.transpositionTable != nil && !m.searchState.searchCancelled {
		// Determine correct entry type based on bounds
		var entryType EntryType
		if alphaImproved {
			// We found a move that improved alpha
			if bestScore >= beta {
				// This shouldn't happen as we return early on beta cutoff
				entryType = EntryLowerBound
			} else {
				// Score is between original alpha and beta
				entryType = EntryExact
			}
		} else {
			// No move improved alpha - this is an upper bound
			entryType = EntryUpperBound
		}

		m.transpositionTable.Store(hash, depth, scoreToTT(bestScore, ply), entryType, bestMove)
	}

	return bestScore
}

// probeTT probes the transposition table for the current position. It returns
// the (sanitized) TT move, and — when the entry is deep enough and its bound
// permits — a cutoff score with done=true. It also applies the alpha/beta
// narrowing that a lower/upper-bound entry induces, returned explicitly as
// alphaOut/betaOut (the former in-place mutation of alpha/beta). When the TT is
// absent, the position is not found, or the entry is too shallow, it returns
// the (possibly zero) TT move with done=false and the bounds unchanged. Stats
// counters (TTProbes/TTHits/TTCutoffs) increment in exactly the original order
// and under the original conditions. Each done=true return also classifies the
// node into PVNodes/CutNodes/AllNodes exactly as the normal move-loop
// classification at the bottom of negamax would (score resolved exactly ==
// PV, fail-high == Cut, fail-low == All) -- without this, any node resolved
// via a TT shortcut was invisible to that classification entirely.
func (m *MinimaxEngine) probeTT(hash uint64, depth, ply int, alpha, beta eval.EvaluationScore) (ttMove board.Move, score, alphaOut, betaOut eval.EvaluationScore, done bool) {
	alphaOut, betaOut = alpha, beta
	if m.transpositionTable != nil {
		m.searchState.searchStats.TTProbes++
		if entry, found := m.transpositionTable.Probe(hash); found {
			m.searchState.searchStats.TTHits++
			ttMove = entry.GetMove()

			if !ttMove.IsValid() {
				ttMove = board.Move{}
			}

			if entry.GetDepth() >= depth {
				ttScore := scoreFromTT(entry.Score, ply)
				switch entry.GetType() {
				case EntryExact:
					m.searchState.searchStats.TTCutoffs++
					m.searchState.searchStats.PVNodes++
					return ttMove, ttScore, alphaOut, betaOut, true
				case EntryLowerBound:
					if ttScore >= beta {
						m.searchState.searchStats.TTCutoffs++
						m.searchState.searchStats.CutNodes++
						return ttMove, ttScore, alphaOut, betaOut, true
					}
					if ttScore > alpha {
						alphaOut = ttScore
					}
				case EntryUpperBound:
					if ttScore <= alpha {
						m.searchState.searchStats.AllNodes++
						return ttMove, ttScore, alphaOut, betaOut, true
					}
					if ttScore < beta {
						betaOut = ttScore
					}
				}
			}
		}
	}
	return ttMove, 0, alphaOut, betaOut, false
}

// tryNullMove performs null move pruning. When the position is good enough to
// forfeit a move and still fail high, it returns (beta, true) to signal a
// cutoff; otherwise (0, false). Side effects (NullMoves/NullCutoffs stats, the
// null make/unmake, and the State player/collectPV scratch mutations) are the
// verbatim body from negamax. A cutoff also increments CutNodes: it is a
// fail-high, exactly the same node type the move-loop's own beta-cutoff path
// (recordCutoff) classifies as.
func (m *MinimaxEngine) tryNullMove(ctx context.Context, b *board.Board, player movegen.Player, depth int, beta eval.EvaluationScore, ply, extended int, staticEval eval.EvaluationScore, inCheck bool) (score eval.EvaluationScore, done bool) {
	if m.searchState.searchParams.NullMoveEnabled &&
		depth >= 3 &&
		staticEval >= beta &&
		beta < eval.MateScore-MateDistanceThreshold &&
		beta > -eval.MateScore+MateDistanceThreshold {
		if !inCheck {
			m.searchState.searchStats.NullMoves++

			nullReduction := m.searchState.searchParams.NullMoveReduction
			if depth >= 6 && nullReduction < 3 {
				nullReduction++
			}

			nullUndo := b.MakeNullMove()

			m.searchState.player = oppositePlayer(player)
			m.searchState.collectPV = false
			// Child ply = originalMaxDepth - (depth-1-nullReduction). With
			// originalMaxDepth = ply + depth - extended this is ply+1-extended+nullReduction.
			nullScore := -m.negamax(ctx, depth-1-nullReduction, -beta, -beta+1, ply+1-extended+nullReduction)

			b.UnmakeNullMove(nullUndo)

			if nullScore >= beta {
				if nullScore < eval.MateScore-MateDistanceThreshold {
					m.searchState.searchStats.NullCutoffs++
					m.searchState.searchStats.CutNodes++
					return beta, true
				}
			}
		}
	}
	return 0, false
}

// tryRazoring performs razoring. At low depths far from mate, when static eval
// plus a margin is still below alpha, it verifies with quiescence and returns
// (qScore, true) if the quiescence score confirms the fail-low; otherwise
// (0, false). Stats (RazoringAttempts/Cutoffs/Failed) and the State player
// mutation are the verbatim body from negamax. A confirmed cutoff also
// increments AllNodes: qScore <= alpha is a fail-low, the same node type the
// move-loop's own !alphaImproved path classifies as.
func (m *MinimaxEngine) tryRazoring(ctx context.Context, player movegen.Player, depth int, alpha, beta eval.EvaluationScore, ply, extended int, staticEval eval.EvaluationScore, inCheck bool) (score eval.EvaluationScore, done bool) {
	if m.searchState.searchParams.RazoringEnabled &&
		!inCheck &&
		depth <= m.searchState.searchParams.RazoringMaxDepth &&
		depth > 0 &&
		alpha < eval.MateScore-MateDistanceThreshold &&
		alpha > -eval.MateScore+MateDistanceThreshold {

		razoringMargin := m.searchState.searchParams.RazoringMargins[depth]

		if staticEval+razoringMargin < alpha {
			// Don't check for TT move at all
			m.searchState.searchStats.RazoringAttempts++

			m.searchState.player = player
			// Quiescence's ply arg is originalMaxDepth-depth (extended depth) ==
			// ply-extended. Razoring requires !inCheck so extended is 0 here.
			qScore := m.quiescence(ctx, alpha, beta, ply-extended)

			if qScore <= alpha {
				m.searchState.searchStats.RazoringCutoffs++
				m.searchState.searchStats.AllNodes++
				return qScore, true
			}
			m.searchState.searchStats.RazoringFailed++
		}
	}
	return 0, false
}

// recordCutoff performs the bookkeeping for a beta cutoff: move-ordering stats,
// killer/history updates for quiet moves, and the lower-bound TT store. The two
// former `!move.IsCapture` blocks (killer, then history) are merged into a
// single guard — same effects, same order. The caller returns beta after this.
func (m *MinimaxEngine) recordCutoff(move board.Move, depth, ply, legalMoveCount int, hash uint64, bestScore eval.EvaluationScore) {
	// Track move ordering statistics
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
		if m.historyTable != nil {
			m.historyTable.UpdateHistory(move, depth)
		}
	}

	if m.transpositionTable != nil && !m.searchState.searchCancelled {
		m.transpositionTable.Store(hash, depth, scoreToTT(bestScore, ply), EntryLowerBound, move)
	}
}

// handleNoLegalMoves returns the appropriate score when no legal moves are available
// Returns checkmate score if in check, stalemate score otherwise.
//
// ply is this node's RAW entry ply (originalMaxDepth - entry depth) — the exact
// value probeTT/scoreFromTT will use to decode this entry on any later probe.
// pliesFromRoot is ply-extended, the mate-distance convention the returned score
// uses (unchanged: it must stay consistent with sibling mate-score ordering).
// The TT store MUST use scoreToTT(score, ply) (raw ply), not pliesFromRoot:
// scoreFromTT decodes with the raw ply, so encoding with pliesFromRoot would
// skew every re-probe of a checkmate node by the check-extension count (which
// nearly always fires at a mate node, since it is in check). See scoreToTT.
func (m *MinimaxEngine) handleNoLegalMoves(b *board.Board, player movegen.Player, depth, ply, pliesFromRoot int, hash uint64) eval.EvaluationScore {
	if m.generator.IsKingInCheck(b, player) {
		// Checkmate - the mate distance is how many plies from the root we are.
		// The returned score uses pliesFromRoot (== ply-extended at the call
		// site) to stay consistent with how mate scores propagate/compare.
		// Negative score because it's mate AGAINST the current player.
		score := -eval.MateScore + eval.EvaluationScore(pliesFromRoot)
		if m.transpositionTable != nil {
			// Encode with the RAW ply so the round-trip through
			// scoreFromTT(entry.Score, ply) on a later probe returns exactly
			// this same root-relative score.
			m.transpositionTable.Store(hash, depth, scoreToTT(score, ply), EntryExact, board.Move{})
		}
		return score
	}
	// Stalemate
	if m.transpositionTable != nil {
		m.transpositionTable.Store(hash, depth, eval.DrawScore, EntryExact, board.Move{})
	}
	return eval.DrawScore
}

// calculateLMRReduction calculates the Late Move Reduction amount for a move.
// Returns the number of plies to reduce search depth by, based on depth, move count,
// and history heuristic. Returns 0 if no reduction should be applied.
func (m *MinimaxEngine) calculateLMRReduction(b *board.Board, depth, legalMoveCount int, inCheck bool, move board.Move, ply int) int {
	if !m.searchState.searchParams.LMREnabled ||
		depth < m.searchState.searchParams.LMRMinDepth ||
		legalMoveCount <= m.searchState.searchParams.LMRMinMoves ||
		inCheck ||
		move.IsCapture ||
		move.Promotion != board.Empty ||
		m.isKillerMove(move, ply) {
		return 0
	}

	// Don't reduce moves that give check
	if board.MoveGivesCheck(b, move) {
		return 0
	}

	// Use pre-calculated LMR table for base reduction (performance optimization)
	tableDepth := min(depth, 15)
	tableMoveCount := min(legalMoveCount, 63)
	if tableDepth < 1 || tableMoveCount < 1 {
		return 0
	}
	reduction := LMRTable[tableDepth][tableMoveCount]

	// Adjust reduction based on history heuristic
	historyScore := m.getHistoryScore(move)
	if historyScore > m.searchState.searchParams.HistoryHighThreshold {
		// Very good history - don't reduce
		return 0
	} else if historyScore > m.searchState.searchParams.HistoryMedThreshold && reduction > 0 {
		// Good history - reduce less
		reduction = reduction * 2 / 3
	} else if historyScore < m.searchState.searchParams.HistoryLowThreshold {
		// Bad history - reduce more
		reduction = reduction * 4 / 3
	}

	reduction = max(0, min(reduction, depth-1))

	return reduction
}
