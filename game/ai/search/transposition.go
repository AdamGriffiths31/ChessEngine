package search

import (
	"sync/atomic"
	"unsafe"

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
// Must be 32 bytes for atomic operations (fits in two 64-bit words)
type TranspositionEntry struct {
	Key      uint64             // Upper 32 bits of hash + packed data
	Hash     uint32             // Lower 32 bits of zobrist hash
	Score    ai.EvaluationScore // Score at this position  
	BestMove board.Move         // Best move found (if any)
	// Packed into Key: Depth (8 bits) + Type (2 bits) + Age (6 bits) + HashUpper (32 bits)
}

// TranspositionTable implements a hash table for storing search results
type TranspositionTable struct {
	table      []TranspositionEntry
	size       uint64
	mask       uint64
	hits       atomic.Uint64  // Use atomic for thread-safe statistics
	misses     atomic.Uint64  // Use atomic for thread-safe statistics
	collisions atomic.Uint64  // Use atomic for thread-safe statistics
	age        atomic.Uint32  // Use atomic for thread-safe age updates
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
		// age initialized to 0 automatically for atomic.Uint32
	}
}

// Clear clears all entries in the transposition table
func (tt *TranspositionTable) Clear() {
	for i := range tt.table {
		atomic.StoreUint64((*uint64)(unsafe.Pointer(&tt.table[i].Key)), 0)
		atomic.StoreUint32((*uint32)(unsafe.Pointer(&tt.table[i].Hash)), 0)
		// Clear the entire entry
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
		curHashUpper, curDepth, _, curAge := unpackKey(currentKey)
		curFullHash := uint64(curHashUpper)<<32 | uint64(currentHash)
		
		// Check for hash collision
		if curFullHash != 0 && curFullHash != hash {
			tt.collisions.Add(1)
		}
		
		// Always replace if same position with deeper or equal search
		if curFullHash == hash && depth >= curDepth {
			if atomic.CompareAndSwapUint64((*uint64)(unsafe.Pointer(&entry.Key)), currentKey, newKey) {
				atomic.StoreUint32((*uint32)(unsafe.Pointer(&entry.Hash)), hashLower)
				atomic.StoreInt64((*int64)(unsafe.Pointer(&entry.Score)), int64(score))
				// Store bestMove as raw bytes - Move is a struct
				*(*board.Move)(unsafe.Pointer(&entry.BestMove)) = bestMove
				return
			}
			continue // Retry if entry changed
		}
		
		// For different positions, use replacement logic
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
	// Base value from depth (deeper searches are more valuable)
	value := depth * 16
	
	// Age bonus (recent entries get bonus, very old entries get penalty)
	currentAge := tt.age.Load()
	ageDiff := int(currentAge - age)
	if ageDiff <= 1 {
		value += 32 // Recent entries get significant bonus
	} else if ageDiff <= 3 {
		value += 16 // Somewhat recent entries get moderate bonus
	} else if ageDiff > 8 {
		value -= ageDiff * 4 // Very old entries get significant penalty
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

// Probe looks up a position in the transposition table using lock-free operations
func (tt *TranspositionTable) Probe(hash uint64) (*TranspositionEntry, bool) {
	index := hash & tt.mask
	entry := &tt.table[index]
	
	// Load entry atomically
	key := atomic.LoadUint64((*uint64)(unsafe.Pointer(&entry.Key)))
	hashLower := atomic.LoadUint32((*uint32)(unsafe.Pointer(&entry.Hash)))
	
	if key == 0 {
		tt.misses.Add(1)
		return nil, false
	}
	
	// Reconstruct full hash
	hashUpper, _, _, _ := unpackKey(key)
	fullHash := uint64(hashUpper)<<32 | uint64(hashLower)
	
	if fullHash == hash {
		tt.hits.Add(1)
		// Return the existing entry directly - no allocation!
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
	tt.age.Add(1)  // Thread-safe atomic increment, no lock needed
}

// GetStats returns hit/miss statistics
func (tt *TranspositionTable) GetStats() (hits, misses, collisions uint64, hitRate float64) {
	// No lock needed for atomic reads
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
	entrySize := uint64(40) // Approximate
	return int((tt.size * entrySize) / (1024 * 1024))
}

// GetDetailedStats returns detailed statistics about table usage
func (tt *TranspositionTable) GetDetailedStats() (hits, misses, collisions, filled, averageDepth uint64, hitRate, fillRate float64) {
	hits = tt.hits.Load()
	misses = tt.misses.Load()
	collisions = tt.collisions.Load()

	// Calculate fill rate and average depth
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
