package evaluation

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestLVALogicCorrectness(t *testing.T) {
	// Create a position where we can verify the LVA logic step by step
	// Position: White queen and pawn both attack black knight on e5
	// LVA should select the pawn first, not the queen
	fen := "4k3/8/8/4n3/3P4/8/3Q4/4K3 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	// Target square e5
	target := board.Square{Rank: 4, File: 4}
	targetSquareIndex := board.FileRankToSquare(target.File, target.Rank)
	
	// Get all white attackers to e5
	whiteAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardWhite)
	occupied := b.AllPieces
	
	t.Logf("Position: White pawn on d4 and queen on d2 both attack knight on e5")
	t.Logf("White attackers bitboard: %064b", uint64(whiteAttackers))
	t.Logf("Occupied squares: %064b", uint64(occupied))
	
	// Let's manually check what pieces are attacking
	pawnBitboard := b.GetPieceBitboard(board.WhitePawn)
	queenBitboard := b.GetPieceBitboard(board.WhiteQueen)
	
	t.Logf("White pawn bitboard: %064b", uint64(pawnBitboard))
	t.Logf("White queen bitboard: %064b", uint64(queenBitboard))
	
	// Check the current logic: pieceBitboard & *attackers & occupied
	pawnAttackingPieces := pawnBitboard & whiteAttackers & occupied
	queenAttackingPieces := queenBitboard & whiteAttackers & occupied
	
	t.Logf("Pawn attacking pieces (current logic): %064b", uint64(pawnAttackingPieces))
	t.Logf("Queen attacking pieces (current logic): %064b", uint64(queenAttackingPieces))
	
	// Check the proposed logic: (pieceBitboard & occupied) & *attackers  
	pawnOnBoard := pawnBitboard & occupied
	queenOnBoard := queenBitboard & occupied
	pawnAttackingProposed := pawnOnBoard & whiteAttackers
	queenAttackingProposed := queenOnBoard & whiteAttackers
	
	t.Logf("Pawn on board: %064b", uint64(pawnOnBoard))
	t.Logf("Queen on board: %064b", uint64(queenOnBoard))
	t.Logf("Pawn attacking pieces (proposed logic): %064b", uint64(pawnAttackingProposed))
	t.Logf("Queen attacking pieces (proposed logic): %064b", uint64(queenAttackingProposed))
	
	// Both logics should give the same result due to associativity of bitwise AND
	if pawnAttackingPieces != pawnAttackingProposed {
		t.Errorf("❌ Pawn results differ! Current: %064b, Proposed: %064b", 
			uint64(pawnAttackingPieces), uint64(pawnAttackingProposed))
	} else {
		t.Logf("✅ Pawn results are identical")
	}
	
	if queenAttackingPieces != queenAttackingProposed {
		t.Errorf("❌ Queen results differ! Current: %064b, Proposed: %064b", 
			uint64(queenAttackingPieces), uint64(queenAttackingProposed))
	} else {
		t.Logf("✅ Queen results are identical")
	}
}

