// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"sync/atomic"
	"unsafe"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
)

// EntryType represents the type of transposition table entry
type EntryType uint8

// Entry type constants for transposition table entries
const (
	EntryExact EntryType = iota
	EntryLowerBound
	EntryUpperBound
)

// TranspositionEntry represents a single entry in the transposition table
type TranspositionEntry struct {
	Key      uint64
	Hash     uint32
	Score    ai.EvaluationScore
	BestMove board.Move
}

// TranspositionTable implements a hash table for storing search results
type TranspositionTable struct {
	table      []TranspositionEntry
	size       uint64
	mask       uint64
	hits       atomic.Uint64
	misses     atomic.Uint64
	collisions atomic.Uint64
	age        atomic.Uint32
}

// NewTranspositionTable creates a new transposition table with the given size in MB
func NewTranspositionTable(sizeMB int) *TranspositionTable {
	entrySize := uint64(40)
	numEntries := (uint64(sizeMB) * 1024 * 1024) / entrySize

	size := uint64(1)
	for size*2 <= numEntries {
		size *= 2
	}

	return &TranspositionTable{
		table: make([]TranspositionEntry, size),
		size:  size,
		mask:  size - 1,
	}
}

// Clear clears all entries in the transposition table
func (tt *TranspositionTable) Clear() {
	for i := range tt.table {
		atomic.StoreUint64((*uint64)(unsafe.Pointer(&tt.table[i].Key)), 0)
		atomic.StoreUint32((*uint32)(unsafe.Pointer(&tt.table[i].Hash)), 0)
		tt.table[i] = TranspositionEntry{}
	}

	tt.hits.Store(0)
	tt.misses.Store(0)
	tt.collisions.Store(0)
	tt.age.Store(0)
}

// packKey packs depth, type, age, and upper hash bits into a single uint64
func packKey(hashUpper uint32, depth int, entryType EntryType, age uint32) uint64 {
	return uint64(hashUpper)<<32 | uint64(depth&0xFF)<<24 | uint64(entryType&0x3)<<22 | uint64(age&0x3F)<<16
}

// unpackKey unpacks the key into its components
func unpackKey(key uint64) (hashUpper uint32, depth int, entryType EntryType, age uint32) {
	hashUpper = uint32(key >> 32)
	depth = int((key >> 24) & 0xFF)
	entryType = EntryType((key >> 22) & 0x3)
	age = uint32((key >> 16) & 0x3F)
	return
}

// Store stores a position in the transposition table using lock-free atomic operations
func (tt *TranspositionTable) Store(hash uint64, depth int, score ai.EvaluationScore,
	entryType EntryType, bestMove board.Move) {

	index := hash & tt.mask
	entry := &tt.table[index]

	hashUpper := uint32(hash >> 32)
	hashLower := uint32(hash)
	currentAge := tt.age.Load()

	// Create new packed key
	newKey := packKey(hashUpper, depth, entryType, currentAge)

	for {
		// Load current entry atomically
		currentKey := atomic.LoadUint64((*uint64)(unsafe.Pointer(&entry.Key)))
		currentHash := atomic.LoadUint32((*uint32)(unsafe.Pointer(&entry.Hash)))

		// Check if slot is empty
		if currentKey == 0 {
			// Try to claim empty slot
			if atomic.CompareAndSwapUint64((*uint64)(unsafe.Pointer(&entry.Key)), 0, newKey) {
				atomic.StoreUint32((*uint32)(unsafe.Pointer(&entry.Hash)), hashLower)
				atomic.StoreInt64((*int64)(unsafe.Pointer(&entry.Score)), int64(score))
				// Store bestMove as raw bytes - Move is a struct
				*(*board.Move)(unsafe.Pointer(&entry.BestMove)) = bestMove
				return
			}
			continue // Retry if someone else claimed it
		}

		// Unpack current entry
		curHashUpper, curDepth, curEntryType, curAge := unpackKey(currentKey)
		curFullHash := uint64(curHashUpper)<<32 | uint64(currentHash)

		// Check for hash collision
		if curFullHash != 0 && curFullHash != hash {
			tt.collisions.Add(1)
		}

		if curFullHash == hash {
			shouldReplace := false

			// Always replace if new search is strictly deeper
			if depth > curDepth {
				shouldReplace = true
			} else if depth == curDepth {
				// At same depth, use entry type priority
				// EntryExact > EntryLowerBound > EntryUpperBound
				// This ensures refutations (usually EntryExact) replace speculative scores

				if entryType == EntryExact {
					// Exact scores always replace at same depth
					shouldReplace = true
				} else if entryType == EntryLowerBound && curEntryType == EntryUpperBound {
					// Lower bounds are better than upper bounds
					shouldReplace = true
				} else if entryType == EntryUpperBound && curEntryType == EntryLowerBound {
					// CRITICAL: Upper bounds should NOT replace lower bounds at the same depth
					// This violates search invariants and causes score/move corruption
					shouldReplace = false
				} else if entryType == curEntryType {
					// Same type at same depth - keep the existing one to avoid thrashing
					// Unless the score difference is significant (indicates a refutation)
					scoreDiff := int(score) - int(entry.Score)
					if scoreDiff < 0 {
						scoreDiff = -scoreDiff
					}
					// Replace if score changed dramatically (> 500 centipawns)
					// This catches refutations that drastically change the evaluation
					if scoreDiff > 500 {
						shouldReplace = true
					}
				}
			}

			if shouldReplace {
				if atomic.CompareAndSwapUint64((*uint64)(unsafe.Pointer(&entry.Key)), currentKey, newKey) {
					atomic.StoreUint32((*uint32)(unsafe.Pointer(&entry.Hash)), hashLower)
					atomic.StoreInt64((*int64)(unsafe.Pointer(&entry.Score)), int64(score))
					// Store bestMove as raw bytes - Move is a struct
					*(*board.Move)(unsafe.Pointer(&entry.BestMove)) = bestMove
					return
				}
				continue // Retry if entry changed
			}
			// Don't replace - keep existing entry
			return
		}

		// For different positions (hash collision), use replacement logic
		if curFullHash != hash {
			existingScore := tt.calculateEntryValueFromPacked(curDepth, curAge)
			newScore := tt.calculateNewEntryValue(depth, entryType)

			// Replace if new entry is more valuable
			if newScore > existingScore {
				if atomic.CompareAndSwapUint64((*uint64)(unsafe.Pointer(&entry.Key)), currentKey, newKey) {
					atomic.StoreUint32((*uint32)(unsafe.Pointer(&entry.Hash)), hashLower)
					atomic.StoreInt64((*int64)(unsafe.Pointer(&entry.Score)), int64(score))
					// Store bestMove as raw bytes - Move is a struct
					*(*board.Move)(unsafe.Pointer(&entry.BestMove)) = bestMove
					return
				}
				continue // Retry if entry changed
			}
		}

		// Entry not replaced, exit
		return
	}
}

