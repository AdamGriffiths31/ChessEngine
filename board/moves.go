package board

import (
	"errors"
	"fmt"
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
	// Ensure move.Piece is set for proper undo
	if move.Piece == Empty {
		move.Piece = b.GetPiece(move.From.Rank, move.From.File)
	}
	
	// Store current state for undo
	undo := MoveUndo{
		Move: move, // This now has the correct piece
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
		return b.handleCastling(move)
	}
	
	// Handle en passant
	if move.IsEnPassant {
		return b.handleEnPassant(move)
	}
	
	// Update game state
	b.updateGameState(move, piece, capturedPiece)
	
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

func (b *Board) updateGameState(move Move, piece Piece, capturedPiece Piece) {
	// Switch side to move
	if b.sideToMove == "w" {
		b.sideToMove = "b"
	} else {
		b.sideToMove = "w"
		b.fullMoveNumber++
	}
	
	// Update half move clock - reset if pawn move or capture
	if piece == WhitePawn || piece == BlackPawn || capturedPiece != Empty {
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
	if piece == WhiteKing {
		b.castlingRights = strings.ReplaceAll(b.castlingRights, "K", "")
		b.castlingRights = strings.ReplaceAll(b.castlingRights, "Q", "")
	} else if piece == BlackKing {
		b.castlingRights = strings.ReplaceAll(b.castlingRights, "k", "")
		b.castlingRights = strings.ReplaceAll(b.castlingRights, "q", "")
	}
	
	// If rook moved from initial position, remove corresponding castling right
	if piece == WhiteRook {
		if move.From.Rank == 0 && move.From.File == 0 {
			b.castlingRights = strings.ReplaceAll(b.castlingRights, "Q", "") // queenside
		} else if move.From.Rank == 0 && move.From.File == 7 {
			b.castlingRights = strings.ReplaceAll(b.castlingRights, "K", "") // kingside
		}
	} else if piece == BlackRook {
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
		// The original piece was a pawn - determine color from the promoted piece
		// or from the destination square piece (which should be the promoted piece)
		promotedPiece := b.GetPiece(move.To.Rank, move.To.File)
		
		// If the move.Piece field is set and valid, use it to determine color
		if move.Piece != Empty && isValidPiece(move.Piece) {
			if isWhitePiece(move.Piece) {
				movedPiece = WhitePawn
			} else {
				movedPiece = BlackPawn
			}
		} else if promotedPiece != Empty && isValidPiece(promotedPiece) {
			// Fallback: determine from the promoted piece on the board
			if isWhitePiece(promotedPiece) {
				movedPiece = WhitePawn
			} else {
				movedPiece = BlackPawn
			}
		} else {
			// Ultimate fallback: determine from board position
			// Promotions only happen on ranks 1 and 8
			if move.To.Rank == 7 {
				// White pawn promoted to rank 8
				movedPiece = WhitePawn
			} else if move.To.Rank == 0 {
				// Black pawn promoted to rank 1
				movedPiece = BlackPawn
			} else {
				// This shouldn't happen, but default to the piece that was there
				movedPiece = b.GetPiece(move.To.Rank, move.To.File)
			}
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
	
	// Validate board consistency after unmake (debug builds only)
	if !b.validateBoardConsistency() {
		// This is a critical error but we can't return an error from UnmakeMove
		// Log it for debugging purposes
		fmt.Printf("CRITICAL: Board inconsistency detected after UnmakeMove!\n")
	}
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