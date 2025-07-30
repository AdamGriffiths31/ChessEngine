package moves

import (
	"fmt"
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator()
	if gen == nil {
		t.Fatal("Expected generator to be non-nil")
	}
}

func TestGenerateAllMoves_InitialPosition(t *testing.T) {
	gen := NewGenerator()
	b, _ := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	// Test white moves - should include pawns (16) + knights (4) + king (0) = 20 moves
	// Rooks, bishops, queens are blocked by pawns in initial position
	// King has no legal moves (blocked by pawns)
	whiteMoves := gen.GenerateAllMoves(b, White)
	if whiteMoves.Count != 20 {
		t.Errorf("Expected 20 white moves from initial position (16 pawn + 4 knight + 0 king), got %d", whiteMoves.Count)
	}
	
	// Test black moves - should include pawns (16) + knights (4) + king (0) = 20 moves
	blackMoves := gen.GenerateAllMoves(b, Black)
	if blackMoves.Count != 20 {
		t.Errorf("Expected 20 black moves from initial position (16 pawn + 4 knight + 0 king), got %d", blackMoves.Count)
	}
}


func TestMoveList_AddMove(t *testing.T) {
	ml := NewMoveList()
	
	if ml.Count != 0 {
		t.Errorf("Expected empty move list to have count 0, got %d", ml.Count)
	}
	
	move := board.Move{
		From: board.Square{File: 4, Rank: 1},
		To:   board.Square{File: 4, Rank: 3},
	}
	
	ml.AddMove(move)
	
	if ml.Count != 1 {
		t.Errorf("Expected move list count to be 1 after adding move, got %d", ml.Count)
	}
	
	if len(ml.Moves) != 1 {
		t.Errorf("Expected moves slice length to be 1, got %d", len(ml.Moves))
	}
}

func TestMoveList_Contains(t *testing.T) {
	ml := NewMoveList()
	
	move := board.Move{
		From:      board.Square{File: 4, Rank: 1},
		To:        board.Square{File: 4, Rank: 3},
		Promotion: board.Empty,
	}
	
	ml.AddMove(move)
	
	if !ml.Contains(move) {
		t.Error("Expected move list to contain the added move")
	}
	
	differentMove := board.Move{
		From:      board.Square{File: 4, Rank: 1},
		To:        board.Square{File: 4, Rank: 2},
		Promotion: board.Empty,
	}
	
	if ml.Contains(differentMove) {
		t.Error("Expected move list to not contain different move")
	}
}

func TestMovesEqual(t *testing.T) {
	move1 := board.Move{
		From:      board.Square{File: 4, Rank: 1},
		To:        board.Square{File: 4, Rank: 3},
		Promotion: board.Empty,
	}
	
	move2 := board.Move{
		From:      board.Square{File: 4, Rank: 1},
		To:        board.Square{File: 4, Rank: 3},
		Promotion: board.Empty,
	}
	
	if !MovesEqual(move1, move2) {
		t.Error("Expected identical moves to be equal")
	}
	
	move3 := board.Move{
		From:      board.Square{File: 4, Rank: 1},
		To:        board.Square{File: 4, Rank: 2},
		Promotion: board.Empty,
	}
	
	if MovesEqual(move1, move3) {
		t.Error("Expected different moves to not be equal")
	}
}

