package search

import (
	"context"
	"sort"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/evaluation"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
	"github.com/AdamGriffiths31/ChessEngine/game/openings"
)

const (
	// MinEval represents the worst possible evaluation score
	MinEval = ai.EvaluationScore(-1000000)
	// MaxKillerDepth defines the maximum depth for killer move storage
	MaxKillerDepth = 64
	// MateDistanceThreshold defines the distance from mate scores where special handling is needed
	MateDistanceThreshold = 1000
)

// MinimaxEngine implements negamax search with alpha-beta pruning, transposition table,
// history heuristic, null move pruning, and opening book support
type MinimaxEngine struct {
	evaluator          ai.Evaluator
	generator          *moves.Generator
	bookService        *openings.BookLookupService
	killerTable        [MaxKillerDepth][2]board.Move
	debugMoveOrdering  bool
	debugMoveOrder     []board.Move
	transpositionTable *TranspositionTable
	zobrist            *openings.ZobristHash
	historyTable       *HistoryTable
}

// NewMinimaxEngine creates a new minimax search engine
func NewMinimaxEngine() *MinimaxEngine {
	return &MinimaxEngine{
		evaluator:          evaluation.NewEvaluator(),
		generator:          moves.NewGenerator(),
		bookService:        nil,
		transpositionTable: nil,
		zobrist:            openings.GetPolyglotHash(),
		historyTable:       NewHistoryTable(),
	}
}

// GetHashDelta implements the board.HashUpdater interface
// Calculates the zobrist hash delta for a move from old state to new state
func (m *MinimaxEngine) GetHashDelta(b *board.Board, move board.Move, oldState board.BoardState) uint64 {
	var hashDelta uint64

	// Always flip side to move
	hashDelta ^= m.zobrist.GetSideKey()

	// Handle piece movement (remove from source, add to destination)
	fromSquare := move.From.Rank*8 + move.From.File
	toSquare := move.To.Rank*8 + move.To.File

	// Remove piece from source square
	if move.Piece != board.Empty {
		pieceIndex := m.zobrist.GetPieceIndex(move.Piece)
		hashDelta ^= m.zobrist.GetPieceKey(fromSquare, pieceIndex)
	}

	// Add piece to destination square (or promotion piece)
	var destPiece board.Piece
	if move.Promotion != board.Empty && move.Promotion != 0 {
		destPiece = move.Promotion
	} else {
		destPiece = move.Piece
	}
	if destPiece != board.Empty {
		pieceIndex := m.zobrist.GetPieceIndex(destPiece)
		hashDelta ^= m.zobrist.GetPieceKey(toSquare, pieceIndex)
	}

	// Handle captured piece
	if move.IsCapture && move.Captured != board.Empty {
		capturedPieceIndex := m.zobrist.GetPieceIndex(move.Captured)
		if move.IsEnPassant {
			// En passant capture is on a different square
			var captureRank int
			if move.Piece == board.WhitePawn {
				captureRank = 4 // Black pawn on rank 5
			} else {
				captureRank = 3 // White pawn on rank 4
			}
			captureSquare := captureRank*8 + move.To.File
			hashDelta ^= m.zobrist.GetPieceKey(captureSquare, capturedPieceIndex)
		} else {
			hashDelta ^= m.zobrist.GetPieceKey(toSquare, capturedPieceIndex)
		}
	}

	// Handle castling rook movement
	if move.IsCastling {
		var rookFrom, rookTo int
		switch move.To.File {
		case 6: // King-side castling
			rookFrom = move.From.Rank*8 + 7
			rookTo = move.From.Rank*8 + 5
		case 2: // Queen-side castling
			rookFrom = move.From.Rank*8 + 0
			rookTo = move.From.Rank*8 + 3
		}
		
		// Remove rook from original square, add to new square
		var rook board.Piece
		if move.From.Rank == 0 {
			rook = board.BlackRook
		} else {
			rook = board.WhiteRook
		}
		rookIndex := m.zobrist.GetPieceIndex(rook)
		hashDelta ^= m.zobrist.GetPieceKey(rookFrom, rookIndex)
		hashDelta ^= m.zobrist.GetPieceKey(rookTo, rookIndex)
	}

	// Handle castling rights changes
	if oldState.CastlingRights != b.GetCastlingRights() {
		oldRights := m.zobrist.GetCastlingKey(oldState.CastlingRights)
		newRights := m.zobrist.GetCastlingKey(b.GetCastlingRights())
		hashDelta ^= oldRights ^ newRights
	}

	// Handle en passant target changes
	// Only include en passant in hash if there are adjacent pawns that can capture
	if (oldState.EnPassantTarget == nil) != (b.GetEnPassantTarget() == nil) ||
		(oldState.EnPassantTarget != nil && b.GetEnPassantTarget() != nil &&
			oldState.EnPassantTarget.File != b.GetEnPassantTarget().File) {
		
		if oldState.EnPassantTarget != nil && m.hasAdjacentCapturingPawn(b, oldState.EnPassantTarget, oldState.SideToMove) {
			hashDelta ^= m.zobrist.GetEnPassantKey(oldState.EnPassantTarget.File)
		}
		if b.GetEnPassantTarget() != nil && m.hasAdjacentCapturingPawn(b, b.GetEnPassantTarget(), b.GetSideToMove()) {
			hashDelta ^= m.zobrist.GetEnPassantKey(b.GetEnPassantTarget().File)
		}
	}

	return hashDelta
}

