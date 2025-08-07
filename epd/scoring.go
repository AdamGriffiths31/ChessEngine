package epd

import (
	"context"
	"fmt"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// STSResult represents the result of testing a single EPD position
type STSResult struct {
	Position       *EPDPosition
	EngineMove     board.Move
	EngineMoveStr  string
	Score          int // Points scored (0-10)
	SearchResult   ai.SearchResult
	TestDuration   time.Duration
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
	engine  ai.Engine
	config  ai.SearchConfig
	verbose bool
}

// NewSTSScorer creates a new STS scorer with the given engine and search configuration
func NewSTSScorer(engine ai.Engine, config ai.SearchConfig, verbose bool) *STSScorer {
	return &STSScorer{
		engine:  engine,
		config:  config,
		verbose: verbose,
	}
}

// ScorePosition tests a single EPD position and returns the result
func (scorer *STSScorer) ScorePosition(ctx context.Context, position *EPDPosition) STSResult {
	startTime := time.Now()
	
	// Create timeout context for this position if MaxTime is configured
	searchCtx := ctx
	if scorer.config.MaxTime > 0 {
		var cancel context.CancelFunc
		searchCtx, cancel = context.WithTimeout(ctx, scorer.config.MaxTime)
		defer cancel()
	}
	
	// Find the best move using our engine
	searchResult := scorer.engine.FindBestMove(searchCtx, position.Board, moves.White, scorer.config)
	
	duration := time.Since(startTime)
	
	// Convert engine move to string for comparison
	engineMoveStr := moveToString(searchResult.BestMove)
	
	// Convert engine move to algebraic notation for better comparison
	algebraicMove := scorer.moveToAlgebraic(position.Board, searchResult.BestMove)
	
	// Calculate score based on whether engine found the best move
	score := scorer.calculateScore(position, searchResult.BestMove, engineMoveStr, algebraicMove)
	
	return STSResult{
		Position:       position,
		EngineMove:     searchResult.BestMove,
		EngineMoveStr:  engineMoveStr + "/" + algebraicMove, // Show both notations
		Score:          score,
		SearchResult:   searchResult,
		TestDuration:   duration,
	}
}

// calculateScore determines the points awarded based on engine's move choice
func (scorer *STSScorer) calculateScore(position *EPDPosition, engineMove board.Move, engineMoveStr, algebraicMove string) int {
	// If we have detailed move scores from STS format, use them
	if len(position.MoveScores) > 0 {
		return scorer.calculateSTSScore(position, engineMove, engineMoveStr, algebraicMove)
	}
	
	// Fallback to simple scoring for basic EPD format
	return scorer.calculateSimpleScore(position, engineMove, engineMoveStr, algebraicMove)
}

// calculateSTSScore uses the detailed STS scoring system
func (scorer *STSScorer) calculateSTSScore(position *EPDPosition, engineMove board.Move, engineMoveStr, algebraicMove string) int {
	// Check each scored move using multiple comparison methods
	for _, moveScore := range position.MoveScores {
		if engineMoveStr == moveScore.Move || 
		   algebraicMove == moveScore.Move ||
		   scorer.moveMatches(engineMove, moveScore.Move) {
			return moveScore.Points
		}
	}
	
	// Handle avoid move penalty
	if position.AvoidMove != "" {
		if engineMoveStr == position.AvoidMove || 
		   algebraicMove == position.AvoidMove ||
		   scorer.moveMatches(engineMove, position.AvoidMove) {
			return 0 // No points for avoided move
		}
	}
	
	// No match in the scored moves - give minimal points
	return 1 // Minimal points for any legal move not explicitly scored
}

// calculateSimpleScore uses basic best move scoring
func (scorer *STSScorer) calculateSimpleScore(position *EPDPosition, engineMove board.Move, engineMoveStr, algebraicMove string) int {
	// Handle best move comparison
	if position.BestMove != "" {
		// Try multiple comparison methods
		if engineMoveStr == position.BestMove || 
		   algebraicMove == position.BestMove ||
		   scorer.moveMatches(engineMove, position.BestMove) {
			return 10 // Full points for match
		}
	}
	
	// Handle avoid move penalty
	if position.AvoidMove != "" {
		if engineMoveStr == position.AvoidMove || 
		   algebraicMove == position.AvoidMove ||
		   scorer.moveMatches(engineMove, position.AvoidMove) {
			return 0 // No points for avoided move
		}
	}
	
	// If no exact match but not an avoided move, give partial credit
	return 1 // Minimal points for any legal move that's not avoided
}

// moveToString converts a Move to string notation
func moveToString(move board.Move) string {
	return move.From.String() + move.To.String()
}

// moveMatches checks if the engine move matches the expected move in various formats
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

// parseAndMatchAlgebraic tries to match algebraic notation moves
func (scorer *STSScorer) parseAndMatchAlgebraic(engineMove board.Move, expectedMove string) bool {
	if len(expectedMove) < 2 {
		return false
	}
	
	// Get destination square from the expected move
	// For moves like "Rad1", "Nf3", "exd5", the destination is the last 2 chars
	if len(expectedMove) >= 2 {
		destSquareStr := expectedMove[len(expectedMove)-2:]
		
		// Parse destination square
		if len(destSquareStr) == 2 && 
		   destSquareStr[0] >= 'a' && destSquareStr[0] <= 'h' &&
		   destSquareStr[1] >= '1' && destSquareStr[1] <= '8' {
			
			expectedFile := int(destSquareStr[0] - 'a')
			expectedRank := int(destSquareStr[1] - '1')
			
			// Check if engine move goes to the same destination
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
						
						// Check if piece types match
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

// ScoreSuite runs a complete test suite and returns aggregated results
func (scorer *STSScorer) ScoreSuite(ctx context.Context, positions []*EPDPosition, suiteName string) STSSuiteResult {
	results := make([]STSResult, 0, len(positions))
	totalScore := 0
	maxScore := len(positions) * 10 // Maximum possible score
	totalTime := time.Duration(0)
	
	for i, position := range positions {
		result := scorer.ScorePosition(ctx, position)
		results = append(results, result)
		totalScore += result.Score
		totalTime += result.TestDuration
		
		// Print result immediately after each position (if verbose mode)
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

// printPositionResult prints the result of a single position immediately
func (scorer *STSScorer) printPositionResult(posNum int, result STSResult) {
	// Show both comment and ID if available
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

	fmt.Printf("%-4d %-10s %-10s %-6d %-8v %s\n",
		posNum,
		result.Position.BestMove,
		result.EngineMoveStr,
		result.Score,
		result.TestDuration.Round(time.Millisecond),
		display)
}

// moveToAlgebraic converts a coordinate move to algebraic notation
func (scorer *STSScorer) moveToAlgebraic(board *board.Board, move board.Move) string {
	// Get the piece that's moving
	piece := board.GetPiece(move.From.Rank, move.From.File)
	
	// Handle castling
	if move.IsCastling {
		if move.To.File == 6 { // Kingside
			return "O-O"
		} else { // Queenside
			return "O-O-O"
		}
	}
	
	// Basic algebraic conversion
	pieceChar := ""
	switch piece {
	case 'K', 'k': // King
		pieceChar = "K"
	case 'Q', 'q': // Queen
		pieceChar = "Q"  
	case 'R', 'r': // Rook
		pieceChar = "R"
	case 'B', 'b': // Bishop
		pieceChar = "B"
	case 'N', 'n': // Knight
		pieceChar = "N"
	case 'P', 'p': // Pawn
		pieceChar = "" // Pawns don't have piece letters
	}
	
	// Destination square
	toSquare := move.To.String()
	
	// Check if it's a capture
	isCapture := board.GetPiece(move.To.Rank, move.To.File) != '.' || move.IsEnPassant
	
	// Basic algebraic notation
	algebraic := pieceChar
	
	// For pawn captures, include the file
	if piece == 'P' || piece == 'p' {
		if isCapture {
			algebraic = string('a' + rune(move.From.File)) + "x"
		}
	} else if isCapture {
		algebraic += "x"
	}
	
	algebraic += toSquare
	
	// Handle pawn promotion
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