package bench

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// MatePosition is one matetrack EPD entry: a position and its known best
// mate distance in moves (positive: the side to move mates; negative: it is
// mated).
type MatePosition struct {
	FEN      string
	BestMate int
}

var bmMatePattern = regexp.MustCompile(`\bbm #(-?\d+)`)

// ParseMatetrack reads a matetrack-format EPD (e.g. vondele/matetrack's
// matetrack.epd): four FEN fields followed by "bm #N;" and other opcodes.
// Lines without a bm #N opcode are skipped. Returned FENs get " 0 1" move
// counters appended.
func ParseMatetrack(r io.Reader) ([]MatePosition, error) {
	var positions []MatePosition
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		m := bmMatePattern.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil || n == 0 {
			return nil, fmt.Errorf("bad bm #%s in line %q", m[1], line)
		}
		positions = append(positions, MatePosition{
			FEN:      strings.Join(fields[:4], " ") + " 0 1",
			BestMate: n,
		})
	}
	return positions, sc.Err()
}

// MateOutcome classifies one search result against the known best mate.
type MateOutcome int

const (
	// MateNotFound means the engine did not report a mate for the winning side.
	MateNotFound MateOutcome = iota
	// MateFound means the engine reported a mate, but longer than the best.
	MateFound
	// MateBest means the engine reported exactly the best mate distance.
	MateBest
	// MateWrong means the engine reported a mate shorter than the proven
	// best, or for the wrong side: a search or mate-scoring bug.
	MateWrong
)

// ClassifyMate compares an engine's reported mate distance (moves, as in UCI
// "score mate N"; 0 when the score is not a mate) with the known best.
func ClassifyMate(engineMate, bestMate int) MateOutcome {
	switch {
	case engineMate == 0:
		return MateNotFound
	case (engineMate > 0) != (bestMate > 0):
		return MateWrong
	case abs(engineMate) < abs(bestMate):
		return MateWrong
	case engineMate == bestMate:
		return MateBest
	default:
		return MateFound
	}
}

// MateSummary tallies ClassifyMate outcomes for a suite run.
type MateSummary struct {
	Positions int
	Found     int // includes Best
	Best      int
	Wrong     int
}

// Add records one outcome.
func (s *MateSummary) Add(o MateOutcome) {
	s.Positions++
	switch o {
	case MateBest:
		s.Best++
		s.Found++
	case MateFound:
		s.Found++
	case MateWrong:
		s.Wrong++
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
