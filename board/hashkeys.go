package board

import (
	"github.com/AdamGriffiths31/ChessEngine/data"
)

func GeneratePositionKey(pos *data.Board) uint64 {
	var finalKey uint64 = 0
	piece := data.Empty

	for sq := 0; sq < 120; sq++ {
		piece = pos.Pieces[sq]
		if sq != data.NoSqaure && piece != data.Empty && piece != data.OffBoard {
			finalKey ^= data.PieceKeys[piece][sq]
		}
	}

	if pos.Side == data.White {
		finalKey ^= data.SideKey
	}

	if pos.EnPas != data.NoSqaure {
		finalKey ^= data.PieceKeys[data.Empty][pos.EnPas]
	}

	finalKey ^= data.CastleKeys[pos.CastlePermission]

	return finalKey
}
