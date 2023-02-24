package board

import (
	"fmt"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
)

func CheckBoard(pos *data.Board) {
	//TODO Need a config to turn this on off (perf is hit when its on and also not need when not deving)
	return
	var pieceNumber = [13]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	var bigPiece = [2]int{0, 0}
	var majorPiece = [2]int{0, 0}
	var minPiece = [2]int{0, 0}
	var material = [2]int{0, 0}
	var pawns = [3]uint64{0, 0}

	pawns[data.White] = pos.Pawns[data.White]
	pawns[data.Black] = pos.Pawns[data.Black]
	pawns[data.Both] = pos.Pawns[data.Both]

	for piece := data.WP; piece <= data.BK; piece++ {
		for pieceNum := 0; pieceNum < pos.PieceNumber[piece]; pieceNum++ {
			sq120 := pos.PieceList[piece][pieceNum]
			if pos.Pieces[sq120] != piece {
				panic(fmt.Errorf("CheckBoard: %v does not equal %v", pos.Pieces[sq120], piece))
			}
		}
	}

	for sq64 := 0; sq64 < 64; sq64++ {
		sq120 := data.Sqaure64ToSquare120[sq64]
		piece := pos.Pieces[sq120]
		pieceNumber[piece]++
		color := data.PieceCol[piece]
		if data.PieceBig[piece] == data.True {
			bigPiece[color]++
		}
		if data.PieceMajor[piece] == data.True {
			majorPiece[color]++
		}
		if data.PieceMin[piece] == data.True {
			minPiece[color]++
		}
		if data.PieceVal[piece] > 0 {
			material[color] += data.PieceVal[piece]
		}
	}

	for piece := data.WP; piece <= data.BK; piece++ {
		if pieceNumber[piece] != pos.PieceNumber[piece] {
			panic(fmt.Errorf("CheckBoard: piece %v [%v] does not match count %v - was %v", piece, data.Pieces[piece], pos.PieceNumber[piece], pieceNumber[piece]))
		}
	}

	pawnCountWhite := CountBits(pawns[data.White])
	if pawnCountWhite != pos.PieceNumber[data.WP] {
		panic(fmt.Errorf("CheckBoard: white pawn error wanted %v but got %v", pos.PieceNumber[data.WP], pawnCountWhite))
	}

	pawnCountBlack := CountBits(pawns[data.Black])
	if pawnCountBlack != pos.PieceNumber[data.BP] {
		panic(fmt.Errorf("CheckBoard: black pawn error wanted %v but got %v", pos.PieceNumber[data.BP], pawnCountBlack))
	}

	pawnCountBoth := CountBits(pawns[data.Both])
	if pawnCountBoth != pawnCountBlack+pawnCountWhite {
		panic(fmt.Errorf("CheckBoard: both pawn error wanted %v but got %v", pawnCountBlack+pawnCountWhite, pawnCountBoth))
	}

	for i := uint64(0); i < 64; i++ {
		if (pawns[data.White] & (1 << i)) != 0 {
			sq64 := PopBit(&pawns[data.White])
			if pos.Pieces[data.Sqaure64ToSquare120[sq64]] != data.WP {
				panic(fmt.Errorf("CheckBoard: PopBit white  wanted %v but got %v", data.WP, pos.Pieces[data.Sqaure64ToSquare120[sq64]]))
			}
		}
		if (pawns[data.Black] & (1 << i)) != 0 {
			sq64 := PopBit(&pawns[data.Black])
			if pos.Pieces[data.Sqaure64ToSquare120[sq64]] != data.BP {
				panic(fmt.Errorf("CheckBoard: PopBit black  wanted %v but got %v", data.BP, pos.Pieces[data.Sqaure64ToSquare120[sq64]]))
			}
		}
		if (pawns[data.Black] & (1 << i)) != 0 {
			sq64 := PopBit(&pawns[data.Both])
			if pos.Pieces[data.Sqaure64ToSquare120[sq64]] != data.BP || pos.Pieces[data.Sqaure120ToSquare64[sq64]] != data.WP {
				panic(fmt.Errorf("CheckBoard: PopBit both wanted %v but got %v", data.BP, pos.Pieces[data.Sqaure64ToSquare120[sq64]]))
			}
		}
	}

	if material[data.White] != pos.Material[data.White] {
		panic(fmt.Errorf("CheckBoard: White material was %v but wanted %v ", material[data.White], pos.Material[data.White]))
	}

	if material[data.Black] != pos.Material[data.Black] {
		panic(fmt.Errorf("CheckBoard: Black material was %v but wanted %v ", material[data.Black], pos.Material[data.Black]))
	}

	if majorPiece[data.White] != pos.MajorPiece[data.White] {
		panic(fmt.Errorf("CheckBoard: White major Piece was %v but wanted %v ", majorPiece[data.White], pos.MajorPiece[data.White]))
	}

	if majorPiece[data.Black] != pos.MajorPiece[data.Black] {
		panic(fmt.Errorf("CheckBoard: Black major Piece was %v but wanted %v ", majorPiece[data.Black], pos.MajorPiece[data.Black]))
	}

	if minPiece[data.White] != pos.MinPiece[data.White] {
		panic(fmt.Errorf("CheckBoard: White min Piece was %v but wanted %v ", minPiece[data.White], pos.MinPiece[data.White]))
	}

	if minPiece[data.Black] != pos.MinPiece[data.Black] {
		io.PrintBoard(pos)
		panic(fmt.Errorf("CheckBoard: Black min Piece was %v but wanted %v ", minPiece[data.Black], pos.MinPiece[data.Black]))
	}

	if bigPiece[data.White] != pos.BigPiece[data.White] {
		panic(fmt.Errorf("CheckBoard: White big Piece was %v but wanted %v ", bigPiece[data.White], pos.BigPiece[data.White]))
	}

	if bigPiece[data.Black] != pos.BigPiece[data.Black] {
		panic(fmt.Errorf("CheckBoard: Black big Piece was %v but wanted %v ", bigPiece[data.Black], pos.BigPiece[data.Black]))
	}

	if pos.Side != data.White && pos.Side != data.Black {
		panic(fmt.Errorf("CheckBoard: side was %v", pos.Side))
	}

	if pos.PosistionKey != GeneratePositionKey(pos) {
		panic(fmt.Errorf("CheckBoard: PosistionKey: %v did not match %v", pos.PosistionKey, GeneratePositionKey(pos)))
	}

	if pos.EnPas != data.NoSqaure {
		if pos.Side == data.White && data.RanksBoard[pos.EnPas] != data.Rank6 {
			panic(fmt.Errorf("CheckBoard: white EnPas error: %v was not on rank %v", pos.EnPas, data.Rank6))
		}

		if pos.Side == data.Black && data.RanksBoard[pos.EnPas] != data.Rank3 {
			panic(fmt.Errorf("CheckBoard: black EnPas error: %v was not on rank %v", pos.EnPas, data.Rank3))
		}
	}

	if pos.Pieces[pos.KingSqaure[data.White]] != data.WK {
		panic(fmt.Errorf("CheckBoard: White king was not found at %v instead %v was found", pos.KingSqaure[data.White], pos.Pieces[pos.KingSqaure[data.White]]))
	}

	if pos.Pieces[pos.KingSqaure[data.Black]] != data.BK {
		panic(fmt.Errorf("CheckBoard: Black king was not found at %v instead %v was found ", io.SqaureString(pos.KingSqaure[data.Black]), pos.Pieces[pos.KingSqaure[data.Black]]))
	}
}

