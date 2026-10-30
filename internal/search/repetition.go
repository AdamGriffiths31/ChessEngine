package search

func (m *MinimaxEngine) setupRepetitionHistory(rootHash uint64) {
	m.zobristHistoryPly = 0
	m.repetitionRootPly = 0
	m.zobristHistory[m.zobristHistoryPly] = rootHash
}

func (m *MinimaxEngine) seedRepetitionHistory(history []uint64) {
	if len(history) > MaxGamePly {
		history = history[len(history)-MaxGamePly:]
	}
	copy(m.zobristHistory[:len(history)], history)
	m.zobristHistoryPly = uint16(len(history) - 1)
	m.repetitionRootPly = m.zobristHistoryPly
}

func (m *MinimaxEngine) addHistory(hash uint64) {
	if m.zobristHistoryPly < MaxGamePly-1 {
		m.zobristHistoryPly++
		m.zobristHistory[m.zobristHistoryPly] = hash
	}
}

func (m *MinimaxEngine) removeHistory() {
	if m.zobristHistoryPly > 0 {
		m.zobristHistoryPly--
	}
}

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
