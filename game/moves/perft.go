package moves

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// PerftTestData represents the structure of the JSON test data
type PerftTestData struct {
	Positions []PerftPosition `json:"positions"`
}

// PerftPosition represents a single test position
type PerftPosition struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	FEN         string       `json:"fen"`
	Depths      []PerftDepth `json:"depths"`
}

// PerftDepth represents expected node count and detailed statistics at a specific depth
type PerftDepth struct {
	Depth      int   `json:"depth"`
	Nodes      int64 `json:"nodes"`
	Captures   int64 `json:"captures,omitempty"`
	EnPassant  int64 `json:"en_passant,omitempty"`
	Castles    int64 `json:"castles,omitempty"`
	Promotions int64 `json:"promotions,omitempty"`
}

// PerftStats holds detailed statistics about a perft run
type PerftStats struct {
	Nodes     int64
	Captures  int64
	Castles   int64
	EnPassant int64
	Promotions int64
}

// Perft calculates the number of possible moves at a given depth
func Perft(b *board.Board, depth int, player Player) int64 {
	if depth == 0 {
		return 1
	}

	// Performance optimization: if depth is 1, just return move count
	if depth == 1 {
		generator := NewGenerator()
		moves := generator.GenerateAllMoves(b, player)
		return int64(moves.Count)
	}

	generator := NewGenerator()
	moves := generator.GenerateAllMoves(b, player)

	var nodeCount int64

	for _, move := range moves.Moves {
		// Save board state
		gameState := saveGameState(b)
		
		// Make the move
		makePerftMove(b, move)

		// Recursively calculate nodes for next player
		nextPlayer := White
		if player == White {
			nextPlayer = Black
		}

		nodeCount += Perft(b, depth-1, nextPlayer)

		// Restore board state
		restoreGameState(b, gameState)
	}

	return nodeCount
}

// PerftWithStats calculates perft with detailed statistics
func PerftWithStats(b *board.Board, depth int, player Player) PerftStats {
	stats := PerftStats{}
	
	if depth == 0 {
		stats.Nodes = 1
		return stats
	}

	generator := NewGenerator()
	moves := generator.GenerateAllMoves(b, player)

	for _, move := range moves.Moves {
		// Save board state
		gameState := saveGameState(b)
		
		// Make the move
		makePerftMove(b, move)

		// Recursively calculate nodes for next player
		nextPlayer := White
		if player == White {
			nextPlayer = Black
		}

		if depth == 1 {
			// At depth 1, count the move types
			stats.Nodes++
			if move.IsCapture {
				stats.Captures++
			}
			if move.IsCastling {
				stats.Castles++
			}
			if move.IsEnPassant {
				stats.EnPassant++
			}
			if move.Promotion != board.Empty {
				stats.Promotions++
			}
		} else {
			// Recursive case
			subStats := PerftWithStats(b, depth-1, nextPlayer)
			stats.Nodes += subStats.Nodes
			stats.Captures += subStats.Captures
			stats.Castles += subStats.Castles
			stats.EnPassant += subStats.EnPassant
			stats.Promotions += subStats.Promotions
		}

		// Restore board state
		restoreGameState(b, gameState)
	}

	return stats
}

// GameState stores complete board state for perft make/unmake
type GameState struct {
	squares         [8][8]board.Piece
	castlingRights  string
	enPassantTarget *board.Square
	halfMoveClock   int
	fullMoveNumber  int
	sideToMove      string
}

// saveGameState saves the current board state
func saveGameState(b *board.Board) GameState {
	state := GameState{
		castlingRights:  b.GetCastlingRights(),
		halfMoveClock:   b.GetHalfMoveClock(),
		fullMoveNumber:  b.GetFullMoveNumber(),
		sideToMove:      b.GetSideToMove(),
	}
	
	// Copy en passant target
	if target := b.GetEnPassantTarget(); target != nil {
		state.enPassantTarget = &board.Square{File: target.File, Rank: target.Rank}
	}
	
	// Copy board squares
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			state.squares[rank][file] = b.GetPiece(rank, file)
		}
	}
	
	return state
}

// restoreGameState restores the board state
func restoreGameState(b *board.Board, state GameState) {
	b.SetCastlingRights(state.castlingRights)
	b.SetEnPassantTarget(state.enPassantTarget)
	b.SetHalfMoveClock(state.halfMoveClock)
	b.SetFullMoveNumber(state.fullMoveNumber)
	b.SetSideToMove(state.sideToMove)
	
	// Restore board squares
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			b.SetPiece(rank, file, state.squares[rank][file])
		}
	}
}

