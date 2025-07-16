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
	pseudoLegalMoves := g.generateAllPseudoLegalMoves(b, player)
	legalMoves := NewMoveList()
	
	// Filter out moves that would leave the king in check
	for _, move := range pseudoLegalMoves.Moves {
		if g.isMoveLegal(b, move, player) {
			legalMoves.AddMove(move)
		}
	}
	
	return legalMoves
}

// MoveHistory stores information needed to undo a move
type MoveHistory struct {
	Move             board.Move
	CapturedPiece    board.Piece
	CastlingRights   string
	EnPassantTarget  *board.Square
	HalfMoveClock    int
	FullMoveNumber   int
	WasEnPassant     bool
	WasCastling      bool
}

// isMoveLegal checks if a move is legal (doesn't leave king in check)
func (g *Generator) isMoveLegal(b *board.Board, move board.Move, player Player) bool {
	// Early exit optimizations for obviously legal moves
	if g.canSkipKingSafetyCheck(b, move, player) {
		return true
	}
	
	// Make the move and get undo information
	history := g.makeMove(b, move)
	
	// Check if the king is in check after the move
	legal := !g.IsKingInCheck(b, player)
	
	// Undo the move
	g.unmakeMove(b, history)
	
	return legal
}

// canSkipKingSafetyCheck determines if a move obviously doesn't affect king safety
func (g *Generator) canSkipKingSafetyCheck(b *board.Board, move board.Move, player Player) bool {
	// Never skip king moves - they always affect king safety
	piece := b.GetPiece(move.From.Rank, move.From.File)
	if piece == board.WhiteKing || piece == board.BlackKing {
		return false
	}
	
	// Never skip if king is currently in check - must verify we're getting out of check
	if g.IsKingInCheck(b, player) {
		return false
	}
	
	// Find king position
	kingSquare := g.findKing(b, player)
	if kingSquare == nil {
		return false
	}
	
	// If the moving piece is not on the same rank, file, or diagonal as the king,
	// and it's not a discovered check situation, we can potentially skip
	if !g.isOnSameLineAsDiagonal(move.From, *kingSquare) {
		// Check if there's a potential discovered attack
		if !g.couldCauseDiscoveredCheck(b, move, player, *kingSquare) {
			return true
		}
	}
	
	return false
}

// isOnSameLineAsDiagonal checks if two squares are on the same rank, file, or diagonal
func (g *Generator) isOnSameLineAsDiagonal(sq1, sq2 board.Square) bool {
	// Same rank
	if sq1.Rank == sq2.Rank {
		return true
	}
	
	// Same file
	if sq1.File == sq2.File {
		return true
	}
	
	// Same diagonal
	rankDiff := abs(sq1.Rank - sq2.Rank)
	fileDiff := abs(sq1.File - sq2.File)
	if rankDiff == fileDiff {
		return true
	}
	
	return false
}

// couldCauseDiscoveredCheck checks if moving a piece could cause a discovered check
func (g *Generator) couldCauseDiscoveredCheck(b *board.Board, move board.Move, player Player, kingSquare board.Square) bool {
	// Check if there's a potential attacking piece behind the moving piece
	// This is a simplified check - we look for sliding pieces that could attack the king
	// if the moving piece is removed
	
	var enemyRook, enemyBishop, enemyQueen board.Piece
	if player == White {
		enemyRook = board.BlackRook
		enemyBishop = board.BlackBishop
		enemyQueen = board.BlackQueen
	} else {
		enemyRook = board.WhiteRook
		enemyBishop = board.WhiteBishop
		enemyQueen = board.WhiteQueen
	}
	
	// Check if the moving piece is between the king and an enemy sliding piece
	// For each direction from the king, see if removing the moving piece would expose the king
	for _, dir := range QueenDirections {
		if g.wouldExposeKingInDirection(b, move.From, kingSquare, dir, enemyRook, enemyBishop, enemyQueen) {
			return true
		}
	}
	
	return false
}

