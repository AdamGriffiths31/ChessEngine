package benchmark

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/game"
	"github.com/AdamGriffiths31/ChessEngine/game/ai"
	"github.com/AdamGriffiths31/ChessEngine/game/ai/search"
	"github.com/AdamGriffiths31/ChessEngine/game/moves"
	"github.com/AdamGriffiths31/ChessEngine/uci"
)

// stockfishDefaultMinElo and stockfishDefaultMaxElo are used if the
// UCI_Elo spin option can't be parsed from the engine's "uci" response.
const (
	stockfishDefaultMinElo = 1320
	stockfishDefaultMaxElo = 3190
)

// EloRunner orchestrates an Elo benchmark run against Stockfish.
type EloRunner struct {
	engineManager *EngineManager
	rootPath      string
}

// NewEloRunner creates a new Elo benchmark runner.
func NewEloRunner(rootPath string) *EloRunner {
	return &EloRunner{
		engineManager: NewEngineManager(rootPath),
		rootPath:      rootPath,
	}
}

// LoadEngines loads engine configurations from engines.json.
func (r *EloRunner) LoadEngines() error {
	return r.engineManager.LoadEngines()
}

// Run executes an Elo benchmark with the given configuration.
func (r *EloRunner) Run(config EloConfig) (*EloResult, error) {
	stockfish, err := r.engineManager.FindEngineByCommandSubstring("stockfish")
	if err != nil {
		return nil, fmt.Errorf("failed to find stockfish in engines.json: %w", err)
	}

	minElo, maxElo, err := probeUCIEloRange(stockfish.Command)
	if err != nil {
		return nil, fmt.Errorf("failed to determine stockfish UCI_Elo range: %w", err)
	}

	anchors := computeAnchors(config.AnchorCenter, config.AnchorCount, config.AnchorStep, minElo, maxElo)
	totalGames := len(anchors) * config.GamesPerLevel

	startTime := time.Now()
	var games []EloGameResult

	for _, anchorElo := range anchors {
		client, err := NewStockfishClient(stockfish.Command)
		if err != nil {
			return nil, fmt.Errorf("failed to start stockfish for anchor %d: %w", anchorElo, err)
		}

		if err := configureStockfish(client, anchorElo); err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("failed to configure stockfish for anchor %d: %w", anchorElo, err)
		}

		for g := 0; g < config.GamesPerLevel; g++ {
			chessEngineWhite := g%2 == 0
			openingIndex := len(games) % len(eloOpenings)

			result, err := playEloGame(client, anchorElo, chessEngineWhite, openingIndex, config)
			if err != nil {
				_ = client.Close()
				return nil, fmt.Errorf("aborting benchmark: anchor %d game %d: %w", anchorElo, g+1, err)
			}
			games = append(games, result)
			fmt.Printf("Game %d/%d (anchor %d, %s): %s (%d plies, depth %d)\n",
				len(games), totalGames, anchorElo, result.Color, outcomeLabel(result.Result), result.Plies, result.AvgDepth)
		}

		_ = client.Close()
	}

	gameAnchors, scores := gameAnchorsAndScores(games)

	return &EloResult{
		EstimatedElo: estimateRating(gameAnchors, scores),
		Anchors:      anchors,
		Games:        games,
		Config:       config,
		Duration:     time.Since(startTime),
		Timestamp:    time.Now().Format("20060102_150405"),
	}, nil
}

// probeUCIEloRange starts a throwaway Stockfish process to read its
// UCI_Elo spin option's min/max from the "uci" handshake.
func probeUCIEloRange(binaryPath string) (minElo, maxElo int, err error) {
	client, err := NewStockfishClient(binaryPath)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = client.Close() }()

	if err := client.Send("uci"); err != nil {
		return 0, 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	lines, err := client.WaitForAll(ctx, "uciok")
	if err != nil {
		return 0, 0, err
	}

	if minElo, maxElo, found := parseUCIEloRange(lines); found {
		return minElo, maxElo, nil
	}
	return stockfishDefaultMinElo, stockfishDefaultMaxElo, nil
}

