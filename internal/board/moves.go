package board

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"
)

// Move represents a chess move with all necessary information
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

// HasPromotion reports whether m promotes.
func (m Move) HasPromotion() bool {
	return m.Promotion != Empty && m.Promotion != 0
}

// IsQuiet reports whether m is neither a capture nor a promotion.
func (m Move) IsQuiet() bool {
	return !m.IsCapture && !m.HasPromotion()
}

// EnPassantCapturedSquare returns the square of the pawn an en passant move captures.
func (m Move) EnPassantCapturedSquare() Square {
	return Square{File: m.To.File, Rank: m.From.Rank}
}

// IsValid reports whether m looks like a plausible, non-degenerate move.
func (m Move) IsValid() bool {
	return m.onBoard() && m.From != m.To
}

func (m Move) onBoard() bool {
	return m.From.File >= 0 && m.From.File <= 7 && m.From.Rank >= 0 && m.From.Rank <= 7 &&
		m.To.File >= 0 && m.To.File <= 7 && m.To.Rank >= 0 && m.To.Rank <= 7
}

// UCI returns the move in UCI notation (e.g. "e2e4", "e7e8q"), or "0000" when
// a square is off the board.
func (m Move) UCI() string {
	if !m.onBoard() {
		return "0000"
	}
	uciMove := m.From.String() + m.To.String()
	if m.HasPromotion() {
		uciMove += strings.ToLower(string(m.Promotion))
	}
	return uciMove
}

