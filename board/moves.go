package board

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Move struct {
	From        Square
	To          Square
	Piece       Piece
	Captured    Piece
	Promotion   Piece
	IsCapture   bool
	IsCastling  bool
	IsEnPassant bool
}

// MoveUndo stores the state needed to undo a move
type MoveUndo struct {
	Move            Move
	CapturedPiece   Piece
	CastlingRights  string
	EnPassantTarget *Square
	HalfMoveClock   int
	FullMoveNumber  int
	SideToMove      string
}

// NullMove represents a "pass turn" move used for null move pruning
var NullMove = Move{
	From:        Square{File: -1, Rank: -1},
	To:          Square{File: -1, Rank: -1},
	Piece:       Empty,
	Captured:    Empty,
	Promotion:   Empty,
	IsCapture:   false,
	IsCastling:  false,
	IsEnPassant: false,
}

func ParseSquare(notation string) (Square, error) {
	if len(notation) != 2 {
		return Square{}, errors.New("invalid square notation: must be 2 characters")
	}

	file := int(notation[0] - 'a')
	rank := int(notation[1] - '1')

	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return Square{}, errors.New("invalid square notation: out of bounds")
	}

	return Square{File: file, Rank: rank}, nil
}

func (s Square) String() string {
	return string(rune('a'+s.File)) + string(rune('1'+s.Rank))
}

// MakeMoveWithUndo makes a move and returns undo information
func (b *Board) MakeMoveWithUndo(move Move) (MoveUndo, error) {
	// Ensure move.Piece is set for proper undo
	if move.Piece == Empty || move.Piece == 0 {
		move.Piece = b.GetPiece(move.From.Rank, move.From.File)
	}

	// For en passant moves, we need to capture the piece from the correct square
	var capturedPiece Piece
	if move.IsEnPassant {
		// The captured pawn is not on the destination square
		var captureRank int
		if move.Piece == WhitePawn {
			captureRank = 4 // Black pawn on rank 5
		} else {
			captureRank = 3 // White pawn on rank 4
		}
		capturedPiece = b.GetPiece(captureRank, move.To.File)
	} else {
		capturedPiece = b.GetPiece(move.To.Rank, move.To.File)
	}

	// Store current state for undo
	undo := MoveUndo{
		Move:            move, // This now has the correct piece
		CapturedPiece:   capturedPiece,
		CastlingRights:  b.castlingRights,
		EnPassantTarget: b.enPassantTarget,
		HalfMoveClock:   b.halfMoveClock,
		FullMoveNumber:  b.fullMoveNumber,
		SideToMove:      b.sideToMove,
	}

	// Push current hash to history for undo
	b.PushHash()

	// Capture old state for incremental hash update
	oldState := b.GetCurrentBoardState()

	// Make the move
	err := b.MakeMove(move)

	// Update hash incrementally if hash updater is available
	if b.hashUpdater != nil && err == nil {
		hashDelta := b.hashUpdater.GetHashDelta(b, move, oldState)
		b.UpdateHash(hashDelta)
	}

	return undo, err
}

// MakeNullMove makes a null move (passes the turn) and returns undo information
func (b *Board) MakeNullMove() MoveUndo {
	// Store current state for undo
	undo := MoveUndo{
		Move:            NullMove,
		CapturedPiece:   Empty,
		CastlingRights:  b.castlingRights,
		EnPassantTarget: b.enPassantTarget,
		HalfMoveClock:   b.halfMoveClock,
		FullMoveNumber:  b.fullMoveNumber,
		SideToMove:      b.sideToMove,
	}

	// Switch side to move
	if b.sideToMove == "w" {
		b.sideToMove = "b"
	} else {
		b.sideToMove = "w"
		b.fullMoveNumber++
	}

	// Increment half move clock (null moves don't reset 50-move rule)
	b.halfMoveClock++

	// Clear en passant target (opportunity expires when turn passes)
	b.enPassantTarget = nil

	// Castling rights remain unchanged (no pieces moved)

	return undo
}

// UnmakeNullMove reverses a null move using the stored undo information
func (b *Board) UnmakeNullMove(undo MoveUndo) {
	// Restore all board state (no pieces to move since it was a null move)
	b.castlingRights = undo.CastlingRights
	b.enPassantTarget = undo.EnPassantTarget
	b.halfMoveClock = undo.HalfMoveClock
	b.fullMoveNumber = undo.FullMoveNumber
	b.sideToMove = undo.SideToMove
}

