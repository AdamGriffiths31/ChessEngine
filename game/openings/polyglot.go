package openings

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

const (
	// PolyglotEntrySize is the size of each entry in bytes
	PolyglotEntrySize = 16
	
	// Move encoding masks and shifts
	ToSquareMask     = 0x003F  // bits 0-5: destination square
	FromSquareMask   = 0x0FC0  // bits 6-11: origin square  
	PromotionMask    = 0x7000  // bits 12-14: promotion piece
	FromSquareShift  = 6
	PromotionShift   = 12
	
	// Promotion piece values
	PromotionKnight = 1
	PromotionBishop = 2
	PromotionRook   = 3
	PromotionQueen  = 4
)

// PolyglotBook implements the OpeningBook interface for Polyglot binary format
type PolyglotBook struct {
	entries   []PolyglotEntry
	info      BookInfo
	isLoaded  bool
	zobrist   *ZobristHash
}

// NewPolyglotBook creates a new Polyglot opening book
func NewPolyglotBook() *PolyglotBook {
	return &PolyglotBook{
		entries: make([]PolyglotEntry, 0),
		zobrist: GetPolyglotHash(),
	}
}

// LoadFromFile loads a Polyglot opening book from a binary file
func (pb *PolyglotBook) LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open book file: %w", err)
	}
	defer file.Close()
	
	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	
	fileSize := stat.Size()
	if fileSize%PolyglotEntrySize != 0 {
		return ErrInvalidBookFile
	}
	
	entryCount := int(fileSize / PolyglotEntrySize)
	entries := make([]PolyglotEntry, entryCount)
	
	// Read all entries
	for i := 0; i < entryCount; i++ {
		var entry PolyglotEntry
		
		// Read in big-endian format
		if err := binary.Read(file, binary.BigEndian, &entry.Hash); err != nil {
			return fmt.Errorf("failed to read hash at entry %d: %w", i, err)
		}
		if err := binary.Read(file, binary.BigEndian, &entry.Move); err != nil {
			return fmt.Errorf("failed to read move at entry %d: %w", i, err)
		}
		if err := binary.Read(file, binary.BigEndian, &entry.Weight); err != nil {
			return fmt.Errorf("failed to read weight at entry %d: %w", i, err)
		}
		if err := binary.Read(file, binary.BigEndian, &entry.Learn); err != nil {
			return fmt.Errorf("failed to read learn at entry %d: %w", i, err)
		}
		
		entries[i] = entry
	}
	
	// Verify entries are sorted by hash (required for binary search)
	if !sort.SliceIsSorted(entries, func(i, j int) bool {
		return entries[i].Hash < entries[j].Hash
	}) {
		return fmt.Errorf("book file is not sorted by hash: %w", ErrInvalidBookFile)
	}
	
	pb.entries = entries
	pb.info = BookInfo{
		Filename:   filename,
		EntryCount: entryCount,
		FileSize:   fileSize,
	}
	pb.isLoaded = true
	
	return nil
}

// LookupMove finds book moves for the given position hash
func (pb *PolyglotBook) LookupMove(hash uint64, b *board.Board) ([]BookMove, error) {
	if !pb.isLoaded {
		return nil, ErrBookNotLoaded
	}
	
	// Find first entry with matching hash using binary search
	startIdx := sort.Search(len(pb.entries), func(i int) bool {
		return pb.entries[i].Hash >= hash
	})
	
	if startIdx >= len(pb.entries) || pb.entries[startIdx].Hash != hash {
		return nil, ErrPositionNotFound
	}
	
	// Find all entries with the same hash
	var bookMoves []BookMove
	for i := startIdx; i < len(pb.entries) && pb.entries[i].Hash == hash; i++ {
		entry := pb.entries[i]
		
		// Decode the move
		move, err := pb.decodeMove(entry.Move, b)
		if err != nil {
			continue // Skip invalid moves
		}
		
		bookMove := BookMove{
			Move:   move,
			Weight: entry.Weight,
			Learn:  entry.Learn,
		}
		
		bookMoves = append(bookMoves, bookMove)
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

// GetBookInfo returns information about the loaded book
func (pb *PolyglotBook) GetBookInfo() BookInfo {
	return pb.info
}

// decodeMove converts a 16-bit Polyglot move encoding to a board.Move
func (pb *PolyglotBook) decodeMove(encoded uint16, b *board.Board) (board.Move, error) {
	// Extract components
	toSquare := int(encoded & ToSquareMask)
	fromSquare := int((encoded & FromSquareMask) >> FromSquareShift)
	promotionPiece := int((encoded & PromotionMask) >> PromotionShift)
	
	// Convert square indices to board coordinates
	fromFile := fromSquare % 8
	fromRank := fromSquare / 8
	toFile := toSquare % 8
	toRank := toSquare / 8
	
	// Get the piece being moved
	movingPiece := b.GetPiece(fromRank, fromFile)
	if movingPiece == board.Empty {
		return board.Move{}, fmt.Errorf("no piece at from square %c%d", 'a'+fromFile, fromRank+1)
	}
	
	// Get any piece being captured
	capturedPiece := b.GetPiece(toRank, toFile)
	
	move := board.Move{
		From:      board.Square{File: fromFile, Rank: fromRank},
		To:        board.Square{File: toFile, Rank: toRank},
		Piece:     movingPiece,
		Captured:  capturedPiece,
		IsCapture: capturedPiece != board.Empty,
	}
	
	// Handle promotion
	if promotionPiece > 0 {
		switch promotionPiece {
		case PromotionKnight:
			move.Promotion = board.WhiteKnight // Will be corrected based on side
		case PromotionBishop:
			move.Promotion = board.WhiteBishop
		case PromotionRook:
			move.Promotion = board.WhiteRook
		case PromotionQueen:
			move.Promotion = board.WhiteQueen
		default:
			return board.Move{}, fmt.Errorf("invalid promotion piece: %d", promotionPiece)
		}
	}
	
	return move, nil
}

// encodeMove converts a board.Move to a 16-bit Polyglot encoding
func (pb *PolyglotBook) encodeMove(move board.Move) uint16 {
	// Convert coordinates to square indices
	fromSquare := move.From.Rank*8 + move.From.File
	toSquare := move.To.Rank*8 + move.To.File
	
	encoded := uint16(toSquare) | (uint16(fromSquare) << FromSquareShift)
	
	// Handle promotion
	if move.Promotion != board.Empty {
		var promotionValue uint16
		switch move.Promotion {
		case board.WhiteKnight, board.BlackKnight:
			promotionValue = PromotionKnight
		case board.WhiteBishop, board.BlackBishop:
			promotionValue = PromotionBishop
		case board.WhiteRook, board.BlackRook:
			promotionValue = PromotionRook
		case board.WhiteQueen, board.BlackQueen:
			promotionValue = PromotionQueen
		}
		encoded |= promotionValue << PromotionShift
	}
	
	return encoded
}

// WriteEntry writes a single entry to a writer (for creating test books)
func WriteEntry(w io.Writer, entry PolyglotEntry) error {
	if err := binary.Write(w, binary.BigEndian, entry.Hash); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, entry.Move); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, entry.Weight); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, entry.Learn); err != nil {
		return err
	}
	return nil
}