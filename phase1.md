# Phase 1: Foundation - Player vs Computer Mode

## Overview
This document provides a detailed implementation guide for Phase 1 of the Player vs Computer mode. The goal is to establish the foundational architecture and get a basic working computer opponent that can play legal moves using simple evaluation and search.

## Phase 1 Goals
- ✅ Define core AI interfaces
- ✅ Implement basic material evaluation
- ✅ Create simple minimax search (depth 3-4)
- ✅ Integrate with existing game engine
- ✅ Create Mode 2 (Player vs Computer)
- ✅ Ensure all moves are legal and games can complete

## Step-by-Step Implementation Guide

### Step 1: Create Core AI Types and Interfaces

#### 1.1 Create `game/ai/types.go`
```go
package ai

import (
    "time"
    "github.com/AdamGriffiths31/ChessEngine/board"
    "github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// SearchStats tracks statistics during search
type SearchStats struct {
    NodesSearched   int64
    Depth          int
    Time           time.Duration
    PrincipalVariation []board.Move
}

// SearchConfig configures the search parameters
type SearchConfig struct {
    MaxDepth       int
    MaxTime        time.Duration
    MaxNodes       int64
    UseAlphaBeta   bool
}

// EvaluationScore represents the score of a position
type EvaluationScore int32

const (
    // Special scores
    MateScore     EvaluationScore = 100000
    DrawScore     EvaluationScore = 0
    UnknownScore  EvaluationScore = -1000000
)

// SearchResult contains the result of a search
type SearchResult struct {
    BestMove board.Move
    Score    EvaluationScore
    Stats    SearchStats
}
```

#### 1.2 Create `game/ai/engine.go`
```go
package ai

import (
    "context"
    "github.com/AdamGriffiths31/ChessEngine/board"
    "github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// Engine defines the interface for a chess AI engine
type Engine interface {
    // FindBestMove searches for the best move in the given position
    FindBestMove(ctx context.Context, b *board.Board, player moves.Player, config SearchConfig) SearchResult
    
    // SetEvaluator sets the position evaluator
    SetEvaluator(eval Evaluator)
    
    // GetName returns the engine name
    GetName() string
}
```

#### 1.3 Create `game/ai/evaluation/evaluator.go`
```go
package evaluation

import (
    "github.com/AdamGriffiths31/ChessEngine/board"
    "github.com/AdamGriffiths31/ChessEngine/game/ai"
    "github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// Evaluator defines the interface for position evaluation
type Evaluator interface {
    // Evaluate returns the score for the position from the given player's perspective
    Evaluate(b *board.Board, player moves.Player) ai.EvaluationScore
    
    // GetName returns the evaluator name
    GetName() string
}
```

### Step 2: Implement Basic Material Evaluation

#### 2.1 Create `game/ai/evaluation/material.go`
```go
package evaluation

import (
    "github.com/AdamGriffiths31/ChessEngine/board"
    "github.com/AdamGriffiths31/ChessEngine/game/ai"
    "github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// PieceValues defines the standard piece values in centipawns
var PieceValues = map[board.Piece]int{
    board.WhitePawn:   100,
    board.WhiteKnight: 320,
    board.WhiteBishop: 330,
    board.WhiteRook:   500,
    board.WhiteQueen:  900,
    board.WhiteKing:   0, // King has no material value
    
    board.BlackPawn:   -100,
    board.BlackKnight: -320,
    board.BlackBishop: -330,
    board.BlackRook:   -500,
    board.BlackQueen:  -900,
    board.BlackKing:   0,
}

// MaterialEvaluator evaluates positions based only on material balance
type MaterialEvaluator struct{}

// NewMaterialEvaluator creates a new material-only evaluator
func NewMaterialEvaluator() *MaterialEvaluator {
    return &MaterialEvaluator{}
}

// Evaluate returns the material balance from the given player's perspective
func (m *MaterialEvaluator) Evaluate(b *board.Board, player moves.Player) ai.EvaluationScore {
    score := 0
    
    // Sum up all piece values on the board
    for rank := 0; rank < 8; rank++ {
        for file := 0; file < 8; file++ {
            piece := b.GetPiece(rank, file)
            if piece != board.Empty {
                score += PieceValues[piece]
            }
        }
    }
    
    // Return score from player's perspective
    if player == moves.Black {
        score = -score
    }
    
    return ai.EvaluationScore(score)
}

// GetName returns the evaluator name
func (m *MaterialEvaluator) GetName() string {
    return "Material Evaluator"
}
```

