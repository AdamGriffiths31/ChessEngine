package bench

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/epd"
	"github.com/AdamGriffiths31/ChessEngine/internal/search"
)

// PositionRecord is one JSONL line describing a single analyzed position. It
// is written by the STS runner to tools/results/sts_<timestamp>.jsonl (one
// line per scored EPD position).
type PositionRecord struct {
	FEN      string        `json:"fen"`
	Expected string        `json:"expected"`
	Chosen   string        `json:"chosen"`
	Score    int           `json:"score"`
	Stats    *search.Stats `json:"stats,omitempty"`
	Source   string        `json:"source,omitempty"`
}

// NewSTSPositionRecord builds a JSONL position record from one scored EPD
// position.
func NewSTSPositionRecord(r epd.STSResult) PositionRecord {
	stats := r.Result.Stats
	return PositionRecord{
		FEN:      r.Position.Board.ToFEN(),
		Expected: r.Position.BestMove,
		Chosen:   r.EngineMoveStr,
		Score:    r.Score,
		Stats:    &stats,
	}
}

// STSSummaryRecord is the trailing JSONL line of an STS run. Type is always
// "summary" so a reader can distinguish it from the per-position
// PositionRecord lines that precede it in the same file.
type STSSummaryRecord struct {
	Type            string   `json:"type"`
	Timestamp       string   `json:"timestamp"` // "2006-01-02 15:04"
	GitCommit       string   `json:"git_commit"`
	Files           []string `json:"files"`
	FileDescription string   `json:"file_description"`
	PositionCount   int      `json:"position_count"`
	TotalScore      int      `json:"total_score"`
	MaxScore        int      `json:"max_score"`
	ScorePercent    float64  `json:"score_percent"`
	CorrectMoves    int      `json:"correct_moves"`
	STSRating       int      `json:"sts_rating"`
	Depth           int      `json:"depth"`
	TimeoutSeconds  int      `json:"timeout_seconds"`
	AvgTimeMS       int64    `json:"avg_time_ms"`
	TotalTimeMS     int64    `json:"total_time_ms"`
	NodesPerSecond  float64  `json:"nodes_per_second"`
	AverageDepth    float64  `json:"average_depth"`
	Notes           string   `json:"notes"`
}

// STSAggregates holds derived statistics computed from a set of scored STS
// results.
type STSAggregates struct {
	CorrectMoves int
	TotalNodes   int64
	TotalDepth   int
	AvgDepth     float64
	NodesPerSec  float64
}

// ComputeSTSAggregates derives per-run aggregates from a set of scored
// results and the wall-clock time the run took.
func ComputeSTSAggregates(results []epd.STSResult, totalTime time.Duration) STSAggregates {
	var agg STSAggregates
	for _, r := range results {
		if r.Score == 10 {
			agg.CorrectMoves++
		}
		agg.TotalNodes += r.Result.Stats.NodesSearched
		agg.TotalDepth += r.Result.Stats.Depth
	}
	if len(results) > 0 {
		agg.AvgDepth = float64(agg.TotalDepth) / float64(len(results))
	}
	if totalTime > 0 {
		agg.NodesPerSec = float64(agg.TotalNodes) / totalTime.Seconds()
	}
	return agg
}

// JSONLWriter appends one JSON object per line to a results file: the STS
// runner writes position records followed by a trailing summary record.
type JSONLWriter struct {
	file *os.File
	// Path is the JSONL file's path, for callers to report to the user.
	Path string
}

func newJSONLWriter(rootPath, prefix string) (*JSONLWriter, error) {
	dir := filepath.Join(rootPath, "tools", "results")
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create results directory: %w", err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to open results directory: %w", err)
	}
	defer func() { _ = root.Close() }()

	timestamp := time.Now().Format("20060102_150405")
	baseName := fmt.Sprintf("%s_%s", prefix, timestamp)

	for attempt := range 100 {
		var name string
		if attempt == 0 {
			name = baseName + ".jsonl"
		} else {
			name = baseName + fmt.Sprintf("_%d.jsonl", attempt+1)
		}
		path := filepath.Join(dir, name)
		f, err := root.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
		if err == nil {
			return &JSONLWriter{file: f, Path: path}, nil
		}
		if !os.IsExist(err) {
			return nil, fmt.Errorf("failed to create %s: %w", path, err)
		}
	}
	return nil, fmt.Errorf("failed to create collision-free JSONL file after 100 attempts for %s", filepath.Join(dir, baseName))
}

// NewSTSResultsWriter creates tools/results/sts_<timestamp>.jsonl under
// rootPath.
func NewSTSResultsWriter(rootPath string) (*JSONLWriter, error) {
	return newJSONLWriter(rootPath, "sts")
}

// WriteLine marshals v as JSON and appends it as one line.
func (w *JSONLWriter) WriteLine(v interface{}) error {
	line, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSONL record: %w", err)
	}
	if _, err := w.file.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("failed to write JSONL record to %s: %w", w.Path, err)
	}
	return nil
}

// Close closes the underlying file.
func (w *JSONLWriter) Close() error {
	return w.file.Close()
}

// GitCommit returns the short commit hash of the repository at rootPath, or
// "unknown" if it can't be determined (e.g. not a git checkout).
func GitCommit(rootPath string) string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = rootPath
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}
