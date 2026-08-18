// Package search provides chess move search algorithms and transposition table implementation.
package search

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
	"github.com/AdamGriffiths31/ChessEngine/internal/eval"
)

// EntryType represents the type of transposition table entry
type EntryType uint8

// Entry type constants for transposition table entries
const (
	EntryExact EntryType = iota
	EntryLowerBound
	EntryUpperBound
)

// scoreToTT converts a root-relative mate score into a node-relative one for
// storage. Mate scores must be stored relative to the storing node because an
// entry can be probed from a different ply (or a later search) than it was
// stored at; non-mate scores pass through unchanged.
func scoreToTT(score eval.EvaluationScore, ply int) eval.EvaluationScore {
	if score >= eval.MateScore-MateDistanceThreshold {
		return score + eval.EvaluationScore(ply)
	}
	if score <= -eval.MateScore+MateDistanceThreshold {
		return score - eval.EvaluationScore(ply)
	}
	return score
}

// scoreFromTT converts a stored node-relative mate score back into a
// root-relative score at the probing node's ply.
func scoreFromTT(score eval.EvaluationScore, ply int) eval.EvaluationScore {
	if score >= eval.MateScore-MateDistanceThreshold {
		return score - eval.EvaluationScore(ply)
	}
	if score <= -eval.MateScore+MateDistanceThreshold {
		return score + eval.EvaluationScore(ply)
	}
	return score
}

// TranspositionEntry represents a single entry in the transposition table.
// Field order matters: Move (4-byte aligned) before Score/DepthAge keeps the
// struct at 16 bytes, so four entries share a 64-byte cache line.
type TranspositionEntry struct {
	Hash     uint64
	Move     uint32
	Score    eval.EvaluationScore
	DepthAge uint8
}

// TranspositionTable implements a hash table for storing search results
type TranspositionTable struct {
	table           []TranspositionEntry
	size            uint64
	mask            uint64
	hits            uint64
	misses          uint64
	collisions      uint64
	secondBucketUse uint64
	totalStores     uint64
	age             uint32
}

// NewTranspositionTable creates a new transposition table with the given size in MB
func NewTranspositionTable(sizeMB int) *TranspositionTable {
	entrySize := uint64(16) // 8+4+2+1 = 15, padded to 16 bytes
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
		tt.table[i] = TranspositionEntry{}
	}

	tt.hits = 0
	tt.misses = 0
	tt.collisions = 0
	tt.secondBucketUse = 0
	tt.totalStores = 0
	tt.age = 0
}

// packDepthAge packs depth (5 bits), entry type (2 bits), and age (1 bit) into a single byte.
// Bit layout: [depth:5][type:2][age:1]
func packDepthAge(depth int, entryType EntryType, age uint32) uint8 {
	return uint8((depth&0x1F)<<3) | uint8((entryType&0x3)<<1) | uint8(age&0x1)
}

// unpackDepthAge extracts depth, entry type, and age from a packed byte.
func unpackDepthAge(depthAge uint8) (depth int, entryType EntryType, age uint32) {
	depth = int((depthAge >> 3) & 0x1F)
	entryType = EntryType((depthAge >> 1) & 0x3)
	age = uint32(depthAge & 0x1)
	return
}

// packMove compresses a board.Move into 32 bits for efficient storage.
// Encodes from square, to square, move type, and special flags (promotion/en passant).
// Note: Piece and Captured fields are not stored and must be derived from board position.
func packMove(move board.Move) uint32 {
	from := uint32(move.From.Rank*8 + move.From.File)
	to := uint32(move.To.Rank*8 + move.To.File)
	// Pack move information using Blunder's approach (32 bits):
	// Bits 31-26: From square (6 bits)
	// Bits 25-20: To square (6 bits)
	// Bits 19-18: Move type (2 bits): 0=Quiet, 1=Attack, 2=Castle, 3=Promotion
	// Bits 17-16: Special flags (2 bits): Promotion piece type or other flags
	// Bits 15-0:  Unused

	packed := (from&0x3F)<<26 | (to&0x3F)<<20

	// Determine move type
	if move.Promotion != board.Empty {
		// Promotion move
		packed |= 3 << 18 // MoveType = Promotion

		// Store only promotion piece type - capture flag will be derived from board position
		var promotionFlag uint32
		switch move.Promotion {
		case board.WhiteKnight, board.BlackKnight:
			promotionFlag = 0 // Knight
		case board.WhiteBishop, board.BlackBishop:
			promotionFlag = 1 // Bishop
		case board.WhiteRook, board.BlackRook:
			promotionFlag = 2 // Rook
		case board.WhiteQueen, board.BlackQueen:
			promotionFlag = 3 // Queen
		}
		packed |= promotionFlag << 16
	} else if move.IsCastling {
		// Castling move
		packed |= 2 << 18 // MoveType = Castle
	} else if move.IsCapture {
		// Attack move (capture)
		packed |= 1 << 18 // MoveType = Attack
		if move.IsEnPassant {
			packed |= 1 << 16 // Flag = EnPassant
		}
	} else {
		// Quiet move
		packed |= 0 << 18 // MoveType = Quiet
	}

	return packed
}

