package engine

import (
	"fmt"
	"unsafe"

	"github.com/AdamGriffiths31/ChessEngine/data"
)

var TranspositionTable = NewCache()

type CacheEntry struct {
	Age     int
	SMPData uint64
	SMPKey  uint64
}

type Cache struct {
	CacheTable    []CacheEntry
	NumberEntries int
	Hit           int
	Cut           int
	CurrentAge    int
	Stored        int
}

func (c *Cache) BestMove(key uint64, play int) int {
	return c.Probe(key)
}

// Probe for the given Position key return the move stored in the TT
func (c *Cache) Probe(key uint64) int {
	index := key % uint64(c.NumberEntries)
	testKey := key ^ c.CacheTable[index].SMPData
	if testKey == c.CacheTable[index].SMPKey {
		return extractMove(c.CacheTable[index].SMPData)
	}
	return data.NoMove
}

// Store Attempts to store the vale in the TT if a value is not already present or
// the depth of the move is greater than the original
func (c *Cache) Store(key uint64, play int, move, score, flag, depth int) {
	index := key % uint64(c.NumberEntries)
	replace := false

	oldValue := c.CacheTable[index]
	oldKey := oldValue.SMPKey
	oldData := oldValue.SMPData
	oldPosKey := oldKey ^ oldData

	if oldData == 0 {
		replace = true
	} else if oldPosKey == key {
		replace = (flag == data.PVExact) || (uint64(depth) >= extractDepth(c.CacheTable[index].SMPData)-3)
	} else {
		if c.CacheTable[index].Age < c.CurrentAge {
			replace = true
		} else if extractDepth(c.CacheTable[index].SMPData) <= uint64(depth) {
			replace = true
		}
	}

	if replace {
		c.Stored++
		if score > data.Mate {
			score += play
		} else if score < -data.Mate {
			score -= play
		}
		smpData := foldData(uint64(score), uint64(depth), uint64(flag), move)
		smpKey := key ^ smpData
		c.CacheTable[index].Age = c.CurrentAge
		c.CacheTable[index].SMPData = smpData
		c.CacheTable[index].SMPKey = smpKey
	}
}

// Get searches the TT for the given Position key for a move
func (c *Cache) Get(key uint64, play int, move *int, score *int, alpha, beta, depth int) bool {
	index := key % uint64(c.NumberEntries)
	entry := c.CacheTable[index]
	testKey := key ^ entry.SMPData
	if testKey == c.CacheTable[index].SMPKey {
		*move = extractMove(entry.SMPData)
		if int(extractDepth(entry.SMPData)) >= depth {
			c.Hit++
			*score = int(extractScore(entry.SMPData))
			if *score > data.Mate {
				*score -= play
			} else if *score < -data.Mate {
				*score += play
			}
			switch extractFlag(entry.SMPData) {
			case data.PVAlpha:
				if *score <= alpha {
					*score = alpha
					return true
				}
			case data.PVBeta:
				if *score >= beta {
					*score = beta
					return true
				}
			case data.PVExact:
				return true
			default:
				panic(fmt.Errorf("ProbePvTable: flag was not found"))
			}
		}
	}
	return false
}

func extractMove(value uint64) int {
	return int(value >> 25)
}

func extractScore(value uint64) uint64 {
	return value&0xFFFF - data.Infinite
}

func extractDepth(value uint64) uint64 {
	return (value >> 16) & 0x3F
}

func extractFlag(value uint64) uint64 {
	return (value >> 23) & 0x3
}

// foldData hashes the data into a unique key
func foldData(score, depth, flag uint64, move int) uint64 {
	return (score + data.Infinite) | (depth << 16) | (flag << 23) | (uint64(move) << 25)
}

// NewCache allocates the space for a new cache
func NewCache() *Cache {
	size := ((0x100000 * 64) / int(unsafe.Sizeof(CacheEntry{})))
	length := size - 2

	return &Cache{make([]CacheEntry, length), length, 0, 0, 0, 0}
}