// wouldExposeKingInDirection checks if removing a piece would expose the king in a direction
func (g *Generator) wouldExposeKingInDirection(b *board.Board, movingPiece, kingSquare board.Square, dir Direction, enemyRook, enemyBishop, enemyQueen board.Piece) bool {
	// Start from the king and move in the direction
	currentRank := kingSquare.Rank + dir.RankDelta
	currentFile := kingSquare.File + dir.FileDelta
	
	foundMovingPiece := false
	
	for currentRank >= MinRank && currentRank <= MaxRank && currentFile >= MinFile && currentFile <= MaxFile {
		if currentRank == movingPiece.Rank && currentFile == movingPiece.File {
			foundMovingPiece = true
		} else {
			piece := b.GetPiece(currentRank, currentFile)
			if piece != board.Empty {
				// Found a piece - check if it's an attacking piece
				if foundMovingPiece {
					// Check if this piece can attack in this direction
					if g.isDirectionValidForPiece(dir, piece, enemyRook, enemyBishop, enemyQueen) {
						return true
					}
				}
				// Stop searching in this direction after finding any piece
				break
			}
		}
		
		currentRank += dir.RankDelta
		currentFile += dir.FileDelta
	}
	
	return false
}

// isDirectionValidForPiece checks if a piece can attack in a given direction
func (g *Generator) isDirectionValidForPiece(dir Direction, piece, enemyRook, enemyBishop, enemyQueen board.Piece) bool {
	// Queens can attack in any direction
	if piece == enemyQueen {
		return true
	}
	
	// Rooks can attack in straight lines
	if piece == enemyRook && (dir.RankDelta == 0 || dir.FileDelta == 0) {
		return true
	}
	
	// Bishops can attack diagonally
	if piece == enemyBishop && (dir.RankDelta != 0 && dir.FileDelta != 0) {
		return true
	}
	
	return false
}

// copyBoard creates a copy of the board for testing moves
func (g *Generator) copyBoard(b *board.Board) *board.Board {
	newBoard := &board.Board{}
	for rank := MinRank; rank < BoardSize; rank++ {
		for file := MinFile; file < BoardSize; file++ {
			newBoard.SetPiece(rank, file, b.GetPiece(rank, file))
		}
	}
	newBoard.SetCastlingRights(b.GetCastlingRights())
	newBoard.SetEnPassantTarget(b.GetEnPassantTarget())
	newBoard.SetHalfMoveClock(b.GetHalfMoveClock())
	newBoard.SetFullMoveNumber(b.GetFullMoveNumber())
	newBoard.SetSideToMove(b.GetSideToMove())
	return newBoard
}

// makeMove executes a move on the board and returns history for undoing
func (g *Generator) makeMove(b *board.Board, move board.Move) *MoveHistory {
	// Create history to enable undoing
	history := &MoveHistory{
		Move:            move,
		CapturedPiece:   board.Empty,
		CastlingRights:  b.GetCastlingRights(),
		EnPassantTarget: b.GetEnPassantTarget(),
		HalfMoveClock:   b.GetHalfMoveClock(),
		FullMoveNumber:  b.GetFullMoveNumber(),
		WasEnPassant:    move.IsEnPassant,
		WasCastling:     move.IsCastling,
	}
	
	// Handle en passant capture
	if move.IsEnPassant {
		// Remove the captured pawn
		captureRank := move.From.Rank
		history.CapturedPiece = b.GetPiece(captureRank, move.To.File)
		b.SetPiece(captureRank, move.To.File, board.Empty)
	} else if move.IsCapture {
		// Store captured piece for normal captures
		history.CapturedPiece = b.GetPiece(move.To.Rank, move.To.File)
	}
	
	// Handle castling
	if move.IsCastling {
		// Move the rook
		var rookFrom, rookTo board.Square
		if move.To.File == 6 { // Kingside
			rookFrom = board.Square{File: 7, Rank: move.From.Rank}
			rookTo = board.Square{File: 5, Rank: move.From.Rank}
		} else { // Queenside
			rookFrom = board.Square{File: 0, Rank: move.From.Rank}
			rookTo = board.Square{File: 3, Rank: move.From.Rank}
		}
		rook := b.GetPiece(rookFrom.Rank, rookFrom.File)
		b.SetPiece(rookFrom.Rank, rookFrom.File, board.Empty)
		b.SetPiece(rookTo.Rank, rookTo.File, rook)
	}
	
	// Move the piece
	piece := b.GetPiece(move.From.Rank, move.From.File)
	b.SetPiece(move.From.Rank, move.From.File, board.Empty)
	
	// Handle promotion
	if move.Promotion != board.Empty {
		b.SetPiece(move.To.Rank, move.To.File, move.Promotion)
	} else {
		b.SetPiece(move.To.Rank, move.To.File, piece)
	}
	
	// Update board state (castling rights, en passant, etc.)
	g.updateBoardState(b, move)
	
	return history
}

