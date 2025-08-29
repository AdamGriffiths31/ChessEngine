package search

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/evaluation"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
	"github.com/AdamGriffiths31/ChessEngine/game/openings"
)

const (
	// MinEval represents the minimum possible evaluation score
	MinEval = ai.EvaluationScore(-32000)
	// MaxKillerDepth is the maximum depth for killer move tables
	MaxKillerDepth = 128
	// MateDistanceThreshold is the threshold for detecting mate distances
	MateDistanceThreshold = 1000
	// MaxGamePly is the maximum number of plies to track for repetition detection
	MaxGamePly = 1024
)

// LMRTable is a pre-calculated reduction table for Late Move Reductions
// Indexed by [depth][moveCount] to get reduction amount
var LMRTable [16][64]int

func init() {
	for depth := 1; depth < 16; depth++ {
		for moveCount := 1; moveCount < 64; moveCount++ {
			LMRTable[depth][moveCount] = int(math.Log(float64(depth)) * math.Log(float64(moveCount)) / 1.8)
		}
	}
}

// Params holds search parameters
type Params struct {
	LMRDivisor           float64
	NullMoveReduction    int
	HistoryHighThreshold int32
	HistoryMedThreshold  int32
	HistoryLowThreshold  int32

	// Razoring parameters
	RazoringEnabled  bool
	RazoringMargins  [5]ai.EvaluationScore // Margins for depths 1-4 (index 0 unused)
	RazoringMaxDepth int                   // Maximum depth to apply razoring

	// Futility pruning parameters
	FutilityMargins [5]ai.EvaluationScore // Futility margins for depths 1-4 (index 0 unused)

	// Extension thresholds
	CheckExtensionThreshold int                // Whether to extend single checks (0=no, 1=yes)
	SingularExtensionMargin ai.EvaluationScore // Margin for singular extensions
}

// getParams returns well-tuned search parameters
func getParams() Params {
	return Params{
		LMRDivisor:           1.8,  // Standard LMR divisor
		NullMoveReduction:    2,    // Conservative null move
		HistoryHighThreshold: 2000, // Well-tested history values
		HistoryMedThreshold:  500,
		HistoryLowThreshold:  -500,

		// Stockfish-aligned razoring margins - targeting 10-15% attempt rate
		RazoringEnabled:  true,
		RazoringMargins:  [5]ai.EvaluationScore{0, 100, 150, 200, 250},
		RazoringMaxDepth: 3,

		// Standard futility pruning
		FutilityMargins: [5]ai.EvaluationScore{0, 100, 200, 300, 400},

		// Standard extensions
		CheckExtensionThreshold: 1,
		SingularExtensionMargin: 100,
	}
}

// State contains state for search
type State struct {
	killerTable     [MaxKillerDepth][2]board.Move
	moveOrderBuffer []moveScore
	reorderBuffer   []board.Move
	searchStats     ai.SearchStats
	searchParams    Params
	debugMoveOrder  []board.Move
	searchCancelled bool
}

// MinimaxEngine implements negamax search with alpha-beta pruning, transposition table,
// history heuristic, null move pruning, SEE-based move ordering, and opening book support
type MinimaxEngine struct {
	evaluator          ai.Evaluator
	generator          *moves.Generator
	bookService        *openings.BookLookupService
	transpositionTable *TranspositionTable
	zobrist            *openings.ZobristHash
	historyTable       *HistoryTable
	seeCalculator      *evaluation.SEECalculator
	debugMoveOrdering  bool
	searchState        State

	// Repetition detection
	zobristHistory    [MaxGamePly]uint64
	zobristHistoryPly uint16
}

// moveScore holds move index and score for ordering
type moveScore struct {
	index int
	score int
}

// NewMinimaxEngine creates a new minimax search engine
func NewMinimaxEngine() *MinimaxEngine {
	engine := &MinimaxEngine{
		evaluator:          evaluation.NewEvaluator(),
		generator:          moves.NewGenerator(),
		bookService:        nil,
		transpositionTable: nil,
		zobrist:            openings.GetPolyglotHash(),
		historyTable:       NewHistoryTable(),
		seeCalculator:      evaluation.NewSEECalculator(),
		searchState: State{
			killerTable:     [MaxKillerDepth][2]board.Move{},
			moveOrderBuffer: make([]moveScore, 0, 512),
			reorderBuffer:   make([]board.Move, 0, 512),
			searchParams:    getParams(),
			debugMoveOrder:  make([]board.Move, 0),
		},
	}

	return engine
}

