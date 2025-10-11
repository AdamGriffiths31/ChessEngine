// Package search provides chess move search algorithms and transposition table implementation.
package search

// setupRepetitionHistory initializes repetition detection with the root position hash.
// This should be called at the start of each search to establish the baseline for
// repetition detection during the current search tree exploration.
func (m *MinimaxEngine) setupRepetitionHistory(rootHash uint64) {
	m.zobristHistoryPly = 0
	m.zobristHistory[m.zobristHistoryPly] = rootHash
}

// addHistory adds a position hash to the repetition detection history.
// Called when making a move during search to track the path from root to current node.
// Prevents buffer overflow by checking against MaxGamePly capacity.
func (m *MinimaxEngine) addHistory(hash uint64) {
	if m.zobristHistoryPly < MaxGamePly-1 {
		m.zobristHistoryPly++
		m.zobristHistory[m.zobristHistoryPly] = hash
	}
}

// removeHistory removes the latest hash from repetition detection history.
// Called when unmaking a move during search to maintain correct history state.
// Prevents underflow by checking that history is not empty.
func (m *MinimaxEngine) removeHistory() {
	if m.zobristHistoryPly > 0 {
		m.zobristHistoryPly--
	}
}

// isDrawByRepetition checks if the current position hash appears in the search history.
// Returns true if the position repeats any position from the current search path,
// indicating a draw by repetition according to chess rules. Only checks positions
// in the current search tree, not the full game history.
func (m *MinimaxEngine) isDrawByRepetition(currentHash uint64) bool {
	for repPly := uint16(0); repPly < m.zobristHistoryPly; repPly++ {
		if m.zobristHistory[repPly] == currentHash {
			return true
		}
	}
	return false
}
