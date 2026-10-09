package bench

import (
	"reflect"
	"testing"
)

func TestBuildSPRTArgs(t *testing.T) {
	cfg := SPRTConfig{
		BasePath:     "/bin/base-uci",
		DevPath:      "/bin/dev-uci",
		Elo0:         0,
		Elo1:         5,
		Alpha:        0.05,
		Beta:         0.05,
		TimeControl:  "10+0.1",
		OpeningsFile: "/data/openings.epd",
		Concurrency:  1,
		MaxGames:     4000,
	}

	got := BuildSPRTArgs(cfg)
	want := []string{
		"-engine", "cmd=/bin/dev-uci", "name=Dev",
		"-engine", "cmd=/bin/base-uci", "name=Base",
		"-each", "proto=uci", "tc=10+0.1",
		"-games", "4000",
		"-repeat",
		"-openings", "file=/data/openings.epd", "format=epd", "order=random",
		"-concurrency", "1",
		"-sprt", "elo0=0", "elo1=5", "alpha=0.05", "beta=0.05",
		"-maxmoves", "200",
		"-recover",
		"-draw", "movenumber=40", "movecount=8", "score=10",
		"-resign", "movecount=6", "score=800",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildSPRTArgs mismatch:\n got:  %v\n want: %v", got, want)
	}
}

func TestBuildSPRTArgsWithEngineOptions(t *testing.T) {
	cfg := SPRTConfig{
		BasePath:     "/bin/base-uci",
		DevPath:      "/bin/dev-uci",
		DevOptions:   []string{"Skill Level=5", "Threads=1"},
		Elo0:         0,
		Elo1:         5,
		Alpha:        0.05,
		Beta:         0.05,
		TimeControl:  "10+0.1",
		OpeningsFile: "/data/openings.epd",
		Concurrency:  1,
		MaxGames:     4000,
	}

	got := BuildSPRTArgs(cfg)
	want := []string{
		"-engine", "cmd=/bin/dev-uci", "name=Dev",
		"option.Skill Level=5", "option.Threads=1",
		"-engine", "cmd=/bin/base-uci", "name=Base",
		"-each", "proto=uci", "tc=10+0.1",
		"-games", "4000",
		"-repeat",
		"-openings", "file=/data/openings.epd", "format=epd", "order=random",
		"-concurrency", "1",
		"-sprt", "elo0=0", "elo1=5", "alpha=0.05", "beta=0.05",
		"-maxmoves", "200",
		"-recover",
		"-draw", "movenumber=40", "movecount=8", "score=10",
		"-resign", "movecount=6", "score=800",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildSPRTArgs mismatch:\n got:  %v\n want: %v", got, want)
	}
}
