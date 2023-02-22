package engine

import (
	"fmt"
)

func CheckBoard(pos *Board) {
	//TODO Need a config to turn this on off (perf is hit when its on and also not need when not deving)

	var pieceNumber = [13]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	var bigPiece = [2]int{0, 0}
	var majorPiece = [2]int{0, 0}
	var minPiece = [2]int{0, 0}
	var material = [2]int{0, 0}
	var pawns = [3]uint64{0, 0}

	pawns[White] = pos.Pawns[White]
	pawns[Black] = pos.Pawns[Black]
	pawns[Both] = pos.Pawns[Both]

	for piece := WP; piece <= BK; piece++ {
		for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
			sq120 := pos.PieceList[piece][pieceNum]
			if pos.Pieces[sq120] != piece {
				panic(fmt.Errorf("CheckBoard: %v does not equal %v", pos.Pieces[sq120], piece))
			}
		}
	}

	for sq64 := 0; sq64 < 64; sq64++ {
		sq120 := Sqaure64ToSquare120[sq64]
		piece := pos.Pieces[sq120]
		pieceNumber[piece]++
		color := PieceCol[piece]
		if PieceBig[piece] == True {
			bigPiece[color]++
		}
		if PieceMajor[piece] == True {
			majorPiece[color]++
		}
		if PieceMin[piece] == True {
			minPiece[color]++
		}
		if PieceVal[piece] > 0 {
			material[color] += PieceVal[piece]
		}
	}

	for piece := WP; piece <= BK; piece++ {
		if pieceNumber[piece] != pos.PieceNumber[piece] {
			panic(fmt.Errorf("CheckBoard: piece %v [%v] does not match count %v - was %v", piece, Pieces[piece], pos.PieceNumber[piece], pieceNumber[piece]))
		}
	}

	pawnCountWhite := CountBits(pawns[White])
	if pawnCountWhite != pos.PieceNumber[WP] {
		panic(fmt.Errorf("CheckBoard: white pawn error wanted %v but got %v", pos.PieceNumber[WP], pawnCountWhite))
	}

	pawnCountBlack := CountBits(pawns[Black])
	if pawnCountBlack != pos.PieceNumber[BP] {
		panic(fmt.Errorf("CheckBoard: black pawn error wanted %v but got %v", pos.PieceNumber[BP], pawnCountBlack))
	}

	pawnCountBoth := CountBits(pawns[Both])
	if pawnCountBoth != pawnCountBlack+pawnCountWhite {
		panic(fmt.Errorf("CheckBoard: both pawn error wanted %v but got %v", pawnCountBlack+pawnCountWhite, pawnCountBoth))
	}

	for i := uint64(0); i < 64; i++ {
		if (pawns[White] & (1 << i)) != 0 {
			sq64 := PopBit(&pawns[White])
			if pos.Pieces[Sqaure64ToSquare120[sq64]] != WP {
				panic(fmt.Errorf("CheckBoard: PopBit white  wanted %v but got %v", WP, pos.Pieces[Sqaure64ToSquare120[sq64]]))
			}
		}
		if (pawns[Black] & (1 << i)) != 0 {
			sq64 := PopBit(&pawns[Black])
			if pos.Pieces[Sqaure64ToSquare120[sq64]] != BP {
				panic(fmt.Errorf("CheckBoard: PopBit black  wanted %v but got %v", BP, pos.Pieces[Sqaure64ToSquare120[sq64]]))
			}
		}
		if (pawns[Black] & (1 << i)) != 0 {
			sq64 := PopBit(&pawns[Both])
			if pos.Pieces[Sqaure64ToSquare120[sq64]] != BP || pos.Pieces[Sqaure120ToSquare64[sq64]] != WP {
				panic(fmt.Errorf("CheckBoard: PopBit both wanted %v but got %v", BP, pos.Pieces[Sqaure64ToSquare120[sq64]]))
			}
		}
	}

	if material[White] != pos.Material[White] {
		panic(fmt.Errorf("CheckBoard: White material was %v but wanted %v ", material[White], pos.Material[White]))
	}

	if material[Black] != pos.Material[Black] {
		panic(fmt.Errorf("CheckBoard: Black material was %v but wanted %v ", material[Black], pos.Material[Black]))
	}

	if majorPiece[White] != pos.MajorPiece[White] {
		panic(fmt.Errorf("CheckBoard: White major Piece was %v but wanted %v ", majorPiece[White], pos.MajorPiece[White]))
	}

	if majorPiece[Black] != pos.MajorPiece[Black] {
		panic(fmt.Errorf("CheckBoard: Black major Piece was %v but wanted %v ", majorPiece[Black], pos.MajorPiece[Black]))
	}

	if minPiece[White] != pos.MinPiece[White] {
		panic(fmt.Errorf("CheckBoard: White min Piece was %v but wanted %v ", minPiece[White], pos.MinPiece[White]))
	}

	if minPiece[Black] != pos.MinPiece[Black] {
		PrintBoard(pos)
		panic(fmt.Errorf("CheckBoard: Black min Piece was %v but wanted %v ", minPiece[Black], pos.MinPiece[Black]))
	}

	if bigPiece[White] != pos.BigPiece[White] {
		panic(fmt.Errorf("CheckBoard: White big Piece was %v but wanted %v ", bigPiece[White], pos.BigPiece[White]))
	}

	if bigPiece[Black] != pos.BigPiece[Black] {
		panic(fmt.Errorf("CheckBoard: Black big Piece was %v but wanted %v ", bigPiece[Black], pos.BigPiece[Black]))
	}

	if pos.Side != White && pos.Side != Black {
		panic(fmt.Errorf("CheckBoard: side was %v", pos.Side))
	}

	if pos.PosistionKey != GeneratePositionKey(pos) {
		panic(fmt.Errorf("CheckBoard: PosistionKey: %v did not match %v", pos.PosistionKey, GeneratePositionKey(pos)))
	}

	if pos.EnPas != noSqaure {
		if pos.Side == White && RanksBoard[pos.EnPas] != Rank6 {
			panic(fmt.Errorf("CheckBoard: white EnPas error: %v was not on rank %v", pos.EnPas, Rank6))
		}

		if pos.Side == Black && RanksBoard[pos.EnPas] != Rank3 {
			panic(fmt.Errorf("CheckBoard: black EnPas error: %v was not on rank %v", pos.EnPas, Rank3))
		}
	}

	if pos.Pieces[pos.KingSqaure[White]] != WK {
		panic(fmt.Errorf("CheckBoard: White king was not found at %v instead %v was found", pos.KingSqaure[White], pos.Pieces[pos.KingSqaure[White]]))
	}

	if pos.Pieces[pos.KingSqaure[Black]] != BK {
		panic(fmt.Errorf("CheckBoard: Black king was not found at %v instead %v was found ", SqaureString(pos.KingSqaure[Black]), pos.Pieces[pos.KingSqaure[Black]]))
	}
}

