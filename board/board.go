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
	castlingRights  string  // KQkq format
	enPassantTarget *Square // nil if no en passant
	halfMoveClock   int
	fullMoveNumber  int
	sideToMove      string // "w" or "b"

	// Bitboard representation (12 piece types)
	PieceBitboards [12]Bitboard // [WhitePawn, WhiteRook, WhiteKnight, WhiteBishop, WhiteQueen, WhiteKing, BlackPawn, BlackRook, BlackKnight, BlackBishop, BlackQueen, BlackKing]

	// Color bitboards (derived from piece bitboards)
	WhitePieces Bitboard // All white pieces
	BlackPieces Bitboard // All black pieces
	AllPieces   Bitboard // All occupied squares

	// Zobrist hash for incremental updates
	currentHash uint64
	hashHistory []uint64 // Stack for unmake operations
	hashUpdater HashUpdater // For incremental hash updates
}

// Define square color masks if not already available
var (
	LightSquares Bitboard
	DarkSquares  Bitboard
)

func init() {
	// Initialize light and dark square masks
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			square := FileRankToSquare(file, rank)
			if (file+rank)%2 == 0 {
				DarkSquares = DarkSquares.SetBit(square)
			} else {
				LightSquares = LightSquares.SetBit(square)
			}
		}
	}
}

func NewBoard() *Board {
	board := &Board{
		castlingRights:  "KQkq",
		enPassantTarget: nil,
		halfMoveClock:   0,
		fullMoveNumber:  1,
		sideToMove:      "w",
	}

	// Initialize bitboards (all empty)
	for i := 0; i < 12; i++ {
		board.PieceBitboards[i] = 0
	}
	board.WhitePieces = 0
	board.BlackPieces = 0
	board.AllPieces = 0

	// Initialize hash fields
	board.currentHash = 0
	board.hashHistory = make([]uint64, 0)

	return board
}

// GetHash returns the current zobrist hash of the board
func (b *Board) GetHash() uint64 {
	return b.currentHash
}

// SetHash sets the current zobrist hash (for initialization)
func (b *Board) SetHash(hash uint64) {
	b.currentHash = hash
}

// UpdateHash updates the hash incrementally by XORing with the given value
func (b *Board) UpdateHash(delta uint64) {
	b.currentHash ^= delta
}

// PushHash saves the current hash to history for unmake operations
func (b *Board) PushHash() {
	b.hashHistory = append(b.hashHistory, b.currentHash)
}

// PopHash restores the hash from history during unmake operations
func (b *Board) PopHash() {
	if len(b.hashHistory) > 0 {
		b.currentHash = b.hashHistory[len(b.hashHistory)-1]
		b.hashHistory = b.hashHistory[:len(b.hashHistory)-1]
	}
}

// InitializeHashFromPosition initializes the board's hash using an external hash calculator
// This should be called once when the board is set up to establish the baseline hash
func (b *Board) InitializeHashFromPosition(hashFunc func(*Board) uint64) {
	b.currentHash = hashFunc(b)
}

// HashUpdater interface for providing zobrist key updates
type HashUpdater interface {
	GetHashDelta(b *Board, move Move, oldState BoardState) uint64
}

// BoardState captures the board state before a move for hash calculation
type BoardState struct {
	CastlingRights  string
	EnPassantTarget *Square
	SideToMove      string
}

// SetHashUpdater sets the hash updater for incremental updates
func (b *Board) SetHashUpdater(updater HashUpdater) {
	b.hashUpdater = updater
}