func TestKingCache(t *testing.T) {
	gen := NewGenerator()
	
	// Test initial state - cache should be invalid
	if gen.kingCacheValid {
		t.Error("Expected king cache to be invalid initially")
	}
	if gen.whiteKingPos != nil {
		t.Error("Expected white king position to be nil initially")
	}
	if gen.blackKingPos != nil {
		t.Error("Expected black king position to be nil initially")
	}
	
	// Load a standard starting position
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}
	
	// Test cache initialization through findKing
	whiteKingPos := gen.findKing(b, White)
	if whiteKingPos == nil {
		t.Error("Expected to find white king")
	}
	
	// Cache should now be valid
	if !gen.kingCacheValid {
		t.Error("Expected king cache to be valid after findKing call")
	}
	
	// Verify cached positions
	expectedWhiteKing := board.Square{File: 4, Rank: 0} // e1
	expectedBlackKing := board.Square{File: 4, Rank: 7} // e8
	
	if gen.whiteKingPos == nil || *gen.whiteKingPos != expectedWhiteKing {
		t.Errorf("Expected white king at %v, got %v", expectedWhiteKing, gen.whiteKingPos)
	}
	if gen.blackKingPos == nil || *gen.blackKingPos != expectedBlackKing {
		t.Errorf("Expected black king at %v, got %v", expectedBlackKing, gen.blackKingPos)
	}
	
	// Test that subsequent findKing calls use cache (should return same pointer)
	whiteKingPos2 := gen.findKing(b, White)
	if whiteKingPos != whiteKingPos2 {
		t.Error("Expected findKing to return cached position (same pointer)")
	}
	
	// Test cache update when king moves
	move := board.Move{
		From:      board.Square{File: 4, Rank: 0}, // e1
		To:        board.Square{File: 5, Rank: 0}, // f1
		Piece:     board.WhiteKing,
		IsCapture: false,
		Captured:  board.Empty,
		Promotion: board.Empty,
	}
	
	// Update the cache (simulating what happens during move execution)
	gen.updateKingCache(move)
	
	// Verify white king position was updated
	expectedNewWhiteKing := board.Square{File: 5, Rank: 0} // f1
	if gen.whiteKingPos == nil || *gen.whiteKingPos != expectedNewWhiteKing {
		t.Errorf("Expected white king at %v after move, got %v", expectedNewWhiteKing, gen.whiteKingPos)
	}
	
	// Black king position should remain unchanged
	if gen.blackKingPos == nil || *gen.blackKingPos != expectedBlackKing {
		t.Errorf("Expected black king to remain at %v, got %v", expectedBlackKing, gen.blackKingPos)
	}
	
	// Test cache initialization through findKing in GenerateAllMoves context
	gen2 := NewGenerator()
	moves := gen2.GenerateAllMoves(b, White)
	
	// Trigger cache initialization by calling findKing
	whiteKingPos3 := gen2.findKing(b, White)
	blackKingPos3 := gen2.findKing(b, Black)
	
	// Cache should be initialized after findKing calls
	if !gen2.kingCacheValid {
		t.Error("Expected king cache to be valid after findKing calls")
	}
	if gen2.whiteKingPos == nil || *gen2.whiteKingPos != expectedWhiteKing {
		t.Errorf("Expected white king cached at %v, got %v", expectedWhiteKing, gen2.whiteKingPos)
	}
	if gen2.blackKingPos == nil || *gen2.blackKingPos != expectedBlackKing {
		t.Errorf("Expected black king cached at %v, got %v", expectedBlackKing, gen2.blackKingPos)
	}
	
	// Verify the positions returned are correct
	if whiteKingPos3 == nil || *whiteKingPos3 != expectedWhiteKing {
		t.Errorf("Expected findKing to return white king at %v, got %v", expectedWhiteKing, whiteKingPos3)
	}
	if blackKingPos3 == nil || *blackKingPos3 != expectedBlackKing {
		t.Errorf("Expected findKing to return black king at %v, got %v", expectedBlackKing, blackKingPos3)
	}
	
	// Verify we got some moves
	if moves.Count == 0 {
		t.Error("Expected to generate some moves")
	}
}

