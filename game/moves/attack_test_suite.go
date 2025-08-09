package moves

import (
	"fmt"
	"testing"
	"time"
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// AttackTestSuite contains comprehensive test cases for attack calculations
type AttackTestSuite struct {
	debugger *AttackDebugger
}

// NewAttackTestSuite creates a new attack test suite
func NewAttackTestSuite() *AttackTestSuite {
	return &AttackTestSuite{
		debugger: NewAttackDebugger(),
	}
}

// TestComplexPinPatterns tests positions with intricate pin arrangements
func (ats *AttackTestSuite) TestComplexPinPatterns(t *testing.T) {
	testCases := []struct {
		name        string
		fen         string
		description string
	}{
		{
			"Multiple Pins Same Direction",
			"r3k3/8/8/3PPP2/8/8/8/4K2R w - - 0 1",
			"Multiple pieces pinned by same rook",
		},
		{
			"Cross Pins",
			"4k3/8/8/3P4/2KPP3/3P4/8/1q2r3 w - - 0 1", 
			"King pinned by rook and queen from different directions",
		},
		{
			"Diagonal Pin Chain",
			"4k3/8/2b5/3P4/4P3/5P2/6K1/8 w - - 0 1",
			"Multiple pieces pinned on same diagonal",
		},
		{
			"Pin Through Pin", 
			"4k3/8/8/8/1r1PP1R1/8/8/4K3 w - - 0 1",
			"Piece pinned by rook attacking through another pinned piece",
		},
		{
			"Pin With Check",
			"4k3/8/8/8/3r4/2K1P3/8/7R w - - 0 1",
			"King in check with piece pinned by checking rook",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Printf("Testing: %s\n", tc.description)
			fmt.Printf("FEN: %s\n", tc.fen)
			
			// Enable attack comparison mode for detailed validation
			originalMode := AttackComparisonMode
			AttackComparisonMode = true
			defer func() { AttackComparisonMode = originalMode }()
			
			ats.debugger.QuickAttackTest(tc.fen)
		})
	}
}

// TestTacticalPositions tests positions with multiple simultaneous attacks
func (ats *AttackTestSuite) TestTacticalPositions(t *testing.T) {
	testCases := []struct {
		name        string
		fen         string
		description string
	}{
		{
			"Fork Position",
			"rnbqkbnr/pppp1ppp/8/4p3/2B1P3/8/PPPP1PPP/RNBQK1NR b KQkq - 2 2",
			"Bishop attacking multiple pieces (fork potential)",
		},
		{
			"Double Attack",
			"r1bqkb1r/pppp1ppp/2n2n2/1B2p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4",
			"Multiple pieces under attack simultaneously",
		},
		{
			"Discovered Attack Setup",
			"r3k2r/8/8/3b4/2KP4/3B4/8/8 w - - 0 1",
			"Pieces that can create discovered attacks",
		},
		{
			"X-Ray Attack",
			"4k3/8/8/8/3QP3/8/3r4/4K3 w - - 0 1",
			"Queen and rook attacking through pieces",
		},
		{
			"Battery Attack",
			"4k3/8/8/8/8/2QQ4/8/4K3 w - - 0 1",
			"Multiple pieces attacking same direction", 
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Printf("Testing: %s\n", tc.description)
			ats.debugger.QuickAttackTest(tc.fen)
		})
	}
}

// TestSpecialMoveEdgeCases tests en passant and castling attack scenarios
func (ats *AttackTestSuite) TestSpecialMoveEdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		fen         string
		description string
	}{
		{
			"En Passant Pin",
			"4k3/8/8/2KPp1r1/8/8/8/8 w - e6 0 1",
			"En passant move would expose king to attack",
		},
		{
			"En Passant Unpin", 
			"4k3/8/8/r1KPp3/8/8/8/8 w - e6 0 1",
			"En passant capture removes pinning piece",
		},
		{
			"Castling Through Check",
			"r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			"King would move through attacked square while castling",
		},
		{
			"Castling Into Check", 
			"r3k2r/8/8/8/8/8/6r1/R3K2R w KQkq - 0 1",
			"King would end up in check after castling",
		},
		{
			"Castling With Pinned Rook",
			"r3k3/8/8/8/8/8/8/R3K2r w Qq - 0 1",
			"Rook is pinned and cannot castle",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Printf("Testing: %s\n", tc.description)
			ats.debugger.QuickAttackTest(tc.fen)
		})
	}
}

