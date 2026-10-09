package uci

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/game"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

// e2eHardTimeout caps the scripted session. Engine.Run has no cancellation,
// so a hang would otherwise wedge the test process; Run is raced against this.
const e2eHardTimeout = 10 * time.Second

// sessionTimeout is e2eHardTimeout, shortened to fit under the test deadline
// so a hang fails with a clear message instead of a go test panic.
func sessionTimeout(t *testing.T) time.Duration {
	t.Helper()
	timeout := e2eHardTimeout
	if dl, ok := t.Deadline(); ok {
		if remaining := time.Until(dl) - 2*time.Second; remaining < timeout {
			timeout = remaining
		}
	}
	return timeout
}

// TestEngine_EndToEndProtocolSession drives Engine.Run in-process with a
// scripted UCI session and checks what a GUI would see: id lines then uciok,
// readyok, an info line before a bestmove that is legal after 1.e4, and a
// clean exit on "quit". A bad BookFile path set mid-session must not disrupt
// any of it.
func TestEngine_EndToEndProtocolSession(t *testing.T) {
	engine := NewUCIEngine()

	const badBookPath = "/nonexistent/definitely-not-a-real-book/path.bin"
	session := strings.Join([]string{
		"uci",
		"isready",
		"setoption name BookFile value " + badBookPath,
		"position startpos moves e2e4",
		"go depth 4",
		"quit",
		"",
	}, "\n")

	input := strings.NewReader(session)
	var output bytes.Buffer

	done := make(chan error, 1)
	go func() {
		done <- engine.Run(input, &output)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Engine.Run returned an error: %v", err)
		}
	case <-time.After(sessionTimeout(t)):
		t.Fatalf("UCI session did not complete within %s (engine appears to have hung)", e2eHardTimeout)
	}

	lines := nonEmptyLines(output.String())
	if len(lines) == 0 {
		t.Fatalf("engine produced no output at all for session:\n%s", session)
	}

	assertResponseOrder(t, lines)
	assertBestMoveIsLegalAfterE2E4(t, lines)
}

// assertResponseOrder checks the ordering a GUI depends on: id lines before
// uciok, readyok after it, an info line before bestmove, bestmove last.
func assertResponseOrder(t *testing.T, lines []string) {
	t.Helper()

	var (
		idNameIdx    = -1
		idAuthorIdx  = -1
		uciokIdx     = -1
		readyokIdx   = -1
		firstInfoIdx = -1
		bestmoveIdx  = -1
	)

	for i, line := range lines {
		switch {
		case line == "id name ChessEngine":
			idNameIdx = i
		case line == "id author Adam Griffiths":
			idAuthorIdx = i
		case line == "uciok":
			uciokIdx = i
		case line == "readyok":
			readyokIdx = i
		case strings.HasPrefix(line, "info depth") && firstInfoIdx == -1:
			firstInfoIdx = i
		case strings.HasPrefix(line, "bestmove"):
			bestmoveIdx = i
		}
	}

	if idNameIdx == -1 || idAuthorIdx == -1 || uciokIdx == -1 {
		t.Fatalf("missing id/uciok lines in output:\n%s", strings.Join(lines, "\n"))
	}
	if idNameIdx >= uciokIdx || idAuthorIdx >= uciokIdx {
		t.Errorf("id lines must precede uciok: id name=%d id author=%d uciok=%d", idNameIdx, idAuthorIdx, uciokIdx)
	}

	if readyokIdx == -1 {
		t.Fatalf("missing readyok line in output:\n%s", strings.Join(lines, "\n"))
	}
	if readyokIdx <= uciokIdx {
		t.Errorf("readyok (line %d) must come after uciok (line %d)", readyokIdx, uciokIdx)
	}

	if firstInfoIdx == -1 {
		t.Fatalf("no 'info depth ...' line found in output:\n%s", strings.Join(lines, "\n"))
	}
	if bestmoveIdx == -1 {
		t.Fatalf("no 'bestmove ...' line found in output:\n%s", strings.Join(lines, "\n"))
	}
	if firstInfoIdx >= bestmoveIdx {
		t.Errorf("info line (line %d) must precede bestmove (line %d)", firstInfoIdx, bestmoveIdx)
	}
	if bestmoveIdx != len(lines)-1 {
		t.Errorf("bestmove must be the last line of output; got it at line %d of %d:\n%s",
			bestmoveIdx, len(lines)-1, strings.Join(lines, "\n"))
	}
}

// assertBestMoveIsLegalAfterE2E4 checks the final bestmove against legal moves
// generated independently for the position after 1.e4.
func assertBestMoveIsLegalAfterE2E4(t *testing.T, lines []string) {
	t.Helper()

	last := lines[len(lines)-1]
	fields := strings.Fields(last)
	if len(fields) < 2 || fields[0] != "bestmove" {
		t.Fatalf("last line is not a well-formed bestmove line: %q", last)
	}
	bestMoveUCI := fields[1]
	if bestMoveUCI == "0000" || bestMoveUCI == "(none)" {
		t.Fatalf("engine returned a null move for a position with legal moves: %q", last)
	}

	gameEngine := game.NewEngine()
	if err := gameEngine.LoadFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"); err != nil {
		t.Fatalf("LoadFromFEN(startpos) failed: %v", err)
	}
	converter := NewMoveConverter()
	e2e4, err := converter.FromUCI("e2e4", gameEngine.GetState().Board)
	if err != nil {
		t.Fatalf("converting e2e4 failed: %v", err)
	}
	if err := gameEngine.MakeMove(e2e4); err != nil {
		t.Fatalf("making e2e4 failed: %v", err)
	}

	player := gameEngine.GetCurrentPlayer()
	generator := movegen.NewGenerator()
	legalMoves := generator.GenerateAllMoves(gameEngine.GetState().Board, player)
	defer movegen.ReleaseMoveList(legalMoves)

	found := false
	for i := range legalMoves.Count {
		if converter.ToUCI(legalMoves.Moves[i]) == bestMoveUCI {
			found = true
			break
		}
	}
	if !found {
		legalUCI := make([]string, legalMoves.Count)
		for i := range legalMoves.Count {
			legalUCI[i] = converter.ToUCI(legalMoves.Moves[i])
		}
		t.Errorf("bestmove %q is not legal after 1.e4; legal moves: %s", bestMoveUCI, strings.Join(legalUCI, " "))
	}
}

// nonEmptyLines splits output into lines, dropping blanks.
func nonEmptyLines(output string) []string {
	rawLines := strings.Split(output, "\n")
	lines := make([]string, 0, len(rawLines))
	for _, l := range rawLines {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}
	return lines
}
