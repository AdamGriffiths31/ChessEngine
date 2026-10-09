// Package book provides chess opening book functionality with Polyglot format support.
package book

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
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

// Config contains configuration options for opening books
type Config struct {
	Enabled         bool
	BookFiles       []string
	SelectionMode   SelectionMode
	RandomSeed      int64
	WeightThreshold uint16
}

// LookupService provides opening book functionality for the chess engine
type LookupService struct {
	manager *Manager
	config  Config
	rng     *rand.Rand
	zobrist *ZobristHash
}

// NewLookupService creates a new book lookup service
func NewLookupService(config Config) *LookupService {
	seed := config.RandomSeed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	return &LookupService{
		manager: NewManager(),
		config:  config,
		rng:     rand.New(rand.NewSource(seed)), // #nosec G404 - chess move selection doesn't require crypto/rand
		zobrist: GetPolyglotHash(),
	}
}

// LoadBooks loads opening books from the configured file paths
func (bls *LookupService) LoadBooks() error {
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
func (bls *LookupService) FindBookMove(b *board.Board) (*board.Move, error) {
	if !bls.config.Enabled {
		return nil, ErrPositionNotFound
	}

	hash := bls.zobrist.HashPosition(b)

	bookMoves, err := bls.manager.LookupMove(hash, b)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup moves in opening books: %w", err)
	}

	var validMoves []Move
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

func (bls *LookupService) selectMove(bookMoves []Move) Move {
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

func (bls *LookupService) selectBestMove(bookMoves []Move) Move {
	best := bookMoves[0]
	for _, bookMove := range bookMoves[1:] {
		if bookMove.Weight > best.Weight {
			best = bookMove
		}
	}
	return best
}

// selectRandomMove returns a random move (equal probability)
func (bls *LookupService) selectRandomMove(bookMoves []Move) Move {
	index := bls.rng.Intn(len(bookMoves))
	return bookMoves[index]
}

func (bls *LookupService) selectWeightedRandomMove(bookMoves []Move) Move {
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
