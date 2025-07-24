package board

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Piece rune

const (
	Empty       Piece = '.'
	WhitePawn   Piece = 'P'
	WhiteRook   Piece = 'R'
	WhiteKnight Piece = 'N'
	WhiteBishop Piece = 'B'
	WhiteQueen  Piece = 'Q'
	WhiteKing   Piece = 'K'
	BlackPawn   Piece = 'p'
	BlackRook   Piece = 'r'
	BlackKnight Piece = 'n'
	BlackBishop Piece = 'b'
	BlackQueen  Piece = 'q'
	BlackKing   Piece = 'k'
)

// Bitboard indices for piece types
const (
	WhitePawnIndex   = 0
	WhiteRookIndex   = 1
	WhiteKnightIndex = 2
	WhiteBishopIndex = 3
	WhiteQueenIndex  = 4
	WhiteKingIndex   = 5
	BlackPawnIndex   = 6
	BlackRookIndex   = 7
	BlackKnightIndex = 8
	BlackBishopIndex = 9
	BlackQueenIndex  = 10
	BlackKingIndex   = 11
)

type Square struct {
	File int // 0-7 (a-h)
	Rank int // 0-7 (1-8)
}

type Board struct {
	squares      [8][8]Piece
	castlingRights string // KQkq format
	enPassantTarget *Square // nil if no en passant
	halfMoveClock   int
	fullMoveNumber  int
	sideToMove      string // "w" or "b"
	
	// Piece lists for fast lookup
	pieceLists   map[Piece][]Square  // Track positions of each piece type
	pieceCount   map[Piece]int       // Count of each piece type
	
	// Bitboard representation (12 piece types)
	PieceBitboards [12]Bitboard // [WhitePawn, WhiteRook, WhiteKnight, WhiteBishop, WhiteQueen, WhiteKing, BlackPawn, BlackRook, BlackKnight, BlackBishop, BlackQueen, BlackKing]
	
	// Color bitboards (derived from piece bitboards)
	WhitePieces Bitboard // All white pieces
	BlackPieces Bitboard // All black pieces
	AllPieces   Bitboard // All occupied squares
}

func NewBoard() *Board {
	board := &Board{
		castlingRights: "KQkq",
		enPassantTarget: nil,
		halfMoveClock: 0,
		fullMoveNumber: 1,
		sideToMove: "w",
		pieceLists: make(map[Piece][]Square),
		pieceCount: make(map[Piece]int),
	}
	
	// Initialize piece lists for all piece types
	pieces := []Piece{
		WhitePawn, WhiteRook, WhiteKnight, WhiteBishop, WhiteQueen, WhiteKing,
		BlackPawn, BlackRook, BlackKnight, BlackBishop, BlackQueen, BlackKing,
	}
	
	for _, piece := range pieces {
		board.pieceLists[piece] = make([]Square, 0, 16) // Max 16 of any piece type
		board.pieceCount[piece] = 0
	}
	
	// Initialize array representation
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			board.squares[rank][file] = Empty
		}
	}
	
	// Initialize bitboards (all empty)
	for i := 0; i < 12; i++ {
		board.PieceBitboards[i] = 0
	}
	board.WhitePieces = 0
	board.BlackPieces = 0
	board.AllPieces = 0
	
	return board
}

// Bitboard helper functions

// PieceToBitboardIndex returns the bitboard index for a given piece
func PieceToBitboardIndex(piece Piece) int {
	switch piece {
	case WhitePawn:
		return WhitePawnIndex
	case WhiteRook:
		return WhiteRookIndex
	case WhiteKnight:
		return WhiteKnightIndex
	case WhiteBishop:
		return WhiteBishopIndex
	case WhiteQueen:
		return WhiteQueenIndex
	case WhiteKing:
		return WhiteKingIndex
	case BlackPawn:
		return BlackPawnIndex
	case BlackRook:
		return BlackRookIndex
	case BlackKnight:
		return BlackKnightIndex
	case BlackBishop:
		return BlackBishopIndex
	case BlackQueen:
		return BlackQueenIndex
	case BlackKing:
		return BlackKingIndex
	default:
		return -1 // Invalid piece
	}
}

