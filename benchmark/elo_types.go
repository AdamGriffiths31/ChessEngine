package benchmark

import "time"

// EloConfig configures an Elo benchmark run.
type EloConfig struct {
	AnchorCenter  int // default 1500
	AnchorCount   int // default 5
	AnchorStep    int // default 100
	GamesPerLevel int // default 2
	TimeMs        int // default 60000 (1 minute base time)
	IncMs         int // default 1000 (1 second increment)
	MaxPlies      int // default 300, safety cap
}

// EloGameResult holds the outcome of a single game against one anchor.
type EloGameResult struct {
	AnchorElo    int
	Color        string // "white" or "black"
	OpeningIndex int
	Result       float64 // 1.0 win, 0.5 draw, 0.0 loss
	Plies        int
	Nodes        uint64
	AvgDepth     int
}

// EloResult holds the complete outcome of an Elo benchmark run.
type EloResult struct {
	EstimatedElo float64
	Anchors      []int
	Games        []EloGameResult
	Config       EloConfig
	Duration     time.Duration
	Timestamp    string
}
