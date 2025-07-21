package main

import (
	"encoding/binary"
	"os"
	"path/filepath"

	"github.com/AdamGriffiths31/ChessEngine/game/openings"
)

// This utility creates a minimal test Polyglot book for unit testing
func main() {
	// Create testdata directory
	err := os.MkdirAll("testdata", 0755)
	if err != nil {
		panic(err)
	}
	
	// Define test entries - these represent common opening moves
	testEntries := []openings.PolyglotEntry{
		// Starting position - e2-e4 (King's pawn)
		{
			Hash:   0x463b96181691fc9c, // Actual Polyglot hash for starting position
			Move:   0x01CC,             // e2-e4 encoded
			Weight: 100,
			Learn:  0,
		},
		// Starting position - d2-d4 (Queen's pawn)
		{
			Hash:   0x463b96181691fc9c, // Same starting position
			Move:   0x01AB,             // d2-d4 encoded
			Weight: 80,
			Learn:  0,
		},
		// Starting position - Nf3
		{
			Hash:   0x463b96181691fc9c, // Same starting position
			Move:   0x015D,             // Ng1-f3 encoded
			Weight: 60,
			Learn:  0,
		},
		// Different position example
		{
			Hash:   0x123456789abcdef0, // Different position hash
			Move:   0x0234,             // Some move
			Weight: 50,
			Learn:  0,
		},
	}
	
	// Create minimal test book
	createTestBook("testdata/minimal.bin", testEntries)
	
	// Create empty book (for testing edge cases)
	createTestBook("testdata/empty.bin", []openings.PolyglotEntry{})
	
	// Create single entry book
	createTestBook("testdata/single.bin", testEntries[:1])
	
	println("Test books created successfully in testdata/")
}

func createTestBook(filename string, entries []openings.PolyglotEntry) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	
	for _, entry := range entries {
		err := openings.WriteEntry(file, entry)
		if err != nil {
			panic(err)
		}
	}
	
	println("Created", filename, "with", len(entries), "entries")
}

// Alternative approach using raw binary writing (if WriteEntry is not available)
func writeEntryRaw(file *os.File, entry openings.PolyglotEntry) error {
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