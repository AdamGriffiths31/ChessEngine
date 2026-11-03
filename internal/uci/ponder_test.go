package uci

import (
	"bufio"
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

const (
	mateInOneFEN  = "6k1/5ppp/8/8/8/8/8/R5K1 w - - 0 1"
	ponderTimeout = 5 * time.Second
	silenceWindow = 200 * time.Millisecond
)

type ponderSession struct {
	t     *testing.T
	in    *io.PipeWriter
	lines chan string
	done  chan error
}

func startPonderSession(t *testing.T, engine *Engine) *ponderSession {
	t.Helper()
	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()
	s := &ponderSession{t: t, in: inWriter, lines: make(chan string, 4096), done: make(chan error, 1)}

	go func() {
		scanner := bufio.NewScanner(outReader)
		for scanner.Scan() {
			s.lines <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			t.Errorf("reading engine output: %v", err)
		}
		close(s.lines)
	}()
	go func() {
		err := engine.Run(inReader, outWriter)
		_ = outWriter.Close()
		s.done <- err
	}()
	return s
}

func (s *ponderSession) send(commands ...string) {
	s.t.Helper()
	for _, command := range commands {
		if _, err := io.WriteString(s.in, command+"\n"); err != nil {
			s.t.Fatalf("writing %q: %v", command, err)
		}
	}
}

func (s *ponderSession) expectBestMove() string {
	s.t.Helper()
	timeout := time.After(ponderTimeout)
	for {
		select {
		case line, ok := <-s.lines:
			if !ok {
				s.t.Fatal("engine closed its output before sending bestmove")
			}
			if strings.HasPrefix(line, "bestmove") {
				return line
			}
		case <-timeout:
			s.t.Fatalf("no bestmove within %s", ponderTimeout)
		}
	}
}

func (s *ponderSession) expectNoBestMove() {
	s.t.Helper()
	timeout := time.After(silenceWindow)
	for {
		select {
		case line := <-s.lines:
			if strings.HasPrefix(line, "bestmove") {
				s.t.Fatalf("engine sent %q while pondering", line)
			}
		case <-timeout:
			return
		}
	}
}

func (s *ponderSession) quit() {
	s.t.Helper()
	s.send("quit")
	select {
	case err := <-s.done:
		if err != nil {
			s.t.Fatalf("Run returned an error: %v", err)
		}
	case <-time.After(ponderTimeout):
		s.t.Fatal("engine did not exit after quit")
	}
}

func TestPonder_MateHoldsBestMoveUntilPonderHit(t *testing.T) {
	s := startPonderSession(t, NewUCIEngine())
	s.send("position fen "+mateInOneFEN, "go ponder wtime 60000 btime 60000")

	s.expectNoBestMove()

	s.send("ponderhit")
	if got := s.expectBestMove(); !strings.HasPrefix(got, "bestmove a1a8") {
		t.Errorf("bestmove = %q, want a1a8", got)
	}
	s.quit()
}

func TestPonder_HitThinksOnItsOwnClock(t *testing.T) {
	s := startPonderSession(t, NewUCIEngine())
	s.send("position startpos", "go ponder wtime 60000 btime 60000")

	s.expectNoBestMove()
	time.Sleep(2 * time.Second)

	start := time.Now()
	s.send("ponderhit")
	s.expectBestMove()
	elapsed := time.Since(start)
	if elapsed < 250*time.Millisecond {
		t.Errorf("bestmove took %v after ponderhit, want a think of its own", elapsed)
	}
	if elapsed > 3*time.Second {
		t.Errorf("bestmove took %v after ponderhit, want it within the soft budget", elapsed)
	}
	s.quit()
}

func TestPonder_StopEndsSearchAndEngineStaysUsable(t *testing.T) {
	s := startPonderSession(t, NewUCIEngine())
	s.send("position startpos", "go ponder wtime 60000 btime 60000")
	s.expectNoBestMove()

	s.send("stop")
	s.expectBestMove()

	s.send("position startpos moves e2e4", "go depth 3")
	s.expectBestMove()
	s.quit()
}

func TestPonder_CommandsWithoutStopDoNotHang(t *testing.T) {
	for _, next := range []string{"position startpos moves e2e4", "go depth 3", "ucinewgame", "quit"} {
		t.Run(next, func(t *testing.T) {
			s := startPonderSession(t, NewUCIEngine())
			s.send("position startpos", "go ponder wtime 60000 btime 60000")
			s.expectNoBestMove()

			s.send(next)
			s.expectBestMove()
			if next != "quit" {
				s.quit()
			}
		})
	}
}

func TestPonder_BestMoveNamesReplyOnlyWithOptionOn(t *testing.T) {
	tests := []struct {
		name    string
		setup   []string
		wantLen int
	}{
		{"off by default", nil, 2},
		{"on", []string{"setoption name Ponder value true"}, 4},
		{"switched off", []string{"setoption name Ponder value false"}, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := startPonderSession(t, NewUCIEngine())
			s.send(tc.setup...)
			s.send("position startpos", "go depth 6")
			if fields := strings.Fields(s.expectBestMove()); len(fields) != tc.wantLen {
				t.Errorf("bestmove fields = %v, want %d of them", fields, tc.wantLen)
			}
			s.quit()
		})
	}
}

