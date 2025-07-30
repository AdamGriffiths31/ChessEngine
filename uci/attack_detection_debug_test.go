package uci

import (
	"testing"
	"github.com/AdamGriffiths31/ChessEngine/game"
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// TestAttackDetectionAfterMove debugs why attack detection fails after d4e3 move
func TestAttackDetectionAfterMove(t *testing.T) {
	// Exact FEN position where d4e3 illegal move was generated
	fenPosition := "rn1qk2r/1b3ppp/1p2pn2/p2p4/PpPQPb2/5P1P/3K4/RNBQ1BNR w kq - 1 14"
	
	// Create game engine and set position
	engine := game.NewEngine()
	err := engine.LoadFromFEN(fenPosition)
	if err != nil {
		t.Fatalf("Failed to set FEN position: %v", err)
	}
	
	// Get the current board
	gameState := engine.GetState()
	gameBoard := gameState.Board
	
	t.Logf("=== ATTACK DETECTION DEBUG ===")
	t.Logf("Original position: %s", fenPosition)
	
	// Find king and bishop positions BEFORE move
	whiteKingSquare := findKingSquare(gameBoard, board.WhiteKing)
	blackBishopSquare := findPieceSquare(gameBoard, board.BlackBishop, 3, 5) // f4 = file 5, rank 3
	
	t.Logf("BEFORE move:")
	t.Logf("  White king at: %s (square %d)", board.SquareToString(whiteKingSquare), whiteKingSquare)
	t.Logf("  Black bishop at: %s (square %d)", board.SquareToString(blackBishopSquare), blackBishopSquare)
	
	// Check if king is attacked by bishop before move
	isAttackedBefore := gameBoard.IsSquareAttackedByColor(whiteKingSquare, board.BitboardBlack)
	t.Logf("  King attacked by black BEFORE: %v", isAttackedBefore)
	
	// Specifically check if f4 bishop attacks d2 king
	isBishopAttackingKing := checkBishopAttacksSquare(gameBoard, blackBishopSquare, whiteKingSquare)
	t.Logf("  f4 bishop attacking d2 king BEFORE: %v", isBishopAttackingKing)
	
	// Now make the d4e3 move manually
	d4Square := board.FileRankToSquare(3, 3) // d4
	e3Square := board.FileRankToSquare(4, 2) // e3
	
	t.Logf("\nMAKING MOVE d4e3:")
	t.Logf("  Moving Queen from d4 (square %d) to e3 (square %d)", d4Square, e3Square)
	
	// Get pieces before move
	queenPiece := gameBoard.GetPiece(3, 3) // d4
	capturedPiece := gameBoard.GetPiece(2, 4) // e3
	
	t.Logf("  Queen piece: %d", queenPiece)
	t.Logf("  Captured piece: %d", capturedPiece)
	
	// Make the move manually
	gameBoard.SetPiece(3, 3, board.Empty) // Clear d4
	gameBoard.SetPiece(2, 4, queenPiece)  // Place queen on e3
	
	t.Logf("\nAFTER move:")
	
	// Check positions after move
	newQueenSquare := findPieceSquare(gameBoard, board.WhiteQueen, 2, 4) // e3
	bishopStillThere := findPieceSquare(gameBoard, board.BlackBishop, 3, 5) // f4
	
	t.Logf("  White queen now at: %s (square %d)", board.SquareToString(newQueenSquare), newQueenSquare)
	t.Logf("  Black bishop still at: %s (square %d)", board.SquareToString(bishopStillThere), bishopStillThere)
	
	// Check if king is STILL attacked after the queen move
	isAttackedAfter := gameBoard.IsSquareAttackedByColor(whiteKingSquare, board.BitboardBlack)
	t.Logf("  King attacked by black AFTER: %v", isAttackedAfter)
	
	// Check if the specific bishop still attacks the king
	isBishopStillAttacking := checkBishopAttacksSquare(gameBoard, bishopStillThere, whiteKingSquare)
	t.Logf("  f4 bishop attacking d2 king AFTER: %v", isBishopStillAttacking)
	
	// THE BUG: if isAttackedAfter is false, then attack detection is broken
	if !isAttackedAfter {
		t.Errorf("BUG FOUND: King should still be in check after d4e3 move!")
		t.Errorf("The f4 bishop should still be attacking the d2 king")
		
		// Debug the diagonal
		t.Logf("\nDIAGONAL DEBUG:")
		debugDiagonal(t, gameBoard, blackBishopSquare, whiteKingSquare)
	} else {
		t.Logf("SUCCESS: Attack detection correctly identifies king still in check")
	}
}

// findPieceSquare finds a specific piece at expected position
func findPieceSquare(b *board.Board, piece board.Piece, expectedRank, expectedFile int) int {
	actualPiece := b.GetPiece(expectedRank, expectedFile)
	if actualPiece == piece {
		return board.FileRankToSquare(expectedFile, expectedRank)
	}
	return -1
}

// checkBishopAttacksSquare checks if a bishop on one square can attack another square
func checkBishopAttacksSquare(b *board.Board, bishopSquare, targetSquare int) bool {
	if bishopSquare == -1 || targetSquare == -1 {
		return false
	}
	
	// Convert to file/rank
	bishopFile, bishopRank := board.SquareToFileRank(bishopSquare)
	targetFile, targetRank := board.SquareToFileRank(targetSquare)
	
	// Check if on same diagonal
	fileDiff := absInt(targetFile - bishopFile)
	rankDiff := absInt(targetRank - bishopRank)
	
	if fileDiff != rankDiff {
		return false // Not on diagonal
	}
	
	// Check if path is clear
	fileDir := 1
	if targetFile < bishopFile {
		fileDir = -1
	}
	rankDir := 1
	if targetRank < bishopRank {
		rankDir = -1
	}
	
	// Check each square along diagonal
	for i := 1; i < fileDiff; i++ {
		checkFile := bishopFile + i*fileDir
		checkRank := bishopRank + i*rankDir
		
		if b.GetPiece(checkRank, checkFile) != board.Empty {
			return false // Path blocked
		}
	}
	
	return true // Clear diagonal path
}

// debugDiagonal prints information about the diagonal between two squares
func debugDiagonal(t *testing.T, b *board.Board, fromSquare, toSquare int) {
	fromFile, fromRank := board.SquareToFileRank(fromSquare)
	toFile, toRank := board.SquareToFileRank(toSquare)
	
	t.Logf("  From %s (%d,%d) to %s (%d,%d)", 
		board.SquareToString(fromSquare), fromFile, fromRank,
		board.SquareToString(toSquare), toFile, toRank)
	
	fileDiff := absInt(toFile - fromFile)
	rankDiff := absInt(toRank - fromRank)
	
	t.Logf("  File diff: %d, Rank diff: %d", fileDiff, rankDiff)
	
	if fileDiff == rankDiff {
		t.Logf("  On diagonal: YES")
		
		fileDir := 1
		if toFile < fromFile {
			fileDir = -1
		}
		rankDir := 1
		if toRank < fromRank {
			rankDir = -1
		}
		
		t.Logf("  Checking path from (%d,%d) to (%d,%d):", fromFile, fromRank, toFile, toRank)
		
		// Check each square along diagonal
		for i := 1; i < fileDiff; i++ {
			checkFile := fromFile + i*fileDir
			checkRank := fromRank + i*rankDir
			piece := b.GetPiece(checkRank, checkFile)
			squareIndex := board.FileRankToSquare(checkFile, checkRank)
			
			t.Logf("    Square %s (%d,%d): piece %d", 
				board.SquareToString(squareIndex), checkFile, checkRank, piece)
		}
	} else {
		t.Logf("  On diagonal: NO")
	}
}

// absInt returns absolute value
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}