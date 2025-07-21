package board

import (
	"errors"
	"strconv"
	"strings"
)

type Move struct {
	From Square
	To   Square
	Piece Piece
	Captured Piece
	Promotion Piece
	IsCapture bool
	IsCastling bool
	IsEnPassant bool
}

// MoveUndo stores the state needed to undo a move
type MoveUndo struct {
	Move Move
	CapturedPiece Piece
	CastlingRights string
	EnPassantTarget *Square
	HalfMoveClock int
	FullMoveNumber int
	SideToMove string
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
	// Store current state for undo
	undo := MoveUndo{
		Move: move,
		CapturedPiece: b.GetPiece(move.To.Rank, move.To.File),
		CastlingRights: b.castlingRights,
		EnPassantTarget: b.enPassantTarget,
		HalfMoveClock: b.halfMoveClock,
		FullMoveNumber: b.fullMoveNumber,
		SideToMove: b.sideToMove,
	}
	
	// Make the move
	err := b.MakeMove(move)
	return undo, err
}

func (b *Board) MakeMove(move Move) error {
	// Get the piece at the from square
	piece := b.GetPiece(move.From.Rank, move.From.File)
	if piece == Empty {
		return errors.New("no piece at from square")
	}
	
	// Clear the from square
	b.SetPiece(move.From.Rank, move.From.File, Empty)
	
	// Handle promotion
	if move.Promotion != Empty {
		b.SetPiece(move.To.Rank, move.To.File, move.Promotion)
	} else {
		// Place the piece at the to square
		b.SetPiece(move.To.Rank, move.To.File, piece)
	}
	
	// Handle castling
	if move.IsCastling {
		return b.handleCastling(move)
	}
	
	// Handle en passant
	if move.IsEnPassant {
		return b.handleEnPassant(move)
	}
	
	return nil
}

func (b *Board) handleCastling(move Move) error {
	// Determine which rook to move based on the king's destination
	var rookFrom, rookTo Square
	
	if move.To.File == 6 { // King-side castling (O-O)
		rookFrom = Square{File: 7, Rank: move.From.Rank}
		rookTo = Square{File: 5, Rank: move.From.Rank}
	} else if move.To.File == 2 { // Queen-side castling (O-O-O)
		rookFrom = Square{File: 0, Rank: move.From.Rank}
		rookTo = Square{File: 3, Rank: move.From.Rank}
	} else {
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
	captureRank := move.From.Rank
	b.SetPiece(captureRank, move.To.File, Empty)
	return nil
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
		
		return Move{From: from, To: to, Promotion: Empty}, nil
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
	
	// Get the piece that was moved (may be promoted)
	movedPiece := b.GetPiece(move.To.Rank, move.To.File)
	
	// Handle undo of promotion
	if move.Promotion != Empty {
		// The original piece was a pawn
		if isWhitePiece(move.Piece) {
			movedPiece = WhitePawn
		} else {
			movedPiece = BlackPawn
		}
	}
	
	// Move piece back to original square
	b.SetPiece(move.From.Rank, move.From.File, movedPiece)
	
	// Restore captured piece or clear destination
	b.SetPiece(move.To.Rank, move.To.File, undo.CapturedPiece)
	
	// Handle undo of castling
	if move.IsCastling {
		b.undoCastling(move)
	}
	
	// Handle undo of en passant
	if move.IsEnPassant {
		b.undoEnPassant(move, undo.CapturedPiece)
	}
	
	// Restore board state
	b.castlingRights = undo.CastlingRights
	b.enPassantTarget = undo.EnPassantTarget
	b.halfMoveClock = undo.HalfMoveClock
	b.fullMoveNumber = undo.FullMoveNumber
	b.sideToMove = undo.SideToMove
}

// undoCastling reverses the rook movement in castling
func (b *Board) undoCastling(move Move) {
	var rookFrom, rookTo Square
	
	if move.To.File == 6 { // King-side castling
		rookFrom = Square{File: 5, Rank: move.From.Rank}
		rookTo = Square{File: 7, Rank: move.From.Rank}
	} else if move.To.File == 2 { // Queen-side castling  
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
	// Restore the captured pawn on the same rank as the moving pawn
	captureRank := move.From.Rank
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
	
	// Add minimal FEN metadata (assuming white to move, no castling rights, etc.)
	fen.WriteString(" w - - 0 1")
	
	return fen.String()
}