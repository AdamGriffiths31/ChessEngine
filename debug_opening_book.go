package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/AdamGriffiths31/ChessEngine/board"
	"github.com/AdamGriffiths31/ChessEngine/game/openings"
)

func main() {
	fmt.Println("=== Chess Engine Opening Book Debug Tool ===\n")

	// Test 1: Initial position hash
	fmt.Println("Test 1: Initial Position Hash")
	fmt.Println("------------------------------")
	initialBoard, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		fmt.Printf("Error creating initial board: %v\n", err)
		return
	}

	zobrist := openings.GetPolyglotHash()
	initialHash := zobrist.HashPosition(initialBoard)
	fmt.Printf("Initial position hash: 0x%016X\n", initialHash)
	fmt.Printf("Initial FEN: %s\n\n", initialBoard.ToFEN())

	// Test 2: Position after e2e4
	fmt.Println("Test 2: Position After 1.e4")
	fmt.Println("----------------------------")
	e4Board, err := board.FromFEN("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")
	if err != nil {
		fmt.Printf("Error creating e4 board: %v\n", err)
		return
	}

	e4Hash := zobrist.HashPosition(e4Board)
	fmt.Printf("After 1.e4 hash: 0x%016X\n", e4Hash)
	fmt.Printf("After 1.e4 FEN: %s\n\n", e4Board.ToFEN())

	// Test 3: Check what happens when we make the move
	fmt.Println("Test 3: Making Move e2e4")
	fmt.Println("------------------------")
	testBoard := board.NewBoard()
	fmt.Printf("Before move FEN: %s\n", testBoard.ToFEN())

	e2e4Move, err := board.ParseSimpleMove("e2e4")
	if err != nil {
		fmt.Printf("Error parsing e2e4: %v\n", err)
		return
	}

	err = testBoard.MakeMove(e2e4Move)
	if err != nil {
		fmt.Printf("Error making move: %v\n", err)
		return
	}

	afterMoveHash := zobrist.HashPosition(testBoard)
	fmt.Printf("After move FEN: %s\n", testBoard.ToFEN())
	fmt.Printf("After move hash: 0x%016X\n", afterMoveHash)
	fmt.Printf("Hash matches expected?: %v\n\n", afterMoveHash == e4Hash)

	// Test 4: Read first few entries from opening book
	fmt.Println("Test 4: Opening Book First Entries")
	fmt.Println("----------------------------------")
	bookPath := "game/openings/testdata/performance.bin"

	file, err := os.Open(bookPath)
	if err != nil {
		fmt.Printf("Error opening book: %v\n", err)
		return
	}
	defer file.Close()

	// Read first 10 entries
	for i := 0; i < 10; i++ {
		var entry struct {
			Hash   uint64
			Move   uint16
			Weight uint16
			Learn  uint32
		}

		err := binary.Read(file, binary.BigEndian, &entry)
		if err != nil {
			break
		}

		fmt.Printf("Entry %d: Hash=0x%016X Move=0x%04X Weight=%d\n",
			i+1, entry.Hash, entry.Move, entry.Weight)
	}

	fmt.Println("\nTest 5: Search for Initial Position in Book")
	fmt.Println("------------------------------------------")

	// Load the book properly
	book := openings.NewPolyglotBook()
	err = book.LoadFromFile(bookPath)
	if err != nil {
		fmt.Printf("Error loading book: %v\n", err)
		return
	}

	bookInfo := book.GetBookInfo()
	fmt.Printf("Book loaded: %d entries\n", bookInfo.EntryCount)

	// Try to find moves for initial position
	moves, err := book.LookupMove(initialHash, initialBoard)
	if err != nil {
		fmt.Printf("Initial position not found: %v\n", err)
		// Try some known hashes from the book
		fmt.Println("\nSearching for any valid position...")
		testHashes := []uint64{
			0x463b96181691fc9c, // Known Polyglot initial position hash
			0x823c9b50fd114196, // Another common hash
		}
		for _, testHash := range testHashes {
			moves, err := book.LookupMove(testHash, initialBoard)
			if err == nil {
				fmt.Printf("Found position with hash 0x%016X: %d moves\n", testHash, len(moves))
			}
		}
	} else {
		fmt.Printf("Found %d moves for initial position\n", len(moves))
		for i, move := range moves {
			fmt.Printf("  Move %d: Weight=%d\n", i+1, move.Weight)
		}
	}

	// Test 6: Debug hash calculation details
	fmt.Println("\nTest 6: Hash Calculation Details")
	fmt.Println("--------------------------------")
	debugHashCalculation(zobrist, initialBoard, "Initial Position")
	debugHashCalculation(zobrist, e4Board, "After 1.e4")
}

func debugHashCalculation(zobrist *openings.ZobristHash, b *board.Board, desc string) {
	fmt.Printf("\n%s:\n", desc)
	fmt.Printf("Side to move: %s\n", b.GetSideToMove())
	fmt.Printf("Castling rights: %s\n", b.GetCastlingRights())

	ep := b.GetEnPassantTarget()
	if ep != nil {
		fmt.Printf("En passant: %s\n", ep.String())
	} else {
		fmt.Printf("En passant: none\n")
	}

	// Show pieces
	fmt.Println("Pieces:")
	for rank := 7; rank >= 0; rank-- {
		for file := 0; file < 8; file++ {
			piece := b.GetPiece(rank, file)
			if piece != board.Empty {
				square := board.Square{File: file, Rank: rank}
				fmt.Printf("  %s: %c\n", square.String(), piece)
			}
		}
	}
}
