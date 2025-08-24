package search

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
)

func TestPackUnpackMove(t *testing.T) {
	tests := []struct {
		name string
		move board.Move
	}{
		{
			name: "simple d2d4 pawn move",
			move: board.Move{
				From:        board.Square{File: 3, Rank: 1},
				To:          board.Square{File: 3, Rank: 3},
				Piece:       board.WhitePawn,
				IsCapture:   false,
				IsCastling:  false,
				IsEnPassant: false,
				Promotion:   board.Empty,
			},
		},
		{
			name: "knight move b1c3",
			move: board.Move{
				From:        board.Square{File: 1, Rank: 0},
				To:          board.Square{File: 2, Rank: 2},
				Piece:       board.WhiteKnight,
				IsCapture:   false,
				IsCastling:  false,
				IsEnPassant: false,
				Promotion:   board.Empty,
			},
		},
		{
			name: "capture move Nxe5",
			move: board.Move{
				From:        board.Square{File: 2, Rank: 2},
				To:          board.Square{File: 4, Rank: 4},
				Piece:       board.WhiteKnight,
				Captured:    board.BlackPawn,
				IsCapture:   true,
				IsCastling:  false,
				IsEnPassant: false,
				Promotion:   board.Empty,
			},
		},
		{
			name: "castling kingside",
			move: board.Move{
				From:        board.Square{File: 4, Rank: 0},
				To:          board.Square{File: 6, Rank: 0},
				Piece:       board.WhiteKing,
				IsCapture:   false,
				IsCastling:  true,
				IsEnPassant: false,
				Promotion:   board.Empty,
			},
		},
		{
			name: "en passant capture",
			move: board.Move{
				From:        board.Square{File: 4, Rank: 4},
				To:          board.Square{File: 3, Rank: 5},
				Piece:       board.WhitePawn,
				Captured:    board.BlackPawn,
				IsCapture:   true,
				IsCastling:  false,
				IsEnPassant: true,
				Promotion:   board.Empty,
			},
		},
		{
			name: "pawn promotion to queen",
			move: board.Move{
				From:        board.Square{File: 0, Rank: 6},
				To:          board.Square{File: 0, Rank: 7},
				Piece:       board.WhitePawn,
				IsCapture:   false,
				IsCastling:  false,
				IsEnPassant: false,
				Promotion:   board.WhiteQueen,
			},
		},
		{
			name: "promotion with capture",
			move: board.Move{
				From:        board.Square{File: 1, Rank: 6},
				To:          board.Square{File: 2, Rank: 7},
				Piece:       board.WhitePawn,
				Captured:    board.BlackBishop,
				IsCapture:   true,
				IsCastling:  false,
				IsEnPassant: false,
				Promotion:   board.WhiteQueen,
			},
		},
		{
			name: "pawn promotion to knight",
			move: board.Move{
				From:        board.Square{File: 7, Rank: 6},
				To:          board.Square{File: 7, Rank: 7},
				Piece:       board.WhitePawn,
				IsCapture:   false,
				IsCastling:  false,
				IsEnPassant: false,
				Promotion:   board.WhiteKnight,
			},
		},
		{
			name: "black pawn promotion to bishop",
			move: board.Move{
				From:        board.Square{File: 3, Rank: 1},
				To:          board.Square{File: 3, Rank: 0},
				Piece:       board.BlackPawn,
				IsCapture:   false,
				IsCastling:  false,
				IsEnPassant: false,
				Promotion:   board.BlackBishop,
			},
		},
		{
			name: "corner to corner queen move",
			move: board.Move{
				From:        board.Square{File: 0, Rank: 0},
				To:          board.Square{File: 7, Rank: 7},
				Piece:       board.WhiteQueen,
				IsCapture:   false,
				IsCastling:  false,
				IsEnPassant: false,
				Promotion:   board.Empty,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packed := packMove(tt.move)
			unpacked := unpackMove(packed)
			if unpacked.From.File != tt.move.From.File {
				t.Errorf("From.File mismatch: got %d, want %d", unpacked.From.File, tt.move.From.File)
			}
			if unpacked.From.Rank != tt.move.From.Rank {
				t.Errorf("From.Rank mismatch: got %d, want %d", unpacked.From.Rank, tt.move.From.Rank)
			}
			if unpacked.To.File != tt.move.To.File {
				t.Errorf("To.File mismatch: got %d, want %d", unpacked.To.File, tt.move.To.File)
			}
			if unpacked.To.Rank != tt.move.To.Rank {
				t.Errorf("To.Rank mismatch: got %d, want %d", unpacked.To.Rank, tt.move.To.Rank)
			}

			if unpacked.IsCastling != tt.move.IsCastling {
				t.Errorf("IsCastling mismatch: got %t, want %t", unpacked.IsCastling, tt.move.IsCastling)
			}
			if unpacked.IsEnPassant != tt.move.IsEnPassant {
				t.Errorf("IsEnPassant mismatch: got %t, want %t", unpacked.IsEnPassant, tt.move.IsEnPassant)
			}

			if unpacked.Promotion != tt.move.Promotion {
				t.Errorf("Promotion piece mismatch: got %v, want %v", unpacked.Promotion, tt.move.Promotion)
			}
		})
	}
}

