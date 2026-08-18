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
// line per scored EPD position) and by the Elo runner to
// tools/results/elo_moves_<timestamp>.jsonl (one line per move ChessEngine
// played across all games in a run).
type PositionRecord struct {
	FEN      string              `json:"fen"`
	Expected string              `json:"expected"`
	Chosen   string              `json:"chosen"`
	Score    int                 `json:"score"`
	Stats    *search.SearchStats `json:"stats,omitempty"`
	Source   string              `json:"source,omitempty"`
}

// NewSTSPositionRecord builds a JSONL position record from one scored EPD
// position.
func NewSTSPositionRecord(r epd.STSResult) PositionRecord {
	stats := r.SearchResult.Stats
	return PositionRecord{
		FEN:      r.Position.Board.ToFEN(),
		Expected: r.Position.BestMove,
		Chosen:   r.EngineMoveStr,
		Score:    r.Score,
		Stats:    &stats,
	}
}

// NewEloMovePositionRecord builds a JSONL position record for one move
// ChessEngine played during an Elo benchmark game. Score is the search's
// raw evaluation (not an STS 0-10 score); Expected is left blank since
// there is no reference best move for a live game.
func NewEloMovePositionRecord(fen, move string, score int, stats search.SearchStats, source string) PositionRecord {
	return PositionRecord{
		FEN:    fen,
		Chosen: move,
		Score:  score,
		Stats:  &stats,
		Source: source,
	}
}

// STSSummaryRecord is the trailing JSONL line of an STS run: aggregate
// statistics used both for the human-readable summary printed at the end of
// a run and to regenerate the sts_history.md-style table (see `bench
// report`). Type is always "summary" so a reader can distinguish it from the
// per-position PositionRecord lines that precede it in the same file.
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

// STSAggregates holds derived statistics computed once from a set of scored
// STS results, shared by the human-readable summary and the JSONL summary
// record so both are computed from the same numbers.
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
		agg.TotalNodes += r.SearchResult.Stats.NodesSearched
		agg.TotalDepth += r.SearchResult.Stats.Depth
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

// newJSONLWriter creates tools/results/<prefix>_<timestamp>.jsonl under
// rootPath and returns a writer for it. If a file with that timestamp already
// exists (collision when two runs start in the same second), it appends a
// numeric suffix (_2, _3, etc.) to the filename before the .jsonl extension,
// up to 100 attempts before returning an error.
func newJSONLWriter(rootPath, prefix string) (*JSONLWriter, error) {
	dir := filepath.Join(rootPath, "tools", "results")
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create results directory: %w", err)
	}
	timestamp := time.Now().Format("20060102_150405")
	basePath := filepath.Join(dir, fmt.Sprintf("%s_%s", prefix, timestamp))

	for attempt := 0; attempt < 100; attempt++ {
		var path string
		if attempt == 0 {
			path = basePath + ".jsonl"
		} else {
			path = basePath + fmt.Sprintf("_%d.jsonl", attempt+1)
		}
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600) // #nosec G304 - path built from rootPath+timestamp, not user input
		if err == nil {
			return &JSONLWriter{file: f, Path: path}, nil
		}
		if !os.IsExist(err) {
			return nil, fmt.Errorf("failed to create %s: %w", path, err)
		}
		// File exists, try next suffix
	}
	return nil, fmt.Errorf("failed to create collision-free JSONL file after 100 attempts for %s", basePath)
}

// NewSTSResultsWriter creates tools/results/sts_<timestamp>.jsonl under
// rootPath.
func NewSTSResultsWriter(rootPath string) (*JSONLWriter, error) {
	return newJSONLWriter(rootPath, "sts")
}

// NewEloMoveResultsWriter creates tools/results/elo_moves_<timestamp>.jsonl
// under rootPath.
func NewEloMoveResultsWriter(rootPath string) (*JSONLWriter, error) {
	return newJSONLWriter(rootPath, "elo_moves")
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