// TestKingSafetyScenarios tests various king safety edge cases
func (ats *AttackTestSuite) TestKingSafetyScenarios(t *testing.T) {
	testCases := []struct {
		name        string
		fen         string
		description string
	}{
		{
			"King in Corner",
			"7k/6pp/8/8/8/8/8/7K w - - 0 1",
			"King safety when trapped in corner",
		},
		{
			"King on Edge",
			"8/4k3/8/8/8/8/8/4K3 w - - 0 1",
			"King safety on board edge",
		},
		{
			"King Surrounded by Own Pieces",
			"8/8/8/2PPP3/2PKP3/2PPP3/8/8 w - - 0 1",
			"King surrounded by own pieces",
		},
		{
			"King in Open Center",
			"8/8/8/8/3K4/8/8/4k3 w - - 0 1",
			"King exposed in center of board",
		},
		{
			"King Near Enemy Pieces",
			"8/8/8/2rrr3/2rKr3/2rrr3/8/8 w - - 0 1",
			"King surrounded by enemy pieces",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Printf("Testing: %s\n", tc.description)
			ats.debugger.QuickAttackTest(tc.fen)
		})
	}
}

// BenchmarkAttackCalculations benchmarks attack calculation performance
func (ats *AttackTestSuite) BenchmarkAttackCalculations(b *testing.B) {
	positions := []struct {
		name string
		fen  string
	}{
		{"Initial", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"Middlegame", "r2qkb1r/ppp2ppp/2npbn2/1B2p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 0 1"},
		{"Tactical", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - "},
		{"Endgame", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - "},
	}
	
	bmg := NewBitboardMoveGenerator()
	
	for _, pos := range positions {
		b.Run("PinCalculation_"+pos.name, func(b *testing.B) {
			testBoard, _ := board.FromFEN(pos.fen)
			kingSquare := testBoard.GetPieceBitboard(board.WhiteKing).LSB()
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = bmg.calculatePinnedPieces(testBoard, kingSquare, board.BitboardBlack)
			}
		})
		
		b.Run("CheckDetection_"+pos.name, func(b *testing.B) {
			testBoard, _ := board.FromFEN(pos.fen)
			kingSquare := testBoard.GetPieceBitboard(board.WhiteKing).LSB()
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = testBoard.IsSquareAttackedByColor(kingSquare, board.BitboardBlack)
			}
		})
	}
}

// RunFullAttackTestSuite runs all attack calculation tests
func (ats *AttackTestSuite) RunFullAttackTestSuite(t *testing.T) {
	fmt.Printf("=== COMPREHENSIVE ATTACK TEST SUITE ===\n\n")
	
	fmt.Printf("Phase 1: Complex Pin Patterns\n")
	ats.TestComplexPinPatterns(t)
	
	fmt.Printf("\nPhase 2: Tactical Positions\n") 
	ats.TestTacticalPositions(t)
	
	fmt.Printf("\nPhase 3: Special Move Edge Cases\n")
	ats.TestSpecialMoveEdgeCases(t)
	
	fmt.Printf("\nPhase 4: King Safety Scenarios\n")
	ats.TestKingSafetyScenarios(t)
	
	fmt.Printf("\n=== ALL ATTACK TESTS COMPLETE ===\n")
}

// ValidateAttackAccuracy validates attack calculation accuracy across test positions
func (ats *AttackTestSuite) ValidateAttackAccuracy() {
	fmt.Printf("=== ATTACK ACCURACY VALIDATION ===\n")
	
	// Test positions that previously caused issues
	problemPositions := []string{
		"1nbqkbnr/1ppppppp/8/r7/8/8/P1PPPPPP/RNBQKBNR w KQk - 0 3",   // Previous rook diagonal pin bug
		"8/8/3p4/1Pp4r/1K5k/5p2/4P1P1/1R6 w - c6 0 3",                // Previous en passant bug
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - ",  // Complex tactical position
	}
	
	for i, fen := range problemPositions {
		fmt.Printf("\nTesting problem position %d:\n", i+1)
		fmt.Printf("FEN: %s\n", fen)
		
		start := time.Now()
		ats.debugger.QuickAttackTest(fen) 
		duration := time.Since(start)
		
		fmt.Printf("Test completed in: %v\n", duration)
	}
	
	fmt.Printf("\n=== ACCURACY VALIDATION COMPLETE ===\n")
}