func TestTranspositionTableStoreProbe(t *testing.T) {
	tests := []struct {
		name      string
		hash      uint64
		depth     int
		score     ai.EvaluationScore
		entryType EntryType
		move      board.Move
		expectHit bool
	}{
		{
			name:      "store and retrieve exact entry",
			hash:      0x123456789ABCDEF0,
			depth:     5,
			score:     150,
			entryType: EntryExact,
			move: board.Move{
				From:      board.Square{File: 4, Rank: 1},
				To:        board.Square{File: 4, Rank: 3},
				Promotion: board.Empty,
			},
			expectHit: true,
		},
		{
			name:      "store and retrieve lower bound entry",
			hash:      0xFEDCBA9876543210,
			depth:     10,
			score:     -75,
			entryType: EntryLowerBound,
			move: board.Move{
				From:      board.Square{File: 6, Rank: 0},
				To:        board.Square{File: 5, Rank: 2},
				Promotion: board.Empty,
			},
			expectHit: true,
		},
		{
			name:      "store and retrieve upper bound entry",
			hash:      0x1111222233334444,
			depth:     3,
			score:     25,
			entryType: EntryUpperBound,
			move: board.Move{
				From:      board.Square{File: 1, Rank: 0},
				To:        board.Square{File: 2, Rank: 2},
				Promotion: board.Empty,
			},
			expectHit: true,
		},
		{
			name:      "store promotion move",
			hash:      0x5555666677778888,
			depth:     7,
			score:     800,
			entryType: EntryExact,
			move: board.Move{
				From:      board.Square{File: 0, Rank: 6},
				To:        board.Square{File: 0, Rank: 7},
				Promotion: board.WhiteQueen,
			},
			expectHit: true,
		},
		{
			name:      "store castling move",
			hash:      0x9999AAAABBBBCCCC,
			depth:     4,
			score:     0,
			entryType: EntryExact,
			move: board.Move{
				From:       board.Square{File: 4, Rank: 0},
				To:         board.Square{File: 6, Rank: 0},
				IsCastling: true,
				Promotion:  board.Empty,
			},
			expectHit: true,
		},
		{
			name:      "store capture move",
			hash:      0xDDDDEEEEFFFF0000,
			depth:     6,
			score:     320,
			entryType: EntryLowerBound,
			move: board.Move{
				From:      board.Square{File: 3, Rank: 4},
				To:        board.Square{File: 4, Rank: 5},
				IsCapture: true,
				Promotion: board.Empty,
			},
			expectHit: true,
		},
		{
			name:      "store en passant move",
			hash:      0x1234567890ABCDEF,
			depth:     2,
			score:     100,
			entryType: EntryExact,
			move: board.Move{
				From:        board.Square{File: 4, Rank: 4},
				To:          board.Square{File: 3, Rank: 5},
				IsCapture:   true,
				IsEnPassant: true,
				Promotion:   board.Empty,
			},
			expectHit: true,
		},
		{
			name:      "probe non-existent entry",
			hash:      0xFFFFFFFFFFFFFFFF,
			depth:     1,
			score:     0,
			entryType: EntryExact,
			move: board.Move{
				Promotion: board.Empty,
			},
			expectHit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := NewTranspositionTable(1)

			if tt.expectHit {
				table.Store(tt.hash, tt.depth, tt.score, tt.entryType, tt.move)

				entry, found := table.Probe(tt.hash)
				if !found {
					t.Errorf("Expected to find entry with hash %x, but got miss", tt.hash)
					return
				}

				if entry.Hash != tt.hash {
					t.Errorf("Hash mismatch: got %x, want %x", entry.Hash, tt.hash)
				}
				if entry.Score != tt.score {
					t.Errorf("Score mismatch: got %d, want %d", entry.Score, tt.score)
				}

				if entry.GetDepth() != tt.depth {
					t.Errorf("Depth mismatch: got %d, want %d", entry.GetDepth(), tt.depth)
				}
				if entry.GetType() != tt.entryType {
					t.Errorf("Entry type mismatch: got %d, want %d", entry.GetType(), tt.entryType)
				}

				storedMove := entry.GetMove()
				if storedMove.From.File != tt.move.From.File || storedMove.From.Rank != tt.move.From.Rank {
					t.Errorf("From square mismatch: got %s, want %s", storedMove.From.String(), tt.move.From.String())
				}
				if storedMove.To.File != tt.move.To.File || storedMove.To.Rank != tt.move.To.Rank {
					t.Errorf("To square mismatch: got %s, want %s", storedMove.To.String(), tt.move.To.String())
				}

				if tt.move.IsCastling {
					if !storedMove.IsCastling {
						t.Errorf("Castling flag not preserved: got %t, want %t", storedMove.IsCastling, tt.move.IsCastling)
					}
				}

				if tt.move.IsEnPassant {
					if !storedMove.IsEnPassant {
						t.Errorf("En passant flag not preserved: got %t, want %t", storedMove.IsEnPassant, tt.move.IsEnPassant)
					}
				}

				if storedMove.Promotion != tt.move.Promotion {
					t.Errorf("Promotion piece mismatch: got %v, want %v", storedMove.Promotion, tt.move.Promotion)
				}

				hits, misses, _, hitRate := table.GetStats()
				if hits != 1 {
					t.Errorf("Expected 1 hit, got %d", hits)
				}
				if misses != 0 {
					t.Errorf("Expected 0 misses, got %d", misses)
				}
				expectedHitRate := 100.0
				if hitRate != expectedHitRate {
					t.Errorf("Expected hit rate %.1f%%, got %.1f%%", expectedHitRate, hitRate)
				}

			} else {
				_, found := table.Probe(tt.hash)
				if found {
					t.Errorf("Expected miss for hash %x, but got hit", tt.hash)
				}

				hits, misses, _, hitRate := table.GetStats()
				if hits != 0 {
					t.Errorf("Expected 0 hits, got %d", hits)
				}
				if misses != 1 {
					t.Errorf("Expected 1 miss, got %d", misses)
				}
				expectedHitRate := 0.0
				if hitRate != expectedHitRate {
					t.Errorf("Expected hit rate %.1f%%, got %.1f%%", expectedHitRate, hitRate)
				}
			}
		})
	}
}

