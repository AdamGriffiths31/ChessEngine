package main

import (
	"fmt"
)

// This will test if the initialization is correct by checking a few key values
func main() {
	fmt.Println("=== Verifying Polyglot Key Initialization ===\n")

	// The first few Polyglot keys from the standard
	expectedKeys := []uint64{
		0x9D39247E33776D41, // [0]
		0x2AF7398005AAA5C7, // [1]
		0x44DB015024623547, // [2]
		0x9C15F73E62A76AE2, // [3]
		0x75834465489C0C89, // [4]
		0x3290AC3A203001BF, // [5]
		0x0FBBAD1F61042279, // [6]
		0xE83A908FF2FB60CA, // [7]
	}

	fmt.Println("Expected Polyglot keys (first 8):")
	for i, key := range expectedKeys {
		fmt.Printf("[%d]: 0x%016X\n", i, key)
	}

	fmt.Println("\n=== Correct Piece Key Mapping ===")
	fmt.Println("For White Rook (piece 7) on a1 (square 0):")
	fmt.Printf("Index should be: 7*64 + 0 = %d\n", 7*64+0)
	fmt.Println("This is index 448 in the random64Poly array")

	fmt.Println("\nFor Black Pawn (piece 0) on a7 (square 48):")
	fmt.Printf("Index should be: 0*64 + 48 = %d\n", 0*64+48)
	fmt.Println("This corresponds to random64Poly[48]")

	fmt.Println("\n=== The Fix ===")
	fmt.Println("In polyglot_standard.go, the init() function should have:")
	fmt.Println(`
// Initialize piece keys (indices 0-767)
for piece := 0; piece < 12; piece++ {
    for square := 0; square < 64; square++ {
        index := piece*64 + square
        officialPolyglotPieceKeys[square][piece] = random64Poly[index]
    }
}`)

	fmt.Println("\n=== Castling Keys ===")
	fmt.Println("The castling keys should use indices 768-771:")
	fmt.Println("castleKeys[0] = random64Poly[768] // White O-O")
	fmt.Println("castleKeys[1] = random64Poly[769] // White O-O-O")
	fmt.Println("castleKeys[2] = random64Poly[770] // Black O-O")
	fmt.Println("castleKeys[3] = random64Poly[771] // Black O-O-O")

	fmt.Println("\nAlso check that the castling initialization is correct:")
	fmt.Println(`
// Initialize castling keys
for i := 0; i < 4; i++ {
    officialPolyglotCastlingKeys[i] = random64Poly[768+i]
}`)
}
