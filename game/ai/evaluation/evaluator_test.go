package evaluation

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
)

func TestNewEvaluator(t *testing.T) {
	evaluator := NewEvaluator()
	if evaluator == nil {
		t.Fatal("NewEvaluator should not return nil")
	}

	if evaluator.GetName() != "Evaluator" {
		t.Errorf("Expected name 'Evaluator', got '%s'", evaluator.GetName())
	}
}

func TestEvaluateEmptyBoard(t *testing.T) {
	evaluator := NewEvaluator()
	b := board.NewBoard()

	// Empty board should have score 0 (always from White's perspective)
	score := evaluator.Evaluate(b)

	if score != 0 {
		t.Errorf("Expected score 0 for empty board, got %d", score)
	}
}

func TestEvaluateStartingPosition(t *testing.T) {
	evaluator := NewEvaluator()
	
	// Create board from starting position FEN
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from starting FEN: %v", err)
	}

	// Starting position should be roughly equal (score ~0 from White's perspective)
	score := evaluator.Evaluate(b)

	if score < -50 || score > 50 {
		t.Errorf("Expected score close to 0 in starting position, got %d", score)
	}
}

func TestKnightPositionalBonus(t *testing.T) {
	evaluator := NewEvaluator()
	
	tests := []struct {
		name           string
		fen            string
		expectedWhite  ai.EvaluationScore
		expectedBlack  ai.EvaluationScore
		description    string
	}{
		{
			name:           "White knight on corner",
			fen:            "8/8/8/8/8/8/8/N7 w - - 0 1", // a1
			expectedWhite:  ai.EvaluationScore(320 + (-50)), // material + positional
			expectedBlack:  ai.EvaluationScore(-(320 + (-50))),
			description:    "Knight on a1 should have -50 positional penalty",
		},
		{
			name:           "White knight on center",
			fen:            "8/8/8/3N4/8/8/8/8 w - - 0 1", // d5
			expectedWhite:  ai.EvaluationScore(320 + 20), // material + positional
			expectedBlack:  ai.EvaluationScore(-(320 + 20)),
			description:    "Knight on d5 should have +20 positional bonus",
		},
		{
			name:           "White knight on edge",
			fen:            "8/8/8/8/N7/8/8/8 w - - 0 1", // a4
			expectedWhite:  ai.EvaluationScore(320 + (-30)), // material + positional
			expectedBlack:  ai.EvaluationScore(-(320 + (-30))),
			description:    "Knight on a4 should have -30 positional penalty",
		},
		{
			name:           "Black knight on corner from black's perspective",
			fen:            "n7/8/8/8/8/8/8/8 w - - 0 1", // a8 (black's back rank corner)
			expectedWhite:  ai.EvaluationScore(-(320 + (-50))), // Black pieces are negative
			expectedBlack:  ai.EvaluationScore(320 + (-50)),    // From black's perspective
			description:    "Black knight on a8 should have -50 positional penalty",
		},
		{
			name:           "Black knight on center from black's perspective", 
			fen:            "8/8/8/3n4/8/8/8/8 w - - 0 1", // d5
			expectedWhite:  ai.EvaluationScore(-(320 + 20)), // Black pieces are negative
			expectedBlack:  ai.EvaluationScore(320 + 20),    // From black's perspective  
			description:    "Black knight on d5 should have +20 positional bonus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN %s: %v", tt.fen, err)
			}

			score := evaluator.Evaluate(b)

			if score != tt.expectedWhite {
				t.Errorf("%s: Expected score %d, got %d", tt.description, tt.expectedWhite, score)
			}
		})
	}
}

func TestKnightTableValues(t *testing.T) {
	// Test specific knight table values
	tests := []struct {
		square   string
		rank     int
		file     int
		expected int
	}{
		{"a1", 0, 0, -50}, // corner
		{"h1", 0, 7, -50}, // corner
		{"d4", 3, 3, 20},  // center
		{"e4", 3, 4, 20},  // center
		{"d5", 4, 3, 20},  // center
		{"e5", 4, 4, 20},  // center
		{"b1", 0, 1, -40}, // back rank knight
		{"g1", 0, 6, -40}, // back rank knight
	}

	for _, tt := range tests {
		t.Run(tt.square, func(t *testing.T) {
			actual := KnightTable[tt.rank*8+tt.file]
			if actual != tt.expected {
				t.Errorf("Knight table value for %s (rank %d, file %d): expected %d, got %d", 
					tt.square, tt.rank, tt.file, tt.expected, actual)
			}
		})
	}
}