// GetPieceBitboard returns the bitboard for a specific piece type
func (b *Board) GetPieceBitboard(piece Piece) Bitboard {
	index := PieceToBitboardIndex(piece)
	if index == -1 {
		return 0
	}
	return b.PieceBitboards[index]
}

// GetColorBitboard returns the bitboard for all pieces of a given color
func (b *Board) GetColorBitboard(color BitboardColor) Bitboard {
	if color == BitboardWhite {
		return b.WhitePieces
	}
	return b.BlackPieces
}

// updateBitboards updates the derived bitboards (color and all pieces)
func (b *Board) updateBitboards() {
	b.WhitePieces = 0
	b.BlackPieces = 0
	
	// Combine all white piece bitboards
	for i := WhitePawnIndex; i <= WhiteKingIndex; i++ {
		b.WhitePieces |= b.PieceBitboards[i]
	}
	
	// Combine all black piece bitboards
	for i := BlackPawnIndex; i <= BlackKingIndex; i++ {
		b.BlackPieces |= b.PieceBitboards[i]
	}
	
	// All pieces is the union of white and black
	b.AllPieces = b.WhitePieces | b.BlackPieces
}

// setPieceBitboard sets a piece on the bitboards
func (b *Board) setPieceBitboard(rank, file int, piece Piece) {
	square := FileRankToSquare(file, rank)
	index := PieceToBitboardIndex(piece)
	
	if index != -1 {
		b.PieceBitboards[index] = b.PieceBitboards[index].SetBit(square)
	}
	
	b.updateBitboards()
}

// removePieceBitboard removes a piece from the bitboards
func (b *Board) removePieceBitboard(rank, file int, piece Piece) {
	square := FileRankToSquare(file, rank)
	index := PieceToBitboardIndex(piece)
	
	if index != -1 {
		b.PieceBitboards[index] = b.PieceBitboards[index].ClearBit(square)
	}
	
	b.updateBitboards()
}

func (b *Board) GetPiece(rank, file int) Piece {
	if rank < 0 || rank > 7 || file < 0 || file > 7 {
		return Empty
	}
	return b.squares[rank][file]
}

func (b *Board) SetPiece(rank, file int, piece Piece) {
	if rank >= 0 && rank <= 7 && file >= 0 && file <= 7 {
		square := Square{File: file, Rank: rank}
		oldPiece := b.squares[rank][file]
		
		
		// Remove old piece from bitboards
		if oldPiece != Empty {
			b.removePieceBitboard(rank, file, oldPiece)
			b.removePieceFromList(oldPiece, square)
		}
		
		// Update array representation
		b.squares[rank][file] = piece
		
		// Add new piece to bitboards and lists
		if piece != Empty {
			b.setPieceBitboard(rank, file, piece)
			b.addPieceToList(piece, square)
		}
	}
}