// hasAdjacentCapturingPawn checks if there's a pawn adjacent to the en passant target that can capture
// This implements the same logic as the full HashPosition function
func (m *MinimaxEngine) hasAdjacentCapturingPawn(b *board.Board, epTarget *board.Square, sideToMove string) bool {
	// Determine the rank where the capturing pawn should be and what piece to look for
	var pawnRank int
	var pawnPiece board.Piece
	
	if sideToMove == "b" { // Black to move, so white pawn moved and black can capture
		pawnRank = 4 // Black pawn should be on rank 5 (0-indexed rank 4)
		pawnPiece = board.BlackPawn
	} else { // White to move, so black pawn moved and white can capture  
		pawnRank = 3 // White pawn should be on rank 4 (0-indexed rank 3)
		pawnPiece = board.WhitePawn
	}
	
	epFile := epTarget.File
	
	// Check adjacent files for capturing pawn
	for _, df := range []int{-1, 1} {
		adjFile := epFile + df
		if adjFile >= 0 && adjFile < 8 {
			if b.GetPiece(pawnRank, adjFile) == pawnPiece {
				return true
			}
		}
	}
	
	return false
}

// initializeBookService initializes the opening book service based on configuration
func (m *MinimaxEngine) initializeBookService(config ai.SearchConfig) error {
	if !config.UseOpeningBook || len(config.BookFiles) == 0 {
		m.bookService = nil
		return nil
	}

	// Convert AI selection mode to openings selection mode
	var selectionMode openings.SelectionMode
	switch config.BookSelectMode {
	case ai.BookSelectBest:
		selectionMode = openings.SelectBest
	case ai.BookSelectRandom:
		selectionMode = openings.SelectRandom
	case ai.BookSelectWeightedRandom:
		selectionMode = openings.SelectWeightedRandom
	default:
		selectionMode = openings.SelectWeightedRandom
	}

	bookConfig := openings.BookConfig{
		Enabled:         true,
		BookFiles:       config.BookFiles,
		SelectionMode:   selectionMode,
		WeightThreshold: config.BookWeightThreshold,
	}

	service := openings.NewBookLookupService(bookConfig)
	err := service.LoadBooks()
	if err != nil {
		return err
	}

	m.bookService = service

	return nil
}

