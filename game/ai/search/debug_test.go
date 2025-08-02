package search

import (
	"context"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/evaluation"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func TestDebugPosition(t *testing.T) {
	engine := NewMinimaxEngine()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	// Test position from actual UCI log: rnbqkbnr/ppp2ppp/8/3Pp3/8/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 3
	// This is after moves: e2e4 d7d5 e4d5 e7e5
	b, err := board.FromFEN("rnbqkbnr/ppp2ppp/8/3Pp3/8/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 3")
	if err != nil {
		t.Fatalf("Failed to create debug position: %v", err)
	}

	config := ai.SearchConfig{
		MaxDepth:       6, // Match the exact depth from UCI log
		UseOpeningBook: true, // Enable opening book like in real games
	}
	result := engine.FindBestMove(ctx, b, moves.White, config)

	t.Logf("Debug position analysis (replicating UCI log):")
	t.Logf("FEN: rnbqkbnr/ppp2ppp/8/3Pp3/8/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 3")
	t.Logf("Moves: e2e4 d7d5 e4d5 e7e5")
	t.Logf("Best move: %s%s", result.BestMove.From.String(), result.BestMove.To.String())
	t.Logf("Score: %d (UCI log showed: 0)", result.Score)
	t.Logf("Nodes searched: %d (UCI log showed: 1181517)", result.Stats.NodesSearched)
	t.Logf("Depth reached: %d (UCI log showed: 6)", result.Stats.Depth)
	t.Logf("Book move used: %t (UCI log showed: false)", result.Stats.BookMoveUsed)
	t.Logf("Time taken: %v", result.Stats.Time)

	// Should return a valid move
	if result.BestMove.From.File == -1 && result.BestMove.From.Rank == -1 {
		t.Error("FindBestMove should return a valid move")
	}
}

func TestDebugEvaluationBug(t *testing.T) {
	// Test the position where our engine scores -270 but strong engines score +100
	// Position: rnb1kb1r/ppp1pppp/5n2/4q3/2B5/2N5/PPPPNPPP/R1BQK2R w KQkq - 6 6
	
	engine := NewMinimaxEngine()
	
	b, err := board.FromFEN("rnb1kb1r/ppp1pppp/5n2/4q3/2B5/2N5/PPPPNPPP/R1BQK2R w KQkq - 6 6")
	if err != nil {
		t.Fatalf("Failed to create debug position: %v", err)
	}

	// Test 1: Direct evaluation (should be positive for White)
	evaluator := evaluation.NewEvaluator()
	directEval := evaluator.Evaluate(b) // Always returns White's perspective
	
	// Test 2: Fixed depth search
	ctx1 := context.Background()
	config1 := ai.SearchConfig{
		MaxDepth:       1, // Test with depth 1
		UseOpeningBook: false,
	}
	result1 := engine.FindBestMove(ctx1, b, moves.White, config1)

	// Test 3: Time-based search (replicating UCI exactly)
	// Create timeout context like UCI engine does  
	timeLimit := 1000 * time.Millisecond // 1 second timeout like in game
	ctx2, cancel := context.WithTimeout(context.Background(), timeLimit)
	defer cancel()
	
	config2 := ai.SearchConfig{
		MaxDepth:       6, // Match the depth from UCI log: depth=6
		UseOpeningBook: true, // Enable book like real UCI
		BookFiles:      []string{"/home/adam/Documents/git/ChessEngine/game/openings/testdata/performance.bin"},
	}
	result2 := engine.FindBestMove(ctx2, b, moves.White, config2)

	t.Logf("=== EVALUATION BUG ANALYSIS ===")
	t.Logf("FEN: rnb1kb1r/ppp1pppp/5n2/4q3/2B5/2N5/PPPPNPPP/R1BQK2R w KQkq - 6 6")
	t.Logf("Strong engine evaluation: +100 (White better)")
	t.Logf("")
	t.Logf("Direct evaluation: %d", directEval)
	t.Logf("Fixed depth (2): move=%s%s score=%d depth=%d nodes=%d", 
		result1.BestMove.From.String(), result1.BestMove.To.String(), 
		result1.Score, result1.Stats.Depth, result1.Stats.NodesSearched)
	t.Logf("Time-based (1035ms): move=%s%s score=%d depth=%d nodes=%d book=%t time=%v", 
		result2.BestMove.From.String(), result2.BestMove.To.String(), 
		result2.Score, result2.Stats.Depth, result2.Stats.NodesSearched, 
		result2.Stats.BookMoveUsed, result2.Stats.Time)
	t.Logf("UCI log showed: move=a2a3 score=-270 depth=6 nodes=1181517 book=false")
	
	// Should return valid moves
	if result1.BestMove.From.File == -1 && result1.BestMove.From.Rank == -1 {
		t.Error("Fixed depth search should return a valid move")
	}
	if result2.BestMove.From.File == -1 && result2.BestMove.From.Rank == -1 {
		t.Error("Time-based search should return a valid move")
	}
}