### Step 3: Implement Basic Minimax Search

#### 3.1 Create `game/ai/search/minimax.go`
```go
package search

import (
    "context"
    "time"
    "github.com/AdamGriffiths31/ChessEngine/board"
    "github.com/AdamGriffiths31/ChessEngine/game/ai"
    "github.com/AdamGriffiths31/ChessEngine/game/ai/evaluation"
    "github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// MinimaxEngine implements a basic minimax search
type MinimaxEngine struct {
    evaluator evaluation.Evaluator
    generator *moves.Generator
}

// NewMinimaxEngine creates a new minimax search engine
func NewMinimaxEngine() *MinimaxEngine {
    return &MinimaxEngine{
        evaluator: evaluation.NewMaterialEvaluator(),
        generator: moves.NewGenerator(),
    }
}

// FindBestMove searches for the best move using minimax
func (m *MinimaxEngine) FindBestMove(ctx context.Context, b *board.Board, player moves.Player, config ai.SearchConfig) ai.SearchResult {
    startTime := time.Now()
    result := ai.SearchResult{
        Stats: ai.SearchStats{},
    }
    
    // Get all legal moves
    legalMoves := m.generator.GenerateAllMoves(b, player)
    defer moves.ReleaseMoveList(legalMoves)
    
    if legalMoves.Count == 0 {
        // No legal moves - game over
        return result
    }
    
    bestScore := ai.EvaluationScore(-1000000)
    var bestMove board.Move
    
    // Try each move
    for i := 0; i < legalMoves.Count; i++ {
        move := legalMoves.Moves[i]
        
        // Make the move
        if err := b.MakeMove(move); err != nil {
            continue
        }
        
        // Search deeper
        score := -m.minimax(ctx, b, oppositePlayer(player), config.MaxDepth-1, &result.Stats)
        
        // Unmake the move
        // TODO: Implement proper unmake functionality
        
        // Update best move if this is better
        if score > bestScore {
            bestScore = score
            bestMove = move
        }
        
        // Check for cancellation
        select {
        case <-ctx.Done():
            result.BestMove = bestMove
            result.Score = bestScore
            result.Stats.Time = time.Since(startTime)
            return result
        default:
        }
    }
    
    result.BestMove = bestMove
    result.Score = bestScore
    result.Stats.Time = time.Since(startTime)
    result.Stats.Depth = config.MaxDepth
    
    return result
}

// minimax is the recursive minimax search
func (m *MinimaxEngine) minimax(ctx context.Context, b *board.Board, player moves.Player, depth int, stats *ai.SearchStats) ai.EvaluationScore {
    stats.NodesSearched++
    
    // Check for cancellation
    select {
    case <-ctx.Done():
        return 0
    default:
    }
    
    // Terminal node - evaluate position
    if depth == 0 {
        return m.evaluator.Evaluate(b, player)
    }
    
    // Get all legal moves
    legalMoves := m.generator.GenerateAllMoves(b, player)
    defer moves.ReleaseMoveList(legalMoves)
    
    if legalMoves.Count == 0 {
        // No legal moves - check for checkmate or stalemate
        if m.generator.IsKingInCheck(b, player) {
            return -ai.MateScore + ai.EvaluationScore(depth) // Checkmate
        }
        return ai.DrawScore // Stalemate
    }
    
    bestScore := ai.EvaluationScore(-1000000)
    
    // Try each move
    for i := 0; i < legalMoves.Count; i++ {
        move := legalMoves.Moves[i]
        
        // Make the move
        if err := b.MakeMove(move); err != nil {
            continue
        }
        
        // Search deeper
        score := -m.minimax(ctx, b, oppositePlayer(player), depth-1, stats)
        
        // Unmake the move
        // TODO: Implement proper unmake functionality
        
        // Update best score
        if score > bestScore {
            bestScore = score
        }
    }
    
    return bestScore
}

// SetEvaluator sets the position evaluator
func (m *MinimaxEngine) SetEvaluator(eval evaluation.Evaluator) {
    m.evaluator = eval
}

// GetName returns the engine name
func (m *MinimaxEngine) GetName() string {
    return "Minimax Engine"
}

// oppositePlayer returns the opposite player
func oppositePlayer(player moves.Player) moves.Player {
    if player == moves.White {
        return moves.Black
    }
    return moves.White
}
```

