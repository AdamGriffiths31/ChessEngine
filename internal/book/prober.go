package book

import (
	"log/slog"

	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

// Prober wraps a LookupService with caller-side lazy loading and the
// opening-phase move-number gate.
//
// A Prober is constructed once per set of book files and reused across
// moves: the underlying book is loaded from disk at most once (on the first
// Probe call within the opening phase), and a load failure is logged once
// and never retried.
type Prober struct {
	files     []string
	service   *LookupService
	attempted bool
}

// NewProber creates a Prober over the given Polyglot book files. The files
// are not read until the first Probe call that falls within the opening
// phase (see BookMoveLimit); a Prober for an empty file list never attempts
// to load anything and Probe always reports a miss.
func NewProber(files []string) *Prober {
	return &Prober{files: files}
}

// Probe looks up a book move for the given position. It returns (move, true)
// on a hit. It returns (nil, false) when: no book files were configured, the
// position's full-move number is past BookMoveLimit, the book failed to
// load, or the position simply is not in the book - all of these are
// ordinary "no book move, fall back to search" outcomes for the caller.
func (p *Prober) Probe(b *board.Board) (*board.Move, bool) {
	if len(p.files) == 0 {
		return nil, false
	}
	if b.GetFullMoveNumber() > BookMoveLimit {
		return nil, false
	}

	p.ensureLoaded()
	if p.service == nil {
		return nil, false
	}

	move, err := p.service.FindBookMove(b)
	if err != nil || move == nil {
		return nil, false
	}
	return move, true
}

func (p *Prober) ensureLoaded() {
	if p.attempted {
		return
	}
	p.attempted = true

	cfg := Config{
		Enabled:       true,
		BookFiles:     p.files,
		SelectionMode: SelectBest,
	}

	service := NewLookupService(cfg)
	if err := service.LoadBooks(); err != nil {
		slog.Warn("failed to load opening book(s)", "files", p.files, "error", err)
		return
	}
	p.service = service
}