func (b *Board) MakeMove(move Move) error {
	// Use the piece from the Move struct if available, otherwise get from board
	var piece Piece
	if move.Piece != Empty {
		piece = move.Piece
	} else {
		piece = b.GetPiece(move.From.Rank, move.From.File)
		if piece == Empty {
			return errors.New("no piece at from square")
		}
	}

	// Auto-detect en passant if not already set
	if !move.IsEnPassant && (piece == WhitePawn || piece == BlackPawn) {
		targetPiece := b.GetPiece(move.To.Rank, move.To.File)
		// If pawn moves diagonally to empty square and en passant target matches
		if targetPiece == Empty && move.From.File != move.To.File {
			if b.enPassantTarget != nil && b.enPassantTarget.File == move.To.File && b.enPassantTarget.Rank == move.To.Rank {
				move.IsEnPassant = true
			}
		}
	}

	// Store captured piece before making move
	capturedPiece := b.GetPiece(move.To.Rank, move.To.File)

	// Clear the from square
	b.SetPiece(move.From.Rank, move.From.File, Empty)

	// Handle promotion
	if move.Promotion != Empty && move.Promotion != 0 {
		b.SetPiece(move.To.Rank, move.To.File, move.Promotion)
	} else {
		// Place the piece at the to square
		b.SetPiece(move.To.Rank, move.To.File, piece)
	}

	// Handle castling
	if move.IsCastling {
		err := b.handleCastling(move)
		if err != nil {
			return err
		}
	}

	// Handle en passant
	if move.IsEnPassant {
		err := b.handleEnPassant(move)
		if err != nil {
			return err
		}
	}

	// Update game state
	b.updateGameState(move, piece, capturedPiece)

	return nil
}

func (b *Board) handleCastling(move Move) error {
	// Determine which rook to move based on the king's destination
	var rookFrom, rookTo Square

	switch move.To.File {
	case 6: // King-side castling (O-O)
		rookFrom = Square{File: 7, Rank: move.From.Rank}
		rookTo = Square{File: 5, Rank: move.From.Rank}
	case 2: // Queen-side castling (O-O-O)
		rookFrom = Square{File: 0, Rank: move.From.Rank}
		rookTo = Square{File: 3, Rank: move.From.Rank}
	default:
		return errors.New("invalid castling move")
	}

	// Move the rook
	rook := b.GetPiece(rookFrom.Rank, rookFrom.File)
	b.SetPiece(rookFrom.Rank, rookFrom.File, Empty)
	b.SetPiece(rookTo.Rank, rookTo.File, rook)

	return nil
}

func (b *Board) handleEnPassant(move Move) error {
	// Remove the captured pawn
	var captureRank int
	if move.Piece == WhitePawn {
		captureRank = 4
	} else {
		captureRank = 3
	}

	b.SetPiece(captureRank, move.To.File, Empty)
	return nil
}

func (b *Board) updateGameState(move Move, piece Piece, capturedPiece Piece) {
	// Switch side to move
	if b.sideToMove == "w" {
		b.sideToMove = "b"
	} else {
		b.sideToMove = "w"
		b.fullMoveNumber++
	}

	// Update half move clock - reset if pawn move, capture, or en passant
	if piece == WhitePawn || piece == BlackPawn || capturedPiece != Empty || move.IsEnPassant {
		b.halfMoveClock = 0
	} else {
		b.halfMoveClock++
	}

	// Update castling rights if king or rook moved
	b.updateCastlingRights(move, piece)

	// Update en passant target
	b.updateEnPassantTarget(move, piece)
}

