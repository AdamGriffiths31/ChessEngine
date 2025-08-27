// Package modes implements various game execution modes for the chess engine.
package modes

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/epd"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/evaluation"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/search"
	"github.com/AdamGriffiths31/ChessEngine/ui"
)

// STSMode implements Strategic Test Suite benchmark functionality
type STSMode struct {
	prompter *ui.Prompter
	rootPath string
}

// STSConfig holds configuration for STS benchmark runs
type STSConfig struct {
	Depth        int
	Timeout      int
	MaxPositions int
	EPDDir       string
	ResultsFile  string
	Verbose      bool
}

// STSResults contains the results of an STS benchmark run
type STSResults struct {
	Files           []string
	TotalScore      int
	MaxScore        int
	PositionsTested int
	CorrectMoves    int
	TotalTime       time.Duration
	AverageDepth    float64
	NodesPerSecond  int64
	ScorePercent    float64
	STSRating       int
	GitCommit       string
	Config          STSConfig
}

// NewSTSMode creates a new STS benchmark mode
func NewSTSMode() *STSMode {
	rootPath, err := os.Getwd()
	if err != nil {
		rootPath = "."
	}

	prompter := ui.NewPrompter()

	return &STSMode{
		prompter: prompter,
		rootPath: rootPath,
	}
}

// Run executes the STS benchmark mode
func (sm *STSMode) Run() error {
	fmt.Println("STS (Strategic Test Suite) Benchmark Runner with Results Recording")
	fmt.Println("==================================================================")
	fmt.Println("")

	config, err := sm.getSTSConfig()
	if err != nil {
		return fmt.Errorf("failed to get STS configuration: %w", err)
	}

	sm.showConfiguration(config)

	epdFiles, err := sm.findSTSFiles(config.EPDDir)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d STS files:\n", len(epdFiles))
	for _, file := range epdFiles {
		fmt.Printf("  %s\n", filepath.Base(file))
	}
	totalPositions := config.MaxPositions * len(epdFiles)
	fmt.Printf("Total positions to test: %d (%d per file)\n", totalPositions, config.MaxPositions)
	fmt.Println("")

	proceed, err := sm.prompter.PromptForConfirmation("Proceed with STS benchmark?", true)
	if err != nil {
		return err
	}

	if !proceed {
		fmt.Println("STS benchmark cancelled.")
		return nil
	}

	results, err := sm.runSTSBenchmark(config, epdFiles)
	if err != nil {
		return err
	}

	sm.displayResults(results)

	if err := sm.saveResults(results); err != nil {
		fmt.Printf("Warning: Failed to save results: %v\n", err)
	} else {
		fmt.Printf("\nResults saved to %s\n", config.ResultsFile)
	}

	return nil
}

func (sm *STSMode) getSTSConfig() (STSConfig, error) {
	config := STSConfig{
		Depth:        999,
		Timeout:      5,
		MaxPositions: 10,
		EPDDir:       "testdata",
		ResultsFile:  "sts_history.md",
		Verbose:      true,
	}

	timeout, err := sm.prompter.PromptForNumber("Enter timeout per position in seconds", 1, 3600)
	if err != nil {
		return config, err
	}
	config.Timeout = timeout

	maxPositions, err := sm.prompter.PromptForNumber("Enter max positions per file", 1, 1000)
	if err != nil {
		return config, err
	}
	config.MaxPositions = maxPositions

	verbose, err := sm.prompter.PromptForConfirmation("Enable verbose output?", true)
	if err != nil {
		return config, err
	}
	config.Verbose = verbose

	return config, nil
}

func (sm *STSMode) showConfiguration(config STSConfig) {
	fmt.Println("Configuration:")
	fmt.Printf("  Search Depth: %d\n", config.Depth)
	fmt.Printf("  Timeout: %ds per position\n", config.Timeout)
	fmt.Printf("  Max Positions: %d per file\n", config.MaxPositions)
	fmt.Printf("  EPD Directory: %s\n", config.EPDDir)
	fmt.Printf("  Results File: %s\n", config.ResultsFile)
	fmt.Printf("  Verbose: %v\n", config.Verbose)
	fmt.Println("")
}

