package uci

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/game"
	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
)

// e2eHardTimeout is the absolute cap on how long the scripted session below
// may take to run. Engine.Run has no context/cancellation of its own - it
// simply reads lines until "quit" or EOF - so a real hang (e.g. a "go"
// command whose search never returns) would otherwise wedge the test
// process forever instead of failing it. Run is driven from a goroutine and
// raced against this timeout.
const e2eHardTimeout = 10 * time.Second

// sessionTimeout returns e2eHardTimeout, shortened to fit under the test's
// own -timeout deadline (with a 2s margin for cleanup/reporting) when that
// deadline is sooner, so a slow CI box fails with a clear "session did not
// complete" message rather than being killed mid-assertion by `go test`'s
// own panic-on-timeout.
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

// TestEngine_EndToEndProtocolSession drives Engine.Run in-process over an
// io.Reader/io.Writer pair (no real stdin/stdout, no subprocess) with a
// scripted UCI session and asserts on the exact response shape a real GUI
// would observe:
//
//   - the two "id ..." lines followed immediately by "uciok"
//   - "readyok" answering "isready"
//   - a "setoption name BookFile value <bad path>" early in the session,
//     which must not crash the engine or otherwise disrupt the session
//     (see internal/book.Prober.ensureLoaded: a failed load just logs a
//     warning and every Probe reports a miss)
//   - at least one "info ..." line before the final answer
//   - "bestmove ..." as the last line of output, and that its move is legal
//     in the position reached by "position startpos moves e2e4"
//   - clean termination: Run returns nil once "quit" is read, well within
//     the hard timeout.
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

// assertResponseOrder walks the captured output lines and asserts the
// protocol-level ordering a GUI depends on: both "id" lines before "uciok",
// "readyok" answering "isready" before the go-command's own response block,
// at least one "info" line before "bestmove", and "bestmove" as the very
// last line emitted.
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

// assertBestMoveIsLegalAfterE2E4 extracts the move from the final
// "bestmove <uci> [ponder ...]" line and verifies it is one of the legal
// moves generated for the position reached after 1.e4 (the position given
// to the scripted session's "position" command), independently re-deriving
// that position rather than trusting the engine's internal state.
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

	player := movegen.Player(gameEngine.GetCurrentPlayer())
	generator := movegen.NewGenerator()
	legalMoves := generator.GenerateAllMoves(gameEngine.GetState().Board, player)
	defer movegen.ReleaseMoveList(legalMoves)

	found := false
	for i := 0; i < legalMoves.Count; i++ {
		if converter.ToUCI(legalMoves.Moves[i]) == bestMoveUCI {
			found = true
			break
		}
	}
	if !found {
		legalUCI := make([]string, legalMoves.Count)
		for i := 0; i < legalMoves.Count; i++ {
			legalUCI[i] = converter.ToUCI(legalMoves.Moves[i])
		}
		t.Errorf("bestmove %q is not legal after 1.e4; legal moves: %s", bestMoveUCI, strings.Join(legalUCI, " "))
	}
}

// nonEmptyLines splits output on newlines and drops empty trailing/interior
// lines, so a trailing "\n" from the last Fprintln doesn't show up as a
// spurious empty final "line" that would otherwise defeat the
// "bestmove is the last line" check.
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
