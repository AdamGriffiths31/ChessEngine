// Package bench provides chess engine benchmarking utilities: Elo estimation,
// Strategic Test Suite (STS) scoring, and profiling.
package bench

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
