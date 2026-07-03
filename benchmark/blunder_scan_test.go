package benchmark

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/epd"
)

func TestParseInfoScore(t *testing.T) {
	tests := []struct {
		line string
		cp   int
		ok   bool
	}{
		{"info depth 12 seldepth 18 score cp 35 nodes 12345 pv e2e4", 35, true},
		{"info depth 12 score cp -260 nodes 99", -260, true},
		{"info depth 20 score mate 3 pv h5f7", 10000, true},
		{"info depth 20 score mate -2 pv e8d8", -10000, true},
		{"info depth 5 nodes 100", 0, false},
		{"bestmove e2e4", 0, false},
	}
	for _, tt := range tests {
		cp, ok := parseInfoScore(tt.line)
		if ok != tt.ok || cp != tt.cp {
			t.Errorf("parseInfoScore(%q) = (%d,%v), want (%d,%v)", tt.line, cp, ok, tt.cp, tt.ok)
		}
	}
}

func TestFormatBlunderEPD(t *testing.T) {
	b := Blunder{
		FEN:        "r1bqkbnr/pppp1ppp/2n5/4p2Q/2B1P3/8/PPPP1PPP/RNB1K1NR b KQkq - 4 4",
		BestMove:   "g8h6",
		PlayedMove: "g8f6",
		SwingCP:    950,
		Source:     "20260702_211900 anchor2400 game1 ply7",
	}
	line := formatBlunderEPD(b)

	// Must parse through the existing epd package
	pos, err := epd.ParseEPD(line)
	if err != nil {
		t.Fatalf("emitted EPD does not parse: %v\nline: %s", err, line)
	}
	if pos.BestMove != "g8h6" {
		t.Errorf("bm = %q, want g8h6", pos.BestMove)
	}
	// Metadata must NOT be in c0 (the parser mines c0 for move=digit scores)
	if len(pos.MoveScores) != 0 {
		t.Errorf("EPD metadata leaked into c0 move scores: %v", pos.MoveScores)
	}
	if !strings.Contains(line, "g8f6") || !strings.Contains(line, "950") {
		t.Errorf("metadata missing from line: %s", line)
	}
}

func TestScanGameFlagsBlunders(t *testing.T) {
	// Engine plays black from the start position (opening 0 is empty, so
	// every engine ply is scanned). Black's final move is a -840cp blunder
	// according to the fake evaluator; earlier moves are dead equal.
	g := EloGameResult{
		AnchorElo:    2400,
		Color:        "black",
		OpeningIndex: 0,
		Moves:        []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "g8f6"},
	}

	// Fake evaluator: position identified by ply (= len(moveHistory) given).
	nPlies := len(g.Moves)
	eval := func(history []string) (string, int, error) {
		switch len(history) {
		case nPlies - 1: // before black's last move: black slightly worse
			return "d7d6", -30, nil
		case nPlies: // after black's last move: white (to move) winning big
			return "f3g5", 870, nil
		default:
			// All other positions: dead equal, "best" matches whatever was played
			if len(history) < nPlies {
				return g.Moves[len(history)], 0, nil
			}
			return "a2a3", 0, nil
		}
	}

	blunders, err := scanGame(g, "testrun game1", BlunderScanConfig{ThresholdCP: 150, MoveTimeMs: 10}, eval)
	if err != nil {
		t.Fatal(err)
	}
	if len(blunders) != 1 {
		t.Fatalf("expected exactly 1 blunder, got %d: %+v", len(blunders), blunders)
	}
	b := blunders[0]
	if b.PlayedMove != "g8f6" {
		t.Errorf("played move = %q, want g8f6", b.PlayedMove)
	}
	if b.BestMove != "d7d6" {
		t.Errorf("best move = %q, want d7d6", b.BestMove)
	}
	// swing = bestScore - (-scoreAfterFromOpponentPOV) = -30 - (-870) = 840
	if b.SwingCP != 840 {
		t.Errorf("swing = %d, want 840", b.SwingCP)
	}
	if !strings.Contains(b.Source, "testrun game1") {
		t.Errorf("source missing run reference: %q", b.Source)
	}
}

func TestScanRecordedGamesEndToEnd(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "results.jsonl")
	output := filepath.Join(dir, "blunders.epd")

	run := EloResult{
		Timestamp: "20260703_120000",
		Games: []EloGameResult{
			{AnchorElo: 2400, Color: "white", OpeningIndex: 0,
				Moves: append(append([]string{}, eloOpenings[0]...), "g1f3", "b8c6")},
			{AnchorElo: 2400, Color: "black", OpeningIndex: 0}, // no Moves: pre-recording game, must be skipped
		},
	}
	line, err := json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	content := string(line) + "\n" + "{corrupt json\n"
	if err := os.WriteFile(input, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	// Fake evaluator factory: every engine move is a 200cp blunder
	newEval := func() (PositionEval, func(), error) {
		return func(history []string) (string, int, error) {
			if len(history)%2 == 0 {
				return "a2a3", 100, nil // before engine (white) move
			}
			return "a7a6", 100, nil // after: opponent +100 => engine -100, swing 200
		}, func() {}, nil
	}

	summary, err := ScanRecordedGames(input, output, BlunderScanConfig{ThresholdCP: 150, MoveTimeMs: 10}, newEval)
	if err != nil {
		t.Fatal(err)
	}
	if summary.GamesScanned != 1 || summary.GamesSkipped != 1 {
		t.Errorf("scanned/skipped = %d/%d, want 1/1", summary.GamesScanned, summary.GamesSkipped)
	}
	if summary.BlundersFound == 0 {
		t.Error("expected blunders to be found")
	}

	// Output must parse as EPD
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	positions, err := epd.ParseEPDFile(string(data))
	if err != nil {
		t.Fatalf("output EPD does not parse: %v", err)
	}
	if len(positions) != summary.BlundersFound {
		t.Errorf("EPD has %d positions, summary says %d", len(positions), summary.BlundersFound)
	}
}