// calculateEntryValueFromPacked calculates replacement value from packed data
func (tt *TranspositionTable) calculateEntryValueFromPacked(depth int, age uint32) int {
	value := depth * 16

	currentAge := tt.age.Load()
	ageDiff := int(currentAge - age)
	if ageDiff <= 1 {
		value += 32
	} else if ageDiff <= 3 {
		value += 16
	} else if ageDiff > 8 {
		value -= ageDiff * 4
	}

	return value
}

// calculateNewEntryValue calculates the value score for a new entry
func (tt *TranspositionTable) calculateNewEntryValue(depth int, entryType EntryType) int {
	value := depth*16 + 32

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

// Probe looks up a position in the transposition table using lock-free operations
func (tt *TranspositionTable) Probe(hash uint64) (*TranspositionEntry, bool) {
	index := hash & tt.mask
	entry := &tt.table[index]

	key := atomic.LoadUint64((*uint64)(unsafe.Pointer(&entry.Key)))
	hashLower := atomic.LoadUint32((*uint32)(unsafe.Pointer(&entry.Hash)))

	if key == 0 {
		tt.misses.Add(1)
		return nil, false
	}

	hashUpper, _, _, _ := unpackKey(key)
	fullHash := uint64(hashUpper)<<32 | uint64(hashLower)

	if fullHash == hash {
		tt.hits.Add(1)
		return entry, true
	}

	tt.misses.Add(1)
	return nil, false
}

// GetDepth extracts depth from a transposition entry
func (entry *TranspositionEntry) GetDepth() int {
	_, depth, _, _ := unpackKey(entry.Key)
	return depth
}

// GetType extracts entry type from a transposition entry
func (entry *TranspositionEntry) GetType() EntryType {
	_, _, entryType, _ := unpackKey(entry.Key)
	return entryType
}

// GetAge extracts age from a transposition entry
func (entry *TranspositionEntry) GetAge() uint32 {
	_, _, _, age := unpackKey(entry.Key)
	return age
}

// IncrementAge increments the age counter (call at start of each search)
func (tt *TranspositionTable) IncrementAge() {
	tt.age.Add(1)
}

// GetStats returns hit/miss statistics
func (tt *TranspositionTable) GetStats() (hits, misses, collisions uint64, hitRate float64) {
	hits = tt.hits.Load()
	misses = tt.misses.Load()
	collisions = tt.collisions.Load()

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
	entrySize := uint64(40)
	return int((tt.size * entrySize) / (1024 * 1024))
}

// GetDetailedStats returns detailed statistics about table usage
func (tt *TranspositionTable) GetDetailedStats() (hits, misses, collisions, filled, averageDepth uint64, hitRate, fillRate float64) {
	hits = tt.hits.Load()
	misses = tt.misses.Load()
	collisions = tt.collisions.Load()

	var totalDepth uint64
	for i := uint64(0); i < tt.size; i++ {
		key := atomic.LoadUint64((*uint64)(unsafe.Pointer(&tt.table[i].Key)))
		if key != 0 {
			filled++
			_, depth, _, _ := unpackKey(key)
			totalDepth += uint64(depth)
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