// UpdateListMaterial sets the Material Lists
func UpdateListMaterial(pos *Board) {
	for i := 0; i < 120; i++ {
		sq := i
		piece := pos.Pieces[i]
		if piece == OffBoard || piece == Empty {
			continue
		}

		colour := PieceCol[piece]
		pos.Material[colour] += PieceVal[piece]
		pos.PieceList[piece][pos.PieceNumber[piece]] = sq
		pos.PieceNumber[piece]++

		if piece == WK {
			pos.KingSqaure[White] = sq
		}

		if piece == BK {
			pos.KingSqaure[Black] = sq
		}

		if PieceBig[piece] == True {
			pos.BigPiece[colour]++
		}

		if PieceMajor[piece] == True {
			pos.MajorPiece[colour]++
		}
		if PieceMin[piece] == True {
			pos.MinPiece[colour]++
		}

		if piece == WP {
			SetBit(&pos.Pawns[White], Sqaure120ToSquare64[sq])
			SetBit(&pos.Pawns[Both], Sqaure120ToSquare64[sq])
		}

		if piece == BP {
			SetBit(&pos.Pawns[Black], Sqaure120ToSquare64[sq])
			SetBit(&pos.Pawns[Both], Sqaure120ToSquare64[sq])
		}
	}
}

func PrintBoard(pos *Board) {
	println("Printing board...")
	for rank := Rank8; rank >= Rank1; rank-- {
		fmt.Printf("%v ", rank+1)
		for file := FileA; file <= FileH; file++ {
			sq := FileRankToSquare(file, rank)
			piece := pos.Pieces[sq]
			fmt.Printf("%3v", PceChar[piece])
		}
		fmt.Print("\n")

	}
	fmt.Print("  ")
	for file := FileA; file <= FileH; file++ {
		fmt.Printf("%3c", 'a'+file)
	}
	fmt.Print("\n")
	fmt.Printf("Side:%v\n", SideChar[pos.Side])
	fmt.Printf("EnPas:%v\n", SqaureString(pos.EnPas))
	fmt.Printf("PosKey:%11X (%v)\n", pos.PosistionKey, pos.PosistionKey)
}

// resetBoard restores Board to a default state
func resetBoard(pos *Board) {
	for i := 0; i < 120; i++ {
		pos.Pieces[i] = OffBoard
	}

	for i := 0; i < 64; i++ {
		pos.Pieces[Sqaure64ToSquare120[i]] = Empty
	}

	for i := 0; i < 2; i++ {
		pos.BigPiece[i] = 0
		pos.MajorPiece[i] = 0
		pos.MinPiece[i] = 0
		pos.Material[i] = 0
	}

	for i := 0; i < 3; i++ {
		pos.Pawns[i] = 0
	}

	for i := 0; i < 13; i++ {
		pos.PieceNumber[i] = 0
	}

	pos.KingSqaure[White], pos.KingSqaure[Black] = noSqaure, noSqaure

	pos.Side = Both
	pos.EnPas = noSqaure
	pos.FiftyMove = 0

	pos.Play = 0
	pos.HistoryPlay = 0

	pos.CastlePermission = 0

	pos.PosistionKey = 0
}

// getPieceType returns returns the corresponding piece type integer
func getPieceType(ch rune) int {
	pieceMap := map[rune]int{
		'p': BP,
		'r': BR,
		'n': BN,
		'b': BB,
		'q': BQ,
		'k': BK,
		'P': WP,
		'R': WR,
		'N': WN,
		'B': WB,
		'Q': WQ,
		'K': WK,
	}

	piece, ok := pieceMap[ch]
	if !ok {
		panic(fmt.Errorf("getPieceType: could not find value for %v", ch))
	}

	return piece
}
