package search

import "testing"

func TestRunBenchDeterministic(t *testing.T) {
	const depth = 4
	first, err := RunBench(depth)
	if err != nil {
		t.Fatalf("RunBench: %v", err)
	}
	second, err := RunBench(depth)
	if err != nil {
		t.Fatalf("RunBench: %v", err)
	}
	if first.Nodes == 0 {
		t.Fatal("RunBench searched 0 nodes")
	}
	if first.Nodes != second.Nodes {
		t.Errorf("bench nodes differ between runs: %d vs %d", first.Nodes, second.Nodes)
	}
}
