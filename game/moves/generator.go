package moves

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Player represents a chess player
type Player int

const (
	White Player = iota
	Black
)

// String returns string representation of player
func (p Player) String() string {
	if p == White {
		return "White"
	}
	return "Black"
}

// MoveGenerator interface for generating legal moves
type MoveGenerator interface {
	GenerateAllMoves(b *board.Board, player Player) *MoveList
	GeneratePawnMoves(b *board.Board, player Player) *MoveList
	GenerateRookMoves(b *board.Board, player Player) *MoveList
	GenerateBishopMoves(b *board.Board, player Player) *MoveList
	GenerateKnightMoves(b *board.Board, player Player) *MoveList
	GenerateQueenMoves(b *board.Board, player Player) *MoveList
	GenerateKingMoves(b *board.Board, player Player) *MoveList
}

// MoveList represents a collection of moves
type MoveList struct {
	Moves []board.Move
	Count int
}

// NewMoveList creates a new empty move list
func NewMoveList() *MoveList {
	return &MoveList{
		Moves: make([]board.Move, 0, 64), // Pre-allocate for performance
		Count: 0,
	}
}

// AddMove adds a move to the list
func (ml *MoveList) AddMove(move board.Move) {
	ml.Moves = append(ml.Moves, move)
	ml.Count++
}

// Contains checks if the move list contains a specific move
func (ml *MoveList) Contains(move board.Move) bool {
	for _, m := range ml.Moves {
		if MovesEqual(m, move) {
			return true
		}
	}
	return false
}

// Clear empties the move list
func (ml *MoveList) Clear() {
	ml.Moves = ml.Moves[:0]
	ml.Count = 0
}

// Generator implements the MoveGenerator interface
type Generator struct{}

// NewGenerator creates a new move generator
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateAllMoves generates all legal moves for the given player
// Includes all piece moves: pawn (3a), rook (3b), bishop (3c), knight (3d), queen (3e), and king (3f)
func (g *Generator) GenerateAllMoves(b *board.Board, player Player) *MoveList {
	moveList := NewMoveList()

	pawnMoves := g.GeneratePawnMoves(b, player)
	for _, move := range pawnMoves.Moves {
		moveList.AddMove(move)
	}

	rookMoves := g.GenerateRookMoves(b, player)
	for _, move := range rookMoves.Moves {
		moveList.AddMove(move)
	}

	bishopMoves := g.GenerateBishopMoves(b, player)
	for _, move := range bishopMoves.Moves {
		moveList.AddMove(move)
	}

	knightMoves := g.GenerateKnightMoves(b, player)
	for _, move := range knightMoves.Moves {
		moveList.AddMove(move)
	}

	queenMoves := g.GenerateQueenMoves(b, player)
	for _, move := range queenMoves.Moves {
		moveList.AddMove(move)
	}

	kingMoves := g.GenerateKingMoves(b, player)
	for _, move := range kingMoves.Moves {
		moveList.AddMove(move)
	}

	return moveList
}

// GeneratePawnMoves generates all legal pawn moves for the given player
func (g *Generator) GeneratePawnMoves(b *board.Board, player Player) *MoveList {
	moveList := NewMoveList()

	// Determine pawn piece and direction based on player
	var pawnPiece board.Piece
	var direction int
	var startRank, promotionRank int

	if player == White {
		pawnPiece = board.WhitePawn
		direction = 1     // White pawns move up (increasing rank)
		startRank = 1     // 2nd rank (index 1)
		promotionRank = 7 // 8th rank (index 7)
	} else {
		pawnPiece = board.BlackPawn
		direction = -1    // Black pawns move down (decreasing rank)
		startRank = 6     // 7th rank (index 6)
		promotionRank = 0 // 1st rank (index 0)
	}

	// Scan the board for pawns of the current player
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece == pawnPiece {
				// Generate moves for this pawn
				g.generatePawnMovesFromSquare(b, player, rank, file, direction, startRank, promotionRank, moveList)
			}
		}
	}

	return moveList
}

