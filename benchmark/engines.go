package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// EngineManager handles loading and managing chess engine configurations.
type EngineManager struct {
	config   *EngineConfig
	rootPath string
}

// NewEngineManager creates a new engine manager.
func NewEngineManager(rootPath string) *EngineManager {
	return &EngineManager{
		rootPath: rootPath,
	}
}

// LoadEngines loads engine configurations from engines.json.
func (em *EngineManager) LoadEngines() error {
	configPath := filepath.Join(em.rootPath, "tools", "engines.json")

	data, err := os.ReadFile(configPath) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to read engines.json: %w", err)
	}

	var config EngineConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse engines.json: %w", err)
	}

	em.config = &config
	return nil
}

// GetAvailableEngines returns all available engines.
func (em *EngineManager) GetAvailableEngines() []Engine {
	if em.config == nil {
		return nil
	}
	return em.config.Engines
}

// GetOpponentEngines returns all engines except ChessEngine itself.
func (em *EngineManager) GetOpponentEngines() []Engine {
	engines := em.GetAvailableEngines()
	opponents := make([]Engine, 0, len(engines))

	for _, engine := range engines {
		if engine.Name != "ChessEngine" {
			opponents = append(opponents, engine)
		}
	}

	return opponents
}

// GetChessEngine returns the ChessEngine configuration.
func (em *EngineManager) GetChessEngine() (*Engine, error) {
	engines := em.GetAvailableEngines()

	for _, engine := range engines {
		if engine.Name == "ChessEngine" {
			return &engine, nil
		}
	}

	return nil, fmt.Errorf("ChessEngine not found in configuration")
}

// GetEngineByName returns an engine by its name.
func (em *EngineManager) GetEngineByName(name string) (*Engine, error) {
	engines := em.GetAvailableEngines()

	for _, engine := range engines {
		if engine.Name == name {
			return &engine, nil
		}
	}

	return nil, fmt.Errorf("engine %s not found", name)
}

// FormatEngineOptions returns a slice of engine options for cutechess-cli.
func (em *EngineManager) FormatEngineOptions(engine *Engine) []string {
	if len(engine.Options) == 0 {
		return nil
	}

	options := make([]string, 0, len(engine.Options))
	for key, value := range engine.Options {
		// All options use the same format for simplicity
		options = append(options, fmt.Sprintf("option.%s=%s", key, value))
	}

	return options
}
