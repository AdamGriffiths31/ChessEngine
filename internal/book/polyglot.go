package book

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

const (
	// PolyglotEntrySize is the size of each entry in bytes
	PolyglotEntrySize = 16

	// FromSquareShift is the right shift amount to extract origin square
	FromSquareShift = 6
	// PromotionShift is the right shift amount to extract promotion piece
	PromotionShift = 12
	// PromotionQueen represents queen promotion in polyglot format
	PromotionQueen = 4

	squareMask    = 0x3F
	promotionMask = 0x7
)

var promotionPieces = [PromotionQueen + 1][2]board.Piece{
	1: {board.WhiteKnight, board.BlackKnight},
	2: {board.WhiteBishop, board.BlackBishop},
	3: {board.WhiteRook, board.BlackRook},
	4: {board.WhiteQueen, board.BlackQueen},
}

// PolyglotBook implements the OpeningBook interface for Polyglot binary format
type PolyglotBook struct {
	entries  []PolyglotEntry
	isLoaded bool
}

// NewPolyglotBook creates a new Polyglot opening book
func NewPolyglotBook() *PolyglotBook {
	return &PolyglotBook{}
}

// LoadFromFile loads a Polyglot opening book from a binary file
func (pb *PolyglotBook) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		return fmt.Errorf("failed to read book file: %w", err)
	}

	if len(data)%PolyglotEntrySize != 0 {
		return ErrInvalidBookFile
	}

	entries := make([]PolyglotEntry, len(data)/PolyglotEntrySize)
	for i := range entries {
		record := data[i*PolyglotEntrySize:]
		entries[i] = PolyglotEntry{
			Hash:   binary.BigEndian.Uint64(record),
			Move:   binary.BigEndian.Uint16(record[8:]),
			Weight: binary.BigEndian.Uint16(record[10:]),
			Learn:  binary.BigEndian.Uint32(record[12:]),
		}
	}

	if !sort.SliceIsSorted(entries, func(i, j int) bool {
		return entries[i].Hash < entries[j].Hash
	}) {
		return fmt.Errorf("book file is not sorted by hash: %w", ErrInvalidBookFile)
	}

	pb.entries = entries
	pb.isLoaded = true

	return nil
}

// LookupMove finds book moves for the given position hash
func (pb *PolyglotBook) LookupMove(hash uint64, b *board.Board) ([]Move, error) {
	if !pb.isLoaded {
		return nil, ErrBookNotLoaded
	}

	startIdx := sort.Search(len(pb.entries), func(i int) bool {
		return pb.entries[i].Hash >= hash
	})

	var bookMoves []Move
	for i := startIdx; i < len(pb.entries) && pb.entries[i].Hash == hash; i++ {
		entry := pb.entries[i]

		move, err := pb.decodeMove(entry.Move, b)
		if err != nil {
			continue
		}

		bookMoves = append(bookMoves, Move{Move: move, Weight: entry.Weight})
	}

	if len(bookMoves) == 0 {
		return nil, ErrPositionNotFound
	}

	return bookMoves, nil
}

// IsLoaded returns true if a book is currently loaded
func (pb *PolyglotBook) IsLoaded() bool {
	return pb.isLoaded
}

func (pb *PolyglotBook) decodeMove(encoded uint16, b *board.Board) (board.Move, error) {
	toSquare := int(encoded & squareMask)
	fromSquare := int(encoded >> FromSquareShift & squareMask)
	promotionPiece := int(encoded >> PromotionShift & promotionMask)

	fromFile, fromRank := board.SquareToFileRank(fromSquare)
	toFile, toRank := board.SquareToFileRank(toSquare)

	movingPiece := b.GetPiece(fromRank, fromFile)
	if movingPiece == board.Empty {
		return board.Move{}, fmt.Errorf("no piece at from square %c%d", 'a'+fromFile, fromRank+1)
	}

	// Polyglot encodes castling as the king moving to its OWN rook's square
	// (e.g. White kingside is e1-h1, not e1-g1). Detect that pattern and
	// remap it to the king's actual destination with IsCastling set, so
	// board.Board.MakeMove executes a real castle instead of treating it
	// as the king capturing its own rook.
	isCastling := false
	if (movingPiece == board.WhiteKing || movingPiece == board.BlackKing) && fromFile == 4 &&
		fromRank == toRank && (fromRank == 0 || fromRank == 7) {
		switch toFile {
		case 7:
			toFile, isCastling = 6, true
		case 0:
			toFile, isCastling = 2, true
		}
	}

	capturedPiece := b.GetPiece(toRank, toFile)
	if isCastling {
		capturedPiece = board.Empty
	}

	move := board.Move{
		From:       board.NewSquare(fromFile, fromRank),
		To:         board.NewSquare(toFile, toRank),
		Piece:      movingPiece,
		Captured:   capturedPiece,
		IsCapture:  capturedPiece != board.Empty,
		IsCastling: isCastling,
		Promotion:  board.Empty,
	}

	if promotionPiece > PromotionQueen {
		return board.Move{}, fmt.Errorf("invalid promotion piece: %d", promotionPiece)
	}
	if promotionPiece > 0 {
		side := 1
		if board.IsWhitePiece(movingPiece) {
			side = 0
		}
		move.Promotion = promotionPieces[promotionPiece][side]
	}

	return move, nil
}
