// Package ai provides chess engine AI functionality including search algorithms and evaluation.
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

// SetOpeningBook configures the opening book settings
func (c *ComputerPlayer) SetOpeningBook(enabled bool, bookFiles []string) {
	c.config.UseOpeningBook = enabled
	c.config.BookFiles = bookFiles
	// Set reasonable defaults
	c.config.BookSelectMode = BookSelectWeightedRandom
	c.config.BookWeightThreshold = 1
}

// SetBookSelectionMode sets how moves are selected from opening books
func (c *ComputerPlayer) SetBookSelectionMode(mode BookSelectionMode) {
	c.config.BookSelectMode = mode
}

// SetBookWeightThreshold sets the minimum weight for considering book moves
func (c *ComputerPlayer) SetBookWeightThreshold(threshold uint16) {
	c.config.BookWeightThreshold = threshold
}

// IsUsingOpeningBook returns whether opening books are enabled
func (c *ComputerPlayer) IsUsingOpeningBook() bool {
	return c.config.UseOpeningBook
}

// GetBookFiles returns the list of configured book files
func (c *ComputerPlayer) GetBookFiles() []string {
	return c.config.BookFiles
}
