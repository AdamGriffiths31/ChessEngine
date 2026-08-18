package epd

import (
	"context"
	"log/slog"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/search"
)

// STSResult represents the result of testing a single EPD position
type STSResult struct {
	Position      *Position
	EngineMove    board.Move
	EngineMoveStr string
	Score         int // Points scored (0-10)
	SearchResult  search.SearchResult
	TestDuration  time.Duration
}

// STSSuiteResult represents results from a complete test suite
type STSSuiteResult struct {
	SuiteName     string
	Results       []STSResult
	TotalScore    int
	MaxScore      int
	ScorePercent  float64
	TotalTime     time.Duration
	PositionCount int
}

// STSScorer handles scoring EPD positions using an AI engine
type STSScorer struct {
	engine  search.Engine
	config  search.SearchConfig
	verbose bool
	clearTT bool
}

// NewSTSScorer creates a new STS scorer with the given engine and search configuration
func NewSTSScorer(engine search.Engine, config search.SearchConfig, verbose bool) *STSScorer {
	return &STSScorer{
		engine:  engine,
		config:  config,
		verbose: verbose,
		clearTT: false,
	}
}

// NewSTSScorerWithTTClear creates a new STS scorer with TT clearing option
func NewSTSScorerWithTTClear(engine search.Engine, config search.SearchConfig, verbose bool, clearTT bool) *STSScorer {
	return &STSScorer{
		engine:  engine,
		config:  config,
		verbose: verbose,
		clearTT: clearTT,
	}
}

// ScorePosition tests a single EPD position and returns the result
func (scorer *STSScorer) ScorePosition(ctx context.Context, position *Position) STSResult {
	startTime := time.Now()

	searchCtx := ctx
	if scorer.config.MaxTime > 0 {
		var cancel context.CancelFunc
		searchCtx, cancel = context.WithTimeout(ctx, scorer.config.MaxTime)
		defer cancel()
	}

	var player movegen.Player
	if position.Board.GetSideToMove() == "w" {
		player = movegen.White
	} else {
		player = movegen.Black
	}

	searchResult := scorer.engine.FindBestMove(searchCtx, position.Board, player, scorer.config)

	duration := time.Since(startTime)

	engineMoveStr := moveToString(searchResult.BestMove)

	algebraicMove := scorer.moveToAlgebraic(position.Board, searchResult.BestMove)

	score := scorer.calculateScore(position, searchResult.BestMove, engineMoveStr, algebraicMove)

	return STSResult{
		Position:      position,
		EngineMove:    searchResult.BestMove,
		EngineMoveStr: engineMoveStr + "/" + algebraicMove, // Show both notations
		Score:         score,
		SearchResult:  searchResult,
		TestDuration:  duration,
	}
}

func (scorer *STSScorer) calculateScore(position *Position, engineMove board.Move, engineMoveStr, algebraicMove string) int {
	// If we have detailed move scores from STS format, use them
	if len(position.MoveScores) > 0 {
		return scorer.calculateSTSScore(position, engineMove, engineMoveStr, algebraicMove)
	}

	// Fallback to simple scoring for basic EPD format
	return scorer.calculateSimpleScore(position, engineMove, engineMoveStr, algebraicMove)
}

func (scorer *STSScorer) calculateSTSScore(position *Position, engineMove board.Move, engineMoveStr, algebraicMove string) int {
	// Check each scored move using multiple comparison methods
	for _, moveScore := range position.MoveScores {
		if engineMoveStr == moveScore.Move ||
			algebraicMove == moveScore.Move ||
			scorer.moveMatches(engineMove, moveScore.Move) {
			return moveScore.Points
		}
	}

	if position.AvoidMove != "" {
		if engineMoveStr == position.AvoidMove ||
			algebraicMove == position.AvoidMove ||
			scorer.moveMatches(engineMove, position.AvoidMove) {
			return 0 // No points for avoided move
		}
	}

	return 1 // Minimal points for any legal move not explicitly scored
}

func (scorer *STSScorer) calculateSimpleScore(position *Position, engineMove board.Move, engineMoveStr, algebraicMove string) int {
	if position.BestMove != "" {
		// Try multiple comparison methods
		if engineMoveStr == position.BestMove ||
			algebraicMove == position.BestMove ||
			scorer.moveMatches(engineMove, position.BestMove) {
			return 10 // Full points for match
		}
	}

	if position.AvoidMove != "" {
		if engineMoveStr == position.AvoidMove ||
			algebraicMove == position.AvoidMove ||
			scorer.moveMatches(engineMove, position.AvoidMove) {
			return 0 // No points for avoided move
		}
	}

	return 1 // Minimal points for any legal move that's not avoided
}

func moveToString(move board.Move) string {
	return move.From.String() + move.To.String()
}

func (scorer *STSScorer) moveMatches(engineMove board.Move, expectedMove string) bool {
	// Handle castling notation variations
	if expectedMove == "e1g1" || expectedMove == "0-0" || expectedMove == "O-O" {
		return engineMove.From.File == 4 && engineMove.From.Rank == 0 &&
			engineMove.To.File == 6 && engineMove.To.Rank == 0
	}
	if expectedMove == "e1c1" || expectedMove == "0-0-0" || expectedMove == "O-O-O" {
		return engineMove.From.File == 4 && engineMove.From.Rank == 0 &&
			engineMove.To.File == 2 && engineMove.To.Rank == 0
	}
	if expectedMove == "e8g8" || expectedMove == "o-o" {
		return engineMove.From.File == 4 && engineMove.From.Rank == 7 &&
			engineMove.To.File == 6 && engineMove.To.Rank == 7
	}
	if expectedMove == "e8c8" || expectedMove == "o-o-o" {
		return engineMove.From.File == 4 && engineMove.From.Rank == 7 &&
			engineMove.To.File == 2 && engineMove.To.Rank == 7
	}

	// Try to parse algebraic moves like "Rad1", "Nf3", "exd5"
	return scorer.parseAndMatchAlgebraic(engineMove, expectedMove)
}