func TestDebugNewPosition(t *testing.T) {
	engine := NewMinimaxEngine()
	
	// Test the new FEN position provided by user
	b, err := board.FromFEN("rnbqk1nr/ppppbppp/4p3/8/2PP4/8/PP1NPPPP/R1BQKBNR w KQkq - 3 4")
	if err != nil {
		t.Fatalf("Failed to create debug position: %v", err)
	}

	// Test direct evaluation
	evaluator := evaluation.NewEvaluator()
	directEval := evaluator.Evaluate(b)
	
	// Test at different depths
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Depth 1
	config1 := ai.SearchConfig{
		MaxDepth:       1,
		UseOpeningBook: false,
	}
	result1 := engine.FindBestMove(ctx, b, moves.White, config1)
	
	// Depth 2
	config2 := ai.SearchConfig{
		MaxDepth:       2, 
		UseOpeningBook: false,
	}
	result2 := engine.FindBestMove(ctx, b, moves.White, config2)
	
	// Depth 3
	config3 := ai.SearchConfig{
		MaxDepth:       3,
		UseOpeningBook: false,
	}
	result3 := engine.FindBestMove(ctx, b, moves.White, config3)

	t.Logf("=== NEW POSITION ANALYSIS ===")
	t.Logf("FEN: rnbqk1nr/ppppbppp/4p3/8/2PP4/8/PP1NPPPP/R1BQKBNR w KQkq - 3 4")
	t.Logf("")
	t.Logf("Direct evaluation: %d", directEval)
	t.Logf("Depth 1: move=%s%s score=%d nodes=%d", 
		result1.BestMove.From.String(), result1.BestMove.To.String(), 
		result1.Score, result1.Stats.NodesSearched)
	t.Logf("Depth 2: move=%s%s score=%d nodes=%d", 
		result2.BestMove.From.String(), result2.BestMove.To.String(), 
		result2.Score, result2.Stats.NodesSearched)
	t.Logf("Depth 3: move=%s%s score=%d nodes=%d", 
		result3.BestMove.From.String(), result3.BestMove.To.String(), 
		result3.Score, result3.Stats.NodesSearched)
}

func TestMoveOrderingBias(t *testing.T) {
	engine := NewMinimaxEngine()
	
	// Test the position where engine chooses a2a3 (the problematic one)
	b, err := board.FromFEN("rnb1kb1r/ppp1pppp/5n2/4q3/2B5/2N5/PPPPNPPP/R1BQK2R w KQkq - 6 6")
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	// Test multiple times to see if we always get the same first move
	ctx := context.Background()
	config := ai.SearchConfig{
		MaxDepth:       4,
		UseOpeningBook: false,
	}
	
	// Run the same search multiple times
	moveList := make([]string, 0)
	for i := 0; i < 5; i++ {
		result := engine.FindBestMove(ctx, b, moves.White, config)
		moveStr := result.BestMove.From.String() + result.BestMove.To.String()
		moveList = append(moveList, moveStr)
		t.Logf("Run %d: move=%s score=%d", i+1, moveStr, result.Score)
	}
	
	// Check if we always get the same move (indicating bias)
	allSame := true
	for i := 1; i < len(moveList); i++ {
		if moveList[i] != moveList[0] {
			allSame = false
			break
		}
	}
	
	if allSame {
		t.Logf("WARNING: All 5 runs returned the same move '%s' - possible move ordering bias", moveList[0])
	}
}

func TestUCIReplication(t *testing.T) {
	engine := NewMinimaxEngine()
	
	// Test the exact position where UCI chose a2a3
	b, err := board.FromFEN("rnb1kb1r/ppp1pppp/5n2/4q3/2B5/2N5/PPPPNPPP/R1BQK2R w KQkq - 6 6")
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	// First, let's see what the first few legal moves are
	legalMoves := engine.generator.GenerateAllMoves(b, moves.White)
	t.Logf("First 5 legal moves:")
	maxToShow := 5
	if legalMoves.Count < maxToShow {
		maxToShow = legalMoves.Count
	}
	for i := 0; i < maxToShow; i++ {
		move := legalMoves.Moves[i]
		t.Logf("  %d: %s%s", i+1, move.From.String(), move.To.String())
	}
	moves.ReleaseMoveList(legalMoves)

	// Replicate exact UCI conditions: depth=6, ~1s time, no book
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()
	
	config := ai.SearchConfig{
		MaxDepth:       6, // Exact depth from UCI log
		UseOpeningBook: false, // book=false from UCI log
	}
	
	result := engine.FindBestMove(ctx, b, moves.White, config)

	t.Logf("=== UCI REPLICATION TEST ===")
	t.Logf("FEN: rnb1kb1r/ppp1pppp/5n2/4q3/2B5/2N5/PPPPNPPP/R1BQK2R w KQkq - 6 6")
	t.Logf("UCI log: move=a2a3 score=50 depth=6 nodes=1230876 time=0.978s")
	t.Logf("Our result: move=%s%s score=%d depth=%d nodes=%d time=%v", 
		result.BestMove.From.String(), result.BestMove.To.String(), 
		result.Score, result.Stats.Depth, result.Stats.NodesSearched,
		result.Stats.Time)
	
	if result.BestMove.From.String() == "a2" && result.BestMove.To.String() == "a3" {
		t.Logf("🚨 CONFIRMED: Engine still chooses a2a3 - bug NOT fixed!")
	} else {
		t.Logf("✅ Engine chose different move - potential improvement!")
	}
}