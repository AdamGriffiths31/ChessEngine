package moves

import (
	"fmt"
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// AttackDebugger provides comprehensive debugging tools for attack-related issues
type AttackDebugger struct {
	bmg *BitboardMoveGenerator
}

// NewAttackDebugger creates a new attack debugging instance
func NewAttackDebugger() *AttackDebugger {
	return &AttackDebugger{
		bmg: NewBitboardMoveGenerator(),
	}
}

// FindAttackDiscrepancy automatically finds the first position where attack calculations differ
// This is similar to our successful findDiscrepancy function but focused on attack calculations
func (ad *AttackDebugger) FindAttackDiscrepancy(b *board.Board, depth int, player Player, path string) {
	if depth == 0 {
		return
	}
	
	// Enable attack comparison mode for this position
	originalMode := AttackComparisonMode
	AttackComparisonMode = true
	defer func() { AttackComparisonMode = originalMode }()
	
	// Test current position for attack calculation issues
	ad.validatePositionAttacks(b, player, path)
	
	// Generate moves to continue searching
	moveList := ad.bmg.GenerateAllMovesBitboard(b, player)
	
	// Continue searching deeper positions
	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		
		undo, err := b.MakeMoveWithUndo(move)
		if err != nil {
			continue
		}
		
		nextPlayer := Black
		if player == Black {
			nextPlayer = White
		}
		
		moveStr := fmt.Sprintf("%s%s", move.From.String(), move.To.String())
		newPath := path
		if newPath != "" {
			newPath += " "
		}
		newPath += moveStr
		
		ad.FindAttackDiscrepancy(b, depth-1, nextPlayer, newPath)
		
		b.UnmakeMove(undo)
	}
}

// validatePositionAttacks checks for attack calculation issues in a specific position
func (ad *AttackDebugger) validatePositionAttacks(b *board.Board, player Player, path string) {
	var ourKingPiece board.Piece
	var opponentColor board.BitboardColor
	
	if player == White {
		ourKingPiece = board.WhiteKing
		opponentColor = board.BitboardBlack
	} else {
		ourKingPiece = board.BlackKing
		opponentColor = board.BitboardWhite
	}
	
	// Find king position
	ourKingBitboard := b.GetPieceBitboard(ourKingPiece)
	if ourKingBitboard == 0 {
		return
	}
	kingSquare := ourKingBitboard.LSB()
	if kingSquare == -1 {
		return
	}
	
	// Test pin detection accuracy
	ad.validatePinDetection(b, kingSquare, opponentColor, path)
	
	// Test check detection accuracy
	ad.validateCheckDetection(b, kingSquare, opponentColor, path)
	
	// Test king safety calculations
	ad.validateKingSafety(b, kingSquare, opponentColor, path)
}

// validatePinDetection checks pin detection accuracy
func (ad *AttackDebugger) validatePinDetection(b *board.Board, kingSquare int, opponentColor board.BitboardColor, path string) {
	currentPins := ad.bmg.calculatePinnedPieces(b, kingSquare, opponentColor)
	referencePins := ad.bmg.calculatePinnedPiecesReference(b, kingSquare, opponentColor)
	
	if currentPins != referencePins {
		fmt.Printf("=== PIN DETECTION ISSUE FOUND ===\n")
		fmt.Printf("FEN: %s\n", b.ToFEN())
		fmt.Printf("Path: %s\n", path)
		fmt.Printf("Current pins:  %064b\n", currentPins)
		fmt.Printf("Reference pins: %064b\n", referencePins)
		fmt.Printf("Difference:    %064b\n", currentPins^referencePins)
		
		// Show which pieces are affected
		difference := currentPins ^ referencePins
		for difference != 0 {
			square, newBitboard := difference.PopLSB()
			difference = newBitboard
			
			piece := b.GetPieceOnSquare(square)
			fmt.Printf("Affected piece: %c on %s\n", rune(piece), board.SquareToString(square))
		}
		fmt.Printf("================================\n\n")
	}
}