### Step 4: Create Computer Player Wrapper

#### 4.1 Create `game/ai/computer_player.go`
```go
package ai

import (
    "context"
    "time"
    "github.com/AdamGriffiths31/ChessEngine/board"
    "github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// ComputerPlayer represents a computer chess player
type ComputerPlayer struct {
    engine   Engine
    config   SearchConfig
    name     string
}

// NewComputerPlayer creates a new computer player
func NewComputerPlayer(name string, engine Engine, config SearchConfig) *ComputerPlayer {
    return &ComputerPlayer{
        engine: engine,
        config: config,
        name:   name,
    }
}

// GetMove returns the computer's chosen move for the position
func (c *ComputerPlayer) GetMove(b *board.Board, player moves.Player, timeLimit time.Duration) (board.Move, error) {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), timeLimit)
    defer cancel()
    
    // Update config with time limit
    config := c.config
    config.MaxTime = timeLimit
    
    // Search for best move
    result := c.engine.FindBestMove(ctx, b, player, config)
    
    return result.BestMove, nil
}

// GetName returns the player name
func (c *ComputerPlayer) GetName() string {
    return c.name
}

// SetDifficulty adjusts the computer's playing strength
func (c *ComputerPlayer) SetDifficulty(level string) {
    switch level {
    case "easy":
        c.config.MaxDepth = 2
        c.config.MaxTime = 1 * time.Second
    case "medium":
        c.config.MaxDepth = 4
        c.config.MaxTime = 3 * time.Second
    case "hard":
        c.config.MaxDepth = 6
        c.config.MaxTime = 5 * time.Second
    }
}
```

### Step 5: Implement Player vs Computer Mode