// FindBestMove searches for the best move using minimax with optional opening book
func (m *MinimaxEngine) FindBestMove(ctx context.Context, b *board.Board, player moves.Player, config ai.SearchConfig) ai.SearchResult {
	result := ai.SearchResult{
		Stats: ai.SearchStats{},
	}

	if config.UseOpeningBook {
		m.initializeBookService(config)
	}

	if config.UseOpeningBook && m.bookService != nil {
		bookMove, err := m.bookService.FindBookMove(b)
		if err == nil && bookMove != nil {
			result.BestMove = *bookMove
			result.Score = 0
			result.Stats.BookMoveUsed = true
			return result
		}
	}

	// Initialize incremental hashing for the board
	b.SetHashUpdater(m)
	b.InitializeHashFromPosition(m.zobrist.HashPosition)

	startTime := time.Now()

	if m.transpositionTable != nil {
		m.transpositionTable.IncrementAge()
	}

	if m.historyTable != nil {
		m.historyTable.Age()
	}

	legalMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(legalMoves)

	if legalMoves.Count == 0 {
		result.BestMove = board.Move{From: board.Square{File: -1, Rank: -1}}
		if m.generator.IsKingInCheck(b, player) {
			result.Score = -ai.MateScore
		} else {
			result.Score = ai.DrawScore
		}
		return result
	}

	var rootTTMove board.Move
	if m.transpositionTable != nil {
		hash := b.GetHash()
		if entry, found := m.transpositionTable.Probe(hash); found {
			rootTTMove = entry.BestMove
		}
	}

	m.orderMoves(legalMoves, 0, rootTTMove)

	lastCompletedBestMove := legalMoves.Moves[0]
	lastCompletedScore := ai.EvaluationScore(0)

	for currentDepth := 1; currentDepth <= config.MaxDepth; currentDepth++ {
		select {
		case <-ctx.Done():
			result.BestMove = lastCompletedBestMove
			result.Score = lastCompletedScore
			result.Stats.Time = time.Since(startTime)
			return result
		default:
		}

		bestScore := ai.EvaluationScore(-ai.MateScore - 1)
		bestMove := legalMoves.Moves[0]
		var searchStats ai.SearchStats
		var completed bool

		// Progressive aspiration window implementation
		var windowSize ai.EvaluationScore = 100 // Start with wider window (±100)
		var maxRetries = 3
		
		for retry := 0; retry <= maxRetries && !completed; retry++ {
			var alpha, beta ai.EvaluationScore
			
			// Set initial window based on depth and retry count
			if currentDepth == 1 {
				// Full window for first depth
				alpha = ai.EvaluationScore(-ai.MateScore - 1)
				beta = ai.EvaluationScore(ai.MateScore + 1)
			} else if retry == 0 {
				// Initial aspiration window - adaptive sizing
				adaptiveWindow := m.calculateAdaptiveWindow(currentDepth, windowSize)
				alpha = lastCompletedScore - adaptiveWindow
				beta = lastCompletedScore + adaptiveWindow
			} else {
				// Progressive window widening on retries
				switch retry {
				case 1:
					windowSize = 150 // Widen to ±150
					alpha = lastCompletedScore - windowSize
					beta = lastCompletedScore + windowSize
				case 2:
					windowSize = 300 // Widen to ±300
					alpha = lastCompletedScore - windowSize
					beta = lastCompletedScore + windowSize
				default:
					// Full window as last resort
					alpha = ai.EvaluationScore(-ai.MateScore - 1)
					beta = ai.EvaluationScore(ai.MateScore + 1)
				}
			}
			
			originalAlpha := alpha
			originalBeta := beta
			searchStats = ai.SearchStats{}
			bestScore = ai.EvaluationScore(-ai.MateScore - 1)

			// Search all moves with current window
			searchCancelled := false
			for i := 0; i < legalMoves.Count && !searchCancelled; i++ {
				move := legalMoves.Moves[i]

				select {
				case <-ctx.Done():
					searchCancelled = true
					break
				default:
				}

				undo, err := b.MakeMoveWithUndo(move)
				if err != nil {
					continue
				}

				var moveStats ai.SearchStats
				score := -m.negamaxWithAlphaBeta(ctx, b, oppositePlayer(player), currentDepth-1, -beta, -alpha, currentDepth, config, &moveStats)

				b.UnmakeMove(undo)

				searchStats.NodesSearched += moveStats.NodesSearched

				select {
				case <-ctx.Done():
					searchCancelled = true
					break
				default:
				}

				if score > bestScore {
					bestScore = score
					bestMove = move

					if score > alpha {
						alpha = score
					}
				}
			}
			
			// Check if search was cancelled
			if searchCancelled {
				completed = false
				break
			}
			
			// Check if aspiration window failed and we need to retry
			if currentDepth > 1 && retry < maxRetries && (bestScore <= originalAlpha || bestScore >= originalBeta) {
				// Window failed, continue to next retry iteration
				continue
			} else {
				// Search completed successfully or no more retries
				completed = true
			}
		}

		if completed {
			lastCompletedBestMove = bestMove
			lastCompletedScore = bestScore
			result.Stats.Depth = currentDepth
			result.Stats.NodesSearched = searchStats.NodesSearched

			if m.transpositionTable != nil {
				hash := b.GetHash()
				m.transpositionTable.Store(hash, currentDepth, bestScore, EntryExact, bestMove)
			}
		} else {
			result.BestMove = lastCompletedBestMove
			result.Score = lastCompletedScore
			result.Stats.Time = time.Since(startTime)
			return result
		}
	}

	result.BestMove = lastCompletedBestMove
	result.Score = lastCompletedScore
	result.Stats.Time = time.Since(startTime)
	return result
}

