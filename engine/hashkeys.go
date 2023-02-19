package engine

func GeneratePositionKey(pos *Board) uint64 {
	var finalKey uint64 = 0
	piece := Empty

	for sq := 0; sq < 120; sq++ {
		piece = pos.Pieces[sq]
		if sq != noSqaure && piece != Empty && piece != OffBoard {
			finalKey ^= PieceKeys[piece][sq]
		}
	}

	if pos.Side == White {
		finalKey ^= SideKey
	}

	if pos.EnPas != noSqaure {
		finalKey ^= PieceKeys[Empty][pos.EnPas]
	}

	finalKey ^= CastleKeys[pos.CastlePermission]

	return finalKey
}