#### 5.1 Create `game/modes/mode2.go`
```go
package modes

import (
    "fmt"
    "time"
    "github.com/AdamGriffiths31/ChessEngine/game"
    "github.com/AdamGriffiths31/ChessEngine/game/ai"
    "github.com/AdamGriffiths31/ChessEngine/game/ai/search"
    "github.com/AdamGriffiths31/ChessEngine/ui"
)

// ComputerMode implements Player vs Computer game mode
type ComputerMode struct {
    engine        *game.Engine
    prompter      *ui.Prompter
    parser        *game.MoveParser
    computer      *ai.ComputerPlayer
    humanColor    game.Player
    isRunning     bool
}

// NewComputerMode creates a new Player vs Computer mode
func NewComputerMode() *ComputerMode {
    engine := game.NewEngine()
    prompter := ui.NewPrompter()
    parser := game.NewMoveParser(game.White)
    
    // Create computer player with minimax engine
    aiEngine := search.NewMinimaxEngine()
    config := ai.SearchConfig{
        MaxDepth: 4,
        MaxTime:  3 * time.Second,
    }
    computer := ai.NewComputerPlayer("Computer", aiEngine, config)
    
    return &ComputerMode{
        engine:     engine,
        prompter:   prompter,
        parser:     parser,
        computer:   computer,
        humanColor: game.White, // Human plays white by default
        isRunning:  false,
    }
}

// Run starts the Player vs Computer game
func (cm *ComputerMode) Run() error {
    cm.isRunning = true
    cm.prompter.ShowWelcome()
    
    // Ask player to choose color
    cm.selectPlayerColor()
    
    // Ask for difficulty level
    cm.selectDifficulty()
    
    for cm.isRunning {
        state := cm.engine.GetState()
        
        // Show current game state
        cm.prompter.ShowGameState(state)
        
        // Check game over
        if state.GameOver {
            cm.handleGameOver(state)
            break
        }
        
        // Handle current player's turn
        if state.CurrentTurn == cm.humanColor {
            cm.handleHumanTurn()
        } else {
            cm.handleComputerTurn()
        }
    }
    
    cm.prompter.ShowGoodbye()
    return nil
}

// selectPlayerColor asks the player to choose their color
func (cm *ComputerMode) selectPlayerColor() {
    fmt.Println("\nChoose your color:")
    fmt.Println("1. White (play first)")
    fmt.Println("2. Black (play second)")
    
    // Simple implementation - can be enhanced with proper input handling
    cm.humanColor = game.White
}

// selectDifficulty asks the player to choose difficulty
func (cm *ComputerMode) selectDifficulty() {
    fmt.Println("\nChoose difficulty:")
    fmt.Println("1. Easy")
    fmt.Println("2. Medium") 
    fmt.Println("3. Hard")
    
    // Simple implementation - can be enhanced
    cm.computer.SetDifficulty("medium")
}

// handleHumanTurn processes the human player's move
func (cm *ComputerMode) handleHumanTurn() {
    state := cm.engine.GetState()
    cm.parser.SetCurrentPlayer(state.CurrentTurn)
    
    input, err := cm.prompter.PromptForMove(state.CurrentTurn)
    if err != nil {
        cm.prompter.ShowError(err)
        return
    }
    
    // Parse and validate move
    move, err := cm.parser.ParseMove(input, state.Board)
    if err != nil {
        cm.handleSpecialCommand(err.Error())
        return
    }
    
    // Apply the move
    err = cm.engine.ValidateAndMakeMove(move)
    if err != nil {
        cm.prompter.ShowError(err)
        return
    }
    
    cm.prompter.ShowMoveValidated()
}

// handleComputerTurn processes the computer's move
func (cm *ComputerMode) handleComputerTurn() {
    state := cm.engine.GetState()
    
    fmt.Println("Computer is thinking...")
    
    // Get computer's move
    move, err := cm.computer.GetMove(state.Board, state.CurrentTurn, 3*time.Second)
    if err != nil {
        cm.prompter.ShowError(err)
        return
    }
    
    // Apply the move
    err = cm.engine.MakeMove(move)
    if err != nil {
        cm.prompter.ShowError(err)
        return
    }
    
    fmt.Printf("Computer plays: %s%s\n", move.From.String(), move.To.String())
}

// handleSpecialCommand handles special commands like quit, reset, etc.
func (cm *ComputerMode) handleSpecialCommand(command string) {
    // Similar to mode1 implementation
    switch command {
    case "QUIT":
        if cm.prompter.ConfirmQuit() {
            cm.isRunning = false
        }
    case "RESET":
        if cm.prompter.ConfirmReset() {
            cm.engine.Reset()
            cm.selectPlayerColor()
            cm.selectDifficulty()
        }
    // ... other commands
    }
}

// handleGameOver handles the end of the game
func (cm *ComputerMode) handleGameOver(state *game.GameState) {
    if state.Winner == cm.humanColor {
        fmt.Println("Congratulations! You won!")
    } else {
        fmt.Println("Computer wins! Better luck next time.")
    }
}
```

### Step 6: Update Main to Support Mode 2

