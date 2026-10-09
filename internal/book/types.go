package book

import (
	"errors"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// BookMoveLimit is the last full-move number on which the engine consults the
// opening book; past this point it always searches instead.
const BookMoveLimit = 10

var (
	// ErrPositionNotFound means the position has no entry in the book.
	ErrPositionNotFound = errors.New("position not found in opening book")
	// ErrInvalidBookFile means the file is not a valid, hash-sorted Polyglot book.
	ErrInvalidBookFile = errors.New("invalid opening book file format")
	// ErrBookNotLoaded means a lookup was attempted before a book was loaded.
	ErrBookNotLoaded = errors.New("opening book not loaded")
)

// OpeningBook defines the interface for opening book implementations
type OpeningBook interface {
	// LookupMove finds book moves for the given position hash
	LookupMove(hash uint64, b *board.Board) ([]Move, error)

	// LoadFromFile loads the opening book from a file
	LoadFromFile(filename string) error

	// IsLoaded returns true if a book is currently loaded
	IsLoaded() bool

	// GetBookInfo returns information about the loaded book
	GetBookInfo() Info
}

// Move represents a move from an opening book with associated metadata
type Move struct {
	Move board.Move

	// Weight represents the relative frequency/strength of this move
	Weight uint16

	// Learn contains learning data (wins, losses, draws)
	Learn uint32
}

// Info contains metadata about a loaded opening book
type Info struct {
	Filename string

	EntryCount int

	// FileSize is the size of the book file in bytes
	FileSize int64
}

// PolyglotEntry represents a single entry in a Polyglot opening book file
type PolyglotEntry struct {
	// Hash is the 64-bit Zobrist hash of the position
	Hash uint64

	// Move is the 16-bit encoded move
	Move uint16

	// Weight is the relative frequency/strength of this move
	Weight uint16

	Learn uint32
}

// Manager manages multiple opening books
type Manager struct {
	books   []OpeningBook
	primary OpeningBook
}

// NewManager creates a new book manager
func NewManager() *Manager {
	return &Manager{
		books: make([]OpeningBook, 0),
	}
}

// AddBook adds an opening book to the manager
func (bm *Manager) AddBook(book OpeningBook) {
	bm.books = append(bm.books, book)
	if bm.primary == nil {
		bm.primary = book
	}
}

// LookupMove searches for moves in all loaded books, starting with primary
func (bm *Manager) LookupMove(hash uint64, b *board.Board) ([]Move, error) {
	if bm.primary != nil && bm.primary.IsLoaded() {
		moves, err := bm.primary.LookupMove(hash, b)
		if err == nil && len(moves) > 0 {
			return moves, nil
		}
	}

	for _, book := range bm.books {
		if book != bm.primary && book.IsLoaded() {
			moves, err := book.LookupMove(hash, b)
			if err == nil && len(moves) > 0 {
				return moves, nil
			}
		}
	}

	return nil, ErrPositionNotFound
}
