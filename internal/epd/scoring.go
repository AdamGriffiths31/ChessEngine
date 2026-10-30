package epd

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
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
	Result        search.Result
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
	config  search.Config
	verbose bool
	clearTT bool
}

// NewSTSScorerWithTTClear creates a new STS scorer with TT clearing option
func NewSTSScorerWithTTClear(engine search.Engine, config search.Config, verbose bool, clearTT bool) *STSScorer {
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
		EngineMoveStr: engineMoveStr + "/" + algebraicMove,
		Score:         score,
		Result:        searchResult,
		TestDuration:  duration,
	}
}

func (scorer *STSScorer) calculateScore(position *Position, engineMove board.Move, engineMoveStr, algebraicMove string) int {
	if len(position.MoveScores) > 0 {
		return scorer.calculateSTSScore(position, engineMove, engineMoveStr, algebraicMove)
	}

	return scorer.calculateSimpleScore(position, engineMove, engineMoveStr, algebraicMove)
}

func (scorer *STSScorer) calculateSTSScore(position *Position, engineMove board.Move, engineMoveStr, algebraicMove string) int {
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
			return 0
		}
	}

	return 1
}

func (scorer *STSScorer) calculateSimpleScore(position *Position, engineMove board.Move, engineMoveStr, algebraicMove string) int {
	if position.BestMove != "" {
		if engineMoveStr == position.BestMove ||
			algebraicMove == position.BestMove ||
			scorer.moveMatches(engineMove, position.BestMove) {
			return 10
		}
	}

	if position.AvoidMove != "" {
		if engineMoveStr == position.AvoidMove ||
			algebraicMove == position.AvoidMove ||
			scorer.moveMatches(engineMove, position.AvoidMove) {
			return 0
		}
	}

	return 1
}

func moveToString(move board.Move) string {
	return move.From.String() + move.To.String()
}

var sanPattern = regexp.MustCompile(`^([KQRBN])?([a-h])?([1-8])?x?([a-h][1-8])(?:=?([QRBNqrbn]))?$`)

func (scorer *STSScorer) moveMatches(engineMove board.Move, expectedMove string) bool {
	expectedMove = strings.TrimRight(expectedMove, "+#!?")

	switch expectedMove {
	case "O-O", "0-0", "o-o":
		return isCastlingTo(engineMove, 6)
	case "O-O-O", "0-0-0", "o-o-o":
		return isCastlingTo(engineMove, 2)
	}

	return matchesSAN(engineMove, expectedMove)
}

func isCastlingTo(move board.Move, kingFile int) bool {
	return (move.Piece == board.WhiteKing || move.Piece == board.BlackKing) &&
		move.From.File == 4 && move.To.File == kingFile && move.From.Rank == move.To.Rank
}

func matchesSAN(move board.Move, san string) bool {
	m := sanPattern.FindStringSubmatch(san)
	if m == nil {
		return false
	}
	piece, fromFile, fromRank, to, promotion := m[1], m[2], m[3], m[4], m[5]

	if move.To.File != int(to[0]-'a') || move.To.Rank != int(to[1]-'1') {
		return false
	}
	if fromFile != "" && move.From.File != int(fromFile[0]-'a') {
		return false
	}
	if fromRank != "" && move.From.Rank != int(fromRank[0]-'1') {
		return false
	}

	isCoordinate := piece == "" && fromFile != "" && fromRank != ""
	if !isCoordinate {
		if piece == "" {
			piece = "P"
		}
		if strings.ToUpper(string(move.Piece)) != piece {
			return false
		}
	}

	if promotion == "" {
		return !move.HasPromotion()
	}
	return move.HasPromotion() && strings.EqualFold(string(move.Promotion), promotion)
}

// ScoreSuite runs a complete test suite and returns aggregated results.
// If onResult is non-nil, it is invoked synchronously after each position is
// scored (e.g. so the caller can stream a JSONL record per position); it may
// be nil if the caller has no per-position side effect to perform.
func (scorer *STSScorer) ScoreSuite(ctx context.Context, positions []*Position, suiteName string, onResult func(STSResult)) STSSuiteResult {
	results := make([]STSResult, 0, len(positions))
	totalScore := 0
	maxScore := len(positions) * 10
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
		if move.To.File == 6 {
			return "O-O"
		}
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
		pieceChar = ""
	}

	toSquare := move.To.String()

	isCapture := board.GetPiece(move.To.Rank, move.To.File) != '.' || move.IsEnPassant

	algebraic := pieceChar

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