// parseUCIEloRange scans "uci" handshake lines for the UCI_Elo spin
// option's declared min/max.
func parseUCIEloRange(lines []string) (minElo, maxElo int, found bool) {
	for _, line := range lines {
		if !strings.HasPrefix(line, "option name UCI_Elo") || !strings.Contains(line, "type spin") {
			continue
		}
		fields := strings.Fields(line)
		for i, f := range fields {
			if f == "min" && i+1 < len(fields) {
				if v, err := strconv.Atoi(fields[i+1]); err == nil {
					minElo = v
				}
			}
			if f == "max" && i+1 < len(fields) {
				if v, err := strconv.Atoi(fields[i+1]); err == nil {
					maxElo = v
				}
			}
		}
		found = true
	}
	return minElo, maxElo, found
}

// configureStockfish completes the "uci" handshake and pins the engine to
// anchorElo via UCI_LimitStrength/UCI_Elo.
func configureStockfish(client *StockfishClient, anchorElo int) error {
	if err := client.Send("uci"); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := client.WaitForAll(ctx, "uciok"); err != nil {
		return err
	}

	for _, opt := range []string{
		"setoption name UCI_LimitStrength value true",
		fmt.Sprintf("setoption name UCI_Elo value %d", anchorElo),
		"setoption name Threads value 1",
	} {
		if err := client.Send(opt); err != nil {
			return err
		}
	}

	return startNewGame(client)
}

// startNewGame sends ucinewgame and waits for the engine to confirm
// readiness.
func startNewGame(client *StockfishClient) error {
	if err := client.Send("ucinewgame"); err != nil {
		return err
	}
	if err := client.Send("isready"); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := client.WaitFor(ctx, "readyok")
	return err
}