// unmakeMove undoes a move using the stored history
func (g *Generator) unmakeMove(b *board.Board, history *MoveHistory) {
	move := history.Move
	
	// Restore the piece to its original position
	piece := b.GetPiece(move.To.Rank, move.To.File)
	if move.Promotion != board.Empty {
		// For promotion, restore the original pawn
		var pawnPiece board.Piece
		if move.To.Rank == 7 { // White promotion
			pawnPiece = board.WhitePawn
		} else { // Black promotion
			pawnPiece = board.BlackPawn
		}
		b.SetPiece(move.From.Rank, move.From.File, pawnPiece)
	} else {
		b.SetPiece(move.From.Rank, move.From.File, piece)
	}
	
	// Restore the target square
	if history.WasEnPassant {
		// For en passant, restore the captured pawn to its original position
		b.SetPiece(move.To.Rank, move.To.File, board.Empty)
		captureRank := move.From.Rank
		b.SetPiece(captureRank, move.To.File, history.CapturedPiece)
	} else if history.CapturedPiece != board.Empty {
		// Restore captured piece
		b.SetPiece(move.To.Rank, move.To.File, history.CapturedPiece)
	} else {
		// Empty the target square
		b.SetPiece(move.To.Rank, move.To.File, board.Empty)
	}
	
	// Undo castling
	if history.WasCastling {
		// Restore the rook
		var rookFrom, rookTo board.Square
		if move.To.File == 6 { // Kingside
			rookFrom = board.Square{File: 7, Rank: move.From.Rank}
			rookTo = board.Square{File: 5, Rank: move.From.Rank}
		} else { // Queenside
			rookFrom = board.Square{File: 0, Rank: move.From.Rank}
			rookTo = board.Square{File: 3, Rank: move.From.Rank}
		}
		rook := b.GetPiece(rookTo.Rank, rookTo.File)
		b.SetPiece(rookTo.Rank, rookTo.File, board.Empty)
		b.SetPiece(rookFrom.Rank, rookFrom.File, rook)
	}
	
	// Restore board state
	b.SetCastlingRights(history.CastlingRights)
	b.SetEnPassantTarget(history.EnPassantTarget)
	b.SetHalfMoveClock(history.HalfMoveClock)
	b.SetFullMoveNumber(history.FullMoveNumber)
}

