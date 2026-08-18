package bench

import "testing"

func TestComputeAnchors(t *testing.T) {
	anchors := computeAnchors(1500, 5, 100, 1320, 3190)
	expected := []int{1320, 1420, 1520, 1620, 1720}

	if len(anchors) != len(expected) {
		t.Fatalf("expected %d anchors, got %d", len(expected), len(anchors))
	}
	for i, a := range anchors {
		if a != expected[i] {
			t.Errorf("anchor %d: expected %d, got %d", i, expected[i], a)
		}
	}
}

func TestExpectedScoreEqualRatings(t *testing.T) {
	score := expectedScore(1500, 1500)
	if diff := score - 0.5; diff > 0.0001 || diff < -0.0001 {
		t.Errorf("expected 0.5 for equal ratings, got %f", score)
	}
}

func TestExpectedScoreHigherRatingWinsMore(t *testing.T) {
	score := expectedScore(1600, 1400)
	if score <= 0.5 {
		t.Errorf("expected a score > 0.5 for the higher-rated player, got %f", score)
	}
}

func TestEstimateRatingDrawAgainstSingleAnchor(t *testing.T) {
	rating := estimateRating([]int{1500}, []float64{0.5})
	if diff := rating - 1500.0; diff > 0.01 || diff < -0.01 {
		t.Errorf("expected ~1500 for a draw against a 1500 anchor, got %f", rating)
	}
}

func TestEstimateRatingMonotonicWithScore(t *testing.T) {
	anchors := []int{1500, 1500}
	low := estimateRating(anchors, []float64{0.0, 0.0})
	high := estimateRating(anchors, []float64{1.0, 1.0})

	if high <= low {
		t.Errorf("expected a higher score to produce a higher rating: low=%f high=%f", low, high)
	}
}

func TestAllocateTimeNormal(t *testing.T) {
	ms := allocateTime(60000, 1000)
	if ms != 3000 {
		t.Errorf("expected 3000ms, got %d", ms)
	}
}

func TestAllocateTimeFloor(t *testing.T) {
	ms := allocateTime(100, 0)
	if ms != 50 {
		t.Errorf("expected the 50ms floor, got %d", ms)
	}
}

func TestGameAnchorsAndScoresMatchGameCount(t *testing.T) {
	// 3 anchor levels, 3 games each: the anchors/scores fed to
	// estimateRating must be parallel arrays of one entry per game, not
	// one entry per distinct anchor level.
	games := []EloGameResult{
		{AnchorElo: 1900, Result: 0.0},
		{AnchorElo: 1900, Result: 1.0},
		{AnchorElo: 1900, Result: 1.0},
		{AnchorElo: 2000, Result: 1.0},
		{AnchorElo: 2000, Result: 1.0},
		{AnchorElo: 2000, Result: 1.0},
		{AnchorElo: 2100, Result: 1.0},
		{AnchorElo: 2100, Result: 0.0},
		{AnchorElo: 2100, Result: 1.0},
	}

	anchors, scores := gameAnchorsAndScores(games)

	if len(anchors) != len(games) || len(scores) != len(games) {
		t.Fatalf("expected %d-element parallel arrays, got %d anchors and %d scores",
			len(games), len(anchors), len(scores))
	}
	for i, g := range games {
		if anchors[i] != g.AnchorElo {
			t.Errorf("anchors[%d]: expected %d, got %d", i, g.AnchorElo, anchors[i])
		}
		if scores[i] != g.Result {
			t.Errorf("scores[%d]: expected %f, got %f", i, g.Result, scores[i])
		}
	}
}

func TestEstimateRatingDoesNotSaturateOnRealMixedResults(t *testing.T) {
	// Regression test for the exact data from a real run: 7 wins, 2 losses
	// against anchors 1900/2000/2100 (3 games each). Before the fix,
	// Run() passed only the 3 distinct anchor levels into estimateRating
	// instead of one anchor per game, so expectedTotal could never exceed
	// ~3.0 while actualScore was 7.0 - forcing the estimate to saturate at
	// the search ceiling (3500) regardless of how well ChessEngine played.
	games := []EloGameResult{
		{AnchorElo: 1900, Result: 0.0},
		{AnchorElo: 1900, Result: 1.0},
		{AnchorElo: 1900, Result: 1.0},
		{AnchorElo: 2000, Result: 1.0},
		{AnchorElo: 2000, Result: 1.0},
		{AnchorElo: 2000, Result: 1.0},
		{AnchorElo: 2100, Result: 1.0},
		{AnchorElo: 2100, Result: 0.0},
		{AnchorElo: 2100, Result: 1.0},
	}

	anchors, scores := gameAnchorsAndScores(games)
	rating := estimateRating(anchors, scores)

	if rating >= 3499.0 {
		t.Errorf("expected a non-saturated estimate for a 7/9 mixed result, got %f", rating)
	}
	if rating <= 2000.0 {
		t.Errorf("expected the estimate to be comfortably above the anchors ChessEngine mostly beat, got %f", rating)
	}
}
