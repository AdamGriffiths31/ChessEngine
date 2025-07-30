package uci

import (
	"testing"
	"github.com/AdamGriffiths31/ChessEngine/game"
	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
)

// TestDebugMoveFiltering analyzes exactly what happens during move filtering
func TestDebugMoveFiltering(t *testing.T) {
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
	
	// Create move generator and get legal moves
	generator := moves.NewGenerator()
	defer generator.Release()
	
	// Generate moves and analyze them
	moveList := generator.GenerateAllMoves(gameBoard, moves.White)
	defer moves.ReleaseMoveList(moveList)
	
	t.Logf("=== MOVE GENERATION DEBUG ===")
	t.Logf("Position: %s", fenPosition)
	t.Logf("Total moves generated: %d", moveList.Count)
	
	// Check if king is in check BEFORE any moves
	isInCheckBefore := generator.IsKingInCheck(gameBoard, moves.White)
	t.Logf("White king in check BEFORE moves: %v", isInCheckBefore)
	
	// Find king position
	whiteKingSquare := findKingSquare(gameBoard, board.WhiteKing)
	if whiteKingSquare != -1 {
		t.Logf("White king position: %s (square %d)", board.SquareToString(whiteKingSquare), whiteKingSquare)
		
		// Check what's attacking the king
		isAttackedByBlack := gameBoard.IsSquareAttackedByColor(whiteKingSquare, board.BitboardBlack)
		t.Logf("King square attacked by black: %v", isAttackedByBlack)
	}
	
	// Create move converter for UCI formatting
	converter := NewMoveConverter()
	
	// Analyze each generated move
	for i := 0; i < moveList.Count; i++ {
		move := moveList.Moves[i]
		moveUCI := converter.ToUCI(move)
		
		t.Logf("--- Move %d: %s ---", i, moveUCI)
		fromIndex := board.FileRankToSquare(move.From.File, move.From.Rank)
		toIndex := board.FileRankToSquare(move.To.File, move.To.Rank)
		t.Logf("  From: %s (%d), To: %s (%d)", 
			move.From.String(), fromIndex,
			move.To.String(), toIndex)
		t.Logf("  Piece: %d, Captured: %d", move.Piece, move.Captured)
		
		// Test this specific move
		if moveUCI == "d4e3" {
			t.Logf("  *** ANALYZING ILLEGAL MOVE d4e3 ***")
			
			// Manually check if this move would leave king in check
			result := analyzeMove(gameBoard, move, generator)
			t.Logf("  Manual check result: %s", result)
		}
	}
}

// findKingSquare finds the king position on the board
func findKingSquare(b *board.Board, kingPiece board.Piece) int {
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			if b.GetPiece(rank, file) == kingPiece {
				return rank*8 + file
			}
		}
	}
	return -1
}

// analyzeMove manually checks if a move would leave the king in check
func analyzeMove(b *board.Board, move board.Move, generator *moves.Generator) string {
	// Create move executor for testing
	moveExecutor := &moves.MoveExecutor{}
	
	// Make the move
	updateBoardState := func(board *board.Board, move board.Move) {
		// Board state update function
	}
	
	history := moveExecutor.MakeMove(b, move, updateBoardState)
	
	// Check if king is in check after the move
	isKingInCheck := generator.IsKingInCheck(b, moves.White)
	
	// Also check directly using bitboard method
	whiteKingBitboard := b.GetPieceBitboard(board.WhiteKing)
	directCheck := false
	if whiteKingBitboard != 0 {
		kingSquare := whiteKingBitboard.LSB()
		if kingSquare != -1 {
			directCheck = b.IsSquareAttackedByColor(kingSquare, board.BitboardBlack)
		}
	}
	
	// Unmake the move
	moveExecutor.UnmakeMove(b, history)
	
	result := ""
	if isKingInCheck {
		result += "ILLEGAL (king in check via generator)"
	} else {
		result += "LEGAL (generator says ok)"
	}
	
	if directCheck {
		result += " | ILLEGAL (direct bitboard check)"
	} else {
		result += " | LEGAL (direct bitboard check)"
	}
	
	return result
}