func (sm *STSMode) findSTSFiles(epdDir string) ([]string, error) {
	if _, err := os.Stat(epdDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("EPD directory '%s' not found", epdDir)
	}

	pattern := filepath.Join(epdDir, "STS*.epd")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find STS files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no STS*.epd files found in '%s'", epdDir)
	}

	sort.Strings(files)
	return files, nil
}

func (sm *STSMode) runSTSBenchmark(config STSConfig, epdFiles []string) (*STSResults, error) {
	fmt.Println("Running STS benchmark...")
	fmt.Println("=======================")

	startTime := time.Now()

	engine := search.NewMinimaxEngine()
	evaluator := evaluation.NewEvaluator()
	engine.SetEvaluator(evaluator)
	engine.SetTranspositionTableSize(256)

	searchConfig := ai.SearchConfig{
		MaxDepth:    config.Depth,
		MaxTime:     time.Duration(config.Timeout) * time.Second,
		DebugMode:   false,
		LMRMinDepth: 3,
		LMRMinMoves: 4,
	}

	scorer := epd.NewSTSScorerWithTTClear(engine, searchConfig, config.Verbose, true)

	results := &STSResults{
		Files:  make([]string, 0, len(epdFiles)),
		Config: config,
	}

	var totalNodes int64
	var totalDepth int

	for _, epdFile := range epdFiles {
		fmt.Printf("Processing %s...\n", filepath.Base(epdFile))
		fmt.Println("==================================")

		content, err := os.ReadFile(epdFile) // #nosec G304
		if err != nil {
			return nil, fmt.Errorf("failed to read EPD file %s: %w", epdFile, err)
		}

		positions, err := epd.ParseEPDFile(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse EPD file %s: %w", epdFile, err)
		}

		if config.MaxPositions > 0 && config.MaxPositions < len(positions) {
			positions = positions[:config.MaxPositions]
		}

		ctx := context.Background()
		fileResults := scorer.ScoreSuite(ctx, positions, filepath.Base(epdFile))

		results.Files = append(results.Files, filepath.Base(epdFile))
		results.TotalScore += fileResults.TotalScore
		results.MaxScore += fileResults.MaxScore
		results.PositionsTested += fileResults.PositionCount

		for _, result := range fileResults.Results {
			if result.Score == 10 {
				results.CorrectMoves++
			}
			totalNodes += result.SearchResult.Stats.NodesSearched
			totalDepth += result.SearchResult.Stats.Depth
		}

		fmt.Println("")
	}

	results.TotalTime = time.Since(startTime)
	results.ScorePercent = float64(results.TotalScore) / float64(results.MaxScore) * 100.0
	results.STSRating = sm.calculateSTSRating(results.ScorePercent)
	results.NodesPerSecond = int64(float64(totalNodes) / results.TotalTime.Seconds())
	results.AverageDepth = float64(totalDepth) / float64(results.PositionsTested)

	gitCommit, err := sm.getGitCommit()
	if err != nil {
		results.GitCommit = "unknown"
	} else {
		results.GitCommit = gitCommit
	}

	return results, nil
}

func (sm *STSMode) calculateSTSRating(scorePercent float64) int {
	scorePercentInt := int(scorePercent)

	switch {
	case scorePercentInt >= 90:
		return 3400 + (scorePercentInt-90)*20/10
	case scorePercentInt >= 80:
		return 3200 + (scorePercentInt-80)*20/10
	case scorePercentInt >= 70:
		return 3000 + (scorePercentInt-70)*20/10
	case scorePercentInt >= 60:
		return 2700 + (scorePercentInt-60)*30/10
	case scorePercentInt >= 50:
		return 2400 + (scorePercentInt-50)*30/10
	default:
		return 2000 + scorePercentInt*8/10
	}
}

func (sm *STSMode) getGitCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = sm.rootPath

	output, err := cmd.Output()
	if err != nil {
		return "unknown", err
	}

	return strings.TrimSpace(string(output)), nil
}

func (sm *STSMode) formatNodes(nodes int64) string {
	if nodes >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(nodes)/1000000.0)
	} else if nodes >= 1000 {
		return fmt.Sprintf("%.1fK", float64(nodes)/1000.0)
	}
	return fmt.Sprintf("%d", nodes)
}

