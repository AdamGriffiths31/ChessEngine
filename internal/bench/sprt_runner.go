package bench

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// RunSPRT runs fastchess with cfg, streaming its stdout to os.Stdout
// (so a long-running SPRT test is observable live) while tracking the
// latest SPRT status, score and pentanomial lines. It returns once fastchess's
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
	if _, err := os.Stat(cfg.Fastchess); err != nil {
		return nil, fmt.Errorf("fastchess not found at %s: %w", cfg.Fastchess, err)
	}

	start := time.Now()
	args := BuildSPRTArgs(cfg)
	cmd := exec.CommandContext(ctx, cfg.Fastchess, args...) // #nosec G204 - binary/args built from local config, not untrusted input

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start fastchess: %w", err)
	}

	var (
		rawLog                     strings.Builder
		latestStatus               sprtStatus
		haveStatus                 bool
		wins, losses, draws, total int
		pentanomial                [5]int
	)

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		rawLog.WriteString(line)
		rawLog.WriteByte('\n')

		if status, ok := parseSPRTStatus(line); ok {
			latestStatus = status
			haveStatus = true
		}
		// Dev is engine 1 (see BuildSPRTArgs's doc comment), so these are
		// directly Dev's results.
		if m := gamesPattern.FindStringSubmatch(line); m != nil {
			n := parseIntMatches(m)
			total, wins, losses, draws = n[0], n[1], n[2], n[3]
		}
		if m := pentanomialPattern.FindStringSubmatch(line); m != nil {
			copy(pentanomial[:], parseIntMatches(m))
		}
	}

	waitErr := cmd.Wait()
	duration := time.Since(start)

	result := &SPRTResult{
		Wins:        wins,
		Losses:      losses,
		Draws:       draws,
		GamesTotal:  total,
		Pentanomial: pentanomial,
		Config:      cfg,
		Timestamp:   start.Format("20060102_150405"),
		Duration:    duration,
		RawLog:      rawLog.String(),
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
		return result, fmt.Errorf("fastchess exited with error: %w\nlast output:\n%s", waitErr, tail(rawLog.String(), 20))
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
