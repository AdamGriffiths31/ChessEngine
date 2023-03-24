package engine

type Game struct {
	position      *Position
	moves         []Move
	numberOfMoves uint16
}

func NewGame(
	position *Position,
	moves []Move,
	numberOfMoves uint16) Game {
	if position == nil {
		position = &Position{PositionHistory: NewPositionHistory(), Positions: map[uint64]int{}}
	}
	return Game{
		position,
		moves,
		numberOfMoves,
	}
}

func (g *Game) Position() *Position {
	return g.position
}

func ParseFen(fen string) Game {
	game := NewGame(nil, nil, 0)
	game.position.ParseFen(fen)
	return game
}

func (p *Position) Copy() *Position {
	copyMap := make(map[uint64]int, len(p.Positions))
	for k, v := range p.Positions {
		copyMap[k] = v
	}

	newPos := &Position{
		Board:            p.Board.copy(),
		Play:             p.Play,
		PositionKey:      p.PositionKey,
		Side:             p.Side,
		CastlePermission: p.CastlePermission,
		EnPassant:        p.EnPassant,
		FailHighFirst:    p.FailHighFirst,
		FailHigh:         p.FailHigh,
		MoveHistory:      p.MoveHistory,
		CurrentScore:     p.CurrentScore,
		FiftyMove:        p.FiftyMove,
		PositionHistory:  NewPositionHistory(),
		Positions:        copyMap,
	}
	return newPos
}