func (m *MinimaxEngine) quiescence(ctx context.Context, b *board.Board, player moves.Player, alpha, beta ai.EvaluationScore, depthFromRoot int, stats *ai.SearchStats) ai.EvaluationScore {
	stats.NodesSearched++

	select {
	case <-ctx.Done():
		eval := m.evaluator.Evaluate(b)
		if player == moves.Black {
			eval = -eval
		}
		return eval
	default:
	}

	hash := b.GetHash()
	if m.transpositionTable != nil {
		if entry, found := m.transpositionTable.Probe(hash); found {
			if entry.Depth >= 0 {
				switch entry.Type {
				case EntryExact:
					return entry.Score
				case EntryLowerBound:
					if entry.Score >= beta {
						return entry.Score
					}
					if entry.Score > alpha {
						alpha = entry.Score
					}
				case EntryUpperBound:
					if entry.Score <= alpha {
						return entry.Score
					}
					if entry.Score < beta {
						beta = entry.Score
					}
				}
			}
		}
	}

	inCheck := m.generator.IsKingInCheck(b, player)
	originalAlpha := alpha

	if !inCheck {
		eval := m.evaluator.Evaluate(b)
		if player == moves.Black {
			eval = -eval
		}

		if eval >= beta {
			return beta
		}

		if eval > alpha {
			alpha = eval
		}
	}

	legalMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(legalMoves)

	if legalMoves.Count == 0 {
		if inCheck {
			return -ai.MateScore + ai.EvaluationScore(depthFromRoot)
		} else {
			return ai.DrawScore
		}
	}

	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]

		if !inCheck && !move.IsCapture {
			continue
		}

		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			continue
		}

		score := -m.quiescence(ctx, b, oppositePlayer(player), -beta, -alpha, depthFromRoot+1, stats)

		b.UnmakeMove(undo)

		if score > alpha {
			alpha = score

			if alpha >= beta {
				return beta
			}
		}
	}

	if m.transpositionTable != nil {
		entryType := EntryExact
		if alpha <= originalAlpha {
			entryType = EntryUpperBound
		} else if alpha >= beta {
			entryType = EntryLowerBound
		}
		m.transpositionTable.Store(hash, 0, alpha, entryType, board.Move{})
	}

	return alpha
}