// updateBoardState updates castling rights, en passant, and move counters
func (g *Generator) updateBoardState(b *board.Board, move board.Move) {
	// Update castling rights based on the move
	castlingRights := b.GetCastlingRights()
	piece := b.GetPiece(move.To.Rank, move.To.File)
	
	// King moves remove all castling rights for that side
	if piece == board.WhiteKing {
		castlingRights = g.removeCastlingRights(castlingRights, "KQ")
	} else if piece == board.BlackKing {
		castlingRights = g.removeCastlingRights(castlingRights, "kq")
	}
	
	// Rook moves remove castling rights for that side
	if piece == board.WhiteRook {
		if move.From.File == 0 && move.From.Rank == 0 { // Queenside rook
			castlingRights = g.removeCastlingRights(castlingRights, "Q")
		} else if move.From.File == 7 && move.From.Rank == 0 { // Kingside rook
			castlingRights = g.removeCastlingRights(castlingRights, "K")
		}
	} else if piece == board.BlackRook {
		if move.From.File == 0 && move.From.Rank == 7 { // Queenside rook
			castlingRights = g.removeCastlingRights(castlingRights, "q")
		} else if move.From.File == 7 && move.From.Rank == 7 { // Kingside rook
			castlingRights = g.removeCastlingRights(castlingRights, "k")
		}
	}
	
	// Captured rook removes castling rights
	if move.IsCapture {
		if move.To.File == 0 && move.To.Rank == 0 { // White queenside rook captured
			castlingRights = g.removeCastlingRights(castlingRights, "Q")
		} else if move.To.File == 7 && move.To.Rank == 0 { // White kingside rook captured
			castlingRights = g.removeCastlingRights(castlingRights, "K")
		} else if move.To.File == 0 && move.To.Rank == 7 { // Black queenside rook captured
			castlingRights = g.removeCastlingRights(castlingRights, "q")
		} else if move.To.File == 7 && move.To.Rank == 7 { // Black kingside rook captured
			castlingRights = g.removeCastlingRights(castlingRights, "k")
		}
	}
	
	b.SetCastlingRights(castlingRights)
	
	// Set en passant target for pawn two-square moves
	if piece == board.WhitePawn || piece == board.BlackPawn {
		if abs(move.To.Rank - move.From.Rank) == 2 {
			// Two-square pawn move - set en passant target
			targetRank := (move.From.Rank + move.To.Rank) / 2
			enPassantTarget := &board.Square{File: move.From.File, Rank: targetRank}
			b.SetEnPassantTarget(enPassantTarget)
		} else {
			b.SetEnPassantTarget(nil)
		}
	} else {
		b.SetEnPassantTarget(nil)
	}
	
	// Update move counters
	halfMoveClock := b.GetHalfMoveClock()
	if move.IsCapture || piece == board.WhitePawn || piece == board.BlackPawn {
		halfMoveClock = 0
	} else {
		halfMoveClock++
	}
	b.SetHalfMoveClock(halfMoveClock)
	
	// Update full move number (increments after black's move)
	if b.GetSideToMove() == "b" {
		b.SetFullMoveNumber(b.GetFullMoveNumber() + 1)
	}
}

// removeCastlingRights removes specific castling rights from the string
func (g *Generator) removeCastlingRights(rights, toRemove string) string {
	result := ""
	for _, r := range rights {
		remove := false
		for _, remove_r := range toRemove {
			if r == remove_r {
				remove = true
				break
			}
		}
		if !remove {
			result += string(r)
		}
	}
	if result == "" {
		return "-"
	}
	return result
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}