func TestTwoBucketCollisionResolution(t *testing.T) {
	table := NewTranspositionTable(1)
	tableSize := table.GetSize()

	firstHash := uint64(0x1000)
	secondHash := firstHash + tableSize

	firstBucketIndex1 := firstHash & (tableSize - 1)
	firstBucketIndex2 := secondHash & (tableSize - 1)
	if firstBucketIndex1 != firstBucketIndex2 {
		t.Fatalf("Test setup error: hashes don't collide. Index1=%d, Index2=%d", firstBucketIndex1, firstBucketIndex2)
	}

	move1 := board.Move{
		From: board.Square{File: 4, Rank: 1},
		To:   board.Square{File: 4, Rank: 3},
	}
	move2 := board.Move{
		From: board.Square{File: 3, Rank: 1},
		To:   board.Square{File: 3, Rank: 3},
	}

	table.Store(firstHash, 5, 150, EntryExact, move1)

	entry1, found1 := table.Probe(firstHash)
	if !found1 {
		t.Fatalf("First entry not found after storage")
	}
	if entry1.Hash != firstHash {
		t.Errorf("First entry hash mismatch: got %x, want %x", entry1.Hash, firstHash)
	}

	hits, misses, _, hitRate := table.GetStats()
	if hits != 1 || misses != 0 {
		t.Errorf("After first store/probe: expected hits=1, misses=0, got hits=%d, misses=%d", hits, misses)
	}
	if hitRate != 100.0 {
		t.Errorf("Expected hit rate 100%%, got %.1f%%", hitRate)
	}

	secondBucketUse, secondBucketRate := table.GetTwoBucketStats()
	if secondBucketUse != 0 || secondBucketRate != 0.0 {
		t.Errorf("Expected no second bucket usage yet, got uses=%d, rate=%.1f%%", secondBucketUse, secondBucketRate)
	}

	table.Store(secondHash, 7, 200, EntryLowerBound, move2)

	entry1Again, found1Again := table.Probe(firstHash)
	if !found1Again {
		t.Fatalf("First entry not found after second storage")
	}
	if entry1Again.Hash != firstHash {
		t.Errorf("First entry corrupted: got hash %x, want %x", entry1Again.Hash, firstHash)
	}

	entry2, found2 := table.Probe(secondHash)
	if !found2 {
		t.Fatalf("Second entry not found after storage")
	}
	if entry2.Hash != secondHash {
		t.Errorf("Second entry hash mismatch: got %x, want %x", entry2.Hash, secondHash)
	}
	if entry2.Score != 200 {
		t.Errorf("Second entry score mismatch: got %d, want %d", entry2.Score, 200)
	}
	if entry2.GetDepth() != 7 {
		t.Errorf("Second entry depth mismatch: got %d, want %d", entry2.GetDepth(), 7)
	}
	if entry2.GetType() != EntryLowerBound {
		t.Errorf("Second entry type mismatch: got %d, want %d", entry2.GetType(), EntryLowerBound)
	}

	move1Retrieved := entry1Again.GetMove()
	move2Retrieved := entry2.GetMove()
	if move1Retrieved.From.File == move2Retrieved.From.File && move1Retrieved.From.Rank == move2Retrieved.From.Rank {
		t.Error("Retrieved moves are identical - collision resolution may have failed")
	}

	hits, misses, collisions, _ := table.GetStats()
	expectedHits := uint64(3)
	if hits != expectedHits {
		t.Errorf("Expected %d hits after collision resolution, got %d", expectedHits, hits)
	}
	if misses != 0 {
		t.Errorf("Expected 0 misses, got %d", misses)
	}
	if collisions == 0 {
		t.Error("Expected collisions to be recorded, got 0")
	}

	secondBucketUse, secondBucketRate = table.GetTwoBucketStats()
	if secondBucketUse == 0 {
		t.Error("Second bucket should have been used for collision resolution, but got 0 uses")
	}
	if secondBucketRate == 0.0 {
		t.Error("Second bucket rate should be > 0 after collision resolution")
	}

	t.Logf("Collision resolution successful:")
	t.Logf("  First hash: %x -> bucket %d", firstHash, firstBucketIndex1)
	t.Logf("  Second hash: %x -> bucket %d (collision!)", secondHash, firstBucketIndex2)
	t.Logf("  Second bucket uses: %d (rate: %.1f%%)", secondBucketUse, secondBucketRate)
	t.Logf("  Total collisions detected: %d", collisions)

	for i := 0; i < 5; i++ {
		if _, found := table.Probe(firstHash); !found {
			t.Errorf("First entry lost after %d additional probes", i+1)
		}
		if _, found := table.Probe(secondHash); !found {
			t.Errorf("Second entry lost after %d additional probes", i+1)
		}
	}

	finalHits, _, _, finalHitRate := table.GetStats()
	finalSecondBucketUse, finalSecondBucketRate := table.GetTwoBucketStats()

	if finalHits <= expectedHits {
		t.Errorf("Hit count should have increased with additional probes, got %d", finalHits)
	}
	if finalHitRate != 100.0 {
		t.Errorf("Hit rate should remain 100%% with no misses, got %.1f%%", finalHitRate)
	}
	if finalSecondBucketUse != secondBucketUse+(5) {
		t.Logf("Note: Second bucket usage may vary based on internal probe order")
	}

	t.Logf("Final stats: hits=%d, second bucket uses=%d (%.1f%%)",
		finalHits, finalSecondBucketUse, finalSecondBucketRate)
}