// unpackMove decompresses a 32-bit packed move back into a board.Move structure.
// Reconstructs from square, to square, move type, and special flags.
// Piece and Captured fields remain empty and must be filled from board context.
func unpackMove(packed uint32) board.Move {
	from := int((packed >> 26) & 0x3F)
	to := int((packed >> 20) & 0x3F)
	moveType := (packed >> 18) & 0x3
	flags := (packed >> 16) & 0x3

	move := board.Move{
		From: board.Square{File: from % 8, Rank: from / 8},
		To:   board.Square{File: to % 8, Rank: to / 8},
		// Note: Piece and Captured are not stored - will be derived from board position
	}

	switch moveType {
	case 0: // Quiet
		move.IsCapture = false
		move.IsCastling = false
		move.IsEnPassant = false
		move.Promotion = board.Empty
	case 1: // Attack (capture)
		move.IsCapture = true
		move.IsCastling = false
		move.IsEnPassant = flags == 1 // Flag indicates en passant
		move.Promotion = board.Empty
	case 2: // Castle
		move.IsCapture = false
		move.IsCastling = true
		move.IsEnPassant = false
		move.Promotion = board.Empty
	case 3: // Promotion
		move.IsCapture = false // Will be derived from board position when move is retrieved
		move.IsCastling = false
		move.IsEnPassant = false

		// Infer promotion piece color from destination rank
		// Rank 0 (1st rank) = Black promotion, Rank 7 (8th rank) = White promotion
		isBlackPromotion := move.To.Rank == 0

		// Decode promotion piece type from flags and apply correct color
		switch flags {
		case 0:
			if isBlackPromotion {
				move.Promotion = board.BlackKnight
			} else {
				move.Promotion = board.WhiteKnight
			}
		case 1:
			if isBlackPromotion {
				move.Promotion = board.BlackBishop
			} else {
				move.Promotion = board.WhiteBishop
			}
		case 2:
			if isBlackPromotion {
				move.Promotion = board.BlackRook
			} else {
				move.Promotion = board.WhiteRook
			}
		case 3:
			if isBlackPromotion {
				move.Promotion = board.BlackQueen
			} else {
				move.Promotion = board.WhiteQueen
			}
		}
	}

	// Note: Piece and Captured are not stored - they should be derived from board position
	// Promotion piece color is inferred from destination rank (rank 0=Black, rank 7=White)
	return move
}

// Store stores a position in the transposition table using two-bucket collision resolution
func (tt *TranspositionTable) Store(hash uint64, depth int, score eval.EvaluationScore,
	entryType EntryType, bestMove board.Move) {

	// Try first bucket
	firstIndex := tt.getFirstBucketIndex(hash)
	if tt.shouldReplace(&tt.table[firstIndex], hash, depth) {
		tt.storeEntry(&tt.table[firstIndex], hash, depth, score, entryType, bestMove)
		tt.totalStores++
		return
	}

	// Try second bucket
	secondIndex := tt.getSecondBucketIndex(firstIndex)
	if tt.shouldReplace(&tt.table[secondIndex], hash, depth) {
		tt.storeEntry(&tt.table[secondIndex], hash, depth, score, entryType, bestMove)
		tt.secondBucketUse++
		tt.totalStores++
	}
}

