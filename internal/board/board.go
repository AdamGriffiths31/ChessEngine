package board

import (
	"log/slog"

	"github.com/AdamGriffiths31/ChessEngine/internal/eval/values"
)

// Piece represents a chess piece using standard FEN notation
type Piece byte

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
	File int8 // 0-7 (a-h)
	Rank int8 // 0-7 (1-8)
}

// NewSquare builds a Square from file and rank indexes.
func NewSquare(file, rank int) Square {
	return Square{File: int8(file), Rank: int8(rank)}
}

// Index returns the square's 0-63 board index.
func (s Square) Index() int {
	return FileRankToSquare(int(s.File), int(s.Rank))
}

// EvalState stores incremental evaluation scores for undo operations.
type EvalState struct {
	MaterialScore int
	PSTScore      int
	PSTScoreEG    int
}

// Board represents a chess board with piece positions and game state
type Board struct {
	castlingRights  CastlingRights
	enPassantSquare Square
	hasEnPassant    bool
	halfMoveClock   int
	fullMoveNumber  int
	blackToMove     bool

	PieceBitboards [12]Bitboard

	Mailbox [64]Piece

	WhitePieces Bitboard
	BlackPieces Bitboard
	AllPieces   Bitboard

	currentHash uint64
	hashHistory []uint64
	hashUpdater HashUpdater

	materialScore int
	pstScore      int
	pstScoreEG    int
	evalHistory   []EvalState
}

// LightSquares and DarkSquares are the masks of each square color.
var (
	LightSquares Bitboard
	DarkSquares  Bitboard
)