// generatePawnMovesFromSquare generates all moves for a pawn at a specific square
func (g *Generator) generatePawnMovesFromSquare(b *board.Board, player Player, rank, file, direction, startRank, promotionRank int, moveList *MoveList) {
	fromSquare := board.Square{File: file, Rank: rank}

	// Forward moves
	g.generatePawnForwardMoves(b, fromSquare, direction, startRank, promotionRank, moveList)

	// Capture moves
	g.generatePawnCaptures(b, player, fromSquare, direction, promotionRank, moveList)

	// En passant moves (will be implemented later when we add en passant tracking)
	// g.generateEnPassantMoves(b, player, fromSquare, direction, moveList)
}

// generatePawnForwardMoves generates forward pawn moves (1 and 2 squares)
func (g *Generator) generatePawnForwardMoves(b *board.Board, from board.Square, direction, startRank, promotionRank int, moveList *MoveList) {
	// One square forward
	newRank := from.Rank + direction
	if newRank >= 0 && newRank <= 7 {
		if b.GetPiece(newRank, from.File) == board.Empty {
			to := board.Square{File: from.File, Rank: newRank}

			// Check for promotion
			if newRank == promotionRank {
				g.addPromotionMoves(from, to, moveList)
			} else {
				move := board.Move{
					From:      from,
					To:        to,
					Promotion: board.Empty,
				}
				moveList.AddMove(move)
			}

			// Two squares forward (only from starting position)
			if from.Rank == startRank {
				newRank2 := newRank + direction
				if newRank2 >= 0 && newRank2 <= 7 && b.GetPiece(newRank2, from.File) == board.Empty {
					to2 := board.Square{File: from.File, Rank: newRank2}
					move := board.Move{
						From:      from,
						To:        to2,
						Promotion: board.Empty,
					}
					moveList.AddMove(move)
				}
			}
		}
	}
}

// generatePawnCaptures generates diagonal capture moves
func (g *Generator) generatePawnCaptures(b *board.Board, player Player, from board.Square, direction, promotionRank int, moveList *MoveList) {
	newRank := from.Rank + direction
	if newRank < 0 || newRank > 7 {
		return
	}

	// Left capture
	if from.File > 0 {
		piece := b.GetPiece(newRank, from.File-1)
		if piece != board.Empty && g.isEnemyPiece(piece, player) {
			to := board.Square{File: from.File - 1, Rank: newRank}

			if newRank == promotionRank {
				g.addPromotionMoves(from, to, moveList)
			} else {
				move := board.Move{
					From:      from,
					To:        to,
					IsCapture: true,
					Captured:  piece,
					Promotion: board.Empty,
				}
				moveList.AddMove(move)
			}
		}
	}

	// Right capture
	if from.File < 7 {
		piece := b.GetPiece(newRank, from.File+1)
		if piece != board.Empty && g.isEnemyPiece(piece, player) {
			to := board.Square{File: from.File + 1, Rank: newRank}

			if newRank == promotionRank {
				g.addPromotionMoves(from, to, moveList)
			} else {
				move := board.Move{
					From:      from,
					To:        to,
					IsCapture: true,
					Captured:  piece,
					Promotion: board.Empty,
				}
				moveList.AddMove(move)
			}
		}
	}
}

// addPromotionMoves adds all four promotion moves (Q, R, B, N)
func (g *Generator) addPromotionMoves(from, to board.Square, moveList *MoveList) {
	promotionPieces := []board.Piece{
		board.WhiteQueen, board.WhiteRook, board.WhiteBishop, board.WhiteKnight,
	}

	for _, piece := range promotionPieces {
		move := board.Move{
			From:      from,
			To:        to,
			Promotion: piece,
		}
		moveList.AddMove(move)
	}
}

// isEnemyPiece checks if a piece belongs to the enemy
func (g *Generator) isEnemyPiece(piece board.Piece, player Player) bool {
	if player == White {
		// White player - enemy pieces are lowercase (black)
		return piece >= 'a' && piece <= 'z'
	} else {
		// Black player - enemy pieces are uppercase (white)
		return piece >= 'A' && piece <= 'Z'
	}
}

// MovesEqual compares two moves for equality
func MovesEqual(a, b board.Move) bool {
	return a.From.File == b.From.File &&
		a.From.Rank == b.From.Rank &&
		a.To.File == b.To.File &&
		a.To.Rank == b.To.Rank &&
		a.Promotion == b.Promotion
}