func TestPonder_MissLeavesBookTrackingUntouched(t *testing.T) {
	const bookOption = "setoption name BookFile value ../book/testdata/performance.bin"

	played := NewUCIEngine()
	s := startPonderSession(t, played)
	s.send(bookOption, "position startpos", "go depth 1")
	s.expectBestMove()
	s.quit()
	if !played.prevMoveFromBook {
		t.Fatal("startpos was not a book hit, so this test proves nothing")
	}

	missed := NewUCIEngine()
	s = startPonderSession(t, missed)
	s.send(bookOption, "position startpos", "go ponder wtime 60000 btime 60000")
	s.expectNoBestMove()
	s.send("stop")
	s.expectBestMove()
	s.quit()
	if missed.prevMoveFromBook {
		t.Error("a ponder miss marked the previous move as played from the book")
	}
}

func TestPonder_OnlyPlayedMovesAreNumbered(t *testing.T) {
	engine := NewUCIEngine()
	s := startPonderSession(t, engine)

	s.send("position startpos", "go ponder wtime 60000 btime 60000")
	s.expectNoBestMove()
	s.send("stop")
	s.expectBestMove()

	s.send("position startpos", "go ponder wtime 60000 btime 60000")
	s.expectNoBestMove()
	s.send("ponderhit")
	s.expectBestMove()

	s.send("position startpos", "go depth 1")
	s.expectBestMove()
	s.quit()

	if engine.moveNumber != 2 {
		t.Errorf("moveNumber = %d, want 2 (one hit and one normal search, the miss not counted)", engine.moveNumber)
	}
}

func TestPonderState_HitWhileSearchingArmsHardLimit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := &ponderState{cancel: cancel, hitLimit: 20 * time.Millisecond}

	p.hit()
	if ctx.Err() != nil {
		t.Fatal("hit stopped a search that still had time to use")
	}
	if p.markReadyAndCheckHeld() {
		t.Error("markReadyAndCheckHeld = true after ponderhit, want the search free to move")
	}

	select {
	case <-ctx.Done():
	case <-time.After(ponderTimeout):
		t.Fatal("hard limit never cancelled the search")
	}
}

func TestPonderState_HitAfterReadyCancelsAtOnce(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := &ponderState{cancel: cancel, hitLimit: time.Hour}

	if !p.markReadyAndCheckHeld() {
		t.Fatal("markReadyAndCheckHeld = false before ponderhit, want the move held")
	}
	if ctx.Err() != nil {
		t.Fatal("markReadyAndCheckHeld cancelled the search")
	}

	p.hit()
	if ctx.Err() == nil {
		t.Error("hit did not cancel a search that was ready to move")
	}
}

func TestPonderState_CancelUnlessHitSparesAHitSearch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := &ponderState{cancel: cancel, hitLimit: time.Hour}

	timer := p.hit()
	defer timer.Stop()
	p.cancelUnlessHit()
	if ctx.Err() != nil {
		t.Error("cancelUnlessHit cancelled a search already on the clock")
	}

	waiting := &ponderState{cancel: cancel}
	waiting.cancelUnlessHit()
	if ctx.Err() == nil {
		t.Error("cancelUnlessHit left a waiting search running")
	}
}

func TestPonderState_SoftClockStartsAtHit(t *testing.T) {
	p := &ponderState{cancel: func() {}, hitLimit: time.Hour}
	if got := p.softElapsed(); got != 0 {
		t.Errorf("softElapsed before ponderhit = %v, want 0", got)
	}

	p.hit().Stop()
	time.Sleep(5 * time.Millisecond)
	if got := p.softElapsed(); got <= 0 {
		t.Errorf("softElapsed after ponderhit = %v, want it running", got)
	}
}

func TestPonderState_OutcomeIsRecorded(t *testing.T) {
	var none *ponderState
	if got := none.outcome(); got != PonderNone {
		t.Errorf("outcome of a normal search = %q, want %q", got, PonderNone)
	}

	p := &ponderState{cancel: func() {}, hitLimit: time.Hour}
	if got := p.outcome(); got != PonderMiss {
		t.Errorf("outcome before ponderhit = %q, want %q", got, PonderMiss)
	}
	p.hit().Stop()
	if got := p.outcome(); got != PonderHit {
		t.Errorf("outcome after ponderhit = %q, want %q", got, PonderHit)
	}
}

func TestProtocolHandler_ParseGoPonder(t *testing.T) {
	params := NewProtocolHandler().ParseGo([]string{"ponder", "wtime", "1000", "btime", "2000"})
	if !params.Ponder || params.WTime != time.Second || params.BTime != 2*time.Second {
		t.Errorf("ParseGo = %+v, want ponder with the clocks", params)
	}
}

func TestProtocolHandler_FormatBestMovePonder(t *testing.T) {
	handler := NewProtocolHandler()
	if got := handler.FormatBestMove("e2e4", "e7e5"); got != "bestmove e2e4 ponder e7e5" {
		t.Errorf("with ponder = %q", got)
	}
	if got := handler.FormatBestMove("e2e4", ""); got != "bestmove e2e4" {
		t.Errorf("without ponder = %q", got)
	}
}