// MoveUndo stores the state needed to undo a move
type MoveUndo struct {
	Move            Move
	CapturedPiece   Piece
	CastlingRights  CastlingRights
	EnPassantSquare Square
	HasEnPassant    bool
	HalfMoveClock   int
	FullMoveNumber  int
	BlackToMove     bool
	Hash            uint64
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

func (s Square) String() string {
	return string(rune('a'+s.File)) + string(rune('1'+s.Rank))
}

// MakeMoveWithUndo makes a move and returns undo information
func (b *Board) MakeMoveWithUndo(move Move) (MoveUndo, error) {
	if move.Piece == Empty || move.Piece == 0 {
		move.Piece = b.PieceAt(move.From)
	}

	var capturedPiece Piece
	if move.IsEnPassant {
		capturedPiece = b.PieceAt(move.EnPassantCapturedSquare())
	} else {
		capturedPiece = b.PieceAt(move.To)
	}

	undo := MoveUndo{
		Move:            move,
		CapturedPiece:   capturedPiece,
		CastlingRights:  b.castlingRights,
		EnPassantSquare: b.enPassantSquare,
		HasEnPassant:    b.hasEnPassant,
		HalfMoveClock:   b.halfMoveClock,
		FullMoveNumber:  b.fullMoveNumber,
		BlackToMove:     b.blackToMove,
		Hash:            b.GetHash(),
	}

	b.PushHash()

	b.PushEvalState()

	oldState := b.GetCurrentBoardState()

	err := b.MakeMove(move)

	if b.hashUpdater != nil && err == nil {
		hashDelta := b.hashUpdater.GetHashDelta(b, move, oldState)
		b.UpdateHash(hashDelta)
	}

	return undo, err
}

// MakeNullMove makes a null move (passes the turn) and returns undo information
func (b *Board) MakeNullMove() MoveUndo {
	undo := MoveUndo{
		Move:            NullMove,
		CapturedPiece:   Empty,
		CastlingRights:  b.castlingRights,
		EnPassantSquare: b.enPassantSquare,
		HasEnPassant:    b.hasEnPassant,
		HalfMoveClock:   b.halfMoveClock,
		FullMoveNumber:  b.fullMoveNumber,
		BlackToMove:     b.blackToMove,
		Hash:            b.GetHash(),
	}

	b.halfMoveClock++

	var nullDelta uint64
	if b.hashUpdater != nil {
		nullDelta = b.hashUpdater.GetNullMoveDelta(b)
	}

	if b.blackToMove {
		b.fullMoveNumber++
	}
	b.blackToMove = !b.blackToMove

	b.hasEnPassant = false

	if b.hashUpdater != nil {
		b.UpdateHash(nullDelta)
	}

	return undo
}

// UnmakeNullMove reverses a null move using the stored undo information
func (b *Board) UnmakeNullMove(undo MoveUndo) {
	b.castlingRights = undo.CastlingRights
	b.enPassantSquare = undo.EnPassantSquare
	b.hasEnPassant = undo.HasEnPassant
	b.halfMoveClock = undo.HalfMoveClock
	b.fullMoveNumber = undo.FullMoveNumber
	b.blackToMove = undo.BlackToMove

	b.SetHash(undo.Hash)
}

// MakeMove executes a chess move on the board, updating the position
func (b *Board) MakeMove(move Move) error {
	var piece Piece
	if move.Piece != Empty {
		piece = move.Piece
	} else {
		piece = b.PieceAt(move.From)
		if piece == Empty {
			return errors.New("no piece at from square")
		}
	}

	if !move.IsEnPassant && (piece == WhitePawn || piece == BlackPawn) {
		targetPiece := b.PieceAt(move.To)
		if targetPiece == Empty && move.From.File != move.To.File {
			if b.hasEnPassant && b.enPassantSquare.File == move.To.File && b.enPassantSquare.Rank == move.To.Rank {
				move.IsEnPassant = true
			}
		}
	}

	capturedPiece := b.PieceAt(move.To)

	b.SetPieceAt(move.From, Empty)

	if move.Promotion != Empty && move.Promotion != 0 {
		b.SetPieceAt(move.To, move.Promotion)
	} else {
		b.SetPieceAt(move.To, piece)
	}

	if move.IsCastling {
		err := b.handleCastling(move)
		if err != nil {
			return err
		}
	}

	if move.IsEnPassant {
		err := b.handleEnPassant(move)
		if err != nil {
			return err
		}
	}

	b.updateGameState(move, piece, capturedPiece)

	return nil
}

func (b *Board) handleCastling(move Move) error {
	var rookFrom, rookTo Square

	switch move.To.File {
	case 6:
		rookFrom = Square{File: 7, Rank: move.From.Rank}
		rookTo = Square{File: 5, Rank: move.From.Rank}
	case 2:
		rookFrom = Square{File: 0, Rank: move.From.Rank}
		rookTo = Square{File: 3, Rank: move.From.Rank}
	default:
		return errors.New("invalid castling move")
	}

	rook := b.PieceAt(rookFrom)
	b.SetPieceAt(rookFrom, Empty)
	b.SetPieceAt(rookTo, rook)

	return nil
}

func (b *Board) handleEnPassant(move Move) error {
	b.SetPieceAt(move.EnPassantCapturedSquare(), Empty)
	return nil
}

func (b *Board) updateGameState(move Move, piece Piece, capturedPiece Piece) {
	if b.blackToMove {
		b.fullMoveNumber++
	}
	b.blackToMove = !b.blackToMove

	if piece == WhitePawn || piece == BlackPawn || capturedPiece != Empty || move.IsEnPassant {
		b.halfMoveClock = 0
	} else {
		b.halfMoveClock++
	}

	b.UpdateCastlingRights(move, piece)

	b.updateEnPassantTarget(move, piece)
}

// UpdateCastlingRights clears whichever castling rights move invalidates:
// the king or rook moving, or a rook being captured on its home square by
// any piece.
func (b *Board) UpdateCastlingRights(move Move, piece Piece) {
	if b.castlingRights == 0 {
		return
	}

	switch piece {
	case WhiteKing:
		b.castlingRights &^= WhiteKingside
		b.castlingRights &^= WhiteQueenside
	case BlackKing:
		b.castlingRights &^= BlackKingside
		b.castlingRights &^= BlackQueenside
	}

	switch piece {
	case WhiteRook:
		if move.From.Rank == 0 && move.From.File == 0 {
			b.castlingRights &^= WhiteQueenside
		} else if move.From.Rank == 0 && move.From.File == 7 {
			b.castlingRights &^= WhiteKingside
		}
	case BlackRook:
		if move.From.Rank == 7 && move.From.File == 0 {
			b.castlingRights &^= BlackQueenside
		} else if move.From.Rank == 7 && move.From.File == 7 {
			b.castlingRights &^= BlackKingside
		}
	}

	if move.IsCapture {
		switch {
		case move.To.Rank == 0 && move.To.File == 0:
			b.castlingRights &^= WhiteQueenside
		case move.To.Rank == 0 && move.To.File == 7:
			b.castlingRights &^= WhiteKingside
		case move.To.Rank == 7 && move.To.File == 0:
			b.castlingRights &^= BlackQueenside
		case move.To.Rank == 7 && move.To.File == 7:
			b.castlingRights &^= BlackKingside
		}
	}

}

func (b *Board) updateEnPassantTarget(move Move, piece Piece) {
	b.hasEnPassant = false

	if piece == WhitePawn && move.From.Rank == 1 && move.To.Rank == 3 {
		b.enPassantSquare = Square{File: move.From.File, Rank: 2}
		b.hasEnPassant = true
	} else if piece == BlackPawn && move.From.Rank == 6 && move.To.Rank == 4 {
		b.enPassantSquare = Square{File: move.From.File, Rank: 5}
		b.hasEnPassant = true
	}
}

// UnmakeMove reverses a move using the undo information
func (b *Board) UnmakeMove(undo MoveUndo) {
	move := undo.Move

	var movedPiece Piece
	if move.Promotion != Empty && move.Promotion != 0 {
		switch move.To.Rank {
		case 7:
			movedPiece = WhitePawn
		case 0:
			movedPiece = BlackPawn
		default:
			movedPiece = move.Piece
		}
	} else {
		movedPiece = move.Piece
	}

	if movedPiece == Empty {
		slog.Warn("UnmakeMove could not determine piece to restore",
			"from", move.From.String(), "to", move.To.String())
		return
	}

	if move.IsCastling {
		b.undoCastling(move)
	}

	if move.IsEnPassant {
		b.undoEnPassant(move, undo.CapturedPiece)
	}

	b.restorePiece(move.From, movedPiece)

	if move.IsEnPassant {
		b.restorePiece(move.To, Empty)
	} else {
		b.restorePiece(move.To, undo.CapturedPiece)
	}

	b.castlingRights = undo.CastlingRights
	b.enPassantSquare = undo.EnPassantSquare
	b.hasEnPassant = undo.HasEnPassant
	b.halfMoveClock = undo.HalfMoveClock
	b.fullMoveNumber = undo.FullMoveNumber
	b.blackToMove = undo.BlackToMove

	b.PopHash()

	b.PopEvalState()
}

func (b *Board) undoCastling(move Move) {
	var rookFrom, rookTo Square

	switch move.To.File {
	case 6:
		rookFrom = Square{File: 5, Rank: move.From.Rank}
		rookTo = Square{File: 7, Rank: move.From.Rank}
	case 2:
		rookFrom = Square{File: 3, Rank: move.From.Rank}
		rookTo = Square{File: 0, Rank: move.From.Rank}
	}

	rook := b.PieceAt(rookFrom)
	b.restorePiece(rookFrom, Empty)
	b.restorePiece(rookTo, rook)
}

func (b *Board) undoEnPassant(move Move, capturedPiece Piece) {
	b.restorePiece(move.EnPassantCapturedSquare(), capturedPiece)
}

// ToFEN converts the board position to FEN (Forsyth-Edwards Notation) string
func (b *Board) ToFEN() string {
	var fen strings.Builder

	for rankIndex := range 8 {
		rank := 7 - rankIndex
		emptyCount := 0

		for file := range 8 {
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

	fen.WriteString(" ")
	fen.WriteString(b.GetSideToMove())
	fen.WriteString(" ")
	fen.WriteString(b.castlingRights.String())
	fen.WriteString(" ")
	if b.hasEnPassant {
		fen.WriteString(b.enPassantSquare.String())
	} else {
		fen.WriteString("-")
	}
	fen.WriteString(" ")
	fen.WriteString(strconv.Itoa(b.halfMoveClock))
	fen.WriteString(" ")
	fen.WriteString(strconv.Itoa(b.fullMoveNumber))

	return fen.String()
}