func TestPositionalBonusFunction(t *testing.T) {
	tests := []struct {
		piece    board.Piece
		rank     int
		file     int
		expected int
	}{
		{board.WhiteKnight, 0, 0, -50}, // a1
		{board.WhiteKnight, 3, 3, 20},  // d4
		{board.BlackKnight, 7, 0, 50},  // a8 (corner penalty, negated for black)
		{board.BlackKnight, 4, 3, -20}, // d5 (center bonus, negated for black)
		{board.WhitePawn, 1, 1, 10},    // Pawn on b2 should have +10 starting rank bonus
		{board.WhiteRook, 4, 4, 0},     // Rooks on middle ranks center should have no bonus
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			actual := getPositionalBonus(tt.piece, tt.rank, tt.file)
			if actual != tt.expected {
				t.Errorf("getPositionalBonus(%v, %d, %d): expected %d, got %d", 
					tt.piece, tt.rank, tt.file, tt.expected, actual)
			}
		})
	}
}

func TestKnightPositionPreference(t *testing.T) {
	evaluator := NewEvaluator()

	// Test with knight in corner vs center
	cornerFEN := "8/8/8/8/8/8/8/N7 w - - 0 1" // a1
	centerFEN := "8/8/8/3N4/8/8/8/8 w - - 0 1" // d5

	cornerBoard, _ := board.FromFEN(cornerFEN)
	centerBoard, _ := board.FromFEN(centerFEN)

	// Evaluator should prefer center over corner
	cornerScore := evaluator.Evaluate(cornerBoard)
	centerScore := evaluator.Evaluate(centerBoard)

	if centerScore <= cornerScore {
		t.Errorf("Evaluator should prefer knight in center (%d) over corner (%d)", 
			centerScore, cornerScore)
	}

	// The difference should be exactly the positional bonus difference
	expectedDiff := 20 - (-50) // center bonus - corner penalty = 70
	actualDiff := int(centerScore - cornerScore)
	if actualDiff != expectedDiff {
		t.Errorf("Expected positional difference of %d, got %d", expectedDiff, actualDiff)
	}
}

func TestComplexPositionWithKnights(t *testing.T) {
	evaluator := NewEvaluator()
	
	// Position with knights in different locations
	// White knight on d4 (center), Black knight on a8 (corner)
	fen := "n7/8/8/8/3N4/8/8/8 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	score := evaluator.Evaluate(b)

	// White knight on d4: 320 + 20 = 340
	// Black knight on a8: -(320 + (-50)) = -270
	// Total from White's perspective: 340 - 270 = 70
	expectedScore := ai.EvaluationScore(70)

	if score != expectedScore {
		t.Errorf("Expected score %d, got %d", expectedScore, score)
	}
}

