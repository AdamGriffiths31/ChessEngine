// Package openings provides chess opening book functionality with Polyglot format support.
package openings

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// SelectionMode defines how to select moves when multiple options exist
type SelectionMode int

const (
	// SelectRandom chooses randomly based on weights
	SelectRandom SelectionMode = iota

	// SelectBest always chooses the highest-weighted move
	SelectBest

	// SelectWeightedRandom uses weighted random selection
	SelectWeightedRandom
)

// BookConfig contains configuration options for opening books
type BookConfig struct {
	Enabled         bool
	BookFiles       []string
	SelectionMode   SelectionMode
	RandomSeed      int64
	WeightThreshold uint16
}

// DefaultBookConfig returns a default book configuration
func DefaultBookConfig() BookConfig {
	return BookConfig{
		Enabled:         true,
		BookFiles:       []string{},
		SelectionMode:   SelectWeightedRandom,
		RandomSeed:      0,
		WeightThreshold: 1,
	}
}

// BookLookupService provides opening book functionality for the chess engine
type BookLookupService struct {
	manager *BookManager
	config  BookConfig
	rng     *rand.Rand
	zobrist *ZobristHash
}

// NewBookLookupService creates a new book lookup service
func NewBookLookupService(config BookConfig) *BookLookupService {
	seed := config.RandomSeed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	return &BookLookupService{
		manager: NewBookManager(),
		config:  config,
		rng:     rand.New(rand.NewSource(seed)), // #nosec G404 - chess move selection doesn't require crypto/rand
		zobrist: GetPolyglotHash(),
	}
}

// LoadBooks loads opening books from the configured file paths
func (bls *BookLookupService) LoadBooks() error {
	if !bls.config.Enabled {
		return nil
	}

	for _, filename := range bls.config.BookFiles {
		book := NewPolyglotBook()
		if err := book.LoadFromFile(filename); err != nil {
			return fmt.Errorf("failed to load book %s: %w", filename, err)
		}

		bls.manager.AddBook(book)
	}

	return nil
}

// FindBookMove searches for a move in the opening books
func (bls *BookLookupService) FindBookMove(b *board.Board) (*board.Move, error) {
	if !bls.config.Enabled {
		return nil, ErrPositionNotFound
	}

	hash := bls.zobrist.HashPosition(b)

	bookMoves, err := bls.manager.LookupMove(hash, b)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup moves in opening books: %w", err)
	}

	var validMoves []BookMove
	for _, bookMove := range bookMoves {
		if bookMove.Weight >= bls.config.WeightThreshold {
			validMoves = append(validMoves, bookMove)
		}
	}

	if len(validMoves) == 0 {
		return nil, ErrPositionNotFound
	}

	selectedMove := bls.selectMove(validMoves)
	return &selectedMove.Move, nil
}

// selectMove chooses a move from the available options based on selection mode
func (bls *BookLookupService) selectMove(bookMoves []BookMove) BookMove {
	if len(bookMoves) == 1 {
		return bookMoves[0]
	}

	switch bls.config.SelectionMode {
	case SelectBest:
		return bls.selectBestMove(bookMoves)

	case SelectRandom:
		return bls.selectRandomMove(bookMoves)

	case SelectWeightedRandom:
		return bls.selectWeightedRandomMove(bookMoves)

	default:
		return bls.selectWeightedRandomMove(bookMoves)
	}
}

// selectBestMove returns the move with the highest weight
func (bls *BookLookupService) selectBestMove(bookMoves []BookMove) BookMove {
	best := bookMoves[0]
	for _, bookMove := range bookMoves[1:] {
		if bookMove.Weight > best.Weight {
			best = bookMove
		}
	}
	return best
}

// selectRandomMove returns a random move (equal probability)
func (bls *BookLookupService) selectRandomMove(bookMoves []BookMove) BookMove {
	index := bls.rng.Intn(len(bookMoves))
	return bookMoves[index]
}

// selectWeightedRandomMove returns a move selected randomly based on weights
func (bls *BookLookupService) selectWeightedRandomMove(bookMoves []BookMove) BookMove {
	var totalWeight uint32
	for _, bookMove := range bookMoves {
		totalWeight += uint32(bookMove.Weight)
	}

	if totalWeight == 0 {
		return bls.selectRandomMove(bookMoves)
	}

	r := uint32(bls.rng.Int63n(int64(totalWeight)))

	var accumulatedWeight uint32
	for _, bookMove := range bookMoves {
		accumulatedWeight += uint32(bookMove.Weight)
		if r < accumulatedWeight {
			return bookMove
		}
	}

	return bookMoves[len(bookMoves)-1]
}

// GetLoadedBooks returns information about loaded books
func (bls *BookLookupService) GetLoadedBooks() []BookInfo {
	var infos []BookInfo
	for _, book := range bls.manager.books {
		if book.IsLoaded() {
			infos = append(infos, book.GetBookInfo())
		}
	}
	return infos
}

// IsEnabled returns whether book lookup is enabled
func (bls *BookLookupService) IsEnabled() bool {
	return bls.config.Enabled
}

// SetEnabled enables or disables book lookup
func (bls *BookLookupService) SetEnabled(enabled bool) {
	bls.config.Enabled = enabled
}

// SetSelectionMode changes the move selection mode
func (bls *BookLookupService) SetSelectionMode(mode SelectionMode) {
	bls.config.SelectionMode = mode
}

// SetWeightThreshold sets the minimum weight for considering moves
func (bls *BookLookupService) SetWeightThreshold(threshold uint16) {
	bls.config.WeightThreshold = threshold
}

// ValidateMove checks if a move exists in the opening books for verification
func (bls *BookLookupService) ValidateMove(b *board.Board, move board.Move) bool {
	if !bls.config.Enabled {
		return false
	}

	hash := bls.zobrist.HashPosition(b)
	bookMoves, err := bls.manager.LookupMove(hash, b)
	if err != nil {
		return false
	}

	for _, bookMove := range bookMoves {
		if movesEqual(bookMove.Move, move) {
			return true
		}
	}

	return false
}

// movesEqual compares two moves for equality
func movesEqual(a, b board.Move) bool {
	return a.From.File == b.From.File &&
		a.From.Rank == b.From.Rank &&
		a.To.File == b.To.File &&
		a.To.Rank == b.To.Rank &&
		a.Promotion == b.Promotion
}
