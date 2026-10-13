// Package search provides chess move search algorithms and transposition table implementation.
package search

// setupRepetitionHistory initializes repetition detection with the root position hash.
// This should be called at the start of each search to establish the baseline for
// repetition detection during the current search tree exploration.
func (m *MinimaxEngine) setupRepetitionHistory(rootHash uint64) {
	m.zobristHistoryPly = 0
	m.repetitionRootPly = 0
	m.zobristHistory[m.zobristHistoryPly] = rootHash
}

// seedRepetitionHistory seeds the engine's repetition history with a
// real-game hash sequence (oldest first, ending with the current position),
// instead of collapsing it to just the current root. See
// Config.RepetitionHistory for the full contract. Sequences longer
// than the array capacity are truncated to their most recent MaxGamePly
// entries; older positions cannot affect repetition detection near the
// current search horizon.
func (m *MinimaxEngine) seedRepetitionHistory(history []uint64) {
	if len(history) > MaxGamePly {
		history = history[len(history)-MaxGamePly:]
	}
	copy(m.zobristHistory[:len(history)], history)
	m.zobristHistoryPly = uint16(len(history) - 1)
	m.repetitionRootPly = m.zobristHistoryPly
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

// isDrawByRepetition reports whether the position just entered (its hash is the
// newest history entry) is a repetition the search should score as a draw.
//
// A repeat of the root or of anything below it is a draw on the first match:
// the side that could avoid the repeat is assumed not to, as engines
// conventionally do. A match before the root is a real game position, and the
// opponent need not walk into a second visit, so it only counts once the
// position has already occurred twice (this would be its third occurrence).
// Treating one earlier visit as a draw makes the engine see a free draw in
// every line that returns to a previous position and drift into worse ones.
func (m *MinimaxEngine) isDrawByRepetition(currentHash uint64) bool {
	gameMatches := 0
	for repPly := range m.zobristHistoryPly {
		if m.zobristHistory[repPly] != currentHash {
			continue
		}
		if repPly >= m.repetitionRootPly {
			return true
		}
		gameMatches++
	}
	return gameMatches >= 2
}