func TestBishopPositionalBonus(t *testing.T) {
	evaluator := NewEvaluator()
	
	tests := []struct {
		name        string
		fen         string
		expected    ai.EvaluationScore
		description string
	}{
		{
			name:        "White bishop on corner",
			fen:         "8/8/8/8/8/8/8/B7 w - - 0 1", // a1
			expected:    ai.EvaluationScore(330 + (-20)), // material + positional
			description: "Bishop on a1 should have -20 positional penalty",
		},
		{
			name:        "White bishop on center",
			fen:         "8/8/8/3B4/8/8/8/8 w - - 0 1", // d5
			expected:    ai.EvaluationScore(330 + 10), // material + positional
			description: "Bishop on d5 should have +10 positional bonus",
		},
		{
			name:        "White bishop on good square",
			fen:         "8/8/8/8/2B5/8/8/8 w - - 0 1", // c4
			expected:    ai.EvaluationScore(330 + 5), // material + positional
			description: "Bishop on c4 should have +5 positional bonus",
		},
		{
			name:        "Black bishop on corner",
			fen:         "b7/8/8/8/8/8/8/8 w - - 0 1", // a8 (black's back rank corner)
			expected:    ai.EvaluationScore(-(330 + (-20))), // Black pieces are negative from White's perspective
			description: "Black bishop on a8 should have -20 positional penalty (negative for Black)",
		},
		{
			name:        "Black bishop on center", 
			fen:         "8/8/8/3b4/8/8/8/8 w - - 0 1", // d5
			expected:    ai.EvaluationScore(-(330 + 10)), // Black pieces are negative from White's perspective
			description: "Black bishop on d5 should have +10 positional bonus (negative for Black)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN %s: %v", tt.fen, err)
			}

			score := evaluator.Evaluate(b)

			if score != tt.expected {
				t.Errorf("%s: Expected score %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestBishopTableValues(t *testing.T) {
	// Test specific bishop table values
	tests := []struct {
		square   string
		rank     int
		file     int
		expected int
	}{
		{"a1", 0, 0, -20}, // corner
		{"h1", 0, 7, -20}, // corner
		{"d4", 3, 3, 10},  // center
		{"e4", 3, 4, 10},  // center
		{"d5", 4, 3, 10},  // center
		{"e5", 4, 4, 10},  // center
		{"c4", 3, 2, 5},   // good square
		{"f4", 3, 5, 5},   // good square
	}

	for _, tt := range tests {
		t.Run(tt.square, func(t *testing.T) {
			actual := BishopTable[tt.rank*8+tt.file]
			if actual != tt.expected {
				t.Errorf("Bishop table value for %s (rank %d, file %d): expected %d, got %d", 
					tt.square, tt.rank, tt.file, tt.expected, actual)
			}
		})
	}
}

func TestBishopPositionPreference(t *testing.T) {
	evaluator := NewEvaluator()

	// Test with bishop in corner vs center
	cornerFEN := "8/8/8/8/8/8/8/B7 w - - 0 1" // a1
	centerFEN := "8/8/8/3B4/8/8/8/8 w - - 0 1" // d5

	cornerBoard, _ := board.FromFEN(cornerFEN)
	centerBoard, _ := board.FromFEN(centerFEN)

	// Evaluator should prefer center over corner
	cornerScore := evaluator.Evaluate(cornerBoard)
	centerScore := evaluator.Evaluate(centerBoard)

	if centerScore <= cornerScore {
		t.Errorf("Evaluator should prefer bishop in center (%d) over corner (%d)", 
			centerScore, cornerScore)
	}

	// The difference should be exactly the positional bonus difference
	expectedDiff := 10 - (-20) // center bonus - corner penalty = 30
	actualDiff := int(centerScore - cornerScore)
	if actualDiff != expectedDiff {
		t.Errorf("Expected positional difference of %d, got %d", expectedDiff, actualDiff)
	}
}

func TestMixedPiecePositions(t *testing.T) {
	evaluator := NewEvaluator()
	
	// Position with both knights and bishops in different locations
	// White knight on d4 (center), White bishop on a1 (corner)
	// Black knight on a8 (corner), Black bishop on e5 (center)
	fen := "n7/8/8/4b3/3N4/8/8/B7 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	score := evaluator.Evaluate(b)

	// White knight on d4: 320 + 20 = 340
	// White bishop on a1: 330 + (-20) = 310
	// Black knight on a8: -(320 + (-50)) = -270
	// Black bishop on e5: -(330 + 10) = -340
	// Total from White's perspective: 340 + 310 - 270 - 340 = 40
	expectedScore := ai.EvaluationScore(40)

	if score != expectedScore {
		t.Errorf("Expected score %d, got %d", expectedScore, score)
	}
}

func TestRookPositionalBonus(t *testing.T) {
	evaluator := NewEvaluator()
	
	tests := []struct {
		name        string
		fen         string
		expected    ai.EvaluationScore
		description string
	}{
		{
			name:        "White rook on a file",
			fen:         "8/8/8/8/8/R7/8/8 w - - 0 1", // a3
			expected:    ai.EvaluationScore(500 + (-5)), // material + positional
			description: "Rook on a3 should have -5 positional penalty (avoid a column)",
		},
		{
			name:        "White rook on 7th rank",
			fen:         "8/3R4/8/8/8/8/8/8 w - - 0 1", // d7
			expected:    ai.EvaluationScore(500 + 0), // material + positional
			description: "Rook on d7 should have 0 positional bonus",
		},
		{
			name:        "White rook on 2nd rank center",
			fen:         "8/8/8/8/8/8/3R4/8 w - - 0 1", // d2
			expected:    ai.EvaluationScore(500 + 10), // material + positional
			description: "Rook on d2 should have +10 positional bonus",
		},
		{
			name:        "White rook on back rank center",
			fen:         "8/8/8/8/8/8/8/3R4 w - - 0 1", // d1
			expected:    ai.EvaluationScore(500 + 0), // material + positional
			description: "Rook on d1 should have 0 positional bonus",
		},
		{
			name:        "Black rook on a file",
			fen:         "8/8/8/8/8/r7/8/8 w - - 0 1", // a3 from black's perspective (rank 5 flipped)
			expected:    ai.EvaluationScore(-(500 + (-5))), // Black pieces are negative from White's perspective
			description: "Black rook on a3 should have -5 positional penalty (negative for Black)",
		},
		{
			name:        "Black rook on 2nd rank", 
			fen:         "8/3r4/8/8/8/8/8/8 w - - 0 1", // d7 (black's 2nd rank)
			expected:    ai.EvaluationScore(-(500 + 10)), // Black pieces are negative from White's perspective
			description: "Black rook on d7 should have +10 positional bonus (negative for Black)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN %s: %v", tt.fen, err)
			}

			score := evaluator.Evaluate(b)

			if score != tt.expected {
				t.Errorf("%s: Expected score %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestRookTableValues(t *testing.T) {
	// Test specific rook table values
	tests := []struct {
		square   string
		rank     int
		file     int
		expected int
	}{
		{"a1", 0, 0, 0},   // back rank
		{"d1", 0, 3, 0},   // back rank center
		{"a2", 1, 0, 5},   // 2nd rank edge
		{"d2", 1, 3, 10},  // 2nd rank center (good)
		{"a3", 2, 0, -5},  // avoid a column
		{"h3", 2, 7, -5},  // avoid h column
		{"d3", 2, 3, 0},   // center file middle ranks
		{"d8", 7, 3, 5},   // 8th rank center
		{"e8", 7, 4, 5},   // 8th rank center
	}

	for _, tt := range tests {
		t.Run(tt.square, func(t *testing.T) {
			actual := RookTable[tt.rank*8+tt.file]
			if actual != tt.expected {
				t.Errorf("Rook table value for %s (rank %d, file %d): expected %d, got %d", 
					tt.square, tt.rank, tt.file, tt.expected, actual)
			}
		})
	}
}

func TestRookPositionPreference(t *testing.T) {
	evaluator := NewEvaluator()

	// Test rook on a-file vs center file
	aFileFEN := "8/8/8/8/8/R7/8/8 w - - 0 1" // a3
	centerFEN := "8/8/8/8/8/8/3R4/8 w - - 0 1" // d2

	aFileBoard, _ := board.FromFEN(aFileFEN)
	centerBoard, _ := board.FromFEN(centerFEN)

	// Evaluator should prefer center 2nd rank over a-file
	aFileScore := evaluator.Evaluate(aFileBoard)
	centerScore := evaluator.Evaluate(centerBoard)

	if centerScore <= aFileScore {
		t.Errorf("Evaluator should prefer rook on 2nd rank center (%d) over a-file (%d)", 
			centerScore, aFileScore)
	}

	// The difference should be exactly the positional bonus difference
	expectedDiff := 10 - (-5) // 2nd rank center - a-file penalty = 15
	actualDiff := int(centerScore - aFileScore)
	if actualDiff != expectedDiff {
		t.Errorf("Expected positional difference of %d, got %d", expectedDiff, actualDiff)
	}
}

func TestAllPiecePositions(t *testing.T) {
	evaluator := NewEvaluator()
	
	// Complex position with knights, bishops, and rooks
	// White: Knight on d4 (center), Bishop on a1 (corner), Rook on d2 (2nd rank center)
	// Black: Knight on a8 (corner), Bishop on e5 (center), Rook on a7 (7th rank a-file)
	fen := "n7/r7/8/4b3/3N4/8/3R4/B7 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	score := evaluator.Evaluate(b)

	// White knight on d4: 320 + 20 = 340
	// White bishop on a1: 330 + (-20) = 310  
	// White rook on d2: 500 + 10 = 510
	// Black knight on a8: -(320 + (-50)) = -270
	// Black bishop on e5: -(330 + 10) = -340
	// Black rook on a7: -(500 + (-5)) = -505 (a7 = rank 6, flipped to rank 1 = 2nd rank a-file = +5, negated = -5)
	// Total for white: 340 + 310 + 510 - 270 - 340 - 505 = 45
	expectedWhiteScore := ai.EvaluationScore(45)

	if score != expectedWhiteScore {
		t.Errorf("Expected white score %d, got %d", expectedWhiteScore, score)
	}
}

func TestPawnPositionalBonus(t *testing.T) {
	evaluator := NewEvaluator()
	
	tests := []struct {
		name           string
		fen            string
		expectedWhite  ai.EvaluationScore
		expectedBlack  ai.EvaluationScore
		description    string
	}{
		{
			name:           "White pawn on starting rank",
			fen:            "8/8/8/8/8/8/3P4/8 w - - 0 1", // d2
			expectedWhite:  ai.EvaluationScore(100 + (-20)), // material + positional
			expectedBlack:  ai.EvaluationScore(-(100 + (-20))),
			description:    "Pawn on d2 should have -20 positional penalty (unmoved center pawn)",
		},
		{
			name:           "White pawn advanced to 4th rank center",
			fen:            "8/8/8/8/3P4/8/8/8 w - - 0 1", // d4
			expectedWhite:  ai.EvaluationScore(100 + 20), // material + positional
			expectedBlack:  ai.EvaluationScore(-(100 + 20)),
			description:    "Pawn on d4 should have +20 positional bonus (advanced center)",
		},
		{
			name:           "White pawn advanced to 5th rank center",
			fen:            "8/8/8/3P4/8/8/8/8 w - - 0 1", // d5
			expectedWhite:  ai.EvaluationScore(100 + 25), // material + positional
			expectedBlack:  ai.EvaluationScore(-(100 + 25)),
			description:    "Pawn on d5 should have +25 positional bonus (far advanced center)",
		},
		{
			name:           "White pawn on 7th rank near promotion",
			fen:            "8/3P4/8/8/8/8/8/8 w - - 0 1", // d7
			expectedWhite:  ai.EvaluationScore(100 + 50), // material + positional
			expectedBlack:  ai.EvaluationScore(-(100 + 50)),
			description:    "Pawn on d7 should have +50 positional bonus (near promotion)",
		},
		{
			name:           "Black pawn on starting rank from black's perspective",
			fen:            "8/3p4/8/8/8/8/8/8 w - - 0 1", // d7 (black's starting rank)
			expectedWhite:  ai.EvaluationScore(-(100 + (-20))), // Black pieces are negative, penalty becomes positive
			expectedBlack:  ai.EvaluationScore(100 + (-20)),    // From black's perspective  
			description:    "Black pawn on d7 should have -20 positional penalty (unmoved center pawn)",
		},
		{
			name:           "Black pawn advanced to 4th rank from black's perspective", 
			fen:            "8/8/8/8/3p4/8/8/8 w - - 0 1", // d4 (black's 4th rank from their perspective)
			expectedWhite:  ai.EvaluationScore(-(100 + 25)), // Black pieces are negative
			expectedBlack:  ai.EvaluationScore(100 + 25),    // From black's perspective  
			description:    "Black pawn on d4 should have +25 positional bonus (advanced from black's perspective)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN %s: %v", tt.fen, err)
			}

			score := evaluator.Evaluate(b)

			if score != tt.expectedWhite {
				t.Errorf("%s: Expected score %d, got %d", tt.description, tt.expectedWhite, score)
			}
		})
	}
}

func TestPawnTableValues(t *testing.T) {
	// Test specific pawn table values
	tests := []struct {
		square   string
		rank     int
		file     int
		expected int
	}{
		{"d2", 1, 3, -20}, // starting rank center penalty (unmoved)
		{"e2", 1, 4, -20}, // starting rank center penalty (unmoved)
		{"d3", 2, 3, 0},   // 3rd rank center
		{"e3", 2, 4, 0},   // 3rd rank center
		{"d4", 3, 3, 20},  // 4th rank center
		{"e4", 3, 4, 20},  // 4th rank center
		{"d5", 4, 3, 25},  // 5th rank center
		{"e5", 4, 4, 25},  // 5th rank center
		{"a3", 2, 0, 5},   // 3rd rank edge
		{"h3", 2, 7, 5},   // 3rd rank edge
		{"d7", 6, 3, 50},  // 7th rank center - near promotion bonus
		{"e7", 6, 4, 50},  // 7th rank center - near promotion bonus
		{"c6", 5, 2, 20},  // 6th rank
		{"f6", 5, 5, 20},  // 6th rank
	}

	for _, tt := range tests {
		t.Run(tt.square, func(t *testing.T) {
			actual := PawnTable[tt.rank*8+tt.file]
			if actual != tt.expected {
				t.Errorf("Pawn table value for %s (rank %d, file %d): expected %d, got %d", 
					tt.square, tt.rank, tt.file, tt.expected, actual)
			}
		})
	}
}

func TestPawnPositionPreference(t *testing.T) {
	evaluator := NewEvaluator()

	// Test pawn on starting position vs advanced center
	startingFEN := "8/8/8/8/8/8/3P4/8 w - - 0 1" // d2
	advancedFEN := "8/8/8/8/3P4/8/8/8 w - - 0 1"  // d4

	startingBoard, _ := board.FromFEN(startingFEN)
	advancedBoard, _ := board.FromFEN(advancedFEN)

	// Evaluator should prefer advanced position over starting position (20 > -20)
	startingScore := evaluator.Evaluate(startingBoard)
	advancedScore := evaluator.Evaluate(advancedBoard)

	if advancedScore <= startingScore {
		t.Errorf("Evaluator should prefer advanced pawn (%d) over starting rank (%d)", 
			advancedScore, startingScore)
	}

	// The difference should be exactly the positional bonus difference
	expectedDiff := 20 - (-20) // advanced bonus - starting penalty = 40
	actualDiff := int(advancedScore - startingScore)
	if actualDiff != expectedDiff {
		t.Errorf("Expected positional difference of %d, got %d", expectedDiff, actualDiff)
	}
}

func TestAdvancedPawnPreference(t *testing.T) {
	evaluator := NewEvaluator()

	// Test advanced pawn vs near-promotion position
	advancedFEN := "8/8/8/3P4/8/8/8/8 w - - 0 1" // d5 (+25)
	nearPromotionFEN := "8/3P4/8/8/8/8/8/8 w - - 0 1"   // d7 (+50)

	advancedBoard, _ := board.FromFEN(advancedFEN)
	nearPromotionBoard, _ := board.FromFEN(nearPromotionFEN)

	// Evaluator should prefer near-promotion position over mid-advanced position
	advancedScore := evaluator.Evaluate(advancedBoard)
	nearPromotionScore := evaluator.Evaluate(nearPromotionBoard)

	if nearPromotionScore <= advancedScore {
		t.Errorf("Evaluator should prefer near-promotion pawn (%d) over mid-advanced (%d)", 
			nearPromotionScore, advancedScore)
	}

	// The difference should be exactly the positional bonus difference
	expectedDiff := 50 - 25 // d7 near-promotion bonus - d5 bonus = 25
	actualDiff := int(nearPromotionScore - advancedScore)
	if actualDiff != expectedDiff {
		t.Errorf("Expected positional difference of %d, got %d", expectedDiff, actualDiff)
	}
}

func TestQueenPositionalBonus(t *testing.T) {
	evaluator := NewEvaluator()
	
	tests := []struct {
		name        string
		fen         string
		expected    ai.EvaluationScore
		description string
	}{
		{
			name:        "White queen on corner",
			fen:         "8/8/8/8/8/8/8/Q7 w - - 0 1", // a1
			expected:    ai.EvaluationScore(900 + (-20)), // material + positional
			description: "Queen on a1 should have -20 positional penalty",
		},
		{
			name:        "White queen on center",
			fen:         "8/8/8/3Q4/8/8/8/8 w - - 0 1", // d5
			expected:    ai.EvaluationScore(900 + 5), // material + positional
			description: "Queen on d5 should have +5 positional bonus",
		},
		{
			name:        "White queen on good square",
			fen:         "8/8/8/8/3Q4/8/8/8 w - - 0 1", // d4
			expected:    ai.EvaluationScore(900 + 5), // material + positional
			description: "Queen on d4 should have +5 positional bonus",
		},
		{
			name:        "Black queen on corner",
			fen:         "q7/8/8/8/8/8/8/8 w - - 0 1", // a8 (black's back rank corner)
			expected:    ai.EvaluationScore(-(900 + (-20))), // Black pieces are negative from White's perspective
			description: "Black queen on a8 should have -20 positional penalty (negative for Black)",
		},
		{
			name:        "Black queen on center", 
			fen:         "8/8/8/3q4/8/8/8/8 w - - 0 1", // d5
			expected:    ai.EvaluationScore(-(900 + 5)), // Black pieces are negative from White's perspective
			description: "Black queen on d5 should have +5 positional bonus (negative for Black)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN %s: %v", tt.fen, err)
			}

			score := evaluator.Evaluate(b)

			if score != tt.expected {
				t.Errorf("%s: Expected score %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestQueenTableValues(t *testing.T) {
	// Test specific queen table values
	tests := []struct {
		square   string
		rank     int
		file     int
		expected int
	}{
		{"a1", 0, 0, -20}, // corner
		{"h1", 0, 7, -20}, // corner
		{"d4", 3, 3, 5},   // center
		{"e4", 3, 4, 5},   // center
		{"d5", 4, 3, 5},   // center
		{"e5", 4, 4, 5},   // center
		{"b1", 0, 1, -10}, // back rank
		{"g1", 0, 6, -10}, // back rank
		{"d1", 0, 3, -5},  // back rank center
		{"e1", 0, 4, -5},  // back rank center
	}

	for _, tt := range tests {
		t.Run(tt.square, func(t *testing.T) {
			actual := QueenTable[tt.rank*8+tt.file]
			if actual != tt.expected {
				t.Errorf("Queen table value for %s (rank %d, file %d): expected %d, got %d", 
					tt.square, tt.rank, tt.file, tt.expected, actual)
			}
		})
	}
}

func TestQueenPositionPreference(t *testing.T) {
	evaluator := NewEvaluator()

	// Test with queen in corner vs center
	cornerFEN := "8/8/8/8/8/8/8/Q7 w - - 0 1" // a1
	centerFEN := "8/8/8/3Q4/8/8/8/8 w - - 0 1" // d5

	cornerBoard, _ := board.FromFEN(cornerFEN)
	centerBoard, _ := board.FromFEN(centerFEN)

	// Evaluator should prefer center over corner
	cornerScore := evaluator.Evaluate(cornerBoard)
	centerScore := evaluator.Evaluate(centerBoard)

	if centerScore <= cornerScore {
		t.Errorf("Evaluator should prefer queen in center (%d) over corner (%d)", 
			centerScore, cornerScore)
	}

	// The difference should be exactly the positional bonus difference
	expectedDiff := 5 - (-20) // center bonus - corner penalty = 25
	actualDiff := int(centerScore - cornerScore)
	if actualDiff != expectedDiff {
		t.Errorf("Expected positional difference of %d, got %d", expectedDiff, actualDiff)
	}
}

func TestQueenPositionalBonusFunction(t *testing.T) {
	tests := []struct {
		piece    board.Piece
		rank     int
		file     int
		expected int
	}{
		{board.WhiteQueen, 0, 0, -20}, // a1
		{board.WhiteQueen, 3, 3, 5},   // d4
		{board.BlackQueen, 7, 0, 20},  // a8 (corner penalty, negated for black)
		{board.BlackQueen, 4, 3, -5},  // d5 (center bonus, negated for black)
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			actual := getPositionalBonus(tt.piece, tt.rank, tt.file)
			if actual != tt.expected {
				t.Errorf("getPositionalBonus(%v, %d, %d): expected %d, got %d", 
					tt.piece, tt.rank, tt.file, tt.expected, actual)
			}
		})
	}
}

func TestKingPositionalBonus(t *testing.T) {
	evaluator := NewEvaluator()
	
	tests := []struct {
		name        string
		fen         string
		expected    ai.EvaluationScore
		description string
	}{
		{
			name:        "White king on back rank corner",
			fen:         "8/8/8/8/8/8/8/K7 w - - 0 1", // a1
			expected:    ai.EvaluationScore(0 + (-30)), // material + positional
			description: "King on a1 should have -30 positional penalty",
		},
		{
			name:        "White king on back rank normal position",
			fen:         "8/8/8/8/8/8/8/6K1 w - - 0 1", // g1
			expected:    ai.EvaluationScore(0 + (-40)), // material + positional
			description: "King on g1 should have -40 positional penalty",
		},
		{
			name:        "White king on center",
			fen:         "8/8/8/8/3K4/8/8/8 w - - 0 1", // d4
			expected:    ai.EvaluationScore(0 + (-50)), // material + positional
			description: "King on d4 should have -50 positional penalty (exposed center)",
		},
		{
			name:        "White king on 8th rank edge (advanced)",
			fen:         "K7/8/8/8/8/8/8/8 w - - 0 1", // a8
			expected:    ai.EvaluationScore(0 + 20), // material + positional
			description: "King on a8 should have +20 positional bonus",
		},
		{
			name:        "Black king on back rank corner",
			fen:         "k7/8/8/8/8/8/8/8 w - - 0 1", // a8 (black's back rank corner)
			expected:    ai.EvaluationScore(-(0 + (-30))), // Black pieces are negative from White's perspective
			description: "Black king on a8 should have -30 positional penalty (negative for Black)",
		},
		{
			name:        "Black king on back rank normal position", 
			fen:         "6k1/8/8/8/8/8/8/8 w - - 0 1", // g8
			expected:    ai.EvaluationScore(-(0 + (-40))), // Black pieces are negative from White's perspective
			description: "Black king on g8 should have -40 positional penalty (negative for Black)",
		},
		{
			name:        "Black king on center", 
			fen:         "8/8/8/8/3k4/8/8/8 w - - 0 1", // d4 (from black's perspective = d5)
			expected:    ai.EvaluationScore(-(0 + (-40))), // Black pieces are negative from White's perspective
			description: "Black king on d4 should have -40 positional penalty (negative for Black)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to create board from FEN %s: %v", tt.fen, err)
			}

			score := evaluator.Evaluate(b)

			if score != tt.expected {
				t.Errorf("%s: Expected score %d, got %d", tt.description, tt.expected, score)
			}
		})
	}
}

