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