// UpdateListMaterial sets the Material Lists
func UpdateListMaterial(pos *data.Board) {
	for i := 0; i < 120; i++ {
		sq := i
		piece := pos.Pieces[i]
		if piece == data.OffBoard || piece == data.Empty {
			continue
		}

		colour := data.PieceCol[piece]
		pos.Material[colour] += data.PieceVal[piece]
		pos.PieceList[piece][pos.PieceNumber[piece]] = sq
		pos.PieceNumber[piece]++

		if piece == data.WK {
			pos.KingSqaure[data.White] = sq
		}

		if piece == data.BK {
			pos.KingSqaure[data.Black] = sq
		}

		if data.PieceBig[piece] == data.True {
			pos.BigPiece[colour]++
		}

		if data.PieceMajor[piece] == data.True {
			pos.MajorPiece[colour]++
		}
		if data.PieceMin[piece] == data.True {
			pos.MinPiece[colour]++
		}

		if piece == data.WP {
			SetBit(&pos.Pawns[data.White], data.Sqaure120ToSquare64[sq])
			SetBit(&pos.Pawns[data.Both], data.Sqaure120ToSquare64[sq])
		}

		if piece == data.BP {
			SetBit(&pos.Pawns[data.Black], data.Sqaure120ToSquare64[sq])
			SetBit(&pos.Pawns[data.Both], data.Sqaure120ToSquare64[sq])
		}
	}
}

