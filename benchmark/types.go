// Package benchmark provides interactive chess engine benchmarking functionality.
package benchmark

import "time"

// Engine represents a chess engine configuration.
type Engine struct {
	Name             string            `json:"name"`
	Command          string            `json:"command"`
	Protocol         string            `json:"protocol"`
	WorkingDirectory string            `json:"workingDirectory"`
	Options          map[string]string `json:"options"`
}

// EngineConfig holds the configuration for all available engines.
type EngineConfig struct {
	Engines []Engine `json:"engines"`
}

// TimeControl represents different time control options.
type TimeControl struct {
	Name        string
	Value       string
	Description string
}

// Settings holds all the user-selected benchmark configuration.
type Settings struct {
	OpponentName string
	Opponent     Engine
	TimeControl  TimeControl
	GameCount    int
	Notes        string
	RecordData   bool
}

// GameResult represents the outcome of a single game.
type GameResult struct {
	Result   string // "1-0", "0-1", "1/2-1/2"
	IsWhite  bool   // true if ChessEngine was white
	Moves    int
	Duration time.Duration
}

// Result contains the complete results of a benchmark run.
type Result struct {
	Settings     Settings
	TotalGames   int
	Wins         int
	Losses       int
	Draws        int
	Score        int // Percentage score
	Duration     time.Duration
	IllegalMoves bool
	PGNFile      string
	Timestamp    string
}

// GetAvailableTimeControls returns the standard time control options.
func GetAvailableTimeControls() []TimeControl {
	return []TimeControl{
		{"Bullet", "2:00+2", "Bullet (2+2) - 2 minutes + 2 second increment"},
		{"Blitz", "3:00+0", "Blitz (3+0) - 3 minutes per game"},
		{"Rapid", "10:00+0", "Rapid (10+0) - 10 minutes per game"},
		{"Long", "30:00+0", "Long (30+0) - 30 minutes per game"},
		{"Fixed30", "st=30", "Fixed time (30s per move)"},
		{"Fixed60", "st=60", "Fixed time (60s per move)"},
	}
}

// GetAvailableGameCounts returns the standard game count options.
func GetAvailableGameCounts() []struct {
	Name  string
	Count int
} {
	return []struct {
		Name  string
		Count int
	}{
		{"Quick test", 3},
		{"Short session", 10},
		{"Medium session", 25},
		{"Long session", 50},
		{"Extended session", 100},
	}
}