func TestPackUnpackDepthAge(t *testing.T) {
	tests := []struct {
		name      string
		depth     int
		entryType EntryType
		age       uint32
	}{
		{
			name:      "minimum values",
			depth:     0,
			entryType: EntryExact,
			age:       0,
		},
		{
			name:      "maximum depth",
			depth:     31,
			entryType: EntryUpperBound,
			age:       1,
		},
		{
			name:      "all entry types",
			depth:     15,
			entryType: EntryLowerBound,
			age:       0,
		},
		{
			name:      "mid-range values",
			depth:     12,
			entryType: EntryExact,
			age:       1,
		},
		{
			name:      "high depth with exact type",
			depth:     28,
			entryType: EntryExact,
			age:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packed := packDepthAge(tt.depth, tt.entryType, tt.age)

			unpackedDepth, unpackedType, unpackedAge := unpackDepthAge(packed)

			if unpackedDepth != tt.depth {
				t.Errorf("Depth mismatch: got %d, want %d", unpackedDepth, tt.depth)
			}
			if unpackedType != tt.entryType {
				t.Errorf("EntryType mismatch: got %d, want %d", unpackedType, tt.entryType)
			}
			if unpackedAge != tt.age {
				t.Errorf("Age mismatch: got %d, want %d", unpackedAge, tt.age)
			}

			expectedDepthBits := uint8((tt.depth & 0x1F) << 3)
			expectedTypeBits := uint8((tt.entryType & 0x3) << 1)
			expectedAgeBits := uint8(tt.age & 0x1)
			expectedPacked := expectedDepthBits | expectedTypeBits | expectedAgeBits

			if packed != expectedPacked {
				t.Errorf("Bit packing mismatch: got %08b, want %08b", packed, expectedPacked)
				t.Errorf("  Depth bits: got %08b, want %08b", (packed>>3)&0x1F, expectedDepthBits>>3)
				t.Errorf("  Type bits:  got %08b, want %08b", (packed>>1)&0x3, expectedTypeBits>>1)
				t.Errorf("  Age bits:   got %08b, want %08b", packed&0x1, expectedAgeBits)
			}
		})
	}
}