func (b *Board) updateCastlingRights(move Move, piece Piece) {
	// If king moved, remove all castling rights for that side
	switch piece {
	case WhiteKing:
		b.castlingRights = strings.ReplaceAll(b.castlingRights, "K", "")
		b.castlingRights = strings.ReplaceAll(b.castlingRights, "Q", "")
	case BlackKing:
		b.castlingRights = strings.ReplaceAll(b.castlingRights, "k", "")
		b.castlingRights = strings.ReplaceAll(b.castlingRights, "q", "")
	}

	// If rook moved from initial position, remove corresponding castling right
	switch piece {
	case WhiteRook:
		if move.From.Rank == 0 && move.From.File == 0 {
			b.castlingRights = strings.ReplaceAll(b.castlingRights, "Q", "") // queenside
		} else if move.From.Rank == 0 && move.From.File == 7 {
			b.castlingRights = strings.ReplaceAll(b.castlingRights, "K", "") // kingside
		}
	case BlackRook:
		if move.From.Rank == 7 && move.From.File == 0 {
			b.castlingRights = strings.ReplaceAll(b.castlingRights, "q", "") // queenside
		} else if move.From.Rank == 7 && move.From.File == 7 {
			b.castlingRights = strings.ReplaceAll(b.castlingRights, "k", "") // kingside
		}
	}

	// If no castling rights remain, set to "-"
	if b.castlingRights == "" {
		b.castlingRights = "-"
	}
}

func (b *Board) updateEnPassantTarget(move Move, piece Piece) {
	// Clear previous en passant target
	b.enPassantTarget = nil

	// Set en passant target if pawn moved two squares
	if piece == WhitePawn && move.From.Rank == 1 && move.To.Rank == 3 {
		square := Square{File: move.From.File, Rank: 2}
		b.enPassantTarget = &square
	} else if piece == BlackPawn && move.From.Rank == 6 && move.To.Rank == 4 {
		square := Square{File: move.From.File, Rank: 5}
		b.enPassantTarget = &square
	}
}

func ParseSimpleMove(notation string) (Move, error) {
	notation = strings.TrimSpace(notation)

	// Handle castling
	if notation == "O-O" || notation == "0-0" {
		return Move{IsCastling: true, Promotion: Empty}, nil
	}
	if notation == "O-O-O" || notation == "0-0-0" {
		return Move{IsCastling: true, Promotion: Empty}, nil
	}

	// Handle simple moves like "e2e4"
	if len(notation) == 4 {
		from, err := ParseSquare(notation[:2])
		if err != nil {
			return Move{}, err
		}

		to, err := ParseSquare(notation[2:4])
		if err != nil {
			return Move{}, err
		}

		return Move{From: from, To: to, Promotion: Empty, Piece: Empty}, nil
	}

	// Handle promotion moves like "e7e8Q"
	if len(notation) == 5 {
		from, err := ParseSquare(notation[:2])
		if err != nil {
			return Move{}, err
		}

		to, err := ParseSquare(notation[2:4])
		if err != nil {
			return Move{}, err
		}

		promotionChar := notation[4]
		promotion, err := charToPiece(promotionChar)
		if err != nil {
			return Move{}, err
		}

		return Move{From: from, To: to, Promotion: promotion}, nil
	}

	return Move{}, errors.New("unsupported move notation format")
}

func charToPiece(char byte) (Piece, error) {
	switch char {
	case 'Q':
		return WhiteQueen, nil
	case 'R':
		return WhiteRook, nil
	case 'B':
		return WhiteBishop, nil
	case 'N':
		return WhiteKnight, nil
	case 'q':
		return BlackQueen, nil
	case 'r':
		return BlackRook, nil
	case 'b':
		return BlackBishop, nil
	case 'n':
		return BlackKnight, nil
	default:
		return Empty, errors.New("invalid piece character")
	}
}