// makePerftMove makes a move on the board for perft testing
func makePerftMove(b *board.Board, move board.Move) {
	// Handle en passant capture
	if move.IsEnPassant {
		// Remove the captured pawn
		captureRank := move.From.Rank
		b.SetPiece(captureRank, move.To.File, board.Empty)
	}
	
	// Handle castling
	if move.IsCastling {
		// Move the rook
		var rookFrom, rookTo board.Square
		if move.To.File == 6 { // Kingside
			rookFrom = board.Square{File: 7, Rank: move.From.Rank}
			rookTo = board.Square{File: 5, Rank: move.From.Rank}
		} else { // Queenside
			rookFrom = board.Square{File: 0, Rank: move.From.Rank}
			rookTo = board.Square{File: 3, Rank: move.From.Rank}
		}
		rook := b.GetPiece(rookFrom.Rank, rookFrom.File)
		b.SetPiece(rookFrom.Rank, rookFrom.File, board.Empty)
		b.SetPiece(rookTo.Rank, rookTo.File, rook)
	}
	
	// Move the piece
	piece := b.GetPiece(move.From.Rank, move.From.File)
	b.SetPiece(move.From.Rank, move.From.File, board.Empty)
	
	// Handle promotion
	if move.Promotion != board.Empty {
		b.SetPiece(move.To.Rank, move.To.File, move.Promotion)
	} else {
		b.SetPiece(move.To.Rank, move.To.File, piece)
	}
	
	// Update game state (simplified for perft)
	// Clear en passant target (will be set by pawn moves if needed)
	b.SetEnPassantTarget(nil)
	
	// Set en passant target for pawn two-square moves
	if piece == board.WhitePawn || piece == board.BlackPawn {
		// Check if it's a two-square pawn move
		if abs(move.To.Rank - move.From.Rank) == 2 {
			// Set en passant target square (the square the pawn passed over)
			var enPassantRank int
			if piece == board.WhitePawn {
				enPassantRank = move.From.Rank + 1
			} else {
				enPassantRank = move.From.Rank - 1
			}
			enPassantTarget := &board.Square{File: move.From.File, Rank: enPassantRank}
			b.SetEnPassantTarget(enPassantTarget)
		}
	}
	
	// Update castling rights if king or rook moved
	updateCastlingRights(b, move)
}

// updateCastlingRights updates castling rights based on move
func updateCastlingRights(b *board.Board, move board.Move) {
	rights := b.GetCastlingRights()
	newRights := ""
	
	for _, r := range rights {
		keepRight := true
		
		// Remove castling rights if king moves
		if move.From.File == 4 && move.From.Rank == 0 && (r == 'K' || r == 'Q') {
			keepRight = false // White king moved
		}
		if move.From.File == 4 && move.From.Rank == 7 && (r == 'k' || r == 'q') {
			keepRight = false // Black king moved
		}
		
		// Remove castling rights if rook moves
		if move.From.File == 0 && move.From.Rank == 0 && r == 'Q' {
			keepRight = false // White queenside rook moved
		}
		if move.From.File == 7 && move.From.Rank == 0 && r == 'K' {
			keepRight = false // White kingside rook moved
		}
		if move.From.File == 0 && move.From.Rank == 7 && r == 'q' {
			keepRight = false // Black queenside rook moved
		}
		if move.From.File == 7 && move.From.Rank == 7 && r == 'k' {
			keepRight = false // Black kingside rook moved
		}
		
		// Remove castling rights if rook is captured
		if move.To.File == 0 && move.To.Rank == 0 && r == 'Q' {
			keepRight = false // White queenside rook captured
		}
		if move.To.File == 7 && move.To.Rank == 0 && r == 'K' {
			keepRight = false // White kingside rook captured
		}
		if move.To.File == 0 && move.To.Rank == 7 && r == 'q' {
			keepRight = false // Black queenside rook captured
		}
		if move.To.File == 7 && move.To.Rank == 7 && r == 'k' {
			keepRight = false // Black kingside rook captured
		}
		
		if keepRight {
			newRights += string(r)
		}
	}
	
	if newRights == "" {
		newRights = "-"
	}
	
	b.SetCastlingRights(newRights)
}

// LoadPerftTestData loads test data from JSON file
func LoadPerftTestData(filePath string) (*PerftTestData, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var testData PerftTestData
	err = json.Unmarshal(data, &testData)
	if err != nil {
		return nil, err
	}

	return &testData, nil
}

// GetTestDataPath returns the path to the test data file
func GetTestDataPath() string {
	return filepath.Join("testdata", "perft_tests.json")
}

