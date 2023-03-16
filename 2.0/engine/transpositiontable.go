package engine

import (
	"fmt"
	"sync/atomic"
	"unsafe"

	"github.com/AdamGriffiths31/ChessEngine/data"
)

var TranspositionTable = NewCache()

type CacheEntry struct {
	Age     int
	SMPData uint64
	SMPKey  uint64
	Gate    int32
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

func (c *Cache) Probe(key uint64) int {
	index := key % uint64(c.NumberEntries)
	testKey := key ^ c.CacheTable[index].SMPData
	if testKey == c.CacheTable[index].SMPKey {
		return extractMove(c.CacheTable[index].SMPData)
	}
	return data.NoMove
}

func (c *Cache) Store(key uint64, play int, move, score, flag, depth int) {
	index := key % uint64(c.NumberEntries)
	replace := false

	oldValue := c.CacheTable[index]
	if atomic.CompareAndSwapInt32(&oldValue.Gate, 0, 1) {
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
		atomic.StoreInt32(&oldValue.Gate, 0)
	}
}

func (c *Cache) Get(key uint64, play int, move *int, score *int, alpha, beta, depth int) bool {
	index := key % uint64(c.NumberEntries)
	entry := c.CacheTable[index]
	if atomic.CompareAndSwapInt32(&entry.Gate, 0, 1) {
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
					}
				case data.PVBeta:
					if *score >= beta {
						*score = beta
					}
				}
				atomic.StoreInt32(&entry.Gate, 0)
				return true
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

func foldData(score, depth, flag uint64, move int) uint64 {
	return (score + data.Infinite) | (depth << 16) | (flag << 23) | (uint64(move) << 25)
}

func NewCache() *Cache {
	size := ((0x100000 * 64) / int(unsafe.Sizeof(CacheEntry{})))
	length := size - 2
	fmt.Printf("%v cache size\n", length)
	return &Cache{make([]CacheEntry, length), length, 0, 0, 0, 0}
}