func (m *MinimaxEngine) negamaxWithAlphaBeta(ctx context.Context, b *board.Board, player moves.Player, depth int, alpha, beta ai.EvaluationScore, originalMaxDepth int, config ai.SearchConfig, stats *ai.SearchStats) ai.EvaluationScore {
	stats.NodesSearched++

	currentDepth := originalMaxDepth - depth
	if currentDepth > stats.Depth {
		stats.Depth = currentDepth
	}

	select {
	case <-ctx.Done():
		eval := m.evaluator.Evaluate(b)
		if player == moves.Black {
			eval = -eval
		}
		return eval
	default:
	}

	var ttMove board.Move
	hash := b.GetHash()

	if m.transpositionTable != nil {
		if entry, found := m.transpositionTable.Probe(hash); found {
			ttMove = entry.BestMove

			if entry.Depth >= depth {
				switch entry.Type {
				case EntryExact:
					return entry.Score
				case EntryLowerBound:
					if entry.Score >= beta {
						return entry.Score
					}
					if entry.Score > alpha {
						alpha = entry.Score
					}
				case EntryUpperBound:
					if entry.Score <= alpha {
						return entry.Score
					}
					if entry.Score < beta {
						beta = entry.Score
					}
				}
			}
		}
	}

	if config.UseNullMove && depth >= 3 && beta < ai.MateScore-MateDistanceThreshold && beta > -ai.MateScore+MateDistanceThreshold {
		inCheck := m.generator.IsKingInCheck(b, player)
		if !inCheck {
			nullReduction := 2
			if depth >= 6 {
				nullReduction = 3
			}

			nullUndo := b.MakeNullMove()

			nullScore := -m.negamaxWithAlphaBeta(ctx, b, oppositePlayer(player),
				depth-1-nullReduction, -beta, -beta+1, originalMaxDepth, config, stats)

			b.UnmakeNullMove(nullUndo)

			// If null move score >= beta, position is too good for opponent
			if nullScore >= beta {
				if nullScore < ai.MateScore-MateDistanceThreshold {
					return beta
				}
			}
		}
	}

	// Terminal node - call quiescence search
	if depth == 0 {
		score := m.quiescence(ctx, b, player, alpha, beta, originalMaxDepth-depth, stats)

		if m.transpositionTable != nil {
			m.transpositionTable.Store(hash, 0, score, EntryExact, board.Move{})
		}

		return score
	}

	legalMoves := m.generator.GenerateAllMoves(b, player)
	defer moves.ReleaseMoveList(legalMoves)

	if legalMoves.Count == 0 {
		if m.generator.IsKingInCheck(b, player) {
			score := -ai.MateScore + ai.EvaluationScore(originalMaxDepth-depth)
			if m.transpositionTable != nil {
				m.transpositionTable.Store(hash, depth, score, EntryExact, board.Move{})
			}
			return score
		}
		if m.transpositionTable != nil {
			m.transpositionTable.Store(hash, depth, ai.DrawScore, EntryExact, board.Move{})
		}
		return ai.DrawScore
	}

	// Sort moves to improve alpha-beta efficiency (TT move, captures, killers, history)
	m.orderMoves(legalMoves, currentDepth, ttMove)

	// Search moves
	bestScore := ai.EvaluationScore(-ai.MateScore - 1)
	bestMove := board.Move{}
	entryType := EntryUpperBound // Assume fail-low initially

	for i := 0; i < legalMoves.Count; i++ {
		move := legalMoves.Moves[i]

		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			continue
		}

		// Search deeper with negamax and alpha-beta
		score := -m.negamaxWithAlphaBeta(ctx, b, oppositePlayer(player), depth-1, -beta, -alpha, originalMaxDepth, config, stats)

		// Unmake the move
		b.UnmakeMove(undo)

		// Update best score
		if score > bestScore {
			bestScore = score
			bestMove = move
		}

		// Update alpha (best score for current player)
		if score > alpha {
			alpha = score
			entryType = EntryExact

			// Beta cutoff - opponent won't allow this line
			if alpha >= beta {
				if !move.IsCapture && currentDepth >= 0 && currentDepth < MaxKillerDepth {
					m.storeKiller(move, currentDepth)
				}

				if !move.IsCapture {
					m.historyTable.UpdateHistory(move, depth)
				}

				if m.transpositionTable != nil {
					m.transpositionTable.Store(hash, depth, beta, EntryLowerBound, move)
				}

				return beta
			}
		}
	}

	if m.transpositionTable != nil {
		m.transpositionTable.Store(hash, depth, bestScore, entryType, bestMove)
	}

	return bestScore
}

// getMVVLVAScore calculates the MVV-LVA score for a capture move
// Higher scores indicate more valuable captures (better moves to try first)
func (m *MinimaxEngine) getMVVLVAScore(move board.Move) int {
	if !move.IsCapture || move.Captured == board.Empty {
		return 0 // Non-captures get score 0
	}

	victimValue := evaluation.PieceValues[move.Captured]
	if victimValue < 0 {
		victimValue = -victimValue
	}

	attackerValue := evaluation.PieceValues[move.Piece]
	if attackerValue < 0 {
		attackerValue = -attackerValue
	}

	return victimValue*10 - attackerValue
}

// orderMoves sorts moves to improve alpha-beta efficiency using TT move, MVV-LVA, killer moves, and history heuristic
func (m *MinimaxEngine) orderMoves(moveList *moves.MoveList, depth int, ttMove board.Move) {
	// Create slice of move indices with their scores
	type moveScore struct {
		index int
		score int
	}

	scores := make([]moveScore, moveList.Count)
	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		var score int

		if ttMove.From.File != -1 && ttMove.From.Rank != -1 &&
			move.From == ttMove.From && move.To == ttMove.To {
			score = 3000000
		} else if move.IsCapture {
			score = 1000000 + m.getMVVLVAScore(move)
		} else if m.isKillerMove(move, depth) {
			if depth >= 0 && depth < MaxKillerDepth &&
				move.From == m.killerTable[depth][0].From && move.To == m.killerTable[depth][0].To {
				score = 500000
			} else {
				score = 490000
			}
		} else {
			historyScore := m.getHistoryScore(move)
			if historyScore > 0 {
				score = int(historyScore)
			} else {
				score = 0
			}
		}
		scores[i] = moveScore{index: i, score: score}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	for i := 0; i < moveList.Count; i++ {
		targetIndex := scores[i].index
		if targetIndex != i {
			moveList.Moves[i], moveList.Moves[targetIndex] = moveList.Moves[targetIndex], moveList.Moves[i]

			for j := i + 1; j < moveList.Count; j++ {
				if scores[j].index == i {
					scores[j].index = targetIndex
					break
				}
			}
		}
	}

	if m.debugMoveOrdering && depth == 0 {
		m.debugMoveOrder = make([]board.Move, moveList.Count)
		copy(m.debugMoveOrder, moveList.Moves[:moveList.Count])
	}
}

