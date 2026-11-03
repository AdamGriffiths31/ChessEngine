package uci

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"github.com/AdamGriffiths31/ChessEngine/internal/search"
)

const (
	ponderSearching int32 = iota
	ponderMoveHeld
	ponderHitReceived
)

type ponderState struct {
	state    atomic.Int32
	hitAt    atomic.Pointer[time.Time]
	cancel   context.CancelFunc
	hitLimit time.Duration
}

func (p *ponderState) markReadyAndCheckHeld() bool {
	p.state.CompareAndSwap(ponderSearching, ponderMoveHeld)
	return p.state.Load() != ponderHitReceived
}

func (p *ponderState) hit() *time.Timer {
	now := time.Now()
	p.hitAt.Store(&now)
	switch p.state.Swap(ponderHitReceived) {
	case ponderSearching:
		return time.AfterFunc(p.hitLimit, p.cancel)
	case ponderMoveHeld:
		p.cancel()
	}
	return nil
}

func (p *ponderState) softElapsed() time.Duration {
	hitAt := p.hitAt.Load()
	if hitAt == nil {
		return 0
	}
	return time.Since(*hitAt)
}

func (p *ponderState) wasHit() bool {
	return p.state.Load() == ponderHitReceived
}

func (p *ponderState) cancelUnlessHit() {
	if !p.wasHit() {
		p.cancel()
	}
}

func (p *ponderState) outcome() string {
	if p == nil {
		return PonderNone
	}
	if p.wasHit() {
		return PonderHit
	}
	return PonderMiss
}

func (ue *Engine) handlePonderHit() {
	if ue.ponder == nil {
		return
	}
	if timer := ue.ponder.hit(); timer != nil {
		ue.ponderTimer = timer
	}
}

func (ue *Engine) ponderEnabled() bool {
	return strings.EqualFold(ue.options["Ponder"], "true")
}

func (ue *Engine) ponderMove(result search.Result) string {
	pv := result.Stats.PrincipalVariation
	if !ue.ponderEnabled() || len(pv) < 2 {
		return ""
	}
	return ue.converter.ToUCI(pv[1])
}
