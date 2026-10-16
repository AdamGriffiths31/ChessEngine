package uci

import (
	"context"
	"log/slog"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/movegen"
	"github.com/AdamGriffiths31/ChessEngine/internal/search"
)

// Book outcomes reported in the "book" group of a move record.
const (
	// BookHit means the move was played from the opening book.
	BookHit = "hit"
	// BookNotConfigured means no book file is set, so the book was not consulted.
	BookNotConfigured = "not_configured"
	// BookNoEntry means the book was consulted and had no move for the position.
	BookNoEntry = "no_entry"
	// BookLeft means the book had no move for a position that followed a book move.
	BookLeft = "left_book"
)

// bookOutcome says whether the book supplied the move and, if not, why.
type bookOutcome struct {
	Used   bool
	Reason string
}

// moveLog is everything handleGo knows about one move once its search ends.
type moveLog struct {
	fen     string
	player  movegen.Player
	params  SearchParams
	budget  time.Duration
	soft    time.Duration
	elapsed time.Duration
	move    string
	result  search.Result
	book    bookOutcome
}

// logMove writes one structured "move" record at Info level. It is a no-op
// unless the default slog logger has Info enabled (cmd/uci enables it when a
// debug log file is configured), so normal play pays nothing for it.
//
// Scores are in centipawns from the side to move's point of view (score_cp_stm).
// The search group is search.Stats marshaled as-is, so it already carries the
// per-iteration trace, stop reason and pruning counters for this one search;
// tt_cumulative is the table's lifetime counters, not per move.
func (ue *Engine) logMove(m moveLog) {
	ctx := context.Background()
	if !slog.Default().Enabled(ctx, slog.LevelInfo) {
		return
	}

	stats := m.result.Stats
	var nps int64
	if secs := m.elapsed.Seconds(); secs > 0 {
		nps = int64(float64(stats.NodesSearched) / secs)
	}

	side := "black"
	if m.player == movegen.White {
		side = "white"
	}

	attrs := []slog.Attr{
		slog.String("session", ue.session),
		slog.Int("game", ue.gameNumber),
		slog.Int("move_number", ue.moveNumber),
		slog.String("fen", m.fen),
		slog.String("side", side),
		slog.String("move", m.move),
		slog.Int("score_cp_stm", int(m.result.Score)),
		slog.Int("depth", stats.Depth),
		slog.Int64("nodes", stats.NodesSearched),
		slog.Int64("nps", nps),
		slog.Int64("elapsed_ms", m.elapsed.Milliseconds()),
		slog.Int64("budget_ms", m.budget.Milliseconds()),
		slog.Int64("soft_ms", m.soft.Milliseconds()),
		slog.Group("clock",
			slog.Int64("wtime_ms", m.params.WTime.Milliseconds()),
			slog.Int64("btime_ms", m.params.BTime.Milliseconds()),
			slog.Int64("winc_ms", m.params.WInc.Milliseconds()),
			slog.Int64("binc_ms", m.params.BInc.Milliseconds()),
			slog.Int64("movetime_ms", m.params.MoveTime.Milliseconds()),
			slog.Int("depth_limit", m.params.Depth),
			slog.Int("movestogo", m.params.MovesToGo),
			slog.Bool("infinite", m.params.Infinite),
		),
		slog.Group("book",
			slog.Bool("used", m.book.Used),
			slog.String("reason", m.book.Reason),
		),
	}

	if minimax, ok := ue.aiEngine.(*search.MinimaxEngine); ok {
		tt := minimax.TTStats()
		attrs = append(attrs, slog.Group("tt_cumulative",
			slog.Uint64("hits", tt.Hits),
			slog.Uint64("misses", tt.Misses),
			slog.Uint64("collisions", tt.Collisions),
			slog.Float64("hit_rate_pct", tt.HitRate),
		))
	}

	attrs = append(attrs, slog.Any("search", stats))

	slog.LogAttrs(ctx, slog.LevelInfo, "move", attrs...)
}
