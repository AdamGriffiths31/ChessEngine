package benchmark

// eloOpenings are fixed UCI move sequences used to vary opening play
// across Elo benchmark games, ported from CChess's EloBenchmark.
var eloOpenings = [][]string{
	{}, // Starting position
	{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "f8c5"},         // Italian Game
	{"d2d4", "d7d5", "c2c4", "e7e6", "b1c3", "g8f6"},         // Queen's Gambit
	{"e2e4", "c7c5", "g1f3", "d7d6", "d2d4", "c5d4", "f3d4"}, // Sicilian Defense
	{"c2c4", "e7e5", "b1c3", "g8f6", "g1f3", "b8c6"},         // English Opening
	{"e2e4", "e7e6", "d2d4", "d7d5", "b1c3", "g8f6"},         // French Defense
}
