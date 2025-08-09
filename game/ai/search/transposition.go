package search

import (
	"sync"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
)

// EntryType represents the type of transposition table entry
type EntryType uint8

const (
	EntryExact EntryType = iota
	EntryLowerBound
	EntryUpperBound
)

// TranspositionEntry represents a single entry in the transposition table
type TranspositionEntry struct {
	Hash     uint64             // Zobrist hash of the position
	Depth    int                // Search depth
	Score    ai.EvaluationScore // Score at this position  
	Type     EntryType          // Type of entry (exact, lower, upper)
	BestMove board.Move         // Best move found (if any)
	Age      uint32             // Age counter for replacement
}

// TranspositionTable implements a hash table for storing search results
type TranspositionTable struct {
	table      []TranspositionEntry
	size       uint64
	mask       uint64
	mu         sync.RWMutex
	hits       uint64
	misses     uint64
	collisions uint64
	age        uint32
}

// NewTranspositionTable creates a new transposition table with the given size in MB
func NewTranspositionTable(sizeMB int) *TranspositionTable {
	// Calculate number of entries based on memory size
	entrySize := uint64(40) // Approximate size of TranspositionEntry in bytes
	numEntries := (uint64(sizeMB) * 1024 * 1024) / entrySize

	// Round down to nearest power of 2 for efficient masking
	size := uint64(1)
	for size*2 <= numEntries {
		size *= 2
	}

	return &TranspositionTable{
		table: make([]TranspositionEntry, size),
		size:  size,
		mask:  size - 1,
		age:   0,
	}
}

// Clear clears all entries in the transposition table
func (tt *TranspositionTable) Clear() {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	for i := range tt.table {
		tt.table[i] = TranspositionEntry{}
	}

	tt.hits = 0
	tt.misses = 0
	tt.collisions = 0
	tt.age = 0
}

// Store stores a position in the transposition table
func (tt *TranspositionTable) Store(hash uint64, depth int, score ai.EvaluationScore,
	entryType EntryType, bestMove board.Move) {

	tt.mu.Lock()
	defer tt.mu.Unlock()

	index := hash & tt.mask
	entry := &tt.table[index]

	if entry.Hash != 0 && entry.Hash != hash {
		tt.collisions++
	}

	// Replace if:
	// 1. Empty entry
	// 2. Same position but deeper search
	// 3. Different position and older entry
	// 4. Different position but shallower search (prefer recent shallow over old deep)
	shouldReplace := entry.Hash == 0 ||
		entry.Hash == hash && depth >= entry.Depth ||
		entry.Hash != hash && entry.Age < tt.age ||
		entry.Hash != hash && depth >= entry.Depth-2

	if shouldReplace {
		entry.Hash = hash
		entry.Depth = depth
		entry.Score = score
		entry.Type = entryType
		entry.BestMove = bestMove
		entry.Age = tt.age
	}
}

// Probe looks up a position in the transposition table
func (tt *TranspositionTable) Probe(hash uint64) (*TranspositionEntry, bool) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	index := hash & tt.mask
	entry := &tt.table[index]

	if entry.Hash == hash {
		tt.hits++
		entryCopy := *entry
		return &entryCopy, true
	}

	tt.misses++
	return nil, false
}

// IncrementAge increments the age counter (call at start of each search)
func (tt *TranspositionTable) IncrementAge() {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	tt.age++
}

// GetStats returns hit/miss statistics
func (tt *TranspositionTable) GetStats() (hits, misses, collisions uint64, hitRate float64) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	hits = tt.hits
	misses = tt.misses
	collisions = tt.collisions

	total := hits + misses
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	return
}

// GetSize returns the size of the table in entries
func (tt *TranspositionTable) GetSize() uint64 {
	return tt.size
}

// GetMemoryUsage returns approximate memory usage in MB
func (tt *TranspositionTable) GetMemoryUsage() int {
	entrySize := uint64(40) // Approximate
	return int((tt.size * entrySize) / (1024 * 1024))
}
