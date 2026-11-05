package main

import (
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_SkipsCutOffLastLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.txt")
	content := "8/8/8/8/8/8/8/K6k w - - 0 1 | 25 | 0.5\n" +
		"8/8/8/8/8/8/8/K6k b - - 0 1 | -40 | 1.0\n" +
		"8/8/8/8/8/8/8/K6k w - - 0 1 | 12"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	data, err := load(path, 0.3)
	if err != nil {
		t.Fatal(err)
	}
	if data.len() != 2 {
		t.Errorf("loaded %d positions, want the 2 complete lines", data.len())
	}
}

func TestCheckpoint_SaveThenRestoreRoundTrips(t *testing.T) {
	path := filepath.Join(t.TempDir(), "net.ckpt")
	rng := rand.New(rand.NewPCG(1, 2))
	m, adamM, adamV := newModel(rng), newModel(rng), newModel(rng)

	ckpt := checkpoint{Epoch: 4, Step: 99, Positions: 1000, Hidden: hidden,
		Params: m.params(), AdamM: adamM.params(), AdamV: adamV.params()}
	if err := ckpt.save(path); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadCheckpoint(path)
	if err != nil {
		t.Fatal(err)
	}

	restored, restoredM, restoredV := newLike(m), newLike(m), newLike(m)
	if err := loaded.restore(1000, restored, restoredM, restoredV); err != nil {
		t.Fatal(err)
	}
	if loaded.Epoch != 4 || loaded.Step != 99 {
		t.Errorf("epoch, step = %d, %d; want 4, 99", loaded.Epoch, loaded.Step)
	}
	if restored.ftWeights[123] != m.ftWeights[123] || restoredV.outBias[0] != adamV.outBias[0] {
		t.Error("restored weights differ from the saved ones")
	}
}
