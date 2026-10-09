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
func (m *MinimaxEngine) negamax(ctx context.Context, depth int, alpha, beta eval.EvaluationScore, ply, pvPly int) eval.EvaluationScore {
	b := m.searchState.board
	player := m.searchState.player
	pv := m.searchState.pv
	collectPV := m.searchState.collectPV

	m.searchState.searchStats.NodesSearched++

	// ply is this node's distance from the root along the unreduced spine
	// (originalMaxDepth - depth): held constant by check extensions, inflated
	// by LMR reductions. Child calls pass ply + (depth - childDepth).
	// pvPly instead counts true recursion depth (+1 per level), and the pv
	// table is indexed ONLY by pvPly: indexing by ply aliased a
	// check-extended child onto its parent's row, dropping moves from the PV.

	if ply >= 0 && ply < len(m.searchState.searchStats.NodesByDepth) {
		m.searchState.searchStats.NodesByDepth[ply]++
	}

	// Poll ctx every ~1024 nodes rather than per node - per-node selects are
	// measurable overhead. NodesSearched is shared with quiescence, so the
	// mask samples the combined stream.
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
	if score, done := m.tryNullMove(ctx, b, player, depth, beta, ply, extended, pvPly, staticEval, inCheck); done {
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

	// Static eval from the side to move's perspective (Evaluate returns a
	// White-relative score), for futility pruning comparisons against alpha.
	stmStaticEval := staticEval
	if player == movegen.Black {
		stmStaticEval = -staticEval
	}

	// Quiet moves searched so far at this node, for late move pruning.
	quietsSearched := 0

	// Set when quiet-move pruning skips at least one legal move. If every
	// move ends up pruned, reaching the post-loop check with
	// legalMoveCount == 0 does NOT mean stalemate/checkmate - zero searched
	// moves is not zero legal moves - so the terminal-score path must not
	// fire (a false proven draw stored as EntryExact here poisons the TT).
	quietMovesPruned := false

	// Sparse-board guard for quiet-move pruning (same threshold and
	// rationale as null move pruning's): on near-bare boards every quiet
	// move matters equally - zugzwang and dead-drawn positions live here,
	// and skipping "hopeless" quiets lets horizon eval noise stand in for a
	// provable draw. Computed once; make/unmake during the loop restores
	// this exact occupancy between moves.
	sparseBoard := b.AllPieces.PopCount() <= NullMoveMaxPieces

	// Track if we improved alpha to determine correct entry type
	alphaImproved := false

	for i := 0; i < pseudoMoves.Count; i++ {
		m.pickNextMove(pseudoMoves, i, ply)
		move := pseudoMoves.Moves[i]

		if m.searchState.searchCancelled {
			break
		}

		// MoveGivesCheck reads the moving piece from move.From, so it must be
		// evaluated BEFORE tryMove applies the move: on the post-move board the
		// from-square is empty and every move looks like a non-check.
		givesCheck := board.MoveGivesCheck(b, move)

		undo, ok := m.tryMove(move, player)
		if !ok {
			continue
		}

		// Quiet-move pruning: internal futility and late move pruning. A
		// quiet move is skipped outright when either (a) the static eval
		// plus a depth-scaled margin still sits below alpha - the position
		// would need an implausible swing for this move to matter at the
		// remaining depth - or (b) under LMP, enough prior quiets have
		// failed to raise alpha that the remainder are presumed hopeless.
		// Exemptions: in-check nodes (evasions can be forced tactics),
		// sparse boards (see sparseBoard), captures and promotions,
		// giving-check moves, killer moves, and near-mate windows where
		// pruning could mask a mate distance. The unmake happens here
		// because a pruned move is never searched.
		if !sparseBoard && !inCheck && !move.IsCapture && move.Promotion == board.Empty &&
			!givesCheck && !m.isKillerMove(move, ply) &&
			alpha > -eval.MateScore+MateDistanceThreshold &&
			alpha < eval.MateScore-MateDistanceThreshold {

			p := &m.searchState.searchParams
			pruned := false
			if p.FutilityEnabled && depth <= p.FutilityMaxDepth &&
				stmStaticEval+p.FutilityMargins[min(depth, len(p.FutilityMargins)-1)] <= alpha {
				m.searchState.searchStats.FutilityPrunes++
				pruned = true
			} else if p.LMPEnabled && depth <= p.LMPMaxDepth &&
				quietsSearched >= 3+depth*depth {
				m.searchState.searchStats.LMPPrunes++
				pruned = true
			}
			if pruned {
				b.UnmakeMove(undo)
				quietMovesPruned = true
				continue
			}
		}

		m.addHistory(b.GetHash())

		legalMoveCount++
		var score eval.EvaluationScore

		if collectPV {
			pv.clearPly(pvPly + 1)
		}

		// Late Move Reductions: Search later moves at reduced depth since
		// move ordering should place best moves first
		reduction := m.calculateLMRReduction(depth, legalMoveCount, inCheck, givesCheck, move, ply)

		if reduction > 0 {
			m.searchState.searchStats.LMRReductions++

			m.searchState.player = oppositePlayer(player)
			m.searchState.collectPV = false
			// Reduced scout child ply = originalMaxDepth - (depth-1-reduction) =
			// ply+1-extended+reduction (searched shallower, so a larger ply index).
			score = -m.negamax(ctx, depth-1-reduction, -alpha-1, -alpha, ply+1-extended+reduction, pvPly+1)

			if score > alpha {
				m.searchState.searchStats.LMRReSearches++

				m.searchState.player = oppositePlayer(player)
				m.searchState.collectPV = collectPV
				score = -m.negamax(ctx, depth-1, -beta, -alpha, ply+1-extended, pvPly+1)
			}
		} else {
			m.searchState.player = oppositePlayer(player)
			m.searchState.collectPV = collectPV
			// Child ply = originalMaxDepth - (depth-1) = ply+1-extended.
			score = -m.negamax(ctx, depth-1, -beta, -alpha, ply+1-extended, pvPly+1)
		}

		b.UnmakeMove(undo)

		m.removeHistory()

		if !move.IsCapture && move.Promotion == board.Empty {
			quietsSearched++
		}

		if score > bestScore {
			bestScore = score
			bestMove = move
		}

		if score > alpha {
			alpha = score
			alphaImproved = true

			if collectPV {
				pv.store(pvPly, move)
			}

			if alpha >= beta {
				m.recordCutoff(move, depth, ply, legalMoveCount, hash, bestScore, pseudoMoves, i)
				return beta
			}
		}
	}

	if legalMoveCount == 0 {
		if !quietMovesPruned {
			return m.handleNoLegalMoves(b, player, depth, ply, ply-extended, hash)
		}
		// Every legal move was quiet-pruned and none was searched. This is a
		// fail-low with no information beyond the window itself: return alpha
		// (a valid upper bound) without touching the TT. Falling through to
		// handleNoLegalMoves here would claim stalemate/checkmate from zero
		// SEARCHED moves and store a false proven draw as EntryExact.
		return alpha
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
func (m *MinimaxEngine) tryNullMove(ctx context.Context, b *board.Board, player movegen.Player, depth int, beta eval.EvaluationScore, ply, extended, pvPly int, staticEval eval.EvaluationScore, inCheck bool) (score eval.EvaluationScore, done bool) {
	// Convert to side-to-move-relative before comparing against beta:
	// StandardEvaluator.Evaluate returns White-relative scores (see
	// internal/eval/evaluator.go), while beta is from the side to move's
	// perspective. Without this, Black nodes gate on an inverted sign.
	if player == movegen.Black {
		staticEval = -staticEval
	}
	if m.searchState.searchParams.NullMoveEnabled &&
		depth >= 3 &&
		staticEval >= beta &&
		beta < eval.MateScore-MateDistanceThreshold &&
		beta > -eval.MateScore+MateDistanceThreshold {
		if !inCheck {
			if !hasNonPawnMaterial(b, player) {
				m.searchState.searchStats.NullMoveZugzwangSkipped++
			} else if b.AllPieces.PopCount() <= NullMoveMaxPieces {
				// Zugzwang guard #2: hasNonPawnMaterial only protects the
				// mover's OWN pieces, but a side holding even one minor can be
				// in a fatal zugzwang if that piece is immobile (e.g. a bishop
				// locked in by its own pawns). Sparse boards are where these
				// positions live, so stop trusting NMP cutoffs there entirely.
				// Real case caught: kbK5/pp6/1P6/8/8/8/8/R7 w - a pure
				// zugzwang mate-in-2 that NMP silently pruned once
				// SEE-corrected move ordering changed which nodes it fired at.
				m.searchState.searchStats.NullMoveZugzwangSkipped++
			} else {
				nullReduction := m.searchState.searchParams.NullMoveReduction
				if depth >= 6 && nullReduction < 3 {
					nullReduction++
				}

				m.searchState.searchStats.NullMoves++

				nullUndo := b.MakeNullMove()

				m.searchState.player = oppositePlayer(player)
				m.searchState.collectPV = false
				// Child ply = originalMaxDepth - (depth-1-nullReduction). With
				// originalMaxDepth = ply + depth - extended this is ply+1-extended+nullReduction.
				nullScore := -m.negamax(ctx, depth-1-nullReduction, -beta, -beta+1, ply+1-extended+nullReduction, pvPly+1)

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
	}
	return 0, false
}

// hasNonPawnMaterial reports whether player has any knight, bishop, rook, or
// queen on the board. Null-move pruning assumes giving the opponent a free
// move can never help them, which is false in zugzwang positions - most
// commonly bare king-and-pawn endings, where every move can only weaken the
// position. Guarding on non-pawn material is the standard, cheap heuristic
// for this: skip null-move pruning once a side is down to king and pawns.
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

// tryRazoring performs razoring. At low depths far from mate, when static eval
// plus a margin is still below alpha, it verifies with quiescence and returns
// (qScore, true) if the quiescence score confirms the fail-low; otherwise
// (0, false). Stats (RazoringAttempts/Cutoffs/Failed) and the State player
// mutation are the verbatim body from negamax. A confirmed cutoff also
// increments AllNodes: qScore <= alpha is a fail-low, the same node type the
// move-loop's own !alphaImproved path classifies as.
func (m *MinimaxEngine) tryRazoring(ctx context.Context, player movegen.Player, depth int, alpha, beta eval.EvaluationScore, ply, extended int, staticEval eval.EvaluationScore, inCheck bool) (score eval.EvaluationScore, done bool) {
	// Convert to side-to-move-relative before comparing against alpha: see
	// tryNullMove. Razoring requires !inCheck so extended is 0 at the gate.
	if player == movegen.Black {
		staticEval = -staticEval
	}
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
// killer/history updates for quiet moves, and the lower-bound TT store.
//
// History gets the modern two-sided treatment: the cutoff move earns a
// depth-squared bonus while every quiet move tried before it in this node
// takes a matching malus - without maluses, moves that repeatedly fail keep
// high scores from past luck and history loses its discriminating power.
// moveList is the node's (pickNextMove-permuted) pseudo-legal list and
// triedIndex is the cutoff move's position within it; entries before it are
// exactly the moves this node already attempted. The two former
// `!move.IsCapture` blocks (killer, then history) are merged into a single
// guard — same effects, same order. The caller returns beta after this.
func (m *MinimaxEngine) recordCutoff(move board.Move, depth, ply, legalMoveCount int, hash uint64, bestScore eval.EvaluationScore, moveList *movegen.MoveList, triedIndex int) {
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
	}

	if m.historyTable != nil {
		if !move.IsCapture {
			m.historyTable.Bonus(move, depth)
		}
		// Malus every earlier-attempted quiet except the cutoff move itself.
		for j := 0; j < triedIndex; j++ {
			earlier := moveList.Moves[j]
			if earlier.IsCapture || hasPromotion(earlier) {
				continue
			}
			if earlier.From == move.From && earlier.To == move.To && earlier.Promotion == move.Promotion {
				continue
			}
			m.historyTable.Malus(earlier, depth)
		}
	}

	if m.transpositionTable != nil && !m.searchState.searchCancelled {
		m.transpositionTable.Store(hash, depth, scoreToTT(bestScore, ply), EntryLowerBound, move)
	}
}

// hasPromotion reports whether a move promotes. Both Empty ('.') and the
// zero value count as "no promotion": hand-built Move literals leave
// Promotion unset, which is not Empty.
func hasPromotion(move board.Move) bool {
	return move.Promotion != board.Empty && move.Promotion != 0
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
//
// givesCheck must be computed by the caller on the PRE-move board: this
// function runs after the move has been made, when the from-square is already
// vacated and a board-side check test would always report false (see
// board.MoveGivesCheck).
func (m *MinimaxEngine) calculateLMRReduction(depth, legalMoveCount int, inCheck, givesCheck bool, move board.Move, ply int) int {
	if !m.searchState.searchParams.LMREnabled ||
		depth < m.searchState.searchParams.LMRMinDepth ||
		legalMoveCount <= m.searchState.searchParams.LMRMinMoves ||
		inCheck ||
		move.IsCapture ||
		move.Promotion != board.Empty ||
		m.isKillerMove(move, ply) {
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

	// Giving-check moves are reduced only gently rather than exempted: modern
	// engines reduce checks too and lean on the fail-high re-search for safety,
	// because full exemption costs node count and effective depth (measured at
	// roughly -26 Elo here). Cap matches Ethereal's "R -= check" spirit.
	if givesCheck {
		reduction = min(reduction, 1)
	}

	return reduction
}