// GetHashDelta implements the board.HashUpdater interface
// Calculates the zobrist hash delta for a move from old state to new state
func (m *MinimaxEngine) GetHashDelta(b *board.Board, move board.Move, oldState board.State) uint64 {
	var hashDelta uint64

	hashDelta ^= m.zobrist.GetSideKey()

	fromSquare := move.From.Rank*8 + move.From.File
	toSquare := move.To.Rank*8 + move.To.File

	if move.Piece != board.Empty {
		pieceIndex := m.zobrist.GetPieceIndex(move.Piece)
		hashDelta ^= m.zobrist.GetPieceKey(fromSquare, pieceIndex)
	}

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

	if move.IsCapture && move.Captured != board.Empty {
		capturedPieceIndex := m.zobrist.GetPieceIndex(move.Captured)
		if move.IsEnPassant {
			var captureRank int
			if move.Piece == board.WhitePawn {
				captureRank = 4
			} else {
				captureRank = 3
			}
			captureSquare := captureRank*8 + move.To.File
			hashDelta ^= m.zobrist.GetPieceKey(captureSquare, capturedPieceIndex)
		} else {
			hashDelta ^= m.zobrist.GetPieceKey(toSquare, capturedPieceIndex)
		}
	}

	if move.IsCastling {
		var rookFrom, rookTo int
		switch move.To.File {
		case 6:
			rookFrom = move.From.Rank*8 + 7
			rookTo = move.From.Rank*8 + 5
		case 2:
			rookFrom = move.From.Rank*8 + 0
			rookTo = move.From.Rank*8 + 3
		}

		var rook board.Piece
		if move.From.Rank == 0 {
			rook = board.WhiteRook
		} else {
			rook = board.BlackRook
		}
		rookIndex := m.zobrist.GetPieceIndex(rook)
		hashDelta ^= m.zobrist.GetPieceKey(rookFrom, rookIndex)
		hashDelta ^= m.zobrist.GetPieceKey(rookTo, rookIndex)
	}

	if oldState.CastlingRights != b.GetCastlingRights() {
		oldRights := m.zobrist.GetCastlingKey(oldState.CastlingRights)
		newRights := m.zobrist.GetCastlingKey(b.GetCastlingRights())
		hashDelta ^= oldRights ^ newRights
	}

	if (oldState.EnPassantTarget == nil) != (b.GetEnPassantTarget() == nil) ||
		(oldState.EnPassantTarget != nil && b.GetEnPassantTarget() != nil &&
			oldState.EnPassantTarget.File != b.GetEnPassantTarget().File) {

		if oldState.EnPassantTarget != nil && m.hasAdjacentCapturingPawn(b, oldState.EnPassantTarget, getOppositeSide(oldState.SideToMove)) {
			hashDelta ^= m.zobrist.GetEnPassantKey(oldState.EnPassantTarget.File)
		}
		if b.GetEnPassantTarget() != nil && m.hasAdjacentCapturingPawn(b, b.GetEnPassantTarget(), getOppositeSide(b.GetSideToMove())) {
			hashDelta ^= m.zobrist.GetEnPassantKey(b.GetEnPassantTarget().File)
		}
	}

	return hashDelta
}

// GetNullMoveDelta returns the hash delta for a null move (flip side to move)
func (m *MinimaxEngine) GetNullMoveDelta() uint64 {
	return m.zobrist.GetSideKey()
}

