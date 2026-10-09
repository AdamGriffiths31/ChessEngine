package bench

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// finishedGamePattern matches cutechess-cli's cumulative score line, e.g.
// "Score of Dev vs Base: 13 - 10 - 17  [0.537] 40". cutechess-cli names the
// first-registered engine first, so the captures are Dev's wins/losses/draws.
var finishedGamePattern = regexp.MustCompile(`Score of Dev vs Base: (\d+) - (\d+) - (\d+)`)

// RunSPRT runs cutechess-cli with cfg, streaming its stdout to os.Stdout
// (so a long-running SPRT test is observable live) while tracking the
// latest SPRT status and final score line. It returns once cutechess-cli's
// process exits, which happens either when the SPRT bound is reached or
// cfg.MaxGames is hit.
func RunSPRT(ctx context.Context, cfg SPRTConfig) (*SPRTResult, error) {
	if _, err := os.Stat(cfg.BasePath); err != nil {
		return nil, fmt.Errorf("base binary not found at %s: %w", cfg.BasePath, err)
	}
	if _, err := os.Stat(cfg.DevPath); err != nil {
		return nil, fmt.Errorf("dev binary not found at %s: %w", cfg.DevPath, err)
	}
	if _, err := os.Stat(cfg.OpeningsFile); err != nil {
		return nil, fmt.Errorf("openings file not found at %s: %w", cfg.OpeningsFile, err)
	}
	if _, err := os.Stat(cfg.CutechessCli); err != nil {
		return nil, fmt.Errorf("cutechess-cli not found at %s: %w", cfg.CutechessCli, err)
	}

	start := time.Now()
	args := BuildSPRTArgs(cfg)
	cmd := exec.CommandContext(ctx, cfg.CutechessCli, args...) // #nosec G204 - binary/args built from local config, not untrusted input

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start cutechess-cli: %w", err)
	}

	var (
		rawLog                     strings.Builder
		latestStatus               sprtStatus
		haveStatus                 bool
		wins, losses, draws, total int
	)

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		rawLog.WriteString(line)
		rawLog.WriteByte('\n')

		if status, ok := ParseSPRTStatus(line); ok {
			latestStatus = status
			haveStatus = true
		}
		if m := finishedGamePattern.FindStringSubmatch(line); m != nil {
			w, _ := strconv.Atoi(m[1])
			l, _ := strconv.Atoi(m[2])
			d, _ := strconv.Atoi(m[3])
			// Dev is engine 1 (see BuildSPRTArgs's doc comment), so these
			// are directly Dev's wins/losses/draws.
			wins, losses, draws = w, l, d
			total = w + l + d
		}
	}

	waitErr := cmd.Wait()
	duration := time.Since(start)

	result := &SPRTResult{
		Wins:       wins,
		Losses:     losses,
		Draws:      draws,
		GamesTotal: total,
		Config:     cfg,
		Timestamp:  start.Format("20060102_150405"),
		Duration:   duration,
		RawLog:     rawLog.String(),
	}
	if haveStatus {
		result.LLR = latestStatus.llr
		result.LBound = latestStatus.lbound
		result.UBound = latestStatus.ubound
		result.Verdict = verdictFromStatus(latestStatus)
	} else {
		result.Verdict = SPRTContinue
	}

	if ctx.Err() != nil {
		return result, fmt.Errorf("SPRT run cancelled: %w", ctx.Err())
	}
	if waitErr != nil {
		return result, fmt.Errorf("cutechess-cli exited with error: %w\nlast output:\n%s", waitErr, tail(rawLog.String(), 20))
	}

	return result, nil
}

// tail returns the last n lines of s, for error context.
func tail(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}
