package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/game"
	"github.com/AdamGriffiths31/ChessEngine/uci"
)

// mateScoreClampCP is the centipawn value mate scores are clamped to.
const mateScoreClampCP = 10000

// BlunderScanConfig configures a blunder scan.
type BlunderScanConfig struct {
	ThresholdCP int // minimum centipawn loss to flag
	MoveTimeMs  int // Stockfish movetime per evaluation
}

// Blunder is one flagged position: the engine played PlayedMove but the
// reference engine preferred BestMove, costing SwingCP centipawns.
type Blunder struct {
	FEN        string
	BestMove   string // coordinate notation, e.g. "g8h6"
	PlayedMove string
	SwingCP    int
	Source     string // run/anchor/game/ply reference
}

// PositionEval evaluates the position reached by playing moveHistory from
// the start position, returning the reference engine's best move (UCI) and
// score in centipawns from the side to move's perspective.
type PositionEval func(moveHistory []string) (bestMoveUCI string, scoreCP int, err error)

// scanGame replays one recorded game and flags every engine move that lost
// at least cfg.ThresholdCP centipawns versus the reference engine's choice.
// Preset opening plies (eloOpenings[g.OpeningIndex]) are skipped: they are
// book moves, not engine decisions.
func scanGame(g EloGameResult, source string, cfg BlunderScanConfig, eval PositionEval) ([]Blunder, error) {
	engine := game.NewEngine()
	converter := uci.NewMoveConverter()

	openingPlies := 0
	if g.OpeningIndex >= 0 && g.OpeningIndex < len(eloOpenings) {
		openingPlies = len(eloOpenings[g.OpeningIndex])
	}

	engineIsWhite := g.Color == "white"
	var blunders []Blunder

	for ply, uciMove := range g.Moves {
		isEngineMove := (ply%2 == 0) == engineIsWhite

		if isEngineMove && ply >= openingPlies {
			fen := engine.GetState().Board.ToFEN()

			best, scoreBefore, err := eval(g.Moves[:ply])
			if err != nil {
				return blunders, fmt.Errorf("ply %d eval failed: %w", ply, err)
			}
			_, scoreAfter, err := eval(g.Moves[:ply+1])
			if err != nil {
				return blunders, fmt.Errorf("ply %d post-move eval failed: %w", ply, err)
			}

			// scoreAfter is from the opponent's perspective; negate it to get
			// the mover's resulting score.
			swing := scoreBefore - (-scoreAfter)
			if swing >= cfg.ThresholdCP && best != uciMove {
				blunders = append(blunders, Blunder{
					FEN:        fen,
					BestMove:   best,
					PlayedMove: uciMove,
					SwingCP:    swing,
					Source:     fmt.Sprintf("%s ply%d", source, ply),
				})
			}
		}

		move, err := converter.FromUCI(uciMove, engine.GetState().Board)
		if err != nil {
			return blunders, fmt.Errorf("recorded move %q at ply %d does not parse: %w", uciMove, ply, err)
		}
		if err := engine.MakeMove(move); err != nil {
			return blunders, fmt.Errorf("recorded move %q at ply %d is illegal: %w", uciMove, ply, err)
		}
	}

	return blunders, nil
}

// formatBlunderEPD renders one EPD line. bm uses coordinate notation because
// the epd scorer compares from+to strings; metadata uses c1 because the epd
// parser mines c0 for move=digit score pairs and would misparse metadata.
func formatBlunderEPD(b Blunder) string {
	fenFields := strings.Fields(b.FEN)
	epdFEN := strings.Join(fenFields[:4], " ")
	return fmt.Sprintf("%s bm %s; c1 \"played %s swing %dcp\"; id \"%s\";",
		epdFEN, b.BestMove, b.PlayedMove, b.SwingCP, b.Source)
}

// parseInfoScore extracts the centipawn score from a UCI info line.
// Mate scores clamp to +/-mateScoreClampCP.
func parseInfoScore(line string) (int, bool) {
	fields := strings.Fields(line)
	if len(fields) == 0 || fields[0] != "info" {
		return 0, false
	}
	for i := 0; i < len(fields)-2; i++ {
		if fields[i] != "score" {
			continue
		}
		value, err := strconv.Atoi(fields[i+2])
		if err != nil {
			return 0, false
		}
		switch fields[i+1] {
		case "cp":
			return value, true
		case "mate":
			if value >= 0 {
				return mateScoreClampCP, true
			}
			return -mateScoreClampCP, true
		}
	}
	return 0, false
}