#### 6.1 Update `main.go`
```go
package main

import (
    "fmt"
    "os"
    "github.com/AdamGriffiths31/ChessEngine/game/modes"
)

func main() {
    fmt.Println("Chess Engine")
    fmt.Println("============")
    fmt.Println("\nSelect game mode:")
    fmt.Println("1. Manual Play (Player vs Player)")
    fmt.Println("2. Player vs Computer")
    fmt.Print("\nEnter choice (1 or 2): ")
    
    var choice int
    fmt.Scanln(&choice)
    
    var err error
    switch choice {
    case 1:
        manualMode := modes.NewManualMode()
        err = manualMode.Run()
    case 2:
        computerMode := modes.NewComputerMode()
        err = computerMode.Run()
    default:
        fmt.Println("Invalid choice")
        os.Exit(1)
    }
    
    if err != nil {
        fmt.Printf("Error running game: %v\n", err)
        os.Exit(1)
    }
}
```

## Testing Plan

### Unit Tests

#### Test 1: Material Evaluator (`game/ai/evaluation/material_test.go`)
```go
package evaluation

import (
    "testing"
    "github.com/AdamGriffiths31/ChessEngine/board"
    "github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func TestMaterialEvaluator(t *testing.T) {
    testCases := []struct {
        name     string
        fen      string
        expected int // Expected score from white's perspective
    }{
        {
            name:     "starting_position",
            fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
            expected: 0, // Equal material
        },
        {
            name:     "white_up_queen",
            fen:      "rnb1kbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
            expected: 900, // White has extra queen
        },
        {
            name:     "black_up_rook",
            fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBN1 w KQkq - 0 1",
            expected: -500, // Black has extra rook
        },
        {
            name:     "endgame_material",
            fen:      "4k3/8/8/8/8/8/PPPP4/4K3 w - - 0 1",
            expected: 400, // White has 4 pawns
        },
    }
    
    evaluator := NewMaterialEvaluator()
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            b, err := board.FromFEN(tc.fen)
            if err != nil {
                t.Fatalf("Failed to parse FEN: %v", err)
            }
            
            score := evaluator.Evaluate(b, moves.White)
            if int(score) != tc.expected {
                t.Errorf("Expected score %d, got %d", tc.expected, score)
            }
            
            // Test from black's perspective (should be negated)
            blackScore := evaluator.Evaluate(b, moves.Black)
            if int(blackScore) != -tc.expected {
                t.Errorf("Expected black score %d, got %d", -tc.expected, blackScore)
            }
        })
    }
}

func TestPieceValues(t *testing.T) {
    // Test that piece values are symmetric
    if PieceValues[board.WhitePawn] != -PieceValues[board.BlackPawn] {
        t.Error("Pawn values not symmetric")
    }
    if PieceValues[board.WhiteQueen] != -PieceValues[board.BlackQueen] {
        t.Error("Queen values not symmetric")
    }
    
    // Test relative values make sense
    if PieceValues[board.WhiteQueen] <= PieceValues[board.WhiteRook] {
        t.Error("Queen should be worth more than rook")
    }
    if PieceValues[board.WhiteRook] <= PieceValues[board.WhiteBishop] {
        t.Error("Rook should be worth more than bishop")
    }
}
```

