package ai

import (
	"context"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// ComputerPlayer represents a computer chess player
type ComputerPlayer struct {
	engine Engine
	config SearchConfig
	name   string
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
	result, err := c.GetMoveWithStats(b, player, timeLimit)
	if err != nil {
		return board.Move{}, err
	}
	return result.BestMove, nil
}

// GetMoveWithStats returns the computer's chosen move along with search statistics
func (c *ComputerPlayer) GetMoveWithStats(b *board.Board, player moves.Player, timeLimit time.Duration) (SearchResult, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeLimit)
	defer cancel()

	// Update config with time limit
	config := c.config
	config.MaxTime = timeLimit

	// Search for best move
	result := c.engine.FindBestMove(ctx, b, player, config)

	return result, nil
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
	default:
		// Default to medium if unknown level
		c.config.MaxDepth = 4
		c.config.MaxTime = 3 * time.Second
	}
}

// GetDifficulty returns the current difficulty level description
func (c *ComputerPlayer) GetDifficulty() string {
	switch c.config.MaxDepth {
	case 2:
		return "Easy (depth 2, 1s think time)"
	case 4:
		return "Medium (depth 4, 3s think time)"
	case 6:
		return "Hard (depth 6, 5s think time)"
	default:
		return "Custom"
	}
}

// SetDebugMode enables or disables debug output
func (c *ComputerPlayer) SetDebugMode(enabled bool) {
	c.config.DebugMode = enabled
}

// IsDebugMode returns whether debug mode is enabled
func (c *ComputerPlayer) IsDebugMode() bool {
	return c.config.DebugMode
}