package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"sort"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/book"
)

// This utility creates a minimal test Polyglot book for unit testing
func main() {
	startingBoard, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	zobrist := book.GetPolyglotHash()
	actualHash := zobrist.HashPosition(startingBoard)

	testEntries := []book.PolyglotEntry{
		// Starting position - e2-e4 (King's pawn)
		{
			Hash:   actualHash,
			Move:   0x031C, // e2-e4 encoded correctly
			Weight: 100,
			Learn:  0,
		},
		// Starting position - d2-d4 (Queen's pawn)
		{
			Hash:   actualHash, // Same starting position
			Move:   0x02DB,     // d2-d4 encoded correctly
			Weight: 80,
			Learn:  0,
		},
		// Starting position - Nf3
		{
			Hash:   actualHash, // Same starting position
			Move:   0x0195,     // Ng1-f3 encoded correctly
			Weight: 60,
			Learn:  0,
		},
		// Different position example
		{
			Hash:   0x123456789abcdef0, // Different position hash
			Move:   0x0234,
			Weight: 50,
			Learn:  0,
		},
	}

	createTestBook("minimal.bin", testEntries)
	createTestBook("empty.bin", []book.PolyglotEntry{})
	createTestBook("single.bin", testEntries[:1])

}

func createTestBook(filename string, entries []book.PolyglotEntry) {
	// Sort entries by hash (required for binary search)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Hash < entries[j].Hash
	})

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for _, entry := range entries {
		err := book.WriteEntry(file, entry)
		if err != nil {
			panic(err)
		}
	}

}

func writeEntryRaw(file *os.File, entry book.PolyglotEntry) error {
	if err := binary.Write(file, binary.BigEndian, entry.Hash); err != nil {
		return err
	}
	if err := binary.Write(file, binary.BigEndian, entry.Move); err != nil {
		return err
	}
	if err := binary.Write(file, binary.BigEndian, entry.Weight); err != nil {
		return err
	}
	if err := binary.Write(file, binary.BigEndian, entry.Learn); err != nil {
		return err
	}
	return nil
}