func FromFEN(fen string) (*Board, error) {
	if fen == "" {
		return nil, errors.New("invalid FEN: missing board position")
	}
	
	parts := strings.Split(fen, " ")
	if len(parts) < 1 {
		return nil, errors.New("invalid FEN: missing board position")
	}

	boardPart := parts[0]
	ranks := strings.Split(boardPart, "/")
	
	if len(ranks) != 8 {
		return nil, errors.New("invalid FEN: must have exactly 8 ranks")
	}

	board := NewBoard()

	// Parse board position
	for rankIndex, rankStr := range ranks {
		// FEN ranks start from 8 (top) and go down to 1 (bottom)
		// Array index 0 should be rank 1, index 7 should be rank 8
		// So FEN rank 8 (rankIndex 0) goes to array index 7
		actualRank := 7 - rankIndex
		file := 0
		for _, char := range rankStr {
			if file >= 8 {
				return nil, errors.New("invalid FEN: too many files in rank")
			}

			if char >= '1' && char <= '8' {
				emptySquares, _ := strconv.Atoi(string(char))
				for i := 0; i < emptySquares; i++ {
					if file >= 8 {
						return nil, errors.New("invalid FEN: too many files in rank")
					}
					board.SetPiece(actualRank, file, Empty)
					file++
				}
			} else {
				piece := Piece(char)
				if !isValidPiece(piece) {
					return nil, errors.New("invalid FEN: invalid piece character")
				}
				board.SetPiece(actualRank, file, piece)
				file++
			}
		}
		
		if file != 8 {
			return nil, errors.New("invalid FEN: incorrect number of files in rank")
		}
	}

	// Parse additional FEN fields if available
	if len(parts) >= 2 {
		board.sideToMove = parts[1]
	}
	if len(parts) >= 3 {
		board.castlingRights = parts[2]
	}
	if len(parts) >= 4 {
		enPassantStr := parts[3]
		if enPassantStr != "-" {
			file := int(enPassantStr[0] - 'a')
			rank := int(enPassantStr[1] - '1')
			if file >= 0 && file <= 7 && rank >= 0 && rank <= 7 {
				square := Square{File: file, Rank: rank}
				board.enPassantTarget = &square
			}
		}
	}
	if len(parts) >= 5 {
		if halfMove, err := strconv.Atoi(parts[4]); err == nil {
			board.halfMoveClock = halfMove
		}
	}
	if len(parts) >= 6 {
		if fullMove, err := strconv.Atoi(parts[5]); err == nil {
			board.fullMoveNumber = fullMove
		}
	}

	// Generate bitboards from the array representation
	board.generateBitboardsFromArray()

	return board, nil
}

// generateBitboardsFromArray populates bitboards from the current array representation
func (b *Board) generateBitboardsFromArray() {
	// Clear all bitboards
	for i := 0; i < 12; i++ {
		b.PieceBitboards[i] = 0
	}
	
	// Scan the board and set appropriate bitboard bits
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.squares[rank][file]
			if piece != Empty {
				square := FileRankToSquare(file, rank)
				index := PieceToBitboardIndex(piece)
				if index != -1 {
					b.PieceBitboards[index] = b.PieceBitboards[index].SetBit(square)
				}
			}
		}
	}
	
	// Update derived bitboards
	b.updateBitboards()
}

func isValidPiece(piece Piece) bool {
	validPieces := []Piece{
		WhitePawn, WhiteRook, WhiteKnight, WhiteBishop, WhiteQueen, WhiteKing,
		BlackPawn, BlackRook, BlackKnight, BlackBishop, BlackQueen, BlackKing,
	}
	
	for _, validPiece := range validPieces {
		if piece == validPiece {
			return true
		}
	}
	return false
}

// addPieceToList adds a piece to the piece list
func (b *Board) addPieceToList(piece Piece, square Square) {
	if piece == Empty {
		return
	}
	
	b.pieceLists[piece] = append(b.pieceLists[piece], square)
	b.pieceCount[piece]++
	
}

// removePieceFromList removes a piece from the piece list
func (b *Board) removePieceFromList(piece Piece, square Square) {
	if piece == Empty {
		return
	}
	
	list := b.pieceLists[piece]
	
	for i, sq := range list {
		if sq.File == square.File && sq.Rank == square.Rank {
			// Remove by swapping with last element
			list[i] = list[len(list)-1]
			b.pieceLists[piece] = list[:len(list)-1]
			b.pieceCount[piece]--
			break
		}
	}
	
}

// GetPieceList returns all squares containing a specific piece type
func (b *Board) GetPieceList(piece Piece) []Square {
	return b.pieceLists[piece]
}

// GetPieceCount returns the count of a specific piece type
func (b *Board) GetPieceCount(piece Piece) int {
	return b.pieceCount[piece]
}

// Getter methods for board state
func (b *Board) GetCastlingRights() string {
	return b.castlingRights
}

