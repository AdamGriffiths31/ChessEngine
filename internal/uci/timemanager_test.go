package uci

import (
	"testing"
	"time"
)

func TestAllocateTimeInvariants(t *testing.T) {
	clocks := []time.Duration{0, 5 * time.Millisecond, 200 * time.Millisecond, time.Second, 10 * time.Second, 60 * time.Second, 5 * time.Minute}
	incs := []time.Duration{0, 100 * time.Millisecond, 2 * time.Second, 30 * time.Second}
	overheads := []time.Duration{0, 50 * time.Millisecond, 500 * time.Millisecond, 10 * time.Second}
	togos := []int{0, 1, 2, 20, 200}

	for _, tl := range clocks {
		for _, inc := range incs {
			for _, oh := range overheads {
				for _, mtg := range togos {
					b := allocateTime(tl, inc, mtg, oh)
					if b.Soft <= 0 || b.Hard <= 0 {
						t.Fatalf("tl=%v inc=%v oh=%v mtg=%d: non-positive budget %+v", tl, inc, oh, mtg, b)
					}
					if b.Soft > b.Hard {
						t.Fatalf("tl=%v inc=%v oh=%v mtg=%d: soft %v > hard %v", tl, inc, oh, mtg, b.Soft, b.Hard)
					}
					usable := max(tl-oh, time.Millisecond)
					if b.Hard > usable {
						t.Fatalf("tl=%v inc=%v oh=%v mtg=%d: hard %v exceeds usable %v", tl, inc, oh, mtg, b.Hard, usable)
					}
				}
			}
		}
	}
}

func TestAllocateTimeSuddenDeath(t *testing.T) {
	b := allocateTime(60*time.Second, 0, 0, 0)
	if b.Soft != 1500*time.Millisecond {
		t.Errorf("soft = %v, want 1.5s (60s / 40)", b.Soft)
	}
	if b.Hard != 6*time.Second {
		t.Errorf("hard = %v, want 6s (4 x soft)", b.Hard)
	}
}

func TestAllocateTimeUsesMovesToGo(t *testing.T) {
	sudden := allocateTime(60*time.Second, 0, 0, 0)
	short := allocateTime(60*time.Second, 0, 5, 0)
	if short.Soft <= sudden.Soft {
		t.Errorf("movestogo 5 soft %v should exceed sudden-death soft %v", short.Soft, sudden.Soft)
	}
	if capped := allocateTime(60*time.Second, 0, 500, 0); capped.Soft != allocateTime(60*time.Second, 0, maxMovesToGo, 0).Soft {
		t.Errorf("movestogo above %d should be capped", maxMovesToGo)
	}
}

func TestAllocateTimeIncrementAddsToSoft(t *testing.T) {
	without := allocateTime(60*time.Second, 0, 0, 0)
	with := allocateTime(60*time.Second, time.Second, 0, 0)
	if got := with.Soft - without.Soft; got != 800*time.Millisecond {
		t.Errorf("increment added %v to soft, want 800ms", got)
	}
}

func TestAllocateTimeOverheadReducesBudget(t *testing.T) {
	plain := allocateTime(10*time.Second, 0, 0, 0)
	slow := allocateTime(10*time.Second, 0, 0, 2*time.Second)
	if slow.Soft >= plain.Soft || slow.Hard >= plain.Hard {
		t.Errorf("overhead should shrink the budget: plain %+v, with overhead %+v", plain, slow)
	}
}

func TestAllocateTimeBigIncrementTinyClock(t *testing.T) {
	b := allocateTime(time.Second, 10*time.Second, 0, 50*time.Millisecond)
	if b.Hard > 800*time.Millisecond {
		t.Errorf("hard = %v, want at most 80%% of usable clock", b.Hard)
	}
}