// hasAdjacentCapturingPawn checks if there's a pawn adjacent to the en passant target that can capture
// This implements the same logic as the full HashPosition function
func (m *MinimaxEngine) hasAdjacentCapturingPawn(b *board.Board, epTarget *board.Square, sideToMove string) bool {
	var pawnRank int
	var pawnPiece board.Piece

	if sideToMove == "b" {
		pawnRank = 4
		pawnPiece = board.BlackPawn
	} else if sideToMove == "w" {
		pawnRank = 3
		pawnPiece = board.WhitePawn
	} else {
		panic("invalid sideToMove: " + sideToMove)
	}

	epFile := epTarget.File

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

// getOppositeSide returns the opposite side
func getOppositeSide(side string) string {
	if side == "w" {
		return "b"
	}
	return "w"
}

// initializeBookService initializes the opening book service with simple configuration
func (m *MinimaxEngine) initializeBookService(config ai.SearchConfig) error {
	if !config.UseOpeningBook || len(config.BookFiles) == 0 {
		m.bookService = nil
		return nil
	}

	bookConfig := openings.BookConfig{
		Enabled:       true,
		BookFiles:     config.BookFiles,
		SelectionMode: openings.SelectBest, // Always pick best move
	}

	service := openings.NewBookLookupService(bookConfig)
	err := service.LoadBooks()
	if err != nil {
		return fmt.Errorf("failed to load opening books: %w", err)
	}

	m.bookService = service
	return nil
}

// FindBestMove searches for the best move using minimax with optional opening book
func (m *MinimaxEngine) FindBestMove(ctx context.Context, b *board.Board, player moves.Player, config ai.SearchConfig) ai.SearchResult {
	// Initialize book service only once if not already done and still in opening phase
	moveNumber := b.GetFullMoveNumber()
	if config.UseOpeningBook && m.bookService == nil && moveNumber <= 10 {
		if err := m.initializeBookService(config); err != nil {
			fmt.Printf("Warning: failed to initialize opening book service: %v\n", err)
		}
	}

	// Check opening book only in first 10 moves
	if config.UseOpeningBook && m.bookService != nil && moveNumber <= 10 {
		bookMove, err := m.bookService.FindBookMove(b)
		if err == nil && bookMove != nil {
			return ai.SearchResult{
				BestMove: *bookMove,
				Score:    0,
				Stats:    ai.SearchStats{BookMoveUsed: true},
			}
		}
	}

	b.SetHashUpdater(m)
	b.InitializeHashFromPosition(m.zobrist.HashPosition)

	// Setup repetition detection with root position
	m.setupRepetitionHistory(b.GetHash())

	startTime := time.Now()

	if m.transpositionTable != nil {
		m.transpositionTable.IncrementAge()
	}

	if m.historyTable != nil {
		m.historyTable.Age()
	}

	return m.runIterativeDeepening(ctx, b, player, config, startTime)
}

// runIterativeDeepening runs the core iterative deepening search
func (m *MinimaxEngine) runIterativeDeepening(ctx context.Context, b *board.Board, player moves.Player, config ai.SearchConfig, startTime time.Time) ai.SearchResult {

	pseudoMoves := m.generator.GeneratePseudoLegalMoves(b, player)
	defer moves.ReleaseMoveList(pseudoMoves)

	if pseudoMoves.Count == 0 {
		isCheck := m.generator.IsKingInCheck(b, player)
		if isCheck {
			return ai.SearchResult{
				BestMove: board.Move{},
				Score:    -ai.MateScore,
				Stats:    ai.SearchStats{},
			}
		}
		return ai.SearchResult{
			BestMove: board.Move{},
			Score:    ai.DrawScore,
			Stats:    ai.SearchStats{},
		}
	}

	var rootTTMove board.Move
	if m.transpositionTable != nil {
		hash := b.GetHash()
		if entry, found := m.transpositionTable.Probe(hash); found {
			rootTTMove = entry.GetMove()
		}
	}

	m.orderMoves(b, pseudoMoves, 0, rootTTMove)

	lastCompletedBestMove := pseudoMoves.Moves[0]
	lastCompletedScore := ai.EvaluationScore(0)
	lastCompletedDepth := 0
	var finalStats ai.SearchStats

	startingDepth := 1

	for currentDepth := startingDepth; currentDepth <= config.MaxDepth; currentDepth++ {
		m.searchState.searchCancelled = false

		select {
		case <-ctx.Done():
			finalStats.Time = time.Since(startTime)
			finalStats.Depth = lastCompletedDepth
			return ai.SearchResult{
				BestMove: lastCompletedBestMove,
				Score:    lastCompletedScore,
				Stats:    finalStats,
			}
		default:
		}

		if config.MaxTime > 0 && time.Since(startTime) >= config.MaxTime {
			break
		}

		bestScore := -ai.MateScore - 1
		var bestMove board.Move

		alpha := -ai.MateScore - 1
		beta := ai.MateScore + 1

		// Use aspiration windows for depths > 1
		useAspirationWindow := currentDepth > 1
		window := ai.EvaluationScore(50)

		if useAspirationWindow {
			alpha = lastCompletedScore - window
			beta = lastCompletedScore + window
		}

		// Keep trying with wider windows until we get a score within bounds
		for {
			tempBestScore := -ai.MateScore - 1
			tempBestMove := board.Move{}
			moveIndex := 0

			for _, move := range pseudoMoves.Moves[:pseudoMoves.Count] {
				if m.searchState.searchCancelled {
					break
				}

				undo, err := b.MakeMoveWithUndo(move)
				if err != nil {
					continue // Skip invalid move
				}

				// Check if king is in check after move (illegal)
				if m.generator.IsKingInCheck(b, player) {
					b.UnmakeMove(undo)
					continue // Skip illegal move
				}

				// Add position to repetition history
				m.addHistory(b.GetHash())

				var score ai.EvaluationScore
				var moveStats ai.SearchStats

				if moveIndex == 0 {
					score = -m.negamax(ctx, b, oppositePlayer(player), currentDepth-1, -beta, -alpha, currentDepth, config, &moveStats)
				} else {
					score = -m.negamax(ctx, b, oppositePlayer(player), currentDepth-1, -alpha-1, -alpha, currentDepth, config, &moveStats)

					if score > alpha && score < beta {
						score = -m.negamax(ctx, b, oppositePlayer(player), currentDepth-1, -beta, -alpha, currentDepth, config, &moveStats)
					}
				}

				b.UnmakeMove(undo)

				// Remove position from repetition history
				m.removeHistory()

				if score > tempBestScore {
					tempBestScore = score
					tempBestMove = move

					if score > alpha {
						alpha = score
					}
				}

				moveIndex++

				if config.MaxTime > 0 && time.Since(startTime) >= config.MaxTime {
					break
				}
			}

			// Check if we need to re-search with wider window
			if !useAspirationWindow || (tempBestScore > lastCompletedScore-window && tempBestScore < lastCompletedScore+window) {
				// Score is within aspiration window or we're not using aspiration
				bestScore = tempBestScore
				bestMove = tempBestMove
				break
			}

			// Aspiration window failed - widen and retry
			if tempBestScore <= lastCompletedScore-window {
				// Fail low - score is worse than expected
				window *= 2
				alpha = lastCompletedScore - window
				// Keep beta the same to avoid re-searching moves that already failed high
			} else if tempBestScore >= lastCompletedScore+window {
				// Fail high - score is better than expected
				window *= 2
				beta = lastCompletedScore + window
				// Keep alpha at the current best score found
			}

			// Safety: if window gets too large, disable aspiration
			if window > 1000 {
				useAspirationWindow = false
				alpha = -ai.MateScore - 1
				beta = ai.MateScore + 1
			}
		}

		finalStats.NodesSearched = m.searchState.searchStats.NodesSearched
		finalStats.Depth = currentDepth
		finalStats.LMRReductions = m.searchState.searchStats.LMRReductions
		finalStats.LMRReSearches = m.searchState.searchStats.LMRReSearches
		finalStats.LMRNodesSkipped = m.searchState.searchStats.LMRNodesSkipped
		finalStats.NullMoves = m.searchState.searchStats.NullMoves
		finalStats.NullCutoffs = m.searchState.searchStats.NullCutoffs
		finalStats.QNodes = m.searchState.searchStats.QNodes
		finalStats.TTCutoffs = m.searchState.searchStats.TTCutoffs
		finalStats.FirstMoveCutoffs = m.searchState.searchStats.FirstMoveCutoffs
		finalStats.TotalCutoffs = m.searchState.searchStats.TotalCutoffs
		finalStats.DeltaPruned = m.searchState.searchStats.DeltaPruned
		finalStats.RazoringAttempts = m.searchState.searchStats.RazoringAttempts
		finalStats.RazoringCutoffs = m.searchState.searchStats.RazoringCutoffs
		finalStats.RazoringFailed = m.searchState.searchStats.RazoringFailed

		if !m.searchState.searchCancelled && (config.MaxTime == 0 || time.Since(startTime) < config.MaxTime) {
			lastCompletedBestMove = bestMove
			lastCompletedScore = bestScore
			lastCompletedDepth = currentDepth
		} else if m.searchState.searchCancelled {
			break
		}

		// Only break on mate if we found the shortest possible mate (mate in 1)
		// This allows us to search deeper for accurate mate distance calculation
		if bestScore >= ai.MateScore-1 || bestScore <= -ai.MateScore+1 {
			break
		}
	}

	finalStats.Time = time.Since(startTime)
	finalStats.Depth = lastCompletedDepth

	return ai.SearchResult{
		BestMove: lastCompletedBestMove,
		Score:    lastCompletedScore,
		Stats:    finalStats,
	}
}

// negamax performs negamax search with alpha-beta pruning and optimizations
func (m *MinimaxEngine) negamax(ctx context.Context, b *board.Board, player moves.Player, depth int, alpha, beta ai.EvaluationScore, originalMaxDepth int, config ai.SearchConfig, stats *ai.SearchStats) ai.EvaluationScore {
	m.searchState.searchStats.NodesSearched++

	currentDepth := originalMaxDepth - depth
	if currentDepth > stats.Depth {
		stats.Depth = currentDepth
	}

	select {
	case <-ctx.Done():
		m.searchState.searchCancelled = true
		return alpha
	default:
	}

	inCheck := m.generator.IsKingInCheck(b, player)
	if inCheck && depth < originalMaxDepth {
		depth++
	}

	var ttMove board.Move
	hash := b.GetHash()

	// Check for draw by repetition
	if m.isDrawByRepetition(hash) {
		return ai.DrawScore
	}

	if m.transpositionTable != nil {
		if entry, found := m.transpositionTable.Probe(hash); found {
			ttMove = entry.GetMove()

			if ttMove.From.File >= 0 && ttMove.From.File <= 7 &&
				ttMove.To.File >= 0 && ttMove.To.File <= 7 {
				if ttMove.From == ttMove.To {
					ttMove = board.Move{}
				}
			} else {
				ttMove = board.Move{}
			}

			if entry.GetDepth() >= depth {
				switch entry.GetType() {
				case EntryExact:
					m.searchState.searchStats.TTCutoffs++
					return entry.Score
				case EntryLowerBound:
					if entry.Score >= beta {
						m.searchState.searchStats.TTCutoffs++
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

	staticEval := m.evaluator.Evaluate(b)
	if depth >= 3 &&
		staticEval >= beta &&
		beta < ai.MateScore-MateDistanceThreshold &&
		beta > -ai.MateScore+MateDistanceThreshold {
		if !inCheck {
			m.searchState.searchStats.NullMoves++

			nullReduction := m.searchState.searchParams.NullMoveReduction
			if depth >= 6 && nullReduction < 3 {
				nullReduction++
			}

			nullUndo := b.MakeNullMove()

			nullScore := -m.negamax(ctx, b, oppositePlayer(player),
				depth-1-nullReduction, -beta, -beta+1, originalMaxDepth, config, stats)

			b.UnmakeNullMove(nullUndo)

			if nullScore >= beta {
				if nullScore < ai.MateScore-MateDistanceThreshold {
					m.searchState.searchStats.NullCutoffs++
					return beta
				}
			}
		}
	}

	// Simplest approach - razor based on static eval only
	if m.searchState.searchParams.RazoringEnabled &&
		!inCheck &&
		depth <= m.searchState.searchParams.RazoringMaxDepth &&
		depth > 0 {

		razoringMargin := m.searchState.searchParams.RazoringMargins[depth]

		if staticEval+razoringMargin < alpha {
			// Don't check for TT move at all
			m.searchState.searchStats.RazoringAttempts++

			qScore := m.quiescence(ctx, b, player, alpha, beta,
				originalMaxDepth-depth, stats)

			if qScore <= alpha {
				m.searchState.searchStats.RazoringCutoffs++
				return qScore
			}
			m.searchState.searchStats.RazoringFailed++
		}
	}

	if depth <= 0 {
		return m.quiescence(ctx, b, player, alpha, beta, originalMaxDepth-depth, stats)
	}

	pseudoMoves := m.generator.GeneratePseudoLegalMoves(b, player)
	defer moves.ReleaseMoveList(pseudoMoves)

	if pseudoMoves.Count == 0 {
		return m.handleNoLegalMoves(b, player, depth, originalMaxDepth, hash)
	}

	m.orderMoves(b, pseudoMoves, currentDepth, ttMove)

	bestScore := -ai.MateScore - 1
	bestMove := board.Move{}
	legalMoveCount := 0

	// Track if we improved alpha to determine correct entry type
	alphaImproved := false

	for i := 0; i < pseudoMoves.Count; i++ {
		move := pseudoMoves.Moves[i]

		if m.searchState.searchCancelled {
			break
		}

		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			continue // Skip invalid move
		}

		// Check if king is in check after move (illegal)
		if m.generator.IsKingInCheck(b, player) {
			b.UnmakeMove(undo)
			continue // Skip illegal move
		}

		// Add position to repetition history
		m.addHistory(b.GetHash())

		legalMoveCount++
		var score ai.EvaluationScore

		reduction := 0

		if depth >= config.LMRMinDepth &&
			legalMoveCount > config.LMRMinMoves &&
			!inCheck &&
			!move.IsCapture &&
			move.Promotion == board.Empty &&
			!m.isKillerMove(move, currentDepth) {

			givesCheck := board.MoveGivesCheck(b, move)

			if !givesCheck {
				reduction = int(math.Log(float64(min(depth, 15))) * math.Log(float64(min(legalMoveCount, 63))) / m.searchState.searchParams.LMRDivisor)

				historyScore := m.getHistoryScore(move)
				if historyScore > m.searchState.searchParams.HistoryHighThreshold {
					reduction = 0
				} else if historyScore > m.searchState.searchParams.HistoryMedThreshold && reduction > 0 {
					newReduction := reduction * 2 / 3
					if newReduction >= 0 {
						reduction = newReduction
					}
				} else if historyScore < m.searchState.searchParams.HistoryLowThreshold && reduction >= 0 {
					newReduction := reduction * 4 / 3
					if newReduction < depth {
						reduction = newReduction
					} else {
						reduction = depth - 1
					}
				}

				if reduction >= depth {
					reduction = depth - 1
				}
				if reduction < 0 {
					reduction = 0
				}
			}
		}

		if reduction > 0 {
			m.searchState.searchStats.LMRReductions++

			score = -m.negamax(ctx, b, oppositePlayer(player),
				depth-1-reduction, -alpha-1, -alpha, originalMaxDepth, config, stats)

			if score > alpha {
				m.searchState.searchStats.LMRReSearches++

				score = -m.negamax(ctx, b, oppositePlayer(player),
					depth-1, -beta, -alpha, originalMaxDepth, config, stats)
			}
		} else {
			score = -m.negamax(ctx, b, oppositePlayer(player),
				depth-1, -beta, -alpha, originalMaxDepth, config, stats)
		}

		b.UnmakeMove(undo)

		// Remove position from repetition history
		m.removeHistory()

		if score > bestScore {
			bestScore = score
			bestMove = move
		}

		if score > alpha {
			alpha = score
			alphaImproved = true

			if alpha >= beta {
				// Track move ordering statistics
				m.searchState.searchStats.TotalCutoffs++
				if legalMoveCount == 1 {
					m.searchState.searchStats.FirstMoveCutoffs++
				}

				if !move.IsCapture && currentDepth >= 0 && currentDepth < MaxKillerDepth {
					m.storeKiller(move, currentDepth)
				}

				if !move.IsCapture {
					if m.historyTable != nil {
						m.historyTable.UpdateHistory(move, depth)
					}
				}

				if m.transpositionTable != nil && !m.searchState.searchCancelled {
					m.transpositionTable.Store(hash, depth, bestScore, EntryLowerBound, move)
				}

				return beta
			}
		}
	}

	// Check if we had no legal moves
	if legalMoveCount == 0 {
		return m.handleNoLegalMoves(b, player, depth, originalMaxDepth, hash)
	}

	if legalMoveCount == 0 && m.searchState.searchCancelled {
		return alpha
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

		m.transpositionTable.Store(hash, depth, bestScore, entryType, bestMove)
	}

	return bestScore
}

// handleNoLegalMoves returns the appropriate score when no legal moves are available
// Returns checkmate score if in check, stalemate score otherwise
func (m *MinimaxEngine) handleNoLegalMoves(b *board.Board, player moves.Player, depth, originalMaxDepth int, hash uint64) ai.EvaluationScore {
	if m.generator.IsKingInCheck(b, player) {
		// Checkmate - the mate distance is how many plies from the root we are
		// Negative score because it's mate AGAINST the current player
		pliesFromRoot := originalMaxDepth - depth
		score := -ai.MateScore + ai.EvaluationScore(pliesFromRoot)
		if m.transpositionTable != nil {
			m.transpositionTable.Store(hash, depth, score, EntryExact, board.Move{})
		}
		return score
	}
	// Stalemate
	if m.transpositionTable != nil {
		m.transpositionTable.Store(hash, depth, ai.DrawScore, EntryExact, board.Move{})
	}
	return ai.DrawScore
}

// quiescence performs quiescence search
func (m *MinimaxEngine) quiescence(ctx context.Context, b *board.Board, player moves.Player, alpha, beta ai.EvaluationScore, depthFromRoot int, stats *ai.SearchStats) ai.EvaluationScore {
	m.searchState.searchStats.NodesSearched++
	m.searchState.searchStats.QNodes++

	select {
	case <-ctx.Done():
		m.searchState.searchCancelled = true
		return alpha
	default:
	}

	hash := b.GetHash()
	if m.transpositionTable != nil {
		if entry, found := m.transpositionTable.Probe(hash); found {
			if entry.GetDepth() >= 0 {
				switch entry.GetType() {
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

	eval := m.evaluator.Evaluate(b)
	if player == moves.Black {
		eval = -eval
	}

	if !inCheck {
		if eval >= beta {
			return beta
		}
		if eval > alpha {
			alpha = eval
		}
	}

	// Generate moves based on whether we're in check
	var movesToSearch *moves.MoveList

	if inCheck {
		// When in check, we must consider ALL legal moves (including quiet escapes)
		// This is the idiomatic approach used by strong chess engines
		allMoves := m.generator.GeneratePseudoLegalMoves(b, player)
		movesToSearch = allMoves

	} else {
		// Normal quiescence - only captures and promotions
		allMoves := m.generator.GeneratePseudoLegalMoves(b, player)
		defer moves.ReleaseMoveList(allMoves)

		captureList := moves.GetMoveList()
		for i := 0; i < allMoves.Count; i++ {
			move := allMoves.Moves[i]
			if move.IsCapture || move.Promotion != board.Empty {
				captureList.AddMove(move)
			}
		}
		movesToSearch = captureList
	}

	defer moves.ReleaseMoveList(movesToSearch)

	// Order moves appropriately
	if inCheck {
		m.orderMoves(b, movesToSearch, 0, board.Move{}) // Order all moves when in check
	} else {
		m.orderCaptures(b, movesToSearch) // Order captures in normal quiescence
	}

	// Check if we have any legal moves when in check
	legalMoveCount := 0
	bestScore := eval

	for i := 0; i < movesToSearch.Count; i++ {
		move := movesToSearch.Moves[i]

		if m.searchState.searchCancelled {
			break
		}

		// Try the move to see if it's legal
		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			continue // Skip invalid move
		}

		// Check if king is in check after move (illegal)
		if m.generator.IsKingInCheck(b, player) {
			b.UnmakeMove(undo)
			continue // Skip illegal move
		}

		legalMoveCount++

		// Apply pruning only when not in check and only for captures
		if !inCheck && move.IsCapture {
			// Delta pruning - skip captures that can't improve alpha significantly
			captureValue := ai.EvaluationScore(0)
			switch move.Captured {
			case board.WhitePawn, board.BlackPawn:
				captureValue = 100
			case board.WhiteKnight, board.BlackKnight, board.WhiteBishop, board.BlackBishop:
				captureValue = 300
			case board.WhiteRook, board.BlackRook:
				captureValue = 500
			case board.WhiteQueen, board.BlackQueen:
				captureValue = 900
			}

			if move.Promotion != board.Empty {
				captureValue += 800
			}

			margin := ai.EvaluationScore(200)
			if eval+captureValue+margin < alpha {
				m.searchState.searchStats.DeltaPruned++
				b.UnmakeMove(undo)
				continue
			}

			// SEE pruning - skip bad captures
			if seeScore := m.seeCalculator.SEE(b, move); seeScore < -100 {
				b.UnmakeMove(undo)
				continue
			}
		}

		// Move is already made and verified as legal above
		score := -m.quiescence(ctx, b, oppositePlayer(player), -beta, -alpha, depthFromRoot+1, stats)
		b.UnmakeMove(undo)

		if score > bestScore {
			bestScore = score
		}

		if score > alpha {
			alpha = score
			if alpha >= beta {
				break
			}
		}
	}

	// Handle the case where we have no legal moves
	if inCheck && legalMoveCount == 0 {
		// Checkmate - no legal moves when in check
		// The mate distance should be how many plies from the original search root
		// In quiescence, depthFromRoot represents the distance from the search root
		return -ai.MateScore + ai.EvaluationScore(depthFromRoot)
	}

	// If not in check and we had no moves to search (no captures), return static eval
	if !inCheck && legalMoveCount == 0 {
		return eval
	}

	if m.transpositionTable != nil && !m.searchState.searchCancelled {
		var entryType EntryType
		if bestScore <= originalAlpha {
			entryType = EntryUpperBound
		} else if bestScore >= beta {
			entryType = EntryLowerBound
		} else {
			entryType = EntryExact
		}
		m.transpositionTable.Store(hash, 0, bestScore, entryType, board.Move{})
	}

	return bestScore
}

// orderMoves orders moves for search
func (m *MinimaxEngine) orderMoves(b *board.Board, moveList *moves.MoveList, depth int, ttMove board.Move) {
	if moveList.Count <= 1 {
		return
	}

	if cap(m.searchState.moveOrderBuffer) < moveList.Count {
		m.searchState.moveOrderBuffer = make([]moveScore, moveList.Count)
	} else {
		m.searchState.moveOrderBuffer = m.searchState.moveOrderBuffer[:moveList.Count]
	}

	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		score := 0

		if move.From == ttMove.From && move.To == ttMove.To && move.Promotion == ttMove.Promotion {
			score = 3000000
		} else {
			if move.IsCapture {
				score = m.getCaptureScore(b, move)
			}

			if move.Promotion != board.Empty {
				switch move.Promotion {
				case board.WhiteQueen, board.BlackQueen:
					score += 9000
				case board.WhiteRook, board.BlackRook:
					score += 5000
				case board.WhiteBishop, board.BlackBishop, board.WhiteKnight, board.BlackKnight:
					score += 3000
				}
			}

			if !move.IsCapture && m.isKillerMove(move, depth) {
				score = 500000
			}

			if !move.IsCapture && move.Promotion == board.Empty {
				score += int(m.getHistoryScore(move))
			}
		}

		m.searchState.moveOrderBuffer[i] = moveScore{index: i, score: score}
	}

	// Insertion sort - O(n²) worst case but O(n) best case, better cache locality
	for i := 1; i < moveList.Count; i++ {
		key := m.searchState.moveOrderBuffer[i]
		j := i - 1
		for j >= 0 && m.searchState.moveOrderBuffer[j].score < key.score {
			m.searchState.moveOrderBuffer[j+1] = m.searchState.moveOrderBuffer[j]
			j--
		}
		m.searchState.moveOrderBuffer[j+1] = key
	}

	if cap(m.searchState.reorderBuffer) < moveList.Count {
		m.searchState.reorderBuffer = make([]board.Move, moveList.Count)
	} else {
		m.searchState.reorderBuffer = m.searchState.reorderBuffer[:moveList.Count]
	}

	for i := 0; i < moveList.Count; i++ {
		origIndex := m.searchState.moveOrderBuffer[i].index
		m.searchState.reorderBuffer[i] = moveList.Moves[origIndex]
	}

	copy(moveList.Moves[:moveList.Count], m.searchState.reorderBuffer)

	if m.debugMoveOrdering {
		if cap(m.searchState.debugMoveOrder) < moveList.Count {
			m.searchState.debugMoveOrder = make([]board.Move, moveList.Count)
		} else {
			m.searchState.debugMoveOrder = m.searchState.debugMoveOrder[:moveList.Count]
		}
		copy(m.searchState.debugMoveOrder, moveList.Moves[:moveList.Count])
	}
}

// isKillerMove checks if move is a killer move
func (m *MinimaxEngine) isKillerMove(move board.Move, depth int) bool {
	if depth < 0 || depth >= MaxKillerDepth {
		return false
	}

	return (move.From == m.searchState.killerTable[depth][0].From && move.To == m.searchState.killerTable[depth][0].To) ||
		(move.From == m.searchState.killerTable[depth][1].From && move.To == m.searchState.killerTable[depth][1].To)
}

// storeKiller stores a killer move
func (m *MinimaxEngine) storeKiller(move board.Move, depth int) {
	if depth < 0 || depth >= MaxKillerDepth {
		return
	}

	if m.isKillerMove(move, depth) {
		return
	}

	m.searchState.killerTable[depth][1] = m.searchState.killerTable[depth][0]
	m.searchState.killerTable[depth][0] = move
}

// orderCaptures orders captures using SEE scores
func (m *MinimaxEngine) orderCaptures(b *board.Board, moveList *moves.MoveList) {
	if moveList.Count <= 1 {
		return
	}

	if cap(m.searchState.moveOrderBuffer) < moveList.Count {
		m.searchState.moveOrderBuffer = make([]moveScore, moveList.Count)
	} else {
		m.searchState.moveOrderBuffer = m.searchState.moveOrderBuffer[:moveList.Count]
	}

	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		score := m.seeCalculator.SEE(b, move)
		m.searchState.moveOrderBuffer[i] = moveScore{index: i, score: score}
	}

	// Insertion sort - O(n²) worst case but O(n) best case, better cache locality
	for i := 1; i < moveList.Count; i++ {
		key := m.searchState.moveOrderBuffer[i]
		j := i - 1
		for j >= 0 && m.searchState.moveOrderBuffer[j].score < key.score {
			m.searchState.moveOrderBuffer[j+1] = m.searchState.moveOrderBuffer[j]
			j--
		}
		m.searchState.moveOrderBuffer[j+1] = key
	}

	if cap(m.searchState.reorderBuffer) < moveList.Count {
		m.searchState.reorderBuffer = make([]board.Move, moveList.Count)
	} else {
		m.searchState.reorderBuffer = m.searchState.reorderBuffer[:moveList.Count]
	}

	for i := 0; i < moveList.Count; i++ {
		origIndex := m.searchState.moveOrderBuffer[i].index
		m.searchState.reorderBuffer[i] = moveList.Moves[origIndex]
	}

	copy(moveList.Moves[:moveList.Count], m.searchState.reorderBuffer)
}

// getCaptureScore calculates the capture score using SEE for accurate evaluation
// Higher scores indicate more valuable captures (better moves to try first)
// Move ordering priorities:
//  1. TT moves: 3,000,000+
//  2. Good captures (SEE > 0): 1,000,000+
//  3. Equal exchanges (SEE = 0): 900,000
//  4. Killer moves: 500,000
//  5. Good history moves: up to ~50,000
//  6. Slightly bad captures (SEE >= -100): 100,000+
//  7. Terrible captures (SEE < -100): 50,000+
//  8. Quiet moves: 0
func (m *MinimaxEngine) getCaptureScore(b *board.Board, move board.Move) int {
	if !move.IsCapture || move.Captured == board.Empty {
		return 0
	}

	seeValue := m.seeCalculator.SEE(b, move)

	victimValue := evaluation.PieceValues[move.Captured]
	if victimValue < 0 {
		victimValue = -victimValue
	}

	attackerValue := evaluation.PieceValues[move.Piece]
	if attackerValue < 0 {
		attackerValue = -attackerValue
	}

	mvvLvaScore := (victimValue * 10) - attackerValue

	if seeValue > 0 {
		return 1000000 + seeValue + mvvLvaScore
	} else if seeValue == 0 {
		return 900000 + mvvLvaScore
	} else if seeValue >= -100 {
		return 50000 + seeValue + 100 + mvvLvaScore
	}
	return 25000 + seeValue + 1000 + mvvLvaScore
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

// GetTranspositionTableStats returns transposition table statistics if available
func (m *MinimaxEngine) GetTranspositionTableStats() (hits, misses, collisions uint64, hitRate float64) {
	if m.transpositionTable != nil {
		return m.transpositionTable.GetStats()
	}
	return 0, 0, 0, 0
}

// GetDetailedTranspositionTableStats returns detailed TT statistics including fill rate and average depth
func (m *MinimaxEngine) GetDetailedTranspositionTableStats() (hits, misses, collisions, filled, averageDepth uint64, hitRate, fillRate float64) {
	if m.transpositionTable != nil {
		return m.transpositionTable.GetDetailedStats()
	}
	return 0, 0, 0, 0, 0, 0, 0
}

// GetTwoBucketStats returns statistics specific to the two-bucket collision resolution
func (m *MinimaxEngine) GetTwoBucketStats() (secondBucketUse uint64, secondBucketRate float64) {
	if m.transpositionTable != nil {
		return m.transpositionTable.GetTwoBucketStats()
	}
	return 0, 0
}

// GetName returns the engine name
func (m *MinimaxEngine) GetName() string {
	return "Minimax Engine"
}

// ClearSearchState clears transient search state between different positions
func (m *MinimaxEngine) ClearSearchState() {
	for i := 0; i < MaxKillerDepth; i++ {
		m.searchState.killerTable[i][0] = board.Move{}
		m.searchState.killerTable[i][1] = board.Move{}
	}
	m.searchState.searchStats = ai.SearchStats{}
	m.searchState.moveOrderBuffer = make([]moveScore, 0, 256)
	m.searchState.searchCancelled = false

	// Clear repetition history
	m.zobristHistoryPly = 0

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
		m.searchState.debugMoveOrder = nil
	}
}

// GetLastMoveOrder returns the move order from the last orderMoves call (for tests only)
func (m *MinimaxEngine) GetLastMoveOrder() []board.Move {
	return m.searchState.debugMoveOrder
}

// getHistoryScore returns the history score for a move
func (m *MinimaxEngine) getHistoryScore(move board.Move) int32 {
	if m.historyTable == nil {
		return 0
	}
	return m.historyTable.GetHistoryScore(move)
}

// setupRepetitionHistory initializes repetition detection with the root position hash.
// This should be called at the start of each search to establish the baseline for
// repetition detection during the current search tree exploration.
func (m *MinimaxEngine) setupRepetitionHistory(rootHash uint64) {
	m.zobristHistoryPly = 0
	m.zobristHistory[m.zobristHistoryPly] = rootHash
}

// addHistory adds a position hash to the repetition detection history.
// Called when making a move during search to track the path from root to current node.
// Prevents buffer overflow by checking against MaxGamePly capacity.
func (m *MinimaxEngine) addHistory(hash uint64) {
	if m.zobristHistoryPly < MaxGamePly-1 {
		m.zobristHistoryPly++
		m.zobristHistory[m.zobristHistoryPly] = hash
	}
}

// removeHistory removes the latest hash from repetition detection history.
// Called when unmaking a move during search to maintain correct history state.
// Prevents underflow by checking that history is not empty.
func (m *MinimaxEngine) removeHistory() {
	if m.zobristHistoryPly > 0 {
		m.zobristHistoryPly--
	}
}

// isDrawByRepetition checks if the current position hash appears in the search history.
// Returns true if the position repeats any position from the current search path,
// indicating a draw by repetition according to chess rules. Only checks positions
// in the current search tree, not the full game history.
func (m *MinimaxEngine) isDrawByRepetition(currentHash uint64) bool {
	for repPly := uint16(0); repPly < m.zobristHistoryPly; repPly++ {
		if m.zobristHistory[repPly] == currentHash {
			return true
		}
	}
	return false
}