func (b *Board) GetEnPassantTarget() *Square {
	return b.enPassantTarget
}

func (b *Board) GetHalfMoveClock() int {
	return b.halfMoveClock
}

func (b *Board) GetFullMoveNumber() int {
	return b.fullMoveNumber
}

func (b *Board) GetSideToMove() string {
	return b.sideToMove
}

// DebugPieceCounts displays comprehensive piece count information
func (b *Board) DebugPieceCounts(label string) {
	fmt.Printf("\n=== Piece Counts: %s ===\n", label)
	
	// Count pieces by scanning the board
	actualCounts := make(map[Piece]int)
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece != Empty {
				actualCounts[piece]++
			}
		}
	}
	
	// Display counts
	pieces := []Piece{WhitePawn, WhiteRook, WhiteKnight, WhiteBishop, WhiteQueen, WhiteKing,
					 BlackPawn, BlackRook, BlackKnight, BlackBishop, BlackQueen, BlackKing}
	
	for _, piece := range pieces {
		actual := actualCounts[piece]
		tracked := b.GetPieceCount(piece)
		name := getPieceName(piece)
		
		if actual != tracked {
			fmt.Printf("❌ %s: actual=%d, tracked=%d (MISMATCH!)\n", name, actual, tracked)
		} else {
			fmt.Printf("✅ %s: %d\n", name, actual)
		}
	}
	
	// Check total counts
	whiteActual := actualCounts[WhitePawn] + actualCounts[WhiteRook] + actualCounts[WhiteKnight] + 
				   actualCounts[WhiteBishop] + actualCounts[WhiteQueen] + actualCounts[WhiteKing]
	blackActual := actualCounts[BlackPawn] + actualCounts[BlackRook] + actualCounts[BlackKnight] + 
				   actualCounts[BlackBishop] + actualCounts[BlackQueen] + actualCounts[BlackKing]
	
	fmt.Printf("\nTotals: White=%d, Black=%d\n", whiteActual, blackActual)
}

// getPieceName returns a human-readable name for a piece
func getPieceName(piece Piece) string {
	switch piece {
	case WhitePawn: return "White Pawn"
	case WhiteRook: return "White Rook"
	case WhiteKnight: return "White Knight"
	case WhiteBishop: return "White Bishop"
	case WhiteQueen: return "White Queen"
	case WhiteKing: return "White King"
	case BlackPawn: return "Black Pawn"
	case BlackRook: return "Black Rook"
	case BlackKnight: return "Black Knight"
	case BlackBishop: return "Black Bishop"
	case BlackQueen: return "Black Queen"
	case BlackKing: return "Black King"
	default: return "Unknown"
	}
}

// Setter methods for board state
func (b *Board) SetCastlingRights(rights string) {
	b.castlingRights = rights
}

func (b *Board) SetEnPassantTarget(target *Square) {
	b.enPassantTarget = target
}

func (b *Board) SetHalfMoveClock(clock int) {
	b.halfMoveClock = clock
}

func (b *Board) SetFullMoveNumber(num int) {
	b.fullMoveNumber = num
}

func (b *Board) SetSideToMove(side string) {
	b.sideToMove = side
}

// validateBoardConsistency checks if the board state is internally consistent
func (b *Board) validateBoardConsistency() bool {
	// Count pieces by scanning the board
	actualCounts := make(map[Piece]int)
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece != Empty {
				actualCounts[piece]++
			}
		}
	}
	
	// Compare with tracked counts
	pieces := []Piece{WhitePawn, WhiteRook, WhiteKnight, WhiteBishop, WhiteQueen, WhiteKing,
					 BlackPawn, BlackRook, BlackKnight, BlackBishop, BlackQueen, BlackKing}
	
	for _, piece := range pieces {
		actual := actualCounts[piece]
		tracked := b.GetPieceCount(piece)
		
		if actual != tracked {
			fmt.Printf("Validation failed: %c has actual=%d, tracked=%d\n", piece, actual, tracked)
			return false
		}
	}
	
	return true
}