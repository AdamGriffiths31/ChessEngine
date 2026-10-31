package board

// GetCastlingRights returns the current castling rights as a string
func (b *Board) GetCastlingRights() string {
	return b.castlingRights
}

// GetEnPassantTarget returns the current en passant target square if any
func (b *Board) GetEnPassantTarget() (Square, bool) {
	return b.enPassantSquare, b.hasEnPassant
}

// GetHalfMoveClock returns the current half-move clock for the 50-move rule
func (b *Board) GetHalfMoveClock() int {
	return b.halfMoveClock
}

// GetFullMoveNumber returns the current full move number
func (b *Board) GetFullMoveNumber() int {
	return b.fullMoveNumber
}

// GetSideToMove returns which side is to move ("w" or "b")
func (b *Board) GetSideToMove() string {
	return b.sideToMove
}

// EnPassantCapturer returns the rank and pawn from which the given side ("w"
// or "b") could capture en passant.
func EnPassantCapturer(side string) (rank int, pawn Piece) {
	if side == "w" {
		return 4, WhitePawn
	}
	return 3, BlackPawn
}

// HasEnPassantCapturer reports whether an en passant target is set and the side
// to move has a pawn next to it. Hashes only include the target in that case.
func (b *Board) HasEnPassantCapturer() bool {
	if !b.hasEnPassant {
		return false
	}
	rank, pawn := EnPassantCapturer(b.sideToMove)
	file := b.enPassantSquare.File
	return (file > 0 && b.GetPiece(rank, file-1) == pawn) || (file < 7 && b.GetPiece(rank, file+1) == pawn)
}

// SetCastlingRights sets the castling rights string
func (b *Board) SetCastlingRights(rights string) {
	b.castlingRights = rights
}

// SetEnPassantTarget sets the en passant target square
func (b *Board) SetEnPassantTarget(target Square, hasEnPassant bool) {
	b.enPassantSquare = target
	b.hasEnPassant = hasEnPassant
}

// SetHalfMoveClock sets the half-move clock for the 50-move rule
func (b *Board) SetHalfMoveClock(clock int) {
	b.halfMoveClock = clock
}

// SetFullMoveNumber sets the full move number
func (b *Board) SetFullMoveNumber(num int) {
	b.fullMoveNumber = num
}

// SetSideToMove sets which side is to move
func (b *Board) SetSideToMove(side string) {
	b.sideToMove = side
}