// ScanSummary reports what a blunder scan covered.
type ScanSummary struct {
	RunsScanned   int
	GamesScanned  int
	GamesSkipped  int
	BlundersFound int
}

// ScanRecordedGames reads Elo runs from a JSONL file, scans every game that
// has a recorded move list, and appends flagged positions to outPath as EPD.
// newEval creates the reference evaluator (and its cleanup) once per run so
// a wedged engine process doesn't poison subsequent runs. Output is written
// incrementally: a crash keeps everything scanned so far.
func ScanRecordedGames(inputPath, outPath string, cfg BlunderScanConfig, newEval func() (PositionEval, func(), error)) (ScanSummary, error) {
	var summary ScanSummary

	data, err := os.ReadFile(inputPath) // #nosec G304 - path from CLI flag
	if err != nil {
		return summary, fmt.Errorf("failed to read %s: %w", inputPath, err)
	}

	out, err := os.OpenFile(outPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600) // #nosec G304 - path from CLI flag
	if err != nil {
		return summary, fmt.Errorf("failed to open %s: %w", outPath, err)
	}
	defer func() { _ = out.Close() }()

	for lineNum, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var run EloResult
		if err := json.Unmarshal([]byte(line), &run); err != nil {
			fmt.Printf("WARNING: skipping unparseable line %d of %s: %v\n", lineNum+1, inputPath, err)
			continue
		}
		summary.RunsScanned++

		eval, cleanup, err := newEval()
		if err != nil {
			return summary, fmt.Errorf("failed to create evaluator: %w", err)
		}

		for gameIdx, g := range run.Games {
			if len(g.Moves) == 0 {
				summary.GamesSkipped++
				continue
			}
			source := fmt.Sprintf("%s anchor%d game%d", run.Timestamp, g.AnchorElo, gameIdx+1)
			blunders, err := scanGame(g, source, cfg, eval)
			// Write whatever was found even if the scan aborted mid-game
			for _, b := range blunders {
				if _, werr := fmt.Fprintln(out, formatBlunderEPD(b)); werr != nil {
					cleanup()
					return summary, fmt.Errorf("failed to write output: %w", werr)
				}
				summary.BlundersFound++
			}
			if err != nil {
				fmt.Printf("WARNING: %s: scan aborted: %v\n", source, err)
				continue
			}
			summary.GamesScanned++
			fmt.Printf("%s: %d plies, %d blunders\n", source, len(g.Moves), len(blunders))
		}
		cleanup()
	}

	return summary, nil
}

// EvalPosition asks the engine for its best move and score on the position
// reached by moveHistory from the start position, searching for moveTimeMs.
// The score is from the side to move's perspective (UCI convention).
func (c *StockfishClient) EvalPosition(moveHistory []string, moveTimeMs int) (string, int, error) {
	if err := c.Send(buildPositionCommand(moveHistory)); err != nil {
		return "", 0, err
	}
	if err := c.Send(fmt.Sprintf("go movetime %d", moveTimeMs)); err != nil {
		return "", 0, err
	}

	deadline := time.Duration(moveTimeMs)*time.Millisecond + 5*time.Second
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()

	lastScore := 0
	haveScore := false
	for {
		select {
		case line, ok := <-c.lines:
			if !ok {
				return "", 0, fmt.Errorf("engine closed stdout during evaluation")
			}
			if cp, ok := parseInfoScore(line); ok {
				lastScore = cp
				haveScore = true
			}
			if strings.HasPrefix(line, "bestmove") {
				fields := strings.Fields(line)
				if len(fields) < 2 {
					return "", 0, fmt.Errorf("malformed bestmove line: %q", line)
				}
				if !haveScore {
					return "", 0, fmt.Errorf("no score seen before bestmove")
				}
				return fields[1], lastScore, nil
			}
		case <-ctx.Done():
			return "", 0, fmt.Errorf("timed out waiting for evaluation: %w", ctx.Err())
		}
	}
}