func TestKingTableValues(t *testing.T) {
	// Test specific king table values
	tests := []struct {
		square   string
		rank     int
		file     int
		expected int
	}{
		{"a1", 0, 0, -30}, // back rank corner
		{"h1", 0, 7, -30}, // back rank corner
		{"g1", 0, 6, -40}, // back rank normal
		{"b1", 0, 1, -40}, // back rank normal
		{"d1", 0, 3, -50}, // back rank center
		{"e1", 0, 4, -50}, // back rank center
		{"d4", 3, 3, -50}, // center (exposed)
		{"e4", 3, 4, -50}, // center (exposed)
		{"a8", 7, 0, 20},  // 8th rank edge (advanced)
		{"h8", 7, 7, 20},  // 8th rank edge (advanced)
		{"g8", 7, 6, 30},  // 8th rank good position
		{"b8", 7, 1, 30},  // 8th rank good position
	}

	for _, tt := range tests {
		t.Run(tt.square, func(t *testing.T) {
			actual := KingTable[tt.rank*8+tt.file]
			if actual != tt.expected {
				t.Errorf("King table value for %s (rank %d, file %d): expected %d, got %d", 
					tt.square, tt.rank, tt.file, tt.expected, actual)
			}
		})
	}
}

func TestKingPositionPreference(t *testing.T) {
	evaluator := NewEvaluator()

	// Test with king in center vs advanced position
	centerFEN := "8/8/8/8/3K4/8/8/8 w - - 0 1" // d4 (exposed)
	advancedFEN := "6K1/8/8/8/8/8/8/8 w - - 0 1" // g8 (advanced)

	centerBoard, _ := board.FromFEN(centerFEN)
	advancedBoard, _ := board.FromFEN(advancedFEN)

	// Evaluator should prefer advanced position over center
	centerScore := evaluator.Evaluate(centerBoard)
	advancedScore := evaluator.Evaluate(advancedBoard)

	if advancedScore <= centerScore {
		t.Errorf("Evaluator should prefer advanced king (%d) over exposed center (%d)", 
			advancedScore, centerScore)
	}

	// The difference should be exactly the positional bonus difference
	expectedDiff := 30 - (-50) // advanced bonus - center penalty = 80
	actualDiff := int(advancedScore - centerScore)
	if actualDiff != expectedDiff {
		t.Errorf("Expected positional difference of %d, got %d", expectedDiff, actualDiff)
	}
}

