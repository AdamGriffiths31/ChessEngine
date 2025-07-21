package evaluation

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

func TestNewMaterialEvaluator(t *testing.T) {
	evaluator := NewMaterialEvaluator()
	if evaluator == nil {
		t.Fatal("NewMaterialEvaluator should not return nil")
	}

	if evaluator.GetName() != "Material Evaluator" {
		t.Errorf("Expected name 'Material Evaluator', got '%s'", evaluator.GetName())
	}
}

func TestPieceValues(t *testing.T) {
	expectedValues := map[board.Piece]int{
		board.WhitePawn:   100,
		board.WhiteKnight: 320,
		board.WhiteBishop: 330,
		board.WhiteRook:   500,
		board.WhiteQueen:  900,
		board.WhiteKing:   0,
		board.BlackPawn:   -100,
		board.BlackKnight: -320,
		board.BlackBishop: -330,
		board.BlackRook:   -500,
		board.BlackQueen:  -900,
		board.BlackKing:   0,
	}

	for piece, expectedValue := range expectedValues {
		if PieceValues[piece] != expectedValue {
			t.Errorf("Expected value for %v to be %d, got %d", piece, expectedValue, PieceValues[piece])
		}
	}
}

func TestEvaluateEmptyBoard(t *testing.T) {
	evaluator := NewMaterialEvaluator()
	b := board.NewBoard()

	// Empty board should have score 0 for both players
	scoreWhite := evaluator.Evaluate(b, moves.White)
	scoreBlack := evaluator.Evaluate(b, moves.Black)

	if scoreWhite != 0 {
		t.Errorf("Expected score 0 for white on empty board, got %d", scoreWhite)
	}
	if scoreBlack != 0 {
		t.Errorf("Expected score 0 for black on empty board, got %d", scoreBlack)
	}
}

func TestEvaluateStartingPosition(t *testing.T) {
	evaluator := NewMaterialEvaluator()
	
	// Create board from starting position FEN
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from starting FEN: %v", err)
	}

	// Starting position should be equal for both sides
	scoreWhite := evaluator.Evaluate(b, moves.White)
	scoreBlack := evaluator.Evaluate(b, moves.Black)

	if scoreWhite != 0 {
		t.Errorf("Expected score 0 for white in starting position, got %d", scoreWhite)
	}
	if scoreBlack != 0 {
		t.Errorf("Expected score 0 for black in starting position, got %d", scoreBlack)
	}
}

func TestEvaluateWhiteAdvantage(t *testing.T) {
	evaluator := NewMaterialEvaluator()
	b := board.NewBoard()

	// Place a white queen and a black pawn
	b.SetPiece(4, 4, board.WhiteQueen) // e5
	b.SetPiece(3, 3, board.BlackPawn)  // d4

	// White should have advantage (900 - 100 = 800)
	scoreWhite := evaluator.Evaluate(b, moves.White)
	scoreBlack := evaluator.Evaluate(b, moves.Black)

	expectedWhiteScore := ai.EvaluationScore(800)  // 900 - 100
	expectedBlackScore := ai.EvaluationScore(-800) // -(900 - 100)

	if scoreWhite != expectedWhiteScore {
		t.Errorf("Expected white score %d, got %d", expectedWhiteScore, scoreWhite)
	}
	if scoreBlack != expectedBlackScore {
		t.Errorf("Expected black score %d, got %d", expectedBlackScore, scoreBlack)
	}
}

func TestEvaluateBlackAdvantage(t *testing.T) {
	evaluator := NewMaterialEvaluator()
	b := board.NewBoard()

	// Place a black rook and a white knight
	b.SetPiece(7, 0, board.BlackRook)   // a8
	b.SetPiece(0, 1, board.WhiteKnight) // b1

	// Black should have advantage (500 - 320 = 180 for black)
	scoreWhite := evaluator.Evaluate(b, moves.White)
	scoreBlack := evaluator.Evaluate(b, moves.Black)

	expectedWhiteScore := ai.EvaluationScore(-180) // 320 - 500
	expectedBlackScore := ai.EvaluationScore(180)  // -(320 - 500)

	if scoreWhite != expectedWhiteScore {
		t.Errorf("Expected white score %d, got %d", expectedWhiteScore, scoreWhite)
	}
	if scoreBlack != expectedBlackScore {
		t.Errorf("Expected black score %d, got %d", expectedBlackScore, scoreBlack)
	}
}