// GetCurrentBoardState returns the current board state for hash calculations
func (b *Board) GetCurrentBoardState() BoardState {
	return BoardState{
		CastlingRights:  b.castlingRights,
		EnPassantTarget: b.enPassantTarget,
		SideToMove:      b.sideToMove,
	}
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

// BitboardIndexToPiece converts a bitboard index back to a piece type
func BitboardIndexToPiece(index int) Piece {
	switch index {
	case WhitePawnIndex:
		return WhitePawn
	case WhiteRookIndex:
		return WhiteRook
	case WhiteKnightIndex:
		return WhiteKnight
	case WhiteBishopIndex:
		return WhiteBishop
	case WhiteQueenIndex:
		return WhiteQueen
	case WhiteKingIndex:
		return WhiteKing
	case BlackPawnIndex:
		return BlackPawn
	case BlackRookIndex:
		return BlackRook
	case BlackKnightIndex:
		return BlackKnight
	case BlackBishopIndex:
		return BlackBishop
	case BlackQueenIndex:
		return BlackQueen
	case BlackKingIndex:
		return BlackKing
	default:
		return Empty
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
		squareBit := Bitboard(1) << square
		b.PieceBitboards[index] = b.PieceBitboards[index].SetBit(square)

		// Incrementally update derived bitboards
		if IsWhitePiece(piece) {
			b.WhitePieces |= squareBit
		} else {
			b.BlackPieces |= squareBit
		}
		b.AllPieces |= squareBit
	}
}

// removePieceBitboard removes a piece from the bitboards
func (b *Board) removePieceBitboard(rank, file int, piece Piece) {
	square := FileRankToSquare(file, rank)
	index := PieceToBitboardIndex(piece)

	if index != -1 {
		squareBit := Bitboard(1) << square
		b.PieceBitboards[index] = b.PieceBitboards[index].ClearBit(square)

		// Incrementally update derived bitboards
		if IsWhitePiece(piece) {
			b.WhitePieces &= ^squareBit
		} else {
			b.BlackPieces &= ^squareBit
		}
		b.AllPieces &= ^squareBit
	}
}

func (b *Board) GetPiece(rank, file int) Piece {
	square := FileRankToSquare(file, rank)

	// Quick check if square is empty
	if !b.AllPieces.HasBit(square) {
		return Empty
	}

	// Check each piece bitboard to find which piece is on this square
	for i, bitboard := range b.PieceBitboards {
		if bitboard.HasBit(square) {
			return BitboardIndexToPiece(i)
		}
	}

	// Should never reach here if bitboards are consistent
	panic("no piece found")
}

func (b *Board) SetPiece(rank, file int, piece Piece) {
	// Get old piece using bitboards
	oldPiece := b.GetPiece(rank, file)

	// Remove old piece from bitboards
	if oldPiece != Empty {
		b.removePieceBitboard(rank, file, oldPiece)
	}

	// Add new piece to bitboards
	if piece != Empty {
		b.setPieceBitboard(rank, file, piece)
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

	return board, nil
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

// getPieceCountFromBitboard returns the count of a specific piece type using bitboards
func (b *Board) getPieceCountFromBitboard(piece Piece) int {
	index := PieceToBitboardIndex(piece)
	if index == -1 {
		return 0
	}
	return b.PieceBitboards[index].PopCount()
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
		tracked := b.getPieceCountFromBitboard(piece)
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
	case WhitePawn:
		return "White Pawn"
	case WhiteRook:
		return "White Rook"
	case WhiteKnight:
		return "White Knight"
	case WhiteBishop:
		return "White Bishop"
	case WhiteQueen:
		return "White Queen"
	case WhiteKing:
		return "White King"
	case BlackPawn:
		return "Black Pawn"
	case BlackRook:
		return "Black Rook"
	case BlackKnight:
		return "Black Knight"
	case BlackBishop:
		return "Black Bishop"
	case BlackQueen:
		return "Black Queen"
	case BlackKing:
		return "Black King"
	default:
		return "Unknown"
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
		tracked := b.getPieceCountFromBitboard(piece)

		if actual != tracked {
			fmt.Printf("Validation failed: %c has actual=%d, tracked=%d\n", piece, actual, tracked)
			return false
		}
	}

	return true
}