func (sm *STSMode) displayResults(results *STSResults) {
	fmt.Println("Aggregated STS Results Summary")
	fmt.Println("==============================")
	fmt.Printf("Files processed: %d (%s/STS*.epd)\n", len(results.Files), results.Config.EPDDir)
	fmt.Printf("Total positions: %d\n", results.PositionsTested)
	fmt.Printf("Total score: %d/%d (%.0f%%)\n", results.TotalScore, results.MaxScore, results.ScorePercent)
	fmt.Printf("Correct moves: %d/%d\n", results.CorrectMoves, results.PositionsTested)
	fmt.Printf("STS Rating: %d\n", results.STSRating)
	fmt.Printf("Total time: %v\n", results.TotalTime.Round(time.Second))

	avgTimePerPosition := results.TotalTime / time.Duration(results.PositionsTested)
	fmt.Printf("Average time per position: %v\n", avgTimePerPosition.Round(time.Millisecond))
	fmt.Printf("Average depth: %.1f\n", results.AverageDepth)
	fmt.Printf("Nodes per second: %s\n", sm.formatNodes(results.NodesPerSecond))

	sm.showPerformanceCategory(results.STSRating)
}

func (sm *STSMode) showPerformanceCategory(rating int) {
	fmt.Printf("Performance Category: ")

	switch {
	case rating >= 3400:
		fmt.Printf("Elite (GM+ level)\n")
	case rating >= 3200:
		fmt.Printf("Very Strong (Strong GM level)\n")
	case rating >= 3000:
		fmt.Printf("Strong (IM+ level)\n")
	case rating >= 2700:
		fmt.Printf("Good (Expert+ level)\n")
	case rating >= 2400:
		fmt.Printf("Decent (Club level)\n")
	default:
		fmt.Printf("Weak (Beginner level)\n")
	}

	fmt.Printf("Reference: Stockfish 8+ typically scores 3300-3400+\n")
}

func (sm *STSMode) saveResults(results *STSResults) error {
	resultsFile := results.Config.ResultsFile

	// Create file if it doesn't exist
	if _, err := os.Stat(resultsFile); os.IsNotExist(err) {
		err := sm.createResultsFile(resultsFile)
		if err != nil {
			return err
		}
	}

	// Append new result
	f, err := os.OpenFile(resultsFile, os.O_APPEND|os.O_WRONLY, 0600) // #nosec G304
	if err != nil {
		return err
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04")
	epdDescription := fmt.Sprintf("STS1-%d (%d files)", len(results.Files), len(results.Files))
	avgTime := results.TotalTime / time.Duration(results.PositionsTested)
	notes := fmt.Sprintf("depth=%d, timeout=%ds, %d per file, %d files",
		results.Config.Depth, results.Config.Timeout, results.Config.MaxPositions, len(results.Files))

	npsFormatted := sm.formatNodes(results.NodesPerSecond)
	line := fmt.Sprintf("| %s | %s | %s | %d | %d | %d | %.0f%% | %d | %d | %d | %ds | %v | %v | %s | %.1f | %s |\n",
		timestamp,
		results.GitCommit,
		epdDescription,
		results.PositionsTested,
		results.TotalScore,
		results.MaxScore,
		results.ScorePercent,
		results.STSRating,
		results.CorrectMoves,
		results.Config.Depth,
		results.Config.Timeout,
		avgTime.Round(time.Millisecond),
		results.TotalTime.Round(time.Second),
		npsFormatted,
		results.AverageDepth,
		notes)

	_, err = f.WriteString(line)
	return err
}

func (sm *STSMode) createResultsFile(filename string) error {
	f, err := os.Create(filename) // #nosec G304
	if err != nil {
		return err
	}
	defer f.Close()

	header := `# STS Benchmark Results History

This file tracks the STS (Strategic Test Suite) performance of ChessEngine over time.

## Results Summary

| Date | Commit | EPD File | Positions | Score | Max | Percent | STS Rating | Correct | Depth | Timeout | Avg Time | Total Time | NPS | Avg Depth | Notes |
|------|--------|----------|-----------|-------|-----|---------|------------|---------|-------|---------|----------|------------|-----|-----------|-------|
`

	_, err = f.WriteString(header)
	return err
}