func MirrorBoard(pos *data.Board) {
	var pieceArray [64]int
	swapPiece := [13]int{data.Empty, data.BP, data.BN, data.BB, data.BR, data.BQ, data.BK, data.WP, data.WN, data.WB, data.WR, data.WQ, data.WK}
	tempEnPas := data.NoSqaure
	tempCastlePerm := 0
	tempSide := pos.Side ^ 1

	if pos.CastlePermission&data.WhiteKingCastle != 0 {
		tempCastlePerm |= data.BlackKingCastle
	}
	if pos.CastlePermission&data.WhiteQueenCastle != 0 {
		tempCastlePerm |= data.BlackQueenCastle
	}
	if pos.CastlePermission&data.BlackKingCastle != 0 {
		tempCastlePerm |= data.WhiteKingCastle
	}
	if pos.CastlePermission&data.BlackQueenCastle != 0 {
		tempCastlePerm |= data.WhiteQueenCastle
	}

	if pos.EnPas != data.NoSqaure {
		tempEnPas = data.Sqaure64ToSquare120[data.Mirror64[data.Sqaure120ToSquare64[pos.EnPas]]]
	}

	for sq := 0; sq < 64; sq++ {
		pieceArray[sq] = pos.Pieces[data.Sqaure64ToSquare120[data.Mirror64[sq]]]
	}

	resetBoard(pos)

	for sq := 0; sq < 64; sq++ {
		tp := swapPiece[pieceArray[sq]]
		pos.Pieces[data.Sqaure64ToSquare120[sq]] = tp
	}

	pos.Side = tempSide
	pos.CastlePermission = tempCastlePerm
	pos.EnPas = tempEnPas

	pos.PosistionKey = GeneratePositionKey(pos)
	UpdateListMaterial(pos)

	CheckBoard(pos)
}

// resetBoard restores Board to a default state
func resetBoard(pos *data.Board) {
	for i := 0; i < 120; i++ {
		pos.Pieces[i] = data.OffBoard
	}

	for i := 0; i < 64; i++ {
		pos.Pieces[data.Sqaure64ToSquare120[i]] = data.Empty
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

	pos.KingSqaure[data.White], pos.KingSqaure[data.Black] = data.NoSqaure, data.NoSqaure

	pos.Side = data.Both
	pos.EnPas = data.NoSqaure
	pos.FiftyMove = 0

	pos.Play = 0
	pos.HistoryPlay = 0

	pos.CastlePermission = 0

	pos.PosistionKey = 0
}

// getPieceType returns returns the corresponding piece type integer
func getPieceType(ch rune) int {
	pieceMap := map[rune]int{
		'p': data.BP,
		'r': data.BR,
		'n': data.BN,
		'b': data.BB,
		'q': data.BQ,
		'k': data.BK,
		'P': data.WP,
		'R': data.WR,
		'N': data.WN,
		'B': data.WB,
		'Q': data.WQ,
		'K': data.WK,
	}

	piece, ok := pieceMap[ch]
	if !ok {
		panic(fmt.Errorf("getPieceType: could not find value for %v", ch))
	}

	return piece
}