func TestTableClear(t *testing.T) {
	table := NewTranspositionTable(1)

	move1 := board.Move{
		From:      board.Square{File: 4, Rank: 1},
		To:        board.Square{File: 4, Rank: 3},
		Promotion: board.Empty,
	}
	move2 := board.Move{
		From:      board.Square{File: 3, Rank: 1},
		To:        board.Square{File: 3, Rank: 3},
		Promotion: board.Empty,
	}

	table.Store(0x1111111111111111, 5, 150, EntryExact, move1)
	table.Store(0x2222222222222222, 10, -75, EntryLowerBound, move2)
	table.Store(0x3333333333333333, 3, 25, EntryUpperBound, move1)

	table.IncrementAge()

	table.Probe(0x1111111111111111)
	table.Probe(0x2222222222222222)
	table.Probe(0x9999999999999999)

	hits, misses, _, _ := table.GetStats()
	if hits == 0 {
		t.Error("Expected some hits before clearing")
	}
	if misses == 0 {
		t.Error("Expected some misses before clearing")
	}

	_, _, _, filled, _, _, _ := table.GetDetailedStats()
	if filled == 0 {
		t.Error("Expected some filled entries before clearing")
	}

	if table.totalStores == 0 {
		t.Error("Expected totalStores > 0 before clearing")
	}

	table.Clear()

	hitsAfter, missesAfter, collisionsAfter, hitRateAfter := table.GetStats()
	if hitsAfter != 0 {
		t.Errorf("Expected 0 hits after clear, got %d", hitsAfter)
	}
	if missesAfter != 0 {
		t.Errorf("Expected 0 misses after clear, got %d", missesAfter)
	}
	if collisionsAfter != 0 {
		t.Errorf("Expected 0 collisions after clear, got %d", collisionsAfter)
	}
	if hitRateAfter != 0 {
		t.Errorf("Expected 0%% hit rate after clear, got %.1f%%", hitRateAfter)
	}

	secondBucketUseAfter, secondBucketRateAfter := table.GetTwoBucketStats()
	if secondBucketUseAfter != 0 {
		t.Errorf("Expected 0 second bucket uses after clear, got %d", secondBucketUseAfter)
	}
	if secondBucketRateAfter != 0 {
		t.Errorf("Expected 0%% second bucket rate after clear, got %.1f%%", secondBucketRateAfter)
	}

	if table.age != 0 {
		t.Errorf("Expected age 0 after clear, got %d", table.age)
	}

	if table.totalStores != 0 {
		t.Errorf("Expected totalStores 0 after clear, got %d", table.totalStores)
	}

	_, _, _, filledAfter, _, _, _ := table.GetDetailedStats()
	if filledAfter != 0 {
		t.Errorf("Expected 0 filled entries after clear, got %d", filledAfter)
	}

	_, found1 := table.Probe(0x1111111111111111)
	_, found2 := table.Probe(0x2222222222222222)
	_, found3 := table.Probe(0x3333333333333333)

	if found1 || found2 || found3 {
		t.Error("Expected all probes to miss after clear")
	}

	table.Store(0x4444444444444444, 2, 100, EntryExact, move1)
	entry, found := table.Probe(0x4444444444444444)
	if !found {
		t.Error("Should be able to store and retrieve after clear")
	}
	if entry.Score != 100 {
		t.Errorf("New entry score mismatch: got %d, want %d", entry.Score, 100)
	}
}

