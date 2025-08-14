package openings

import (
	"errors"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Common errors
var (
	ErrPositionNotFound = errors.New("position not found in opening book")
	ErrInvalidBookFile  = errors.New("invalid opening book file format")
	ErrBookNotLoaded    = errors.New("opening book not loaded")
)

// OpeningBook defines the interface for opening book implementations
type OpeningBook interface {
	// LookupMove finds book moves for the given position hash
	LookupMove(hash uint64, b *board.Board) ([]BookMove, error)

	// LoadFromFile loads the opening book from a file
	LoadFromFile(filename string) error

	// IsLoaded returns true if a book is currently loaded
	IsLoaded() bool

	// GetBookInfo returns information about the loaded book
	GetBookInfo() BookInfo
}

// BookMove represents a move from an opening book with associated metadata
type BookMove struct {
	// Move is the chess move
	Move board.Move

	// Weight represents the relative frequency/strength of this move
	Weight uint16

	// Learn contains learning data (wins, losses, draws)
	Learn uint32
}

// BookInfo contains metadata about a loaded opening book
type BookInfo struct {
	// Filename is the path to the book file
	Filename string

	// EntryCount is the number of entries in the book
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

	// Learn contains learning data
	Learn uint32
}

// BookManager manages multiple opening books
type BookManager struct {
	books   []OpeningBook
	primary OpeningBook
}

// NewBookManager creates a new book manager
func NewBookManager() *BookManager {
	return &BookManager{
		books: make([]OpeningBook, 0),
	}
}

// AddBook adds an opening book to the manager
func (bm *BookManager) AddBook(book OpeningBook) {
	bm.books = append(bm.books, book)
	if bm.primary == nil {
		bm.primary = book
	}
}

// SetPrimary sets the primary opening book
func (bm *BookManager) SetPrimary(book OpeningBook) {
	bm.primary = book
}

// LookupMove searches for moves in all loaded books, starting with primary
func (bm *BookManager) LookupMove(hash uint64, b *board.Board) ([]BookMove, error) {
	if bm.primary != nil && bm.primary.IsLoaded() {
		moves, err := bm.primary.LookupMove(hash, b)
		if err == nil && len(moves) > 0 {
			return moves, nil
		}
	}

	// Search other books if primary didn't have the position
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
