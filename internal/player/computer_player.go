// Package player provides chess player implementations, including the computer player.
package player

import (
	"context"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/book"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/search"
)

// ComputerPlayer represents a computer chess player
type ComputerPlayer struct {
	engine search.Engine
	config search.SearchConfig
	name   string

	// prober is created lazily the first time an opening book with a
	// non-empty file list is configured, and then reused for the life of
	// this player: the book is loaded from disk at most once. This mirrors
	// the lazy, memoize-after-first-load semantics the search engine used
	// to implement internally before book probing moved out to callers.
	prober *book.Prober
}

// NewComputerPlayer creates a new computer player
func NewComputerPlayer(name string, engine search.Engine, config search.SearchConfig) *ComputerPlayer {
	return &ComputerPlayer{
		engine: engine,
		config: config,
		name:   name,
	}
}

// GetMove returns the computer's chosen move for the position
func (c *ComputerPlayer) GetMove(b *board.Board, player movegen.Player, timeLimit time.Duration) (board.Move, error) {
	result, err := c.GetMoveWithStats(b, player, timeLimit)
	if err != nil {
		return board.Move{}, err
	}
	return result.BestMove, nil
}

// GetMoveWithStats returns the computer's chosen move along with search statistics
func (c *ComputerPlayer) GetMoveWithStats(b *board.Board, player movegen.Player, timeLimit time.Duration) (search.SearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeLimit)
	defer cancel()

	config := c.config
	config.MaxTime = timeLimit

	// Consult the opening book before searching. The prober itself
	// enforces the opening-phase move-number gate and reports a miss for
	// any position not found in the book, so a graceful fallback to a full
	// search is just "prober said no".
	if config.UseOpeningBook && len(config.BookFiles) > 0 {
		if c.prober == nil {
			// prober is built once; later BookFiles changes on a live instance are ignored — matches the old engine's behavior.
			c.prober = book.NewProber(config.BookFiles)
		}
		if bookMove, ok := c.prober.Probe(b); ok {
			return search.SearchResult{
				BestMove: *bookMove,
				Score:    0,
				Stats:    search.SearchStats{BookMoveUsed: true},
			}, nil
		}
	}

	result := c.engine.FindBestMove(ctx, b, player, config)

	return result, nil
}

// GetName returns the player name
func (c *ComputerPlayer) GetName() string {
	return c.name
}

// SetOpeningBook configures the opening book settings
func (c *ComputerPlayer) SetOpeningBook(enabled bool, bookFiles []string) {
	c.config.UseOpeningBook = enabled
	c.config.BookFiles = bookFiles
}

// IsUsingOpeningBook returns whether opening books are enabled
func (c *ComputerPlayer) IsUsingOpeningBook() bool {
	return c.config.UseOpeningBook
}

// GetBookFiles returns the list of configured book files
func (c *ComputerPlayer) GetBookFiles() []string {
	return c.config.BookFiles
}