#### Test 2: Minimax Search (`game/ai/search/minimax_test.go`)
```go
package search

import (
    "context"
    "testing"
    "time"
    "github.com/AdamGriffiths31/ChessEngine/board"
    "github.com/AdamGriffiths31/ChessEngine/game/ai"
    "github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func TestMinimaxFindsObviousMove(t *testing.T) {
    testCases := []struct {
        name         string
        fen          string
        player       moves.Player
        expectedMove string // in format "e2e4"
    }{
        {
            name:         "capture_queen",
            fen:          "rnbqkbnr/pppp1ppp/8/4p3/3P4/8/PPP1PPPP/RNBQKBNR b KQkq - 0 1",
            player:       moves.Black,
            expectedMove: "e5d4", // Capture white pawn
        },
        {
            name:         "avoid_checkmate",
            fen:          "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5Q2/PPPP1PPP/RNB1K1NR b KQkq - 0 1",
            player:       moves.Black,
            expectedMove: "g7g6", // Block checkmate threat
        },
    }
    
    engine := NewMinimaxEngine()
    config := ai.SearchConfig{
        MaxDepth: 3,
        MaxTime:  5 * time.Second,
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            b, err := board.FromFEN(tc.fen)
            if err != nil {
                t.Fatalf("Failed to parse FEN: %v", err)
            }
            
            ctx := context.Background()
            result := engine.FindBestMove(ctx, b, tc.player, config)
            
            moveStr := result.BestMove.From.String() + result.BestMove.To.String()
            if moveStr != tc.expectedMove {
                t.Errorf("Expected move %s, got %s", tc.expectedMove, moveStr)
            }
        })
    }
}

func TestMinimaxDepth(t *testing.T) {
    // Test that deeper search takes more nodes
    fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
    b, _ := board.FromFEN(fen)
    
    engine := NewMinimaxEngine()
    
    var prevNodes int64
    for depth := 1; depth <= 3; depth++ {
        config := ai.SearchConfig{
            MaxDepth: depth,
            MaxTime:  10 * time.Second,
        }
        
        ctx := context.Background()
        result := engine.FindBestMove(ctx, b, moves.White, config)
        
        if result.Stats.NodesSearched <= prevNodes {
            t.Errorf("Depth %d should search more nodes than depth %d", depth, depth-1)
        }
        prevNodes = result.Stats.NodesSearched
        
        t.Logf("Depth %d: %d nodes", depth, result.Stats.NodesSearched)
    }
}

func TestMinimaxTimeout(t *testing.T) {
    fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
    b, _ := board.FromFEN(fen)
    
    engine := NewMinimaxEngine()
    config := ai.SearchConfig{
        MaxDepth: 10, // Very deep to ensure timeout
        MaxTime:  100 * time.Millisecond,
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    
    start := time.Now()
    result := engine.FindBestMove(ctx, b, moves.White, config)
    elapsed := time.Since(start)
    
    if elapsed > 200*time.Millisecond {
        t.Errorf("Search should timeout within 200ms, took %v", elapsed)
    }
    
    // Should still return a valid move
    if result.BestMove.From.File == 0 && result.BestMove.From.Rank == 0 {
        t.Error("Should return a valid move even after timeout")
    }
}

func BenchmarkMinimax(b *testing.B) {
    fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
    board, _ := board.FromFEN(fen)
    
    engine := NewMinimaxEngine()
    config := ai.SearchConfig{
        MaxDepth: 3,
        MaxTime:  10 * time.Second,
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ctx := context.Background()
        engine.FindBestMove(ctx, board, moves.White, config)
    }
}
```

#### Test 3: Computer Player (`game/ai/computer_player_test.go`)
```go
package ai

import (
    "testing"
    "time"
    "github.com/AdamGriffiths31/ChessEngine/board"
    "github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func TestComputerPlayerMakesLegalMoves(t *testing.T) {
    // Test that computer always makes legal moves
    testPositions := []string{
        "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
        "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - ",
        "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - ",
    }
    
    for _, fen := range testPositions {
        t.Run(fen[:20], func(t *testing.T) {
            b, err := board.FromFEN(fen)
            if err != nil {
                t.Fatalf("Failed to parse FEN: %v", err)
            }
            
            engine := NewMockEngine() // Create a mock engine for testing
            config := SearchConfig{MaxDepth: 2}
            computer := NewComputerPlayer("Test", engine, config)
            
            move, err := computer.GetMove(b, moves.White, 1*time.Second)
            if err != nil {
                t.Errorf("Computer failed to find move: %v", err)
            }
            
            // Verify move is legal
            generator := moves.NewGenerator()
            legalMoves := generator.GenerateAllMoves(b, moves.White)
            defer moves.ReleaseMoveList(legalMoves)
            
            found := false
            for i := 0; i < legalMoves.Count; i++ {
                if movesEqual(move, legalMoves.Moves[i]) {
                    found = true
                    break
                }
            }
            
            if !found {
                t.Error("Computer made illegal move")
            }
        })
    }
}
