package search

import "testing"

// TestRunBenchDeterministic guards the bench signature's core property: two
// runs of the same build at the same depth search exactly the same nodes.
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