func TestLVAActualBehavior(t *testing.T) {
	// Test the actual LVA selection to see if it works correctly
	fen := "4k3/8/8/4n3/3P4/8/3Q4/4K3 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	see := NewSEECalculator()
	
	// Target square e5
	target := board.Square{Rank: 4, File: 4}
	targetSquareIndex := board.FileRankToSquare(target.File, target.Rank)
	
	// Get all white attackers
	whiteAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardWhite)
	occupied := b.AllPieces
	
	t.Logf("Testing LVA selection for white attacking e5")
	
	// First LVA call should return the pawn (least valuable)
	firstAttacker := see.getLeastValuableAttacker(b, &whiteAttackers, "w", occupied)
	t.Logf("First attacker selected: %v at square %d", firstAttacker.piece, firstAttacker.square)
	
	if firstAttacker.piece != board.WhitePawn {
		t.Errorf("❌ Expected pawn (least valuable), got %v", firstAttacker.piece)
	} else {
		t.Logf("✅ Correctly selected pawn as least valuable attacker")
	}
	
	// Second LVA call should return the queen (next least valuable)
	secondAttacker := see.getLeastValuableAttacker(b, &whiteAttackers, "w", occupied)
	t.Logf("Second attacker selected: %v at square %d", secondAttacker.piece, secondAttacker.square)
	
	if secondAttacker.piece != board.WhiteQueen {
		t.Errorf("❌ Expected queen (next available), got %v", secondAttacker.piece)
	} else {
		t.Logf("✅ Correctly selected queen as next attacker")
	}
	
	// Third call should return empty (no more attackers)
	thirdAttacker := see.getLeastValuableAttacker(b, &whiteAttackers, "w", occupied)
	t.Logf("Third attacker selected: %v at square %d", thirdAttacker.piece, thirdAttacker.square)
	
	if thirdAttacker.piece != board.Empty {
		t.Errorf("❌ Expected no more attackers (Empty), got %v", thirdAttacker.piece)
	} else {
		t.Logf("✅ Correctly returned empty when no more attackers")
	}
}

func TestBitboardLogicEquivalence(t *testing.T) {
	// Mathematical proof that the operations are equivalent
	// For any bitboards A, B, C: (A & B & C) == ((A & B) & C) == ((A & C) & B)
	
	// Create some test bitboards
	var a, b, c board.Bitboard = 0b11110000, 0b11001100, 0b10101010
	
	result1 := a & b & c
	result2 := (a & b) & c
	result3 := (a & c) & b
	result4 := (b & c) & a
	
	t.Logf("Testing bitboard AND associativity:")
	t.Logf("A: %08b", uint8(a))
	t.Logf("B: %08b", uint8(b))
	t.Logf("C: %08b", uint8(c))
	t.Logf("A & B & C: %08b", uint8(result1))
	t.Logf("(A & B) & C: %08b", uint8(result2))
	t.Logf("(A & C) & B: %08b", uint8(result3))
	t.Logf("(B & C) & A: %08b", uint8(result4))
	
	if result1 != result2 || result1 != result3 || result1 != result4 {
		t.Errorf("❌ Bitboard AND operations are not equivalent!")
	} else {
		t.Logf("✅ All bitboard AND operations are mathematically equivalent")
	}
}

func TestSEEWithActualMove(t *testing.T) {
	// Test SEE calculation to see if the LVA logic produces correct results
	fen := "4k3/8/8/4n3/3P4/8/3Q4/4K3 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	see := NewSEECalculator()
	
	// Test pawn takes knight
	pawnCapture := board.Move{
		From:      board.Square{Rank: 3, File: 3}, // d4
		To:        board.Square{Rank: 4, File: 4}, // e5
		Piece:     board.WhitePawn,
		Captured:  board.BlackKnight,
		IsCapture: true,
	}
	
	pawnSEE := see.SEE(b, pawnCapture)
	t.Logf("Pawn takes knight SEE: %d (should be +320, knight value)", pawnSEE)
	
	// Test queen takes knight  
	queenCapture := board.Move{
		From:      board.Square{Rank: 1, File: 3}, // d2
		To:        board.Square{Rank: 4, File: 4}, // e5
		Piece:     board.WhiteQueen,
		Captured:  board.BlackKnight,
		IsCapture: true,
	}
	
	queenSEE := see.SEE(b, queenCapture)
	t.Logf("Queen takes knight SEE: %d (should be +320, knight value)", queenSEE)
	
	// Both should give the same result since the knight is undefended
	if pawnSEE != 320 {
		t.Errorf("❌ Pawn capture should win knight (+320), got %d", pawnSEE)
	}
	
	if queenSEE != 320 {
		t.Errorf("❌ Queen capture should win knight (+320), got %d", queenSEE)
	}
	
	if pawnSEE == queenSEE && pawnSEE == 320 {
		t.Logf("✅ SEE calculations are correct - LVA logic is working properly")
	}
}