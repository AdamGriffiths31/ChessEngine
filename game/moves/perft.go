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
	Depth int   `json:"depth"`
	Nodes int64 `json:"nodes"`
}


// Perft calculates the number of possible moves at a given depth
func Perft(b *board.Board, depth int, player Player) int64 {
	generator := NewGenerator()
	return PerftWithGenerator(b, depth, player, generator)
}

// PerftWithGenerator calculates the number of possible moves at a given depth using a provided generator
func PerftWithGenerator(b *board.Board, depth int, player Player, generator *Generator) int64 {
	if depth == 0 {
		return 1
	}

	// Performance optimization: if depth is 1, just return move count
	if depth == 1 {
		moves := generator.GenerateAllMoves(b, player)
		defer ReleaseMoveList(moves)
		return int64(moves.Count)
	}

	moves := generator.GenerateAllMoves(b, player)
	defer ReleaseMoveList(moves)

	var nodeCount int64
	
	for _, move := range moves.Moves {
		// Make the move and store history using the generator
		history := generator.makeMove(b, move)

		// Recursively calculate nodes for next player
		nextPlayer := White
		if player == White {
			nextPlayer = Black
		}

		nodeCount += PerftWithGenerator(b, depth-1, nextPlayer, generator)

		// Undo the move using the generator
		generator.unmakeMove(b, history)
	}

	return nodeCount
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