// SetEvaluator sets the position evaluator
func (m *MinimaxEngine) SetEvaluator(eval ai.Evaluator) {
	m.evaluator = eval
}

// SetTranspositionTableSize initializes the transposition table with the given size in MB
func (m *MinimaxEngine) SetTranspositionTableSize(sizeMB int) {
	if sizeMB <= 0 {
		m.transpositionTable = nil
		return
	}
	m.transpositionTable = NewTranspositionTable(sizeMB)
}

// SetTranspositionTableEnabled enables or disables transposition table usage
func (m *MinimaxEngine) SetTranspositionTableEnabled(enabled bool) {
	if m.transpositionTable != nil {
	}
}

// GetTranspositionTableStats returns transposition table statistics if available
func (m *MinimaxEngine) GetTranspositionTableStats() (hits, misses, collisions uint64, hitRate float64) {
	if m.transpositionTable != nil {
		return m.transpositionTable.GetStats()
	}
	return 0, 0, 0, 0
}

// GetName returns the engine name
func (m *MinimaxEngine) GetName() string {
	return "Minimax Engine"
}

// ClearSearchState clears transient search state between different positions
func (m *MinimaxEngine) ClearSearchState() {
	for i := 0; i < MaxKillerDepth; i++ {
		m.killerTable[i][0] = board.Move{}
		m.killerTable[i][1] = board.Move{}
	}
	if m.transpositionTable != nil {
		m.transpositionTable.Clear()
	}
	if m.historyTable != nil {
		m.historyTable.Clear()
	}
}

// SetDebugMoveOrdering enables/disables move ordering debug tracking (for tests only)
func (m *MinimaxEngine) SetDebugMoveOrdering(enabled bool) {
	m.debugMoveOrdering = enabled
	if !enabled {
		m.debugMoveOrder = nil
	}
}

// GetLastMoveOrder returns the move order from the last orderMoves call (for tests only)
func (m *MinimaxEngine) GetLastMoveOrder() []board.Move {
	return m.debugMoveOrder
}

// isKillerMove checks if a move is a killer move at the given depth
func (m *MinimaxEngine) isKillerMove(move board.Move, depth int) bool {
	if depth < 0 || depth >= MaxKillerDepth {
		return false
	}

	// Check both killer slots at this depth
	return (move.From == m.killerTable[depth][0].From && move.To == m.killerTable[depth][0].To) ||
		(move.From == m.killerTable[depth][1].From && move.To == m.killerTable[depth][1].To)
}

// storeKiller stores a move as a killer move, shifting existing killers
func (m *MinimaxEngine) storeKiller(move board.Move, depth int) {
	if depth < 0 || depth >= MaxKillerDepth {
		return
	}

	if m.isKillerMove(move, depth) {
		return
	}

	m.killerTable[depth][1] = m.killerTable[depth][0]
	m.killerTable[depth][0] = move
}

// getHistoryScore returns the history score for a move
func (m *MinimaxEngine) getHistoryScore(move board.Move) int32 {
	if m.historyTable == nil {
		return 0
	}
	return m.historyTable.GetHistoryScore(move)
}

// calculateAdaptiveWindow determines the appropriate aspiration window size
// based on search depth and position characteristics
func (m *MinimaxEngine) calculateAdaptiveWindow(depth int, baseWindow ai.EvaluationScore) ai.EvaluationScore {
	// Start with base window and adjust based on depth
	adaptiveWindow := baseWindow
	
	// Increase window size for deeper searches as they tend to be more volatile
	if depth >= 6 {
		adaptiveWindow += ai.EvaluationScore(depth * 10)
	}
	
	// Cap the maximum window size to prevent excessive re-searches
	maxWindow := ai.EvaluationScore(200)
	if adaptiveWindow > maxWindow {
		adaptiveWindow = maxWindow
	}
	
	return adaptiveWindow
}

// oppositePlayer returns the opposite player
func oppositePlayer(player moves.Player) moves.Player {
	if player == moves.White {
		return moves.Black
	}
	return moves.White
}
