package benchmark

import "testing"

func TestFindEngineByCommandSubstring(t *testing.T) {
	em := &EngineManager{
		config: &EngineConfig{
			Engines: []Engine{
				{Name: "ChessEngine", Command: "tools/bin/uci"},
				{Name: "Stockfish Skill 5", Command: "tools/engines/stockfish/stockfish-windows-x86-64-avx2.exe"},
			},
		},
	}

	engine, err := em.FindEngineByCommandSubstring("stockfish")
	if err != nil {
		t.Fatalf("expected to find an engine, got error: %v", err)
	}
	if engine.Name != "Stockfish Skill 5" {
		t.Errorf("expected to find 'Stockfish Skill 5', got %q", engine.Name)
	}
}

func TestFindEngineByCommandSubstringNotFound(t *testing.T) {
	em := &EngineManager{
		config: &EngineConfig{
			Engines: []Engine{
				{Name: "ChessEngine", Command: "tools/bin/uci"},
			},
		},
	}

	if _, err := em.FindEngineByCommandSubstring("stockfish"); err == nil {
		t.Fatal("expected an error when no engine matches")
	}
}