// playEloGame plays one game between ChessEngine and an already-configured
// Stockfish client, returning the result from ChessEngine's perspective.
// A Stockfish process or protocol failure aborts the game immediately with
// an error, rather than scoring it as a win — crediting ChessEngine with
// undeserved wins would silently inflate the Elo estimate.
func playEloGame(client *StockfishClient, anchorElo int, chessEngineWhite bool, openingIndex int, config EloConfig) (EloGameResult, error) {
	result := EloGameResult{
		AnchorElo:    anchorElo,
		Color:        colorName(chessEngineWhite),
		OpeningIndex: openingIndex,
	}

	if err := startNewGame(client); err != nil {
		return EloGameResult{}, fmt.Errorf("failed to start new game with stockfish: %w", err)
	}

	engine := game.NewEngine()
	converter := uci.NewMoveConverter()
	computer := ai.NewComputerPlayer("ChessEngine", search.NewMinimaxEngine(), ai.SearchConfig{
		MaxDepth:       64,
		UseOpeningBook: true,
		BookFiles:      []string{"game/openings/testdata/performance.bin"},
	})

	var moveHistory []string
	for _, openingMove := range eloOpenings[openingIndex] {
		move, err := converter.FromUCI(openingMove, engine.GetState().Board)
		if err != nil {
			break
		}
		if err := engine.MakeMove(move); err != nil {
			break
		}
		moveHistory = append(moveHistory, openingMove)
	}

	wtime, btime := config.TimeMs, config.TimeMs
	var totalNodes uint64
	totalDepth, searchMoves := 0, 0
	resultSet := false

	for ply := 0; ply < config.MaxPlies; ply++ {
		state := engine.GetState()
		if state.GameOver {
			result.Result = scoreFromGameOver(state, chessEngineWhite)
			resultSet = true
			break
		}

		currentPlayer := engine.GetCurrentPlayer()
		isChessEngineTurn := (currentPlayer == game.White) == chessEngineWhite
		moveStart := time.Now()

		if isChessEngineTurn {
			remaining := wtime
			if currentPlayer == game.Black {
				remaining = btime
			}
			budget := time.Duration(allocateTime(remaining, config.IncMs)) * time.Millisecond

			movesPlayer := moves.White
			if currentPlayer == game.Black {
				movesPlayer = moves.Black
			}

			searchResult, err := computer.GetMoveWithStats(state.Board, movesPlayer, budget)
			if err != nil {
				result.Result = 0.0
				resultSet = true
				break
			}
			if err := engine.MakeMove(searchResult.BestMove); err != nil {
				result.Result = 0.0
				resultSet = true
				break
			}
			moveHistory = append(moveHistory, converter.ToUCI(searchResult.BestMove))
			totalNodes += uint64(searchResult.Stats.NodesSearched)
			totalDepth += searchResult.Stats.Depth
			searchMoves++
		} else {
			if err := client.Send(buildPositionCommand(moveHistory)); err != nil {
				return EloGameResult{}, fmt.Errorf("failed to send position to stockfish: %w", err)
			}
			if err := client.Send(fmt.Sprintf("go wtime %d btime %d winc %d binc %d", wtime, btime, config.IncMs, config.IncMs)); err != nil {
				return EloGameResult{}, fmt.Errorf("failed to send go command to stockfish: %w", err)
			}

			remaining := remainingFor(currentPlayer, wtime, btime)
			deadline := time.Duration(remaining+5000) * time.Millisecond
			ctx, cancel := context.WithTimeout(context.Background(), deadline)
			line, err := client.WaitFor(ctx, "bestmove")
			cancel()
			if err != nil {
				return EloGameResult{}, fmt.Errorf("stockfish did not respond with bestmove: %w", err)
			}

			fields := strings.Fields(line)
			if len(fields) < 2 {
				return EloGameResult{}, fmt.Errorf("malformed bestmove line from stockfish: %q", line)
			}
			move, err := converter.FromUCI(fields[1], state.Board)
			if err != nil {
				return EloGameResult{}, fmt.Errorf("failed to parse stockfish's move %q: %w", fields[1], err)
			}
			if err := engine.MakeMove(move); err != nil {
				return EloGameResult{}, fmt.Errorf("stockfish sent an illegal move %q: %w", fields[1], err)
			}
			moveHistory = append(moveHistory, fields[1])
		}

		elapsed := int(time.Since(moveStart).Milliseconds())
		if currentPlayer == game.White {
			wtime = wtime - elapsed + config.IncMs
		} else {
			btime = btime - elapsed + config.IncMs
		}

		if wtime <= 0 {
			result.Result = boolToScore(!chessEngineWhite)
			resultSet = true
			break
		}
		if btime <= 0 {
			result.Result = boolToScore(chessEngineWhite)
			resultSet = true
			break
		}
	}

	if !resultSet {
		result.Result = 0.5 // MaxPlies safety cap reached: treat as a draw.
	}

	result.Plies = len(moveHistory)
	result.Nodes = totalNodes
	if searchMoves > 0 {
		result.AvgDepth = totalDepth / searchMoves
	}

	return result, nil
}

func colorName(chessEngineWhite bool) string {
	if chessEngineWhite {
		return "white"
	}
	return "black"
}

// outcomeLabel converts a game result (1.0/0.5/0.0) to a human-readable
// outcome from ChessEngine's perspective.
func outcomeLabel(result float64) string {
	switch result {
	case 1.0:
		return "WIN"
	case 0.5:
		return "DRAW"
	default:
		return "LOSS"
	}
}

func boolToScore(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

// scoreFromGameOver converts the engine's terminal game state into a score
// from ChessEngine's perspective.
func scoreFromGameOver(state *game.State, chessEngineWhite bool) float64 {
	if state.IsDraw {
		return 0.5
	}
	chessEngineColor := game.Black
	if chessEngineWhite {
		chessEngineColor = game.White
	}
	return boolToScore(state.Winner == chessEngineColor)
}

func remainingFor(player game.Player, wtime, btime int) int {
	if player == game.White {
		return wtime
	}
	return btime
}

func buildPositionCommand(moveHistory []string) string {
	if len(moveHistory) == 0 {
		return "position startpos"
	}
	return "position startpos moves " + strings.Join(moveHistory, " ")
}
