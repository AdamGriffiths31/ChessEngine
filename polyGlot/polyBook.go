package polyglot

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"os"
	"unsafe"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/moveGen"
)

func InitPolyBook() {
	data.EngineSettings.UseBook = false
	file, err := os.Open("performance.bin")
	if err != nil {
		fmt.Println(os.Getwd())
		data.EngineSettings.UseBook = false
		return
	}
	defer file.Close()

	fi, err := os.Stat("performance.bin")
	if err != nil {
		panic(fmt.Errorf("InitPolyBook: size  error %v", err))
	}
	size := fi.Size()

	NumEntries = uint64(size) / uint64(unsafe.Sizeof(PolyBookEntry{}))

	PolyEntry = make([]PolyBookEntry, NumEntries)

	reader := io.Reader(file)
	_ = binary.Read(reader, binary.LittleEndian, &PolyEntry)

	if NumEntries > 0 {
		data.EngineSettings.UseBook = true
	}
}

func GetBookMove(pos *data.Board) int {
	var bookMoves [32]int
	polyKey := PolyKeyFromBoard(pos)
	count := 0
	for i := 0; i < int(NumEntries); i++ {
		if polyKey == littleEndianToBigEndianUint64(PolyEntry[i].Key) {
			move := littleEndianToBigEndianUint16(PolyEntry[i].Move)
			tempMove := ConvertPolyMove(move, pos)
			if tempMove != data.NoMove {
				bookMoves[count] = tempMove
				count++
				if count > 32 {
					break
				}
			}
		}
	}
	if count != 0 {
		randMove := rand.Intn(count)
		return bookMoves[randMove]
	}
	return data.NoMove
}

func ConvertPolyMove(polyMove uint16, pos *data.Board) int {

	ff := data.FileChars[(polyMove >> 6 & 7)]
	fr := data.RankChars[(polyMove >> 9 & 7)]
	tf := data.FileChars[(polyMove >> 0 & 7)]
	tr := data.RankChars[(polyMove >> 3 & 7)]
	pp := (polyMove >> 12 & 7)
	var move string
	promotedPiece := "q"
	if pp != 0 {
		switch pp {
		case 1:
			promotedPiece = "n"
		case 2:
			promotedPiece = "b"
		case 3:
			promotedPiece = "r"
		}
		move = fmt.Sprintf("%s%s%s%s%s", ff, fr, tf, tr, promotedPiece)
	} else {
		move = fmt.Sprintf("%s%s%s%s", ff, fr, tf, tr)
	}

	return moveGen.ParseMove([]byte(move), pos)
}

func PolyKeyFromBoard(pos *data.Board) uint64 {
	finalKey := uint64(0)
	for sq := 0; sq < 120; sq++ {
		piece := pos.Pieces[sq]
		if piece != data.NoSquare && piece != data.Empty && piece != data.OffBoard {
			if piece > data.BK || piece < data.WP {
				panic(fmt.Errorf("PolyKeyFromBoard: piece error sq:%v value: %v", sq, piece))
			}
			polyPiece := PolyKindOfPiece[piece]
			rank := data.RanksBoard[sq]
			file := data.FilesBoard[sq]

			finalKey ^= Random64Poly[(64*polyPiece)+(8*rank)+file]
		}
	}
	//Castle
	offSet := 768
	if pos.CastlePermission&data.WhiteKingCastle != 0 {
		finalKey ^= Random64Poly[offSet+0]
	}
	if pos.CastlePermission&data.WhiteQueenCastle != 0 {
		finalKey ^= Random64Poly[offSet+1]
	}
	if pos.CastlePermission&data.BlackKingCastle != 0 {
		finalKey ^= Random64Poly[offSet+2]
	}
	if pos.CastlePermission&data.BlackQueenCastle != 0 {
		finalKey ^= Random64Poly[offSet+3]
	}
	//EnPas
	offSet = 772
	if hasPawnForCapture(pos) {
		file := data.FilesBoard[pos.EnPas]
		finalKey ^= Random64Poly[offSet+file]
	}
	//Side
	if pos.Side == data.White {
		offSet = 780
		finalKey ^= Random64Poly[offSet]
	}
	return finalKey
}

func hasPawnForCapture(pos *data.Board) bool {
	target := data.WP
	sqWithPawn := 0
	if pos.Side == data.Black {
		target = data.BP
	}
	if pos.EnPas != data.NoSquare {
		if pos.Side == data.White {
			sqWithPawn = pos.EnPas - 10
		} else {
			sqWithPawn = pos.EnPas + 10
		}

		if pos.Pieces[sqWithPawn+1] == target {
			return true
		} else if pos.Pieces[sqWithPawn-1] == target {
			return true
		}
	}
	return false
}

func clearEntries() {
	for i := range PolyEntry {
		PolyEntry[i].Key = 0
		PolyEntry[i].Move = 0
		PolyEntry[i].Weight = 0
		PolyEntry[i].Learn = 0
	}
}

func littleToBigEndian(l uint32) uint32 {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, l)
	return binary.BigEndian.Uint32(b)
}

func littleEndianToBigEndianUint64(littleEndian uint64) uint64 {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, littleEndian)
	return binary.BigEndian.Uint64(b)
}

func littleEndianToBigEndianUint16(n uint16) uint16 {
	return (n >> 8) | (n << 8)
}