// addPromotionMoves adds all four promotion moves (Q, R, B, N)
func (g *Generator) addPromotionMoves(from, to board.Square, player Player, moveList *MoveList) {
	var promotionPieces []board.Piece
	var pawnPiece board.Piece
	
	if player == White {
		pawnPiece = board.WhitePawn
		promotionPieces = []board.Piece{
			board.WhiteQueen, board.WhiteRook, board.WhiteBishop, board.WhiteKnight,
		}
	} else {
		pawnPiece = board.BlackPawn
		promotionPieces = []board.Piece{
			board.BlackQueen, board.BlackRook, board.BlackBishop, board.BlackKnight,
		}
	}

	for _, piece := range promotionPieces {
		move := board.Move{
			From:      from,
			To:        to,
			Piece:     pawnPiece,
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

// Direction represents a movement direction with rank and file deltas
type Direction struct {
	RankDelta int
	FileDelta int
}

// Board constants
const (
	MinRank = 0
	MaxRank = 7
	MinFile = 0
	MaxFile = 7
	BoardSize = 8
)

// Direction constants for pieces
var (
	// Rook directions (straight lines)
	RookDirections = []Direction{
		{1, 0},  // Up
		{-1, 0}, // Down
		{0, 1},  // Right
		{0, -1}, // Left
	}
	
	// Bishop directions (diagonals)
	BishopDirections = []Direction{
		{1, 1},   // Up-right
		{1, -1},  // Up-left
		{-1, 1},  // Down-right
		{-1, -1}, // Down-left
	}
	
	// Queen directions (combination of rook and bishop)
	QueenDirections = []Direction{
		{1, 0}, {-1, 0}, {0, 1}, {0, -1},     // Rook moves
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1},   // Bishop moves
	}
	
	// Knight directions (L-shaped moves)
	KnightDirections = []Direction{
		{2, 1},   // Up 2, Right 1
		{2, -1},  // Up 2, Left 1
		{-2, 1},  // Down 2, Right 1
		{-2, -1}, // Down 2, Left 1
		{1, 2},   // Up 1, Right 2
		{1, -2},  // Up 1, Left 2
		{-1, 2},  // Down 1, Right 2
		{-1, -2}, // Down 1, Left 2
	}
)

// generateSlidingPieceMoves generates moves for sliding pieces (rook, bishop, queen)
func (g *Generator) generateSlidingPieceMoves(b *board.Board, player Player, pieceType board.Piece, directions []Direction) *MoveList {
	moveList := NewMoveList()
	
	// Scan the board for pieces of the specified type
	for rank := MinRank; rank < BoardSize; rank++ {
		for file := MinFile; file < BoardSize; file++ {
			piece := b.GetPiece(rank, file)
			if piece == pieceType {
				// Generate moves for this piece
				fromSquare := board.Square{File: file, Rank: rank}
				for _, dir := range directions {
					g.generateSlidingMoves(b, player, fromSquare, dir.RankDelta, dir.FileDelta, moveList)
				}
			}
		}
	}
	
	return moveList
}

// generateJumpingPieceMoves generates moves for jumping pieces (knight, king)
func (g *Generator) generateJumpingPieceMoves(b *board.Board, player Player, pieceType board.Piece, directions []Direction) *MoveList {
	moveList := NewMoveList()
	
	// Scan the board for pieces of the specified type
	for rank := MinRank; rank < BoardSize; rank++ {
		for file := MinFile; file < BoardSize; file++ {
			piece := b.GetPiece(rank, file)
			if piece == pieceType {
				// Generate moves for this piece
				fromSquare := board.Square{File: file, Rank: rank}
				for _, dir := range directions {
					g.generateJumpingMove(b, player, fromSquare, dir.RankDelta, dir.FileDelta, moveList)
				}
			}
		}
	}
	
	return moveList
}

// generateJumpingMove generates a single jumping move if valid
func (g *Generator) generateJumpingMove(b *board.Board, player Player, from board.Square, rankDelta, fileDelta int, moveList *MoveList) {
	newRank := from.Rank + rankDelta
	newFile := from.File + fileDelta
	
	// Check if the target square is within board boundaries
	if newRank >= MinRank && newRank <= MaxRank && newFile >= MinFile && newFile <= MaxFile {
		piece := b.GetPiece(newRank, newFile)
		to := board.Square{File: newFile, Rank: newRank}
		
		if piece == board.Empty {
			// Empty square - valid move
			move := g.createMove(from, to, false, board.Empty, board.Empty)
			moveList.AddMove(move)
		} else if g.isEnemyPiece(piece, player) {
			// Enemy piece - valid capture
			move := g.createMove(from, to, true, piece, board.Empty)
			moveList.AddMove(move)
		}
		// Own piece - can't move here
	}
}

// createMove creates a move with the given parameters
func (g *Generator) createMove(from, to board.Square, isCapture bool, captured, promotion board.Piece) board.Move {
	return board.Move{
		From:      from,
		To:        to,
		IsCapture: isCapture,
		Captured:  captured,
		Promotion: promotion,
	}
}

// generateSlidingMoves generates moves in a straight line until blocked or edge reached
func (g *Generator) generateSlidingMoves(b *board.Board, player Player, from board.Square, rankDelta, fileDelta int, moveList *MoveList) {
	currentRank := from.Rank + rankDelta
	currentFile := from.File + fileDelta
	
	// Continue sliding in the direction until we hit the board edge or a piece
	for currentRank >= MinRank && currentRank <= MaxRank && currentFile >= MinFile && currentFile <= MaxFile {
		piece := b.GetPiece(currentRank, currentFile)
		to := board.Square{File: currentFile, Rank: currentRank}
		
		if piece == board.Empty {
			// Empty square - valid move
			move := g.createMove(from, to, false, board.Empty, board.Empty)
			moveList.AddMove(move)
		} else if g.isEnemyPiece(piece, player) {
			// Enemy piece - valid capture, but can't continue sliding
			move := g.createMove(from, to, true, piece, board.Empty)
			moveList.AddMove(move)
			break // Stop sliding in this direction
		} else {
			// Own piece - can't move here and can't continue sliding
			break
		}
		
		// Move to next square in this direction
		currentRank += rankDelta
		currentFile += fileDelta
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

// IsKingInCheck checks if the king of the given player is in check
func (g *Generator) IsKingInCheck(b *board.Board, player Player) bool {
	return g.IsKingInCheckFast(b, player)
}

// IsKingInCheckFast optimized version that avoids repeated king searches
func (g *Generator) IsKingInCheckFast(b *board.Board, player Player) bool {
	kingSquare := g.findKing(b, player)
	if kingSquare == nil {
		return false // No king found
	}
	
	return g.isSquareAttackedByEnemy(b, *kingSquare, player)
}

// isSquareAttackedByEnemy checks if a square is attacked by the enemy player
func (g *Generator) isSquareAttackedByEnemy(b *board.Board, square board.Square, player Player) bool {
	enemyPlayer := Black
	if player == Black {
		enemyPlayer = White
	}
	
	// Check for enemy pawn attacks
	if g.isSquareAttackedByPawns(b, square, enemyPlayer) {
		return true
	}
	
	// Check for enemy knight attacks
	if g.isSquareAttackedByKnights(b, square, enemyPlayer) {
		return true
	}
	
	// Check for enemy sliding piece attacks (rooks, bishops, queens)
	if g.isSquareAttackedBySlidingPieces(b, square, enemyPlayer) {
		return true
	}
	
	// Check for enemy king attacks
	if g.isSquareAttackedByKing(b, square, enemyPlayer) {
		return true
	}
	
	return false
}

// isSquareAttackedByPawns checks if pawns attack the square
func (g *Generator) isSquareAttackedByPawns(b *board.Board, square board.Square, enemyPlayer Player) bool {
	var pawnPiece board.Piece
	var pawnDirection int
	
	if enemyPlayer == White {
		pawnPiece = board.WhitePawn
		pawnDirection = 1 // White pawns attack upward
	} else {
		pawnPiece = board.BlackPawn
		pawnDirection = -1 // Black pawns attack downward
	}
	
	// Check diagonally backwards (where pawns would attack from)
	pawnRank := square.Rank - pawnDirection
	if pawnRank >= MinRank && pawnRank <= MaxRank {
		// Check left diagonal
		if square.File > MinFile && b.GetPiece(pawnRank, square.File-1) == pawnPiece {
			return true
		}
		// Check right diagonal
		if square.File < MaxFile && b.GetPiece(pawnRank, square.File+1) == pawnPiece {
			return true
		}
	}
	
	return false
}

// isSquareAttackedByKnights checks if knights attack the square
func (g *Generator) isSquareAttackedByKnights(b *board.Board, square board.Square, enemyPlayer Player) bool {
	var knightPiece board.Piece
	if enemyPlayer == White {
		knightPiece = board.WhiteKnight
	} else {
		knightPiece = board.BlackKnight
	}
	
	// Check all knight move patterns
	for _, dir := range KnightDirections {
		knightRank := square.Rank - dir.RankDelta
		knightFile := square.File - dir.FileDelta
		
		if knightRank >= MinRank && knightRank <= MaxRank && knightFile >= MinFile && knightFile <= MaxFile {
			if b.GetPiece(knightRank, knightFile) == knightPiece {
				return true
			}
		}
	}
	
	return false
}

// isSquareAttackedBySlidingPieces checks if sliding pieces attack the square
func (g *Generator) isSquareAttackedBySlidingPieces(b *board.Board, square board.Square, enemyPlayer Player) bool {
	var rookPiece, bishopPiece, queenPiece board.Piece
	
	if enemyPlayer == White {
		rookPiece = board.WhiteRook
		bishopPiece = board.WhiteBishop
		queenPiece = board.WhiteQueen
	} else {
		rookPiece = board.BlackRook
		bishopPiece = board.BlackBishop
		queenPiece = board.BlackQueen
	}
	
	// Check rook/queen attacks (straight lines)
	for _, dir := range RookDirections {
		if g.isSquareAttackedFromDirection(b, square, dir, rookPiece, queenPiece) {
			return true
		}
	}
	
	// Check bishop/queen attacks (diagonals)
	for _, dir := range BishopDirections {
		if g.isSquareAttackedFromDirection(b, square, dir, bishopPiece, queenPiece) {
			return true
		}
	}
	
	return false
}

// isSquareAttackedFromDirection checks if a square is attacked from a specific direction
func (g *Generator) isSquareAttackedFromDirection(b *board.Board, square board.Square, dir Direction, piece1, piece2 board.Piece) bool {
	currentRank := square.Rank + dir.RankDelta
	currentFile := square.File + dir.FileDelta
	
	// Slide in the direction until we hit a piece or board edge
	for currentRank >= MinRank && currentRank <= MaxRank && currentFile >= MinFile && currentFile <= MaxFile {
		pieceAtSquare := b.GetPiece(currentRank, currentFile)
		
		if pieceAtSquare != board.Empty {
			// Found a piece - check if it's an attacking piece
			return pieceAtSquare == piece1 || pieceAtSquare == piece2
		}
		
		currentRank += dir.RankDelta
		currentFile += dir.FileDelta
	}
	
	return false
}

// isSquareAttackedByKing checks if the enemy king attacks the square
func (g *Generator) isSquareAttackedByKing(b *board.Board, square board.Square, enemyPlayer Player) bool {
	var kingPiece board.Piece
	if enemyPlayer == White {
		kingPiece = board.WhiteKing
	} else {
		kingPiece = board.BlackKing
	}
	
	// Check all 8 directions around the square
	for _, dir := range QueenDirections { // King moves in all 8 directions like queen but only 1 square
		kingRank := square.Rank - dir.RankDelta
		kingFile := square.File - dir.FileDelta
		
		if kingRank >= MinRank && kingRank <= MaxRank && kingFile >= MinFile && kingFile <= MaxFile {
			if b.GetPiece(kingRank, kingFile) == kingPiece {
				return true
			}
		}
	}
	
	return false
}

// findKing finds the king's position for the given player
func (g *Generator) findKing(b *board.Board, player Player) *board.Square {
	var kingPiece board.Piece
	if player == White {
		kingPiece = board.WhiteKing
	} else {
		kingPiece = board.BlackKing
	}
	
	for rank := MinRank; rank < BoardSize; rank++ {
		for file := MinFile; file < BoardSize; file++ {
			if b.GetPiece(rank, file) == kingPiece {
				return &board.Square{File: file, Rank: rank}
			}
		}
	}
	return nil
}

// generateAllPseudoLegalMoves generates all pseudo-legal moves (without check validation)
func (g *Generator) generateAllPseudoLegalMoves(b *board.Board, player Player) *MoveList {
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

// IsSquareAttacked checks if a square is attacked by the enemy (public method)
func (g *Generator) IsSquareAttacked(b *board.Board, square board.Square, player Player) bool {
	enemyPlayer := Black
	if player == Black {
		enemyPlayer = White
	}
	
	// Generate all enemy moves (without castling to avoid recursion)
	enemyMoves := g.generateMovesWithoutCastling(b, enemyPlayer)
	
	for _, move := range enemyMoves.Moves {
		if move.To.File == square.File && move.To.Rank == square.Rank {
			return true
		}
	}
	
	return false
}