func TestKingPositionalBonusFunction(t *testing.T) {
	tests := []struct {
		piece    board.Piece
		rank     int
		file     int
		expected int
	}{
		{board.WhiteKing, 0, 0, -30}, // a1 corner
		{board.WhiteKing, 0, 6, -40}, // g1 back rank
		{board.WhiteKing, 3, 3, -50}, // d4 center
		{board.WhiteKing, 7, 6, 30},  // g8 advanced
		{board.BlackKing, 7, 0, 30},  // a8 (corner penalty, negated for black)
		{board.BlackKing, 7, 6, 40}, // g8 (back rank penalty -40, negated for black = +40)
		{board.BlackKing, 4, 3, 50},  // d5 from black's perspective (center penalty, negated for black)
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			actual := getPositionalBonus(tt.piece, tt.rank, tt.file)
			if actual != tt.expected {
				t.Errorf("getPositionalBonus(%v, %d, %d): expected %d, got %d", 
					tt.piece, tt.rank, tt.file, tt.expected, actual)
			}
		})
	}
}

func TestKingSafetyPreference(t *testing.T) {
	evaluator := NewEvaluator()

	// Test king safety: corner vs edge vs center
	cornerFEN := "8/8/8/8/8/8/8/K7 w - - 0 1"  // a1 (-30)
	edgeFEN := "8/8/8/8/8/K7/8/8 w - - 0 1"    // a3 (-30)
	centerFEN := "8/8/8/8/3K4/8/8/8 w - - 0 1"  // d4 (-50)

	cornerBoard, _ := board.FromFEN(cornerFEN)
	edgeBoard, _ := board.FromFEN(edgeFEN)
	centerBoard, _ := board.FromFEN(centerFEN)

	cornerScore := evaluator.Evaluate(cornerBoard)
	edgeScore := evaluator.Evaluate(edgeBoard)
	centerScore := evaluator.Evaluate(centerBoard)

	// Both corner and edge should be better than center
	if cornerScore <= centerScore {
		t.Errorf("Corner king (%d) should be better than center king (%d)", cornerScore, centerScore)
	}
	if edgeScore <= centerScore {
		t.Errorf("Edge king (%d) should be better than center king (%d)", edgeScore, centerScore)
	}
}