// validateCheckDetection checks check detection accuracy
func (ad *AttackDebugger) validateCheckDetection(b *board.Board, kingSquare int, opponentColor board.BitboardColor, path string) {
	currentCheck := b.IsSquareAttackedByColor(kingSquare, opponentColor)
	referenceCheck := ad.bmg.isSquareAttackedByColorReference(b, kingSquare, opponentColor)
	
	if currentCheck != referenceCheck {
		fmt.Printf("=== CHECK DETECTION ISSUE FOUND ===\n")
		fmt.Printf("FEN: %s\n", b.ToFEN())
		fmt.Printf("Path: %s\n", path)
		fmt.Printf("King on: %s\n", board.SquareToString(kingSquare))
		fmt.Printf("Current check:  %t\n", currentCheck)
		fmt.Printf("Reference check: %t\n", referenceCheck)
		
		// Show attackers if any
		if currentCheck || referenceCheck {
			attackers := b.GetAttackersToSquare(kingSquare, opponentColor)
			fmt.Printf("Attackers: %064b\n", attackers)
			
			for attackers != 0 {
				attackerSquare, newBitboard := attackers.PopLSB()
				attackers = newBitboard
				
				piece := b.GetPieceOnSquare(attackerSquare)
				fmt.Printf("Attacking piece: %c on %s\n", rune(piece), board.SquareToString(attackerSquare))
			}
		}
		fmt.Printf("==================================\n\n")
	}
}

// validateKingSafety checks king safety calculation accuracy
func (ad *AttackDebugger) validateKingSafety(b *board.Board, kingSquare int, opponentColor board.BitboardColor, path string) {
	// Test squares around the king
	kingRank := kingSquare / 8
	kingFile := kingSquare % 8
	
	safetyIssues := []string{}
	
	for dr := -1; dr <= 1; dr++ {
		for df := -1; df <= 1; df++ {
			if dr == 0 && df == 0 {
				continue
			}
			
			newRank := kingRank + dr
			newFile := kingFile + df
			
			if newRank < 0 || newRank > 7 || newFile < 0 || newFile > 7 {
				continue
			}
			
			testSquare := newRank*8 + newFile
			
			currentAttacked := b.IsSquareAttackedByColor(testSquare, opponentColor)
			referenceAttacked := ad.bmg.isSquareAttackedByColorReference(b, testSquare, opponentColor)
			
			if currentAttacked != referenceAttacked {
				issue := fmt.Sprintf("Square %s: current=%t, reference=%t",
					board.SquareToString(testSquare), currentAttacked, referenceAttacked)
				safetyIssues = append(safetyIssues, issue)
			}
		}
	}
	
	if len(safetyIssues) > 0 {
		fmt.Printf("=== KING SAFETY ISSUES FOUND ===\n")
		fmt.Printf("FEN: %s\n", b.ToFEN())
		fmt.Printf("Path: %s\n", path)
		fmt.Printf("King on: %s\n", board.SquareToString(kingSquare))
		
		for _, issue := range safetyIssues {
			fmt.Printf("  %s\n", issue)
		}
		fmt.Printf("===============================\n\n")
	}
}

// QuickAttackTest runs a fast attack validation test on a specific position
func (ad *AttackDebugger) QuickAttackTest(fen string) {
	fmt.Printf("=== QUICK ATTACK TEST ===\n")
	fmt.Printf("FEN: %s\n", fen)
	
	b, err := board.FromFEN(fen)
	if err != nil {
		fmt.Printf("Error parsing FEN: %v\n", err)
		return
	}
	
	// Test both colors
	for _, player := range []Player{White, Black} {
		fmt.Printf("\nTesting %s to move:\n", player)
		ad.validatePositionAttacks(b, player, "direct-test")
	}
	
	fmt.Printf("========================\n")
}

// PerftAttackComparison runs perft while comparing attack calculations
func (ad *AttackDebugger) PerftAttackComparison(fen string, depth int) {
	fmt.Printf("=== PERFT ATTACK COMPARISON ===\n")
	fmt.Printf("FEN: %s\n", fen)
	fmt.Printf("Depth: %d\n", depth)
	
	b, err := board.FromFEN(fen)
	if err != nil {
		fmt.Printf("Error parsing FEN: %v\n", err)
		return
	}
	
	// Determine starting player from FEN
	player := White
	if b.GetSideToMove() == "b" {
		player = Black
	}
	
	fmt.Printf("Starting search for attack discrepancies...\n\n")
	
	ad.FindAttackDiscrepancy(b, depth, player, "")
	
	fmt.Printf("Attack comparison complete.\n")
	fmt.Printf("==============================\n")
}