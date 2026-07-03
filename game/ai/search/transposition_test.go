package search

import (
	"testing"
	"unsafe"
)

// TestTranspositionEntrySize pins the entry to 16 bytes so four entries fit in a
// 64-byte cache line. Probe latency is dominated by the DRAM fetch of the entry,
// so accidental padding growth (e.g. from field reordering) is a real regression.
func TestTranspositionEntrySize(t *testing.T) {
	const wantSize = 16
	if size := unsafe.Sizeof(TranspositionEntry{}); size != wantSize {
		t.Errorf("TranspositionEntry size = %d bytes, want %d", size, wantSize)
	}
}

// TestTranspositionTableEntryCount verifies the table allocates entries based on
// the actual entry size, so the requested memory budget is used, not overshot.
func TestTranspositionTableEntryCount(t *testing.T) {
	const sizeMB = 8
	tt := NewTranspositionTable(sizeMB)

	budgetBytes := uint64(sizeMB) * 1024 * 1024
	entrySize := uint64(unsafe.Sizeof(TranspositionEntry{}))
	usedBytes := tt.size * entrySize

	if usedBytes > budgetBytes {
		t.Errorf("table uses %d bytes, exceeds %dMB budget", usedBytes, sizeMB)
	}
	// The table is sized to the largest power of two within budget; anything
	// under half the budget means the entry-size constant is out of date.
	if usedBytes*2 <= budgetBytes {
		t.Errorf("table uses only %d of %d budget bytes; entry size constant likely stale", usedBytes, budgetBytes)
	}
}