func TestCompleteEvaluationWithAllPieces(t *testing.T) {
	evaluator := NewEvaluator()
	
	// Complex position with all piece types
	// White: Knight on d4 (+20), Bishop on a1 (-20), Rook on d2 (+10), Pawn on e4 (+25)
	// Black: Knight on a8 (-50), Bishop on e5 (+10), Rook on a7 (+5), Pawn on d5 (+20)
	fen := "n7/r7/8/3pb3/3NP3/8/3R4/B7 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	score := evaluator.Evaluate(b)

	// White pieces:
	// Knight on d4: 320 + 20 = 340
	// Bishop on a1: 330 + (-20) = 310  
	// Rook on d2: 500 + 10 = 510
	// Pawn on e4: 100 + 20 = 120
	// 
	// Black pieces:
	// Knight on a8: -(320 + (-50)) = -270
	// Bishop on e5: -(330 + 10) = -340
	// Rook on a7: -(500 + (-5)) = -505 (a7 = rank 6, flipped to rank 1 = +5, negated = -5)
	// Pawn on d5: -(100 + (-20)) = -120 (d5 = rank 4, flipped to rank 3 = +20, negated = -20)
	// 
	// Total for white: 340 + 310 + 510 + 120 - 270 - 340 - 505 - 120 = 45
	expectedWhiteScore := ai.EvaluationScore(45)

	if score != expectedWhiteScore {
		t.Errorf("Expected white score %d, got %d", expectedWhiteScore, score)
	}
}