// UnmakeMove reverses a move using the undo information
func (b *Board) UnmakeMove(undo MoveUndo) {
	move := undo.Move
	
	// Determine what piece to restore based on move type
	var movedPiece Piece
	if move.Promotion != Empty && move.Promotion != 0 {
		// For promotion moves, restore the original pawn
		if move.To.Rank == 7 { // White promotion (to rank 8)
			movedPiece = WhitePawn
		} else if move.To.Rank == 0 { // Black promotion (to rank 1)
			movedPiece = BlackPawn
		} else {
			movedPiece = move.Piece
		}
	} else {
		movedPiece = move.Piece
	}

	// Validate that we have a valid piece to move back
	if movedPiece == Empty {
		fmt.Printf("WARNING: UnmakeMove could not determine piece to restore for move %s-%s\n",
			move.From.String(), move.To.String())
		return
	}

	// Handle special moves first (order matters)

	// Undo castling - must be done before moving the king back
	if move.IsCastling {
		b.undoCastling(move)
	}

	// Undo en passant - restore the captured pawn
	if move.IsEnPassant {
		b.undoEnPassant(move, undo.CapturedPiece)
	}

	// Move the piece back to its original square
	b.SetPiece(move.From.Rank, move.From.File, movedPiece)

	// Restore the destination square
	// For en passant, the destination square should be empty (pawn was captured elsewhere)
	if move.IsEnPassant {
		b.SetPiece(move.To.Rank, move.To.File, Empty)
	} else {
		// For normal moves, restore whatever was captured (or Empty)
		b.SetPiece(move.To.Rank, move.To.File, undo.CapturedPiece)
	}

	// Restore all board state
	b.castlingRights = undo.CastlingRights
	b.enPassantTarget = undo.EnPassantTarget
	b.halfMoveClock = undo.HalfMoveClock
	b.fullMoveNumber = undo.FullMoveNumber
	b.sideToMove = undo.SideToMove

	// Restore hash from history
	b.PopHash()
}

// undoCastling reverses the rook movement in castling
func (b *Board) undoCastling(move Move) {
	var rookFrom, rookTo Square

	switch move.To.File {
	case 6: // King-side castling
		rookFrom = Square{File: 5, Rank: move.From.Rank}
		rookTo = Square{File: 7, Rank: move.From.Rank}
	case 2: // Queen-side castling
		rookFrom = Square{File: 3, Rank: move.From.Rank}
		rookTo = Square{File: 0, Rank: move.From.Rank}
	}

	// Move the rook back
	rook := b.GetPiece(rookFrom.Rank, rookFrom.File)
	b.SetPiece(rookFrom.Rank, rookFrom.File, Empty)
	b.SetPiece(rookTo.Rank, rookTo.File, rook)
}

// undoEnPassant restores the captured pawn in en passant
func (b *Board) undoEnPassant(move Move, capturedPiece Piece) {
	// In en passant, the captured pawn is not on the destination square
	// It's on the same rank as the capturing pawn started, same file as destination
	
	var captureRank int
	if move.Piece == WhitePawn || (move.Piece == Empty && move.From.Rank < move.To.Rank) {
		// White pawn captured - black pawn was on rank 5 (index 4)
		captureRank = 4
	} else {
		// Black pawn captured - white pawn was on rank 4 (index 3)
		captureRank = 3
	}
	
	// Restore the captured pawn
	b.SetPiece(captureRank, move.To.File, capturedPiece)
}

// isWhitePiece checks if a piece belongs to white
func isWhitePiece(piece Piece) bool {
	return piece == WhitePawn || piece == WhiteRook || piece == WhiteKnight ||
		piece == WhiteBishop || piece == WhiteQueen || piece == WhiteKing
}

func (b *Board) ToFEN() string {
	var fen strings.Builder

	// FEN ranks start from 8 (top) and go down to 1 (bottom)
	for rankIndex := 0; rankIndex < 8; rankIndex++ {
		rank := 7 - rankIndex // Convert FEN rank to board array index
		emptyCount := 0

		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)

			if piece == Empty {
				emptyCount++
			} else {
				if emptyCount > 0 {
					fen.WriteString(strconv.Itoa(emptyCount))
					emptyCount = 0
				}
				fen.WriteRune(rune(piece))
			}
		}

		if emptyCount > 0 {
			fen.WriteString(strconv.Itoa(emptyCount))
		}

		if rankIndex < 7 {
			fen.WriteString("/")
		}
	}

	// Add FEN metadata: side to move, castling rights, en passant, halfmove clock, fullmove number
	fen.WriteString(" ")
	fen.WriteString(b.sideToMove)
	fen.WriteString(" ")
	if b.castlingRights != "" {
		fen.WriteString(b.castlingRights)
	} else {
		fen.WriteString("-")
	}
	fen.WriteString(" ")
	if b.enPassantTarget != nil {
		fen.WriteString(b.enPassantTarget.String())
	} else {
		fen.WriteString("-")
	}
	fen.WriteString(" ")
	fen.WriteString(strconv.Itoa(b.halfMoveClock))
	fen.WriteString(" ")
	fen.WriteString(strconv.Itoa(b.fullMoveNumber))

	return fen.String()
}
