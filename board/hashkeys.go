package board

import (
	"github.com/AdamGriffiths31/ChessEngine/data"
)

// GeneratePositionKey generates a unique key based on the position
func GeneratePositionKey(pos *data.Board) uint64 {
	var finalKey uint64 = 0
	piece := data.Empty

	for sq := 0; sq < 120; sq++ {
		piece = pos.Pieces[sq]
		if sq != data.NoSquare && piece != data.Empty && piece != data.OffBoard {
			finalKey ^= data.PieceKeys[piece][sq]
		}
	}

	if pos.Side == data.White {
		finalKey ^= data.SideKey
	}

	if pos.EnPas != data.NoSquare {
		finalKey ^= data.PieceKeys[data.Empty][pos.EnPas]
	}

	finalKey ^= data.CastleKeys[pos.CastlePermission]

	return finalKey
}