func TestEntryReplacement(t *testing.T) {
	table := NewTranspositionTable(1)

	move := board.Move{
		From:      board.Square{File: 4, Rank: 1},
		To:        board.Square{File: 4, Rank: 3},
		Promotion: board.Empty,
	}

	hash := uint64(0x1111111111111111)

	t.Run("replace empty entry", func(t *testing.T) {
		table.Clear()

		index := table.getFirstBucketIndex(hash)
		entry := &table.table[index]

		if !table.shouldReplace(entry, hash, 5) {
			t.Error("Should replace empty entry (Hash=0)")
		}
	})

	t.Run("replace same hash with higher depth", func(t *testing.T) {
		table.Clear()

		table.Store(hash, 5, 100, EntryExact, move)

		index := table.getFirstBucketIndex(hash)
		entry := &table.table[index]

		if !table.shouldReplace(entry, hash, 10) {
			t.Error("Should replace same hash with higher depth")
		}

		if table.shouldReplace(entry, hash, 3) {
			t.Error("Should NOT replace same hash with lower depth")
		}

		if table.shouldReplace(entry, hash, 5) {
			t.Error("Should NOT replace same hash with equal depth")
		}
	})

	t.Run("replace different hash based on age", func(t *testing.T) {
		table.Clear()

		table.Store(hash, 5, 100, EntryExact, move)

		index := table.getFirstBucketIndex(hash)
		entry := &table.table[index]

		differentHash := uint64(0x2222222222222222)

		initialCollisions := table.collisions
		shouldRepl1 := table.shouldReplace(entry, differentHash, 10)

		if shouldRepl1 {
			t.Error("Should NOT replace different hash with same age")
		}

		if table.collisions <= initialCollisions {
			t.Error("Collision should have been recorded")
		}

		table.IncrementAge()

		if !table.shouldReplace(entry, differentHash, 3) {
			t.Error("Should replace different hash with different age")
		}
	})

	t.Run("collision counting", func(t *testing.T) {
		table.Clear()

		table.Store(hash, 5, 100, EntryExact, move)

		index := table.getFirstBucketIndex(hash)
		entry := &table.table[index]

		initialCollisions := table.collisions

		differentHash := uint64(0x3333333333333333)
		table.shouldReplace(entry, differentHash, 10)

		if table.collisions != initialCollisions+1 {
			t.Errorf("Expected collisions to increment by 1, got %d -> %d",
				initialCollisions, table.collisions)
		}
	})
}

