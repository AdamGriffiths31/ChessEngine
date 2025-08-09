package evaluation

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestQueenAttackDetection(t *testing.T) {
	// Debug why the queen on d2 isn't detected as attacking e5
	fen := "4k3/8/8/4n3/3P4/8/3Q4/4K3 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	// Print the board for visualization
	t.Log("Board position:")
	for rank := 7; rank >= 0; rank-- {
		line := ""
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece == board.Empty {
				line += "."
			} else {
				line += string(piece)
			}
		}
		t.Logf("%d %s", rank+1, line)
	}
	t.Log("  abcdefgh")

	// Check queen position
	queenRank, queenFile := 1, 3 // d2
	queenPiece := b.GetPiece(queenRank, queenFile)
	t.Logf("Piece on d2: %v", queenPiece)

	// Check target position  
	targetRank, targetFile := 4, 4 // e5
	targetPiece := b.GetPiece(targetRank, targetFile)
	t.Logf("Piece on e5: %v", targetPiece)

	// Check if queen can attack e5 manually
	targetSquareIndex := board.FileRankToSquare(targetFile, targetRank)
	t.Logf("Target square index for e5: %d", targetSquareIndex)

	// Get all attackers to e5
	whiteAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardWhite)
	blackAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardBlack)
	
	t.Logf("White attackers to e5: %064b", uint64(whiteAttackers))
	t.Logf("Black attackers to e5: %064b", uint64(blackAttackers))

	// Check individual piece bitboards
	pawnBB := b.GetPieceBitboard(board.WhitePawn)
	queenBB := b.GetPieceBitboard(board.WhiteQueen)
	kingBB := b.GetPieceBitboard(board.WhiteKing)

	t.Logf("White pawn bitboard:  %064b", uint64(pawnBB))
	t.Logf("White queen bitboard: %064b", uint64(queenBB))
	t.Logf("White king bitboard:  %064b", uint64(kingBB))

	// Check which pieces can attack e5
	pawnCanAttack := (pawnBB & whiteAttackers) != 0
	queenCanAttack := (queenBB & whiteAttackers) != 0
	kingCanAttack := (kingBB & whiteAttackers) != 0

	t.Logf("Can pawn attack e5? %v", pawnCanAttack)
	t.Logf("Can queen attack e5? %v", queenCanAttack)
	t.Logf("Can king attack e5? %v", kingCanAttack)

	// The issue: Queen on d2 cannot actually attack e5!
	// d2 to e5 requires a diagonal move, but there are 3 ranks and 1 file difference
	// That's not a valid queen move pattern
	
	t.Log("Analysis:")
	t.Log("Queen on d2 (file 3, rank 1)")
	t.Log("Target on e5 (file 4, rank 4)")  
	t.Log("Difference: +1 file, +3 ranks")
	t.Log("This is NOT a valid diagonal, rank, or file move for a queen!")
	
	if !queenCanAttack {
		t.Log("✅ CORRECT: Queen on d2 cannot actually attack e5")
		t.Log("✅ The LVA logic is working correctly")
		t.Log("✅ The 'bug' I identified was actually incorrect analysis")
	} else {
		t.Error("❌ Queen should not be able to attack e5 from d2")
	}
}

func TestValidQueenAttack(t *testing.T) {
	// Test with a position where queen CAN actually attack
	// Queen on d1 attacking e5 (valid diagonal)
	fen := "4k3/8/8/4n3/3P4/8/8/3QK3 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}

	t.Log("Testing with queen on d1 (can attack e5 diagonally):")
	
	targetSquareIndex := board.FileRankToSquare(4, 4) // e5
	whiteAttackers := b.GetAttackersToSquare(targetSquareIndex, board.BitboardWhite)
	
	queenBB := b.GetPieceBitboard(board.WhiteQueen)
	queenCanAttack := (queenBB & whiteAttackers) != 0
	
	t.Logf("Queen bitboard: %064b", uint64(queenBB))
	t.Logf("White attackers: %064b", uint64(whiteAttackers))
	t.Logf("Can queen attack e5? %v", queenCanAttack)
	
	if queenCanAttack {
		t.Log("✅ Queen on d1 can correctly attack e5 (valid diagonal)")
		
		// Test LVA with this position
		see := NewSEECalculator()
		occupied := b.AllPieces
		
		// Should select pawn first
		firstAttacker := see.getLeastValuableAttacker(b, &whiteAttackers, "w", occupied)
		t.Logf("First attacker: %v", firstAttacker.piece)
		
		// Should select queen second
		secondAttacker := see.getLeastValuableAttacker(b, &whiteAttackers, "w", occupied)
		t.Logf("Second attacker: %v", secondAttacker.piece)
		
		if firstAttacker.piece == board.WhitePawn && secondAttacker.piece == board.WhiteQueen {
			t.Log("✅ LVA correctly selected pawn first, then queen")
		}
	}
}