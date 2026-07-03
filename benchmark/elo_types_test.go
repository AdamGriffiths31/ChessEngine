package benchmark

import (
	"encoding/json"
	"testing"
)

// Moves must round-trip through JSON so recorded games in elo_results.jsonl
// can be replayed by the blunder scanner.
func TestEloGameResultMovesRoundTrip(t *testing.T) {
	in := EloGameResult{
		AnchorElo: 2400,
		Color:     "white",
		Moves:     []string{"e2e4", "e7e5", "g1f3"},
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out EloGameResult
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Moves) != 3 || out.Moves[0] != "e2e4" || out.Moves[2] != "g1f3" {
		t.Errorf("moves did not round-trip: %v", out.Moves)
	}
}