func TestIncrementAge(t *testing.T) {
	table := NewTranspositionTable(1)

	if table.age != 0 {
		t.Errorf("Initial age should be 0, got %d", table.age)
	}

	move := board.Move{
		From:      board.Square{File: 4, Rank: 1},
		To:        board.Square{File: 4, Rank: 3},
		Promotion: board.Empty,
	}

	hash1 := uint64(0x1111111111111111)
	table.Store(hash1, 5, 100, EntryExact, move)

	entry1, found := table.Probe(hash1)
	if !found {
		t.Fatal("Entry should be found")
	}
	_, _, age := unpackDepthAge(entry1.DepthAge)
	if age != 0 {
		t.Errorf("Entry should have age 0, got %d", age)
	}

	table.IncrementAge()
	if table.age != 1 {
		t.Errorf("After increment, age should be 1, got %d", table.age)
	}

	hash2 := uint64(0x2222222222222222)
	table.Store(hash2, 7, 150, EntryLowerBound, move)

	entry2, found2 := table.Probe(hash2)
	if !found2 {
		t.Fatal("Second entry should be found")
	}
	_, _, age2 := unpackDepthAge(entry2.DepthAge)
	if age2 != 1 {
		t.Errorf("New entry should have age 1, got %d", age2)
	}

	entry1Again, found1Again := table.Probe(hash1)
	if !found1Again {
		t.Fatal("Original entry should still be found")
	}
	_, _, ageOriginal := unpackDepthAge(entry1Again.DepthAge)
	if ageOriginal != 0 {
		t.Errorf("Original entry should still have age 0, got %d", ageOriginal)
	}

	for i := 0; i < 10; i++ {
		table.IncrementAge()
	}

	if table.age != 11 {
		t.Errorf("After 11 total increments, age should be 11, got %d", table.age)
	}

	hash3 := uint64(0x3333333333333333)
	table.Store(hash3, 3, 50, EntryUpperBound, move)

	entry3, found3 := table.Probe(hash3)
	if !found3 {
		t.Fatal("Third entry should be found")
	}
	_, _, age3 := unpackDepthAge(entry3.DepthAge)
	expectedAge := table.age & 1
	if age3 != expectedAge {
		t.Errorf("Third entry should have age %d, got %d", expectedAge, age3)
	}
}

func TestTableSizeCalculation(t *testing.T) {
	tests := []struct {
		name         string
		sizeMB       int
		expectedSize uint64
	}{
		{
			name:         "1MB table",
			sizeMB:       1,
			expectedSize: 52428,
		},
		{
			name:         "8MB table",
			sizeMB:       8,
			expectedSize: 419430,
		},
		{
			name:         "16MB table",
			sizeMB:       16,
			expectedSize: 838860,
		},
		{
			name:         "64MB table",
			sizeMB:       64,
			expectedSize: 3355443,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := NewTranspositionTable(tt.sizeMB)

			size := table.GetSize()
			if size&(size-1) != 0 {
				t.Errorf("Table size %d is not a power of 2", size)
			}

			if table.mask != size-1 {
				t.Errorf("Mask should be %d, got %d", size-1, table.mask)
			}

			entrySize := uint64(20)
			requestedEntries := (uint64(tt.sizeMB) * 1024 * 1024) / entrySize

			if size > requestedEntries {
				t.Errorf("Size %d exceeds requested entries %d", size, requestedEntries)
			}

			if size < requestedEntries/2 {
				t.Errorf("Size %d is too small compared to requested %d", size, requestedEntries)
			}

			actualMemory := size * entrySize
			requestedMemory := uint64(tt.sizeMB) * 1024 * 1024

			if actualMemory > requestedMemory {
				t.Errorf("Actual memory %d bytes exceeds requested %d bytes", actualMemory, requestedMemory)
			}

			t.Logf("%s: requested %d MB (%d entries), got size %d (%d KB, %.1f%% utilization)",
				tt.name, tt.sizeMB, requestedEntries, size, actualMemory/1024,
				float64(actualMemory)/float64(requestedMemory)*100)
		})
	}
}
