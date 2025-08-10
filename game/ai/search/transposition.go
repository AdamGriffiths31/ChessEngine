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

// Store stores a position in the transposition table using improved replacement logic
func (tt *TranspositionTable) Store(hash uint64, depth int, score ai.EvaluationScore,
	entryType EntryType, bestMove board.Move) {

	tt.mu.Lock()
	defer tt.mu.Unlock()

	index := hash & tt.mask
	entry := &tt.table[index]

	if entry.Hash != 0 && entry.Hash != hash {
		tt.collisions++
	}

	// Always replace if:
	// 1. Empty entry
	// 2. Same position with deeper or equal depth search
	if entry.Hash == 0 || (entry.Hash == hash && depth >= entry.Depth) {
		entry.Hash = hash
		entry.Depth = depth
		entry.Score = score
		entry.Type = entryType
		entry.BestMove = bestMove
		entry.Age = tt.age
		return
	}

	// For different positions, use sophisticated replacement logic
	if entry.Hash != hash {
		// Calculate replacement score for existing entry (higher = more valuable)
		existingScore := tt.calculateEntryValue(entry)
		
		// Calculate replacement score for new entry
		newScore := tt.calculateNewEntryValue(depth, entryType)
		
		// Replace if new entry is more valuable
		if newScore > existingScore {
			entry.Hash = hash
			entry.Depth = depth
			entry.Score = score
			entry.Type = entryType
			entry.BestMove = bestMove
			entry.Age = tt.age
		}
	}
}

// calculateEntryValue calculates the replacement value of an existing entry
// Higher values mean the entry should be kept longer
func (tt *TranspositionTable) calculateEntryValue(entry *TranspositionEntry) int {
	// Base value from depth (deeper searches are more valuable)
	value := entry.Depth * 16
	
	// Age bonus (recent entries get bonus, very old entries get penalty)
	ageDiff := int(tt.age - entry.Age)
	if ageDiff <= 1 {
		value += 32 // Recent entries get significant bonus
	} else if ageDiff <= 3 {
		value += 16 // Somewhat recent entries get moderate bonus
	} else if ageDiff > 8 {
		value -= ageDiff * 4 // Very old entries get significant penalty
	}
	
	// Entry type bonus (exact values are most valuable)
	switch entry.Type {
	case EntryExact:
		value += 8
	case EntryLowerBound:
		value += 4
	case EntryUpperBound:
		value += 2
	}
	
	return value
}

// calculateNewEntryValue calculates the value score for a new entry
func (tt *TranspositionTable) calculateNewEntryValue(depth int, entryType EntryType) int {
	// New entries always get current age bonus
	value := depth*16 + 32
	
	// Entry type bonus
	switch entryType {
	case EntryExact:
		value += 8
	case EntryLowerBound:
		value += 4
	case EntryUpperBound:
		value += 2
	}
	
	return value
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

// GetDetailedStats returns detailed statistics about table usage
func (tt *TranspositionTable) GetDetailedStats() (hits, misses, collisions, filled, averageDepth uint64, hitRate, fillRate float64) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	hits = tt.hits
	misses = tt.misses
	collisions = tt.collisions

	// Calculate fill rate and average depth
	var totalDepth uint64
	for i := uint64(0); i < tt.size; i++ {
		if tt.table[i].Hash != 0 {
			filled++
			totalDepth += uint64(tt.table[i].Depth)
		}
	}

	total := hits + misses
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}
	
	fillRate = float64(filled) / float64(tt.size) * 100
	
	if filled > 0 {
		averageDepth = totalDepth / filled
	}

	return
}
