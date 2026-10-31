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

func isMateScore(score eval.EvaluationScore) bool {
	return score >= eval.MateScore-MateDistanceThreshold || score <= -eval.MateScore+MateDistanceThreshold
}

func scoreToTT(score eval.EvaluationScore, ply int) eval.EvaluationScore {
	if score >= eval.MateScore-MateDistanceThreshold {
		return score + eval.EvaluationScore(ply)
	}
	if score <= -eval.MateScore+MateDistanceThreshold {
		return score - eval.EvaluationScore(ply)
	}
	return score
}

func scoreFromTT(score eval.EvaluationScore, ply int) eval.EvaluationScore {
	return scoreToTT(score, -ply)
}

// TranspositionEntry represents a single entry in the transposition table.
type TranspositionEntry struct {
	Hash     uint64
	Move     uint32
	Score    eval.EvaluationScore
	DepthAge uint8
}

// TranspositionTable implements a hash table for storing search results
type TranspositionTable struct {
	table           []TranspositionEntry
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
	entrySize := uint64(16)
	numEntries := (uint64(sizeMB) * 1024 * 1024) / entrySize

	size := uint64(1)
	for size*2 <= numEntries {
		size *= 2
	}

	return &TranspositionTable{
		table: make([]TranspositionEntry, size),
		mask:  size - 1,
	}
}

// Clear clears all entries in the transposition table
func (tt *TranspositionTable) Clear() {
	clear(tt.table)

	tt.hits = 0
	tt.misses = 0
	tt.collisions = 0
	tt.secondBucketUse = 0
	tt.totalStores = 0
	tt.age = 0
}

func packDepthAge(depth int, entryType EntryType, age uint32) uint8 {
	if depth > 0x0F {
		depth = 0x0F
	}
	return uint8((depth&0x0F)<<4) | uint8((entryType&0x3)<<2) | uint8(age&0x3)
}

func unpackDepthAge(depthAge uint8) (depth int, entryType EntryType, age uint32) {
	depth = int((depthAge >> 4) & 0x0F)
	entryType = EntryType((depthAge >> 2) & 0x3)
	age = uint32(depthAge & 0x3)
	return
}

func packMove(move board.Move) uint32 {
	from := uint32(board.FileRankToSquare(move.From.File, move.From.Rank))
	to := uint32(board.FileRankToSquare(move.To.File, move.To.Rank))

	packed := (from&0x3F)<<26 | (to&0x3F)<<20

	switch {
	case move.Promotion != board.Empty:
		packed |= 3 << 18

		var promotionFlag uint32
		switch move.Promotion {
		case board.WhiteKnight, board.BlackKnight:
			promotionFlag = 0
		case board.WhiteBishop, board.BlackBishop:
			promotionFlag = 1
		case board.WhiteRook, board.BlackRook:
			promotionFlag = 2
		case board.WhiteQueen, board.BlackQueen:
			promotionFlag = 3
		}
		packed |= promotionFlag << 16
	case move.IsCastling:
		packed |= 2 << 18
	case move.IsCapture:
		packed |= 1 << 18
		if move.IsEnPassant {
			packed |= 1 << 16
		}
	}

	return packed
}

var promotionPieces = [2][4]board.Piece{
	{board.WhiteKnight, board.WhiteBishop, board.WhiteRook, board.WhiteQueen},
	{board.BlackKnight, board.BlackBishop, board.BlackRook, board.BlackQueen},
}

func unpackMove(packed uint32) board.Move {
	from := int((packed >> 26) & 0x3F)
	to := int((packed >> 20) & 0x3F)
	moveType := (packed >> 18) & 0x3
	flags := (packed >> 16) & 0x3

	move := board.Move{
		From:      board.Square{File: from % 8, Rank: from / 8},
		To:        board.Square{File: to % 8, Rank: to / 8},
		Promotion: board.Empty,
	}

	switch moveType {
	case 1:
		move.IsCapture = true
		move.IsEnPassant = flags == 1
	case 2:
		move.IsCastling = true
	case 3:
		side := 0
		if move.To.Rank == 0 {
			side = 1
		}
		move.Promotion = promotionPieces[side][flags]
	}

	return move
}

// Store stores a position in the transposition table using two-bucket collision resolution
func (tt *TranspositionTable) Store(hash uint64, depth int, score eval.EvaluationScore,
	entryType EntryType, bestMove board.Move) {

	firstIndex := tt.getFirstBucketIndex(hash)
	if tt.shouldReplace(&tt.table[firstIndex], hash, depth) {
		tt.storeEntry(&tt.table[firstIndex], hash, depth, score, entryType, bestMove)
		tt.totalStores++
		return
	}

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

func (tt *TranspositionTable) getSecondBucketIndex(firstIndex uint64) uint64 {
	return (firstIndex + 1) & tt.mask
}

func (tt *TranspositionTable) storeEntry(entry *TranspositionEntry, hash uint64, depth int,
	score eval.EvaluationScore, entryType EntryType, bestMove board.Move) {
	currentAge := tt.age & 3

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

	tt.collisions++

	return (tt.age & 3) != curAge
}

// Probe looks up a position in the transposition table using two-bucket collision resolution
func (tt *TranspositionTable) Probe(hash uint64) (*TranspositionEntry, bool) {
	firstIndex := tt.getFirstBucketIndex(hash)
	for _, index := range [2]uint64{firstIndex, tt.getSecondBucketIndex(firstIndex)} {
		entry := &tt.table[index]
		if entry.Hash != 0 && entry.Hash == hash {
			if (hash & 0xFF) == 0 {
				tt.hits += 256
			}
			return entry, true
		}
	}

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
// two-bucket collision resolution.
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