// TestActualIllegalF6F7 tests the actual game position where f6f7 was reported as illegal
// This is the real position from the UCI logs where cutechess-cli rejected f6f7
func TestActualIllegalB3F7(t *testing.T) {
	gen := NewGenerator()
	
	// Start from initial position 
	b, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("Failed to create board from FEN: %v", err)
	}
	
	// Apply the EXACT game moves from the UCI logs - this includes the erroneous f1e1 that causes desync
	// From logs: position startpos moves c2c4 g8f6 g1f3 a7a6 g2g3 a6a5 f1g2 a5a4 d1c2 a4a3 e1g1 b7b6 b2b3 b6b5 b3b4 c7c6 c4c5 d7d6 d2d4 d6d5 c1a3 e7e6 f3e5 g7g6 c2b3 g6g5 a3c1 g5g4 f1e1 h7h6 b1c3 h6h5 c1g5 h5h4 g5h4 b8a6 e5c6 e6e5 d4e5 d5d4 c6d8 d4d3 b3f7
	realGameMoves := []string{
		"c2c4", "g8f6", "g1f3", "a7a6", "g2g3", "a6a5", "f1g2", "a5a4", "d1c2", "a4a3", 
		"e1g1", "b7b6", "b2b3", "b6b5", "b3b4", "c7c6", "c4c5", "d7d6", "d2d4", "d6d5", 
		"c1a3", "e7e6", "f3e5", "g7g6", "c2b3", "g6g5", "a3c1", "g5g4", // problematic next move:
		"f1e1", "h7h6", // NOTE: f1e1 is impossible after e1g1 castling - this is the desync!
		"b1c3", "h6h5", "c1g5", "h5h4", "g5h4", "b8a6", "e5c6", "e6e5", "d4e5", "d5d4", 
		"c6d8", "d4d3", // This should be move 42
		// Move 43 will be "b3f7" - the problematic move
	}
	
	t.Logf("Applying %d moves from the actual game", len(realGameMoves))
	
	// Apply each move using proper UCI conversion (like the engine does)
	for i, moveStr := range realGameMoves {
		// Check what piece is at the from square before conversion
		if len(moveStr) >= 4 {
			fromFile := int(moveStr[0] - 'a')
			fromRank := int(moveStr[1] - '1')
			pieceAtFrom := b.GetPiece(fromRank, fromFile)
			t.Logf("Move %d (%s): Piece at %c%c = %c", i+1, moveStr, moveStr[0], moveStr[1], pieceAtFrom)
		}
		
		// Use proper UCI conversion logic like the engine does
		move, err := func() (board.Move, error) {
			// Manual UCI parsing with proper castling detection
			if len(moveStr) < 4 {
				return board.Move{}, fmt.Errorf("invalid move format: %s", moveStr)
			}
			
			fromFile := int(moveStr[0] - 'a')
			fromRank := int(moveStr[1] - '1')
			toFile := int(moveStr[2] - 'a')
			toRank := int(moveStr[3] - '1')
			
			piece := b.GetPiece(fromRank, fromFile)
			if piece == board.Empty {
				return board.Move{}, fmt.Errorf("no piece at from square %s", moveStr[:2])
			}
			
			captured := b.GetPiece(toRank, toFile)
			
			move := board.Move{
				From:      board.Square{File: fromFile, Rank: fromRank},
				To:        board.Square{File: toFile, Rank: toRank},
				Piece:     piece,
				Captured:  captured,
				IsCapture: captured != board.Empty,
			}
			
			// Handle promotion
			if len(moveStr) == 5 {
				switch moveStr[4] {
				case 'q': 
					if piece == board.BlackPawn {
						move.Promotion = board.BlackQueen
					} else {
						move.Promotion = board.WhiteQueen
					}
				case 'r': 
					if piece == board.BlackPawn {
						move.Promotion = board.BlackRook
					} else {
						move.Promotion = board.WhiteRook
					}
				case 'b': 
					if piece == board.BlackPawn {
						move.Promotion = board.BlackBishop
					} else {
						move.Promotion = board.WhiteBishop
					}
				case 'n': 
					if piece == board.BlackPawn {
						move.Promotion = board.BlackKnight
					} else {
						move.Promotion = board.WhiteKnight
					}
				}
			}
			
			// Detect castling: King moves 2 squares horizontally
			if (piece == board.WhiteKing || piece == board.BlackKing) && 
			   abs(toFile - fromFile) == 2 && toRank == fromRank {
				move.IsCastling = true
			}
			
			// Detect en passant: Pawn moves diagonally to empty square with en passant target
			if (piece == board.WhitePawn || piece == board.BlackPawn) && 
			   captured == board.Empty && abs(toFile - fromFile) == 1 {
				enPassantTarget := b.GetEnPassantTarget()
				if enPassantTarget != nil && 
				   toFile == enPassantTarget.File && toRank == enPassantTarget.Rank {
					move.IsEnPassant = true
				}
			}
			
			return move, nil
		}()
		
		if err != nil {
			t.Fatalf("Failed to parse move %s at index %d: %v", moveStr, i, err)
		}
		
		err = b.MakeMove(move)
		if err != nil {
			t.Fatalf("Failed to make move %s at index %d: %v", moveStr, i, err)
		}
		
		t.Logf("Applied move %d: %s, FEN: %s", i+1, moveStr, b.ToFEN())
	}
	
	t.Logf("Final position after all moves: %s", b.ToFEN())
	
	// Position should be ready for White to move (since it's White's turn)
	
	// This should be the position: r1bNkb1r/5p2/n4q2/1pP1P3/1P4pB/1QNp2P1/P3PPBP/R3R1K1 w kq - 0 22
	finalFEN := b.ToFEN()
	expectedFEN := "r1bNkb1r/5p2/n4q2/1pP1P3/1P4pB/1QNp2P1/P3PPBP/R3R1K1 w kq - 0 22"
	
	t.Logf("Position after %d moves: %s", len(realGameMoves), finalFEN)
	t.Logf("Expected position from logs: %s", expectedFEN)
	
	if finalFEN != expectedFEN {
		t.Logf("WARNING: FEN mismatch! Our manual application differs from engine logs")
	} else {
		t.Logf("SUCCESS: FEN matches the engine logs exactly")
	}
	
	// Print the board visually
	t.Logf("Board position:")
	t.Logf("  a b c d e f g h")
	for rank := 7; rank >= 0; rank-- {
		row := fmt.Sprintf("%d ", rank+1)
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece == board.Empty {
				row += ". "
			} else {
				// Simple piece representation
				pieceChar := "?"
				switch piece {
				case board.WhitePawn: pieceChar = "P"
				case board.BlackPawn: pieceChar = "p"
				case board.WhiteRook: pieceChar = "R"
				case board.BlackRook: pieceChar = "r"
				case board.WhiteKnight: pieceChar = "N"
				case board.BlackKnight: pieceChar = "n"
				case board.WhiteBishop: pieceChar = "B"
				case board.BlackBishop: pieceChar = "b"
				case board.WhiteQueen: pieceChar = "Q"
				case board.BlackQueen: pieceChar = "q"
				case board.WhiteKing: pieceChar = "K"
				case board.BlackKing: pieceChar = "k"
				}
				row += pieceChar + " "
			}
		}
		t.Logf("%s", row)
	}
	t.Logf("  a b c d e f g h")
	
	// Now test if b3f7 should be legal
	t.Logf("\n=== TESTING b3f7 LEGALITY ===")
	
	// Generate all legal moves for White (since it's White to move)
	whiteMoves := gen.GenerateAllMoves(b, White)
	t.Logf("White has %d legal moves:", whiteMoves.Count)
	
	// Check if b3f7 is among the legal moves
	b3f7Found := false
	b3f7Index := -1
	for i := 0; i < whiteMoves.Count; i++ { // Check ALL moves, not just first 20
		move := whiteMoves.Moves[i]
		moveStr := move.From.String() + move.To.String()
		if i < 20 { // Only log first 20 for readability
			t.Logf("  [%d]: %s (From=%s, To=%s, Piece=%d)", i, moveStr, move.From.String(), move.To.String(), move.Piece)
		}
		
		if moveStr == "b3f7" {
			b3f7Found = true
			b3f7Index = i
			t.Logf("  [%d]: %s (From=%s, To=%s, Piece=%d) ← FOUND b3f7!", i, moveStr, move.From.String(), move.To.String(), move.Piece)
		}
	}
	
	if whiteMoves.Count > 20 {
		t.Logf("  ... and %d more moves", whiteMoves.Count - 20)
	}
	
	t.Logf("\n=== RESULT ===")
	if b3f7Found {
		t.Logf("SUCCESS: b3f7 found in legal moves at index %d", b3f7Index)
		t.Logf("This means our engine's rejection was WRONG - the move should be legal")
	} else {
		t.Logf("CONFIRMED: b3f7 NOT found in legal moves")
		t.Logf("This means our engine was correct to reject it, but cutechess-cli sent it anyway")
		
		// Check what piece is actually on b3
		pieceOnB3 := b.GetPiece(2, 1) // rank 2 (3rd rank), file 1 (b file)
		t.Logf("Piece on b3: %d", pieceOnB3)
		
		// Check what piece is on f7
		pieceOnF7 := b.GetPiece(6, 5) // rank 6 (7th rank), file 5 (f file)
		t.Logf("Piece on f7: %d", pieceOnF7)
		
		if pieceOnB3 == board.Empty {
			t.Logf("ERROR: No piece on b3! Move b3f7 is impossible")
		} else if pieceOnB3 != board.WhiteQueen {
			t.Logf("ERROR: Piece on b3 is not White Queen! Move b3f7 might be wrong piece type")
		}
	}
	
	// Check if the Black king is in check
	blackKingPos := gen.findKing(b, Black)
	if blackKingPos == nil {
		t.Fatal("Could not find Black king")
	}
	t.Logf("Black king position: %s", blackKingPos.String())
	
	inCheck := gen.IsKingInCheck(b, Black)
	t.Logf("Black king is in check: %v", inCheck)
	
	// Manual verification: Try to make the b3f7 move and see what happens
	t.Logf("\n=== MANUAL b3f7 VERIFICATION ===")
	manualB3F7 := board.Move{
		From:      board.Square{File: 1, Rank: 2}, // b3
		To:        board.Square{File: 5, Rank: 6}, // f7
		Piece:     board.WhiteQueen,
		Captured:  board.BlackPawn,
		IsCapture: true,
	}
	
	// Test if making this move leaves White king in check
	testBoard, err := board.FromFEN(b.ToFEN())
	if err != nil {
		t.Fatalf("Failed to create test board: %v", err)
	}
	
	err = testBoard.MakeMove(manualB3F7)
	if err != nil {
		t.Logf("Manual b3f7 move failed: %v", err)
	} else {
		t.Logf("Manual b3f7 move succeeded!")
		t.Logf("Position after b3f7: %s", testBoard.ToFEN())
		
		// Check if White king is in check after this move
		whiteInCheckAfter := gen.IsKingInCheck(testBoard, White)
		t.Logf("White king in check after b3f7: %v", whiteInCheckAfter)
		
		if whiteInCheckAfter {
			t.Logf("REASON: b3f7 is illegal because it leaves White king in check!")
		} else {
			t.Logf("PROBLEM: b3f7 should be legal but our move generator doesn't include it!")
		}
	}
	
	// This test reveals the exact nature of the desynchronization bug
}