func init() {
	for rank := range 8 {
		for file := range 8 {
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
		castlingRights: allCastlingRights,
		hasEnPassant:   false,
		halfMoveClock:  0,
		fullMoveNumber: 1,
	}

	for i := range 12 {
		board.PieceBitboards[i] = 0
	}
	board.WhitePieces = 0
	board.BlackPieces = 0
	board.AllPieces = 0

	for i := range 64 {
		board.Mailbox[i] = Empty
	}

	board.currentHash = 0
	board.hashHistory = make([]uint64, 0)

	board.materialScore = 0
	board.pstScore = 0
	board.pstScoreEG = 0
	board.evalHistory = make([]EvalState, 0)

	return board
}

// GetHash returns the current zobrist hash of the board
func (b *Board) GetHash() uint64 {
	return b.currentHash
}

// SetHash sets the current zobrist hash
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

// GetPSTScore returns the current incrementally maintained middlegame
// piece-square table score.
// This value is updated automatically by SetPiece and represents the sum of all
// positional bonuses from White's perspective.
func (b *Board) GetPSTScore() int {
	return b.pstScore
}

// GetPSTScoreEndgame returns the endgame PST score, maintained in parallel
// with the middlegame score.
func (b *Board) GetPSTScoreEndgame() int {
	return b.pstScoreEG
}

// PushEvalState saves current evaluation state to history for later restoration.
// This should be called before making a move to enable proper unmake operations.
func (b *Board) PushEvalState() {
	b.evalHistory = append(b.evalHistory, EvalState{
		MaterialScore: b.materialScore,
		PSTScore:      b.pstScore,
		PSTScoreEG:    b.pstScoreEG,
	})
}

// PopEvalState restores evaluation state from history during move unmake.
// This reverts the evaluation scores to their state before the last PushEvalState call.
func (b *Board) PopEvalState() {
	if len(b.evalHistory) > 0 {
		state := b.evalHistory[len(b.evalHistory)-1]
		b.materialScore = state.MaterialScore
		b.pstScore = state.PSTScore
		b.pstScoreEG = state.PSTScoreEG
		b.evalHistory = b.evalHistory[:len(b.evalHistory)-1]
	}
}

func (b *Board) updateEvalScoresForPiece(square int, oldPiece, newPiece Piece) {
	if index := int(pieceIndexes[oldPiece]); index != -1 {
		score := &pieceScores[index][square]
		b.materialScore -= int(score.material)
		b.pstScore -= int(score.pst)
		b.pstScoreEG -= int(score.pstEndgame)
	}

	if index := int(pieceIndexes[newPiece]); index != -1 {
		score := &pieceScores[index][square]
		b.materialScore += int(score.material)
		b.pstScore += int(score.pst)
		b.pstScoreEG += int(score.pstEndgame)
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
	b.pstScoreEG = 0

	for rank := range 8 {
		for file := range 8 {
			piece := b.GetPiece(rank, file)
			if piece != Empty {
				b.materialScore += values.GetPieceValue(values.Piece(piece))
				b.pstScore += values.GetPositionalBonus(values.Piece(piece), rank, file)
				b.pstScoreEG += values.GetPositionalBonusEndgame(values.Piece(piece), rank, file)
			}
		}
	}
}

// HashUpdater interface for providing zobrist key updates
type HashUpdater interface {
	GetHashDelta(b *Board, move Move, oldState State) uint64
	// GetNullMoveDelta returns the hash delta for a null move about to be
	// made from position b. Called BEFORE any state mutation, so the
	// implementer sees the pre-null side-to-move and a still-live en-passant
	// target: the delta must flip the side-to-move key AND XOR out the
	// en-passant file key when that target was actually hashed (i.e. a
	// capturing pawn existed), since MakeNullMove clears it.
	GetNullMoveDelta(b *Board) uint64
}

// State captures the board state before a move for hash calculation
type State struct {
	CastlingRights  CastlingRights
	EnPassantSquare Square
	HasEnPassant    bool
	BlackToMove     bool
}

// GetPawnHash returns a hash of only pawn positions for pawn structure caching
func (b *Board) GetPawnHash() uint64 {
	whitePawns := b.GetPieceBitboard(WhitePawn)
	blackPawns := b.GetPieceBitboard(BlackPawn)

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
		BlackToMove:     b.blackToMove,
	}
}

// pieceScore is a piece's signed material and PST contribution on one square.
type pieceScore struct {
	material, pst, pstEndgame int32
}

var (
	pieceIndexes = buildPieceIndexes()
	pieceScores  = buildPieceScores()
)

func buildPieceIndexes() [256]int8 {
	var indexes [256]int8
	for i := range indexes {
		indexes[i] = -1
	}
	indexes[WhitePawn] = WhitePawnIndex
	indexes[WhiteRook] = WhiteRookIndex
	indexes[WhiteKnight] = WhiteKnightIndex
	indexes[WhiteBishop] = WhiteBishopIndex
	indexes[WhiteQueen] = WhiteQueenIndex
	indexes[WhiteKing] = WhiteKingIndex
	indexes[BlackPawn] = BlackPawnIndex
	indexes[BlackRook] = BlackRookIndex
	indexes[BlackKnight] = BlackKnightIndex
	indexes[BlackBishop] = BlackBishopIndex
	indexes[BlackQueen] = BlackQueenIndex
	indexes[BlackKing] = BlackKingIndex
	return indexes
}

func buildPieceScores() [12][64]pieceScore {
	var scores [12][64]pieceScore
	for piece, index := range pieceIndexes {
		if index == -1 {
			continue
		}
		valuePiece := values.Piece(piece)
		for square := range 64 {
			file, rank := SquareToFileRank(square)
			scores[index][square] = pieceScore{
				material:   int32(values.GetPieceValue(valuePiece)),
				pst:        int32(values.GetPositionalBonus(valuePiece, rank, file)),
				pstEndgame: int32(values.GetPositionalBonusEndgame(valuePiece, rank, file)),
			}
		}
	}
	return scores
}

// PieceToBitboardIndex returns the bitboard index for a given piece, or -1
// when it is not a piece. It stays a switch so that calls with a constant
// piece compile to a direct index.
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
		return -1
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

func (b *Board) setPieceBitboard(square int, piece Piece) {
	index := int(pieceIndexes[piece])

	if index != -1 {
		squareBit := Bitboard(1) << square
		b.PieceBitboards[index] = b.PieceBitboards[index].SetBit(square)

		if IsWhitePiece(piece) {
			b.WhitePieces |= squareBit
		} else {
			b.BlackPieces |= squareBit
		}
		b.AllPieces |= squareBit
	}
}

func (b *Board) removePieceBitboard(square int, piece Piece) {
	index := int(pieceIndexes[piece])

	if index != -1 {
		squareBit := Bitboard(1) << square
		b.PieceBitboards[index] = b.PieceBitboards[index].ClearBit(square)

		if IsWhitePiece(piece) {
			b.WhitePieces &= ^squareBit
		} else {
			b.BlackPieces &= ^squareBit
		}
		b.AllPieces &= ^squareBit
	}
}

// PieceAt returns the piece on the given square.
func (b *Board) PieceAt(s Square) Piece {
	return b.GetPiece(int(s.Rank), int(s.File))
}

// SetPieceAt places a piece on the given square.
func (b *Board) SetPieceAt(s Square, piece Piece) {
	b.SetPiece(int(s.Rank), int(s.File), piece)
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

	oldPiece := b.Mailbox[square]

	b.updateEvalScoresForPiece(square, oldPiece, piece)

	b.placePiece(square, oldPiece, piece)
}

func (b *Board) placePiece(square int, oldPiece, piece Piece) {
	b.Mailbox[square] = piece

	if oldPiece != Empty {
		b.removePieceBitboard(square, oldPiece)
	}

	if piece != Empty {
		b.setPieceBitboard(square, piece)
	}
}

// restorePiece places a piece without updating the eval scores, for unmake,
// which restores them from history.
func (b *Board) restorePiece(s Square, piece Piece) {
	square := s.Index()
	b.placePiece(square, b.Mailbox[square], piece)
}

func (b *Board) getPieceCountFromBitboard(piece Piece) int {
	index := PieceToBitboardIndex(piece)
	if index == -1 {
		return 0
	}
	return b.PieceBitboards[index].PopCount()
}

// DebugPieceCounts logs comprehensive piece count information, flagging any
// mismatch between the scanned board and the tracked bitboard counts.
func (b *Board) DebugPieceCounts(label string) {
	actualCounts := make(map[Piece]int)
	for rank := range 8 {
		for file := range 8 {
			piece := b.GetPiece(rank, file)
			if piece != Empty {
				actualCounts[piece]++
			}
		}
	}

	pieces := []Piece{WhitePawn, WhiteRook, WhiteKnight, WhiteBishop, WhiteQueen, WhiteKing,
		BlackPawn, BlackRook, BlackKnight, BlackBishop, BlackQueen, BlackKing}

	for _, piece := range pieces {
		actual := actualCounts[piece]
		tracked := b.getPieceCountFromBitboard(piece)
		name := getPieceName(piece)

		if actual != tracked {
			slog.Warn("piece count mismatch", "label", label, "piece", name, "actual", actual, "tracked", tracked)
		} else {
			slog.Debug("piece count", "label", label, "piece", name, "count", actual)
		}
	}

	whiteActual := actualCounts[WhitePawn] + actualCounts[WhiteRook] + actualCounts[WhiteKnight] +
		actualCounts[WhiteBishop] + actualCounts[WhiteQueen] + actualCounts[WhiteKing]
	blackActual := actualCounts[BlackPawn] + actualCounts[BlackRook] + actualCounts[BlackKnight] +
		actualCounts[BlackBishop] + actualCounts[BlackQueen] + actualCounts[BlackKing]

	slog.Debug("piece count totals", "label", label, "white", whiteActual, "black", blackActual)
}

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