func (scorer *STSScorer) parseAndMatchAlgebraic(engineMove board.Move, expectedMove string) bool {
	if len(expectedMove) < 2 {
		return false
	}

	// For moves like "Rad1", "Nf3", "exd5", the destination is the last 2 chars
	if len(expectedMove) >= 2 {
		destSquareStr := expectedMove[len(expectedMove)-2:]

		if len(destSquareStr) == 2 &&
			destSquareStr[0] >= 'a' && destSquareStr[0] <= 'h' &&
			destSquareStr[1] >= '1' && destSquareStr[1] <= '8' {

			expectedFile := int(destSquareStr[0] - 'a')
			expectedRank := int(destSquareStr[1] - '1')

			if engineMove.To.File == expectedFile && engineMove.To.Rank == expectedRank {

				// If it's a piece move (starts with capital letter), check piece type matches
				if len(expectedMove) > 2 {
					firstChar := expectedMove[0]
					if firstChar >= 'A' && firstChar <= 'Z' {
						// Get engine piece type (convert to uppercase)
						enginePiece := engineMove.Piece
						if enginePiece >= 'a' && enginePiece <= 'z' {
							enginePiece = enginePiece - 'a' + 'A'
						}

						return rune(firstChar) == rune(enginePiece)
					}
				}

				// For pawn moves or exact destination matches, accept it
				return true
			}
		}
	}

	return false
}

// ScoreSuite runs a complete test suite and returns aggregated results.
// If onResult is non-nil, it is invoked synchronously after each position is
// scored (e.g. so the caller can stream a JSONL record per position); it may
// be nil if the caller has no per-position side effect to perform.
func (scorer *STSScorer) ScoreSuite(ctx context.Context, positions []*Position, suiteName string, onResult func(STSResult)) STSSuiteResult {
	results := make([]STSResult, 0, len(positions))
	totalScore := 0
	maxScore := len(positions) * 10 // Maximum possible score
	totalTime := time.Duration(0)

	for i, position := range positions {
		if scorer.clearTT {
			if minimaxEngine, ok := scorer.engine.(*search.MinimaxEngine); ok {
				minimaxEngine.ClearSearchState()
			}
		}

		result := scorer.ScorePosition(ctx, position)
		results = append(results, result)
		totalScore += result.Score
		totalTime += result.TestDuration

		if onResult != nil {
			onResult(result)
		}

		if scorer.verbose {
			scorer.printPositionResult(i+1, result)
		}
	}

	scorePercent := float64(totalScore) / float64(maxScore) * 100.0

	return STSSuiteResult{
		SuiteName:     suiteName,
		Results:       results,
		TotalScore:    totalScore,
		MaxScore:      maxScore,
		ScorePercent:  scorePercent,
		TotalTime:     totalTime,
		PositionCount: len(positions),
	}
}

// printPositionResult logs the result of a single position immediately
// (verbose mode).
func (scorer *STSScorer) printPositionResult(posNum int, result STSResult) {
	display := ""
	if result.Position.ID != "" {
		display = result.Position.ID
	}
	if result.Position.Comment != "" {
		if display != "" {
			display += " | " + result.Position.Comment
		} else {
			display = result.Position.Comment
		}
	}

	// Add FEN at the end for easy analysis
	if display != "" {
		display += " | FEN: " + result.Position.Board.ToFEN()
	} else {
		display = "FEN: " + result.Position.Board.ToFEN()
	}

	slog.Info("sts position result",
		"position", posNum,
		"bestMove", result.Position.BestMove,
		"engineMove", result.EngineMoveStr,
		"score", result.Score,
		"duration", result.TestDuration.Round(time.Millisecond),
		"detail", display)
}

func (scorer *STSScorer) moveToAlgebraic(board *board.Board, move board.Move) string {
	piece := board.GetPiece(move.From.Rank, move.From.File)

	if move.IsCastling {
		if move.To.File == 6 { // Kingside
			return "O-O"
		}
		// Queenside
		return "O-O-O"
	}

	pieceChar := ""
	switch piece {
	case 'K', 'k':
		pieceChar = "K"
	case 'Q', 'q':
		pieceChar = "Q"
	case 'R', 'r':
		pieceChar = "R"
	case 'B', 'b':
		pieceChar = "B"
	case 'N', 'n':
		pieceChar = "N"
	case 'P', 'p':
		pieceChar = "" // Pawns don't have piece letters
	}

	toSquare := move.To.String()

	isCapture := board.GetPiece(move.To.Rank, move.To.File) != '.' || move.IsEnPassant

	algebraic := pieceChar

	// For pawn captures, include the file
	if piece == 'P' || piece == 'p' {
		if isCapture {
			algebraic = string('a'+rune(move.From.File)) + "x"
		}
	} else if isCapture {
		algebraic += "x"
	}

	algebraic += toSquare

	if move.Promotion != '.' {
		switch move.Promotion {
		case 'Q', 'q':
			algebraic += "=Q"
		case 'R', 'r':
			algebraic += "=R"
		case 'B', 'b':
			algebraic += "=B"
		case 'N', 'n':
			algebraic += "=N"
		}
	}

	return algebraic
}
