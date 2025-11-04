package board

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/game/ai/evaluation/values"
)

// Piece represents a chess piece using standard FEN notation
type Piece rune

const (
	// Empty represents an empty square
	Empty Piece = '.'
	// WhitePawn represents a white pawn piece
	WhitePawn Piece = 'P'
	// WhiteRook represents a white rook piece
	WhiteRook Piece = 'R'
	// WhiteKnight represents a white knight piece
	WhiteKnight Piece = 'N'
	// WhiteBishop represents a white bishop piece
	WhiteBishop Piece = 'B'
	// WhiteQueen represents a white queen piece
	WhiteQueen Piece = 'Q'
	// WhiteKing represents a white king piece
	WhiteKing Piece = 'K'
	// BlackPawn represents a black pawn piece
	BlackPawn Piece = 'p'
	// BlackRook represents a black rook piece
	BlackRook Piece = 'r'
	// BlackKnight represents a black knight piece
	BlackKnight Piece = 'n'
	// BlackBishop represents a black bishop piece
	BlackBishop Piece = 'b'
	// BlackQueen represents a black queen piece
	BlackQueen Piece = 'q'
	// BlackKing represents a black king piece
	BlackKing Piece = 'k'
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

// Square represents a position on the chess board
type Square struct {
	File int // 0-7 (a-h)
	Rank int // 0-7 (1-8)
}

// EvalState stores incremental evaluation scores for undo operations.
// These scores are maintained incrementally during moves and restored during unmake.
type EvalState struct {
	MaterialScore int
	PSTScore      int
}

// Board represents a chess board with piece positions and game state
type Board struct {
	castlingRights  string // KQkq format
	enPassantSquare Square // en passant target square (only valid if hasEnPassant is true)
	hasEnPassant    bool   // true if en passant is possible
	halfMoveClock   int
	fullMoveNumber  int
	sideToMove      string // "w" or "b"

	// Bitboard representation (12 piece types)
	PieceBitboards [12]Bitboard // [WhitePawn, WhiteRook, WhiteKnight, WhiteBishop, WhiteQueen, WhiteKing, BlackPawn, BlackRook, BlackKnight, BlackBishop, BlackQueen, BlackKing]

	// Mailbox representation for O(1) piece lookup
	Mailbox [64]Piece // Direct piece lookup by square index

	// Color bitboards (derived from piece bitboards)
	WhitePieces Bitboard // All white pieces
	BlackPieces Bitboard // All black pieces
	AllPieces   Bitboard // All occupied squares

	// Zobrist hash for incremental updates
	currentHash uint64
	hashHistory []uint64    // Stack for unmake operations
	hashUpdater HashUpdater // For incremental hash updates

	// Incremental evaluation scores (from White's perspective)
	materialScore int         // Sum of all piece values
	pstScore      int         // Sum of all positional bonuses
	evalHistory   []EvalState // Stack for unmake operations
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

// NewBoard creates a new chess board with the standard starting position
func NewBoard() *Board {
	board := &Board{
		castlingRights: "KQkq",
		hasEnPassant:   false,
		halfMoveClock:  0,
		fullMoveNumber: 1,
		sideToMove:     "w",
	}

	// Initialize bitboards (all empty)
	for i := 0; i < 12; i++ {
		board.PieceBitboards[i] = 0
	}
	board.WhitePieces = 0
	board.BlackPieces = 0
	board.AllPieces = 0

	// Initialize mailbox (all empty)
	for i := 0; i < 64; i++ {
		board.Mailbox[i] = Empty
	}

	// Initialize hash fields
	board.currentHash = 0
	board.hashHistory = make([]uint64, 0)

	// Initialize evaluation fields
	board.materialScore = 0
	board.pstScore = 0
	board.evalHistory = make([]EvalState, 0)

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

// GetMaterialScore returns the current incrementally maintained material score.
// This value is updated automatically by SetPiece and represents the sum of all
// piece values from White's perspective (positive for White, negative for Black).
func (b *Board) GetMaterialScore() int {
	return b.materialScore
}

// GetPSTScore returns the current incrementally maintained piece-square table score.
// This value is updated automatically by SetPiece and represents the sum of all
// positional bonuses from White's perspective.
func (b *Board) GetPSTScore() int {
	return b.pstScore
}

// PushEvalState saves current evaluation state to history for later restoration.
// This should be called before making a move to enable proper unmake operations.
func (b *Board) PushEvalState() {
	b.evalHistory = append(b.evalHistory, EvalState{
		MaterialScore: b.materialScore,
		PSTScore:      b.pstScore,
	})
}

// PopEvalState restores evaluation state from history during move unmake.
// This reverts the evaluation scores to their state before the last PushEvalState call.
func (b *Board) PopEvalState() {
	if len(b.evalHistory) > 0 {
		state := b.evalHistory[len(b.evalHistory)-1]
		b.materialScore = state.MaterialScore
		b.pstScore = state.PSTScore
		b.evalHistory = b.evalHistory[:len(b.evalHistory)-1]
	}
}

// updateEvalScoresForPiece updates the incremental scores after a piece placement/removal
// This is an internal helper called by SetPiece
func (b *Board) updateEvalScoresForPiece(rank, file int, oldPiece, newPiece Piece) {
	// Remove old piece contribution
	if oldPiece != Empty {
		b.materialScore -= values.GetPieceValue(values.Piece(oldPiece))
		b.pstScore -= values.GetPositionalBonus(values.Piece(oldPiece), rank, file)
	}

	// Add new piece contribution
	if newPiece != Empty {
		b.materialScore += values.GetPieceValue(values.Piece(newPiece))
		b.pstScore += values.GetPositionalBonus(values.Piece(newPiece), rank, file)
	}
}

// InitializeHashFromPosition initializes the board's hash using an external hash calculator
// This should be called once when the board is set up to establish the baseline hash
func (b *Board) InitializeHashFromPosition(hashFunc func(*Board) uint64) {
	b.currentHash = hashFunc(b)
}

// InitializeEvalScoresFromPosition calculates initial evaluation scores from the current position.
// This should be called once after board setup (e.g., after FromFEN) to establish the baseline
// scores that will be maintained incrementally during subsequent moves.
func (b *Board) InitializeEvalScoresFromPosition() {
	b.materialScore = 0
	b.pstScore = 0

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece != Empty {
				b.materialScore += values.GetPieceValue(values.Piece(piece))
				b.pstScore += values.GetPositionalBonus(values.Piece(piece), rank, file)
			}
		}
	}
}

// HashUpdater interface for providing zobrist key updates
type HashUpdater interface {
	GetHashDelta(b *Board, move Move, oldState State) uint64
	GetNullMoveDelta() uint64 // Get hash delta for null move (just flip side-to-move)
}

// State captures the board state before a move for hash calculation
type State struct {
	CastlingRights  string
	EnPassantSquare Square
	HasEnPassant    bool
	SideToMove      string
}

// GetPawnHash returns a hash of only pawn positions for pawn structure caching
func (b *Board) GetPawnHash() uint64 {
	// Use bitboards for efficient pawn hash calculation
	whitePawns := b.GetPieceBitboard(WhitePawn)
	blackPawns := b.GetPieceBitboard(BlackPawn)

	// Simple but effective hash: combine white and black pawn bitboards
	// Multiply by different primes to distinguish white vs black pawns
	return uint64(whitePawns)*17 + uint64(blackPawns)*23
}

// SetHashUpdater sets the hash updater for incremental updates
func (b *Board) SetHashUpdater(updater HashUpdater) {
	b.hashUpdater = updater
}

// GetCurrentBoardState returns the current board state for hash calculations
func (b *Board) GetCurrentBoardState() State {
	return State{
		CastlingRights:  b.castlingRights,
		EnPassantSquare: b.enPassantSquare,
		HasEnPassant:    b.hasEnPassant,
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

// GetPiece returns the piece at the specified rank and file coordinates
func (b *Board) GetPiece(rank, file int) Piece {
	if rank < 0 || rank > 7 || file < 0 || file > 7 {
		return Empty
	}

	square := rank*8 + file
	return b.Mailbox[square]
}

// SetPiece places a piece at the specified rank and file coordinates
func (b *Board) SetPiece(rank, file int, piece Piece) {
	square := rank*8 + file

	// Get old piece from mailbox
	oldPiece := b.Mailbox[square]

	// Update incremental evaluation scores
	b.updateEvalScoresForPiece(rank, file, oldPiece, piece)

	// Update mailbox
	b.Mailbox[square] = piece

	// Remove old piece from bitboards
	if oldPiece != Empty {
		b.removePieceBitboard(rank, file, oldPiece)
	}

	// Add new piece to bitboards
	if piece != Empty {
		b.setPieceBitboard(rank, file, piece)
	}
}

// FromFEN creates a new board from FEN (Forsyth-Edwards Notation) string
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
				emptySquares, err := strconv.Atoi(string(char))
				if err != nil {
					return nil, fmt.Errorf("invalid FEN: failed to parse empty squares count: %w", err)
				}
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
				board.enPassantSquare = Square{File: file, Rank: rank}
				board.hasEnPassant = true
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

	// Initialize incremental evaluation scores
	board.InitializeEvalScoresFromPosition()

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