func TestEvaluateComplexPosition(t *testing.T) {
	evaluator := NewMaterialEvaluator()
	
	// Create a position with various pieces
	// FEN: "2r1kb1r/8/8/8/8/8/3PPP2/2QRBN2 w - - 0 1"
	// Black pieces: 2 Rooks (1000), 1 King (0), 1 Bishop (330) = -1330
	// White pieces: 1 Queen (900), 1 Rook (500), 1 Bishop (330), 1 Knight (320), 3 Pawns (300) = 2350  
	// Total: 2350 - 1330 = 1020
	b, err := board.FromFEN("2r1kb1r/8/8/8/8/8/3PPP2/2QRBN2 w - - 0 1")
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}

	scoreWhite := evaluator.Evaluate(b, moves.White)
	scoreBlack := evaluator.Evaluate(b, moves.Black)

	// Calculate expected scores
	// White pieces: Q(900) + R(500) + B(330) + N(320) + 3P(300) = 2350
	// Black pieces: 2R(1000) + B(330) + K(0) = 1330
	// White advantage: 2350 - 1330 = 1020
	expectedWhiteScore := ai.EvaluationScore(1020)
	expectedBlackScore := ai.EvaluationScore(-1020)

	if scoreWhite != expectedWhiteScore {
		t.Errorf("Expected white score %d, got %d", expectedWhiteScore, scoreWhite)
	}
	if scoreBlack != expectedBlackScore {
		t.Errorf("Expected black score %d, got %d", expectedBlackScore, scoreBlack)
	}
}

func TestEvaluateKingsOnly(t *testing.T) {
	evaluator := NewMaterialEvaluator()
	
	// Position with only kings (which have 0 value)
	b, err := board.FromFEN("8/8/8/3k4/3K4/8/8/8 w - - 0 1")
	if err != nil {
		t.Fatalf("Failed to create kings-only position: %v", err)
	}

	scoreWhite := evaluator.Evaluate(b, moves.White)
	scoreBlack := evaluator.Evaluate(b, moves.Black)

	if scoreWhite != 0 {
		t.Errorf("Expected score 0 for white with kings only, got %d", scoreWhite)
	}
	if scoreBlack != 0 {
		t.Errorf("Expected score 0 for black with kings only, got %d", scoreBlack)
	}
}

func TestEvaluateAllPieceTypes(t *testing.T) {
	evaluator := NewMaterialEvaluator()
	b := board.NewBoard()

	// Place one of each white piece type
	b.SetPiece(0, 0, board.WhiteRook)   // a1
	b.SetPiece(0, 1, board.WhiteKnight) // b1
	b.SetPiece(0, 2, board.WhiteBishop) // c1
	b.SetPiece(0, 3, board.WhiteQueen)  // d1
	b.SetPiece(0, 4, board.WhiteKing)   // e1
	b.SetPiece(1, 0, board.WhitePawn)   // a2

	// Calculate expected score for white pieces
	expectedScore := 500 + 320 + 330 + 900 + 0 + 100 // 2150

	scoreWhite := evaluator.Evaluate(b, moves.White)
	if scoreWhite != ai.EvaluationScore(expectedScore) {
		t.Errorf("Expected white score %d, got %d", expectedScore, scoreWhite)
	}

	scoreBlack := evaluator.Evaluate(b, moves.Black)
	if scoreBlack != ai.EvaluationScore(-expectedScore) {
		t.Errorf("Expected black score %d, got %d", -expectedScore, scoreBlack)
	}
}