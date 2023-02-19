package main

import (
	"github.com/AdamGriffiths31/ChessEngine/engine"
)

func main() {

	// engine.SetSquares()
	b := &engine.Board{}
	engine.ParseFEN(engine.StartFEN, b)
	engine.PrintBoard(b)
	// for index := 0; index < 120; index++ {
	// 	if index%10 == 0 {
	// 		fmt.Printf("\n")
	// 	}
	// 	fmt.Printf("%v", engine.Pieces[b.Pieces[index]])
	// }
	// fmt.Printf("\n")
	// fmt.Printf("\n")
	// fmt.Printf("\n")
	// fmt.Printf("side: %v", b.Side)

	// for index := 0; index < 64; index++ {
	// 	if index%8 == 0 {
	// 		fmt.Printf("\n")
	// 	}
	// 	fmt.Printf("%5d", engine.Sqaure64ToSquare120[index])
	// }

	// var bitboard uint64 = 0
	// engine.SetBit(&bitboard, 61)
	// engine.PrintBitboard(bitboard)
	// engine.ClearBit(&bitboard, 61)

	// engine.PrintBitboard(bitboard)

}