func (tt *TranspositionTable) getFirstBucketIndex(hash uint64) uint64 {
	return hash & tt.mask
}

// getSecondBucketIndex calculates the secondary bucket index with wraparound
func (tt *TranspositionTable) getSecondBucketIndex(firstIndex uint64) uint64 {
	return (firstIndex + 1) & tt.mask
}

func (tt *TranspositionTable) storeEntry(entry *TranspositionEntry, hash uint64, depth int,
	score eval.EvaluationScore, entryType EntryType, bestMove board.Move) {
	currentAge := tt.age & 1

	entry.Hash = hash
	entry.Score = score
	entry.Move = packMove(bestMove)
	entry.DepthAge = packDepthAge(depth, entryType, currentAge)
}

func (tt *TranspositionTable) shouldReplace(entry *TranspositionEntry, hash uint64, depth int) bool {
	if entry.Hash == 0 {
		return true
	}

	curDepth, _, curAge := unpackDepthAge(entry.DepthAge)

	if entry.Hash == hash {
		return depth > curDepth
	}

	// At this point we have a hash collision - different position wants this slot
	tt.collisions++

	return (tt.age & 1) != curAge
}

// Probe looks up a position in the transposition table using two-bucket collision resolution
// Optimized version with inlined probeEntry logic and minimal statistics overhead
func (tt *TranspositionTable) Probe(hash uint64) (*TranspositionEntry, bool) {
	// Check first bucket (inline probeEntry logic)
	firstIndex := hash & tt.mask
	entry := &tt.table[firstIndex]
	if entry.Hash != 0 && entry.Hash == hash {
		// Use hash for sampling to avoid counter increment
		if (hash & 0xFF) == 0 {
			tt.hits += 256
		}
		return entry, true
	}

	// Check second bucket (+1 offset typically in same cache line)
	secondIndex := (firstIndex + 1) & tt.mask
	entry = &tt.table[secondIndex]
	if entry.Hash != 0 && entry.Hash == hash {
		if (hash & 0xFF) == 0 {
			tt.hits += 256
		}
		return entry, true
	}

	// Track misses with sampling
	if (hash & 0xFF) == 0 {
		tt.misses += 256
	}
	return nil, false
}

// GetDepth extracts depth from a transposition entry
func (entry *TranspositionEntry) GetDepth() int {
	depth, _, _ := unpackDepthAge(entry.DepthAge)
	return depth
}

// GetType extracts the entry type from a transposition entry
func (entry *TranspositionEntry) GetType() EntryType {
	_, entryType, _ := unpackDepthAge(entry.DepthAge)
	return entryType
}

// GetMove extracts the move from a transposition entry
func (entry *TranspositionEntry) GetMove() board.Move {
	return unpackMove(entry.Move)
}

// IncrementAge increments the age counter (call at start of each search)
func (tt *TranspositionTable) IncrementAge() {
	tt.age++
}

// TTStats aggregates all transposition table statistics: basic hit/miss
// counters and two-bucket collision resolution stats.
type TTStats struct {
	Hits             uint64
	Misses           uint64
	Collisions       uint64
	SecondBucketUse  uint64
	HitRate          float64
	SecondBucketRate float64
}

// Stats returns aggregated statistics about table usage: hit/miss counts and
// two-bucket collision resolution. It deliberately avoids scanning the table
// (an O(table size) walk over what can be millions of entries) since this is
// called on every completed search; fill-rate/average-depth style stats that
// require a full scan belong in a separate, explicitly-opt-in method.
func (tt *TranspositionTable) Stats() TTStats {
	var stats TTStats
	stats.Hits = tt.hits
	stats.Misses = tt.misses
	stats.Collisions = tt.collisions

	total := stats.Hits + stats.Misses
	if total > 0 {
		stats.HitRate = float64(stats.Hits) / float64(total) * 100
	}

	stats.SecondBucketUse = tt.secondBucketUse
	if tt.totalStores > 0 {
		stats.SecondBucketRate = float64(stats.SecondBucketUse) / float64(tt.totalStores) * 100
	}

	return stats
}

// GetSize returns the size of the transposition table
func (tt *TranspositionTable) GetSize() uint64 {
	return tt.size
}
