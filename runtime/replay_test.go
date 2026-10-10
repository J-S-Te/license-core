package runtime

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	core "github.com/J-S-Te/license-core"
)

func TestExactSignedReplayIsDurableWithoutRenewingAuthorization(t *testing.T) {
	f := newFixture(t)
	s := f.snapshot()
	s.CurrentLicenseJWS = f.license(1, f.now.Unix()-10, f.now.Unix()+10, f.binding.Application)
	raw, ctx := f.sign(s), context.Background()
	if err := f.r.ApplySnapshot(ctx, raw, f.now); err != nil {
		t.Fatal(err)
	}
	f.r = f.restart()
	if err := f.r.ApplySnapshot(ctx, raw, f.now.Add(10*time.Second)); err != nil {
		t.Fatal("exact replay failed", err)
	}
	state, err := f.r.ReadState(ctx)
	if err != nil || state.SnapshotJWS != raw || state.ControlRevision != s.Revision || state.HighestLicenseVersion != 1 || state.ClockHighWater != f.now.Unix()+10 {
		t.Fatal("replay changed authority or lost clock persistence", err)
	}
	f.r = f.restart()
	if err := f.r.ApplySnapshot(ctx, raw, f.now.Add(5*time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := f.r.Evaluate(ctx, core.MUTATE_BUSINESS, f.now.Add(5*time.Second)); !errors.Is(err, core.ErrDenied) {
		t.Fatal("replay resurrected expired authority", err)
	}
	for _, op := range []core.Operation{core.READ_HISTORY, core.EXPORT_HISTORY} {
		if err := f.r.Evaluate(ctx, op, f.now.Add(10*time.Second)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestExactSignedReplayCannotClearClockLatch(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	s := f.snapshot()
	s.SnapshotIssuedAt = f.now.Unix() - 1000
	raw := f.sign(s)
	if err := f.r.ApplySnapshot(ctx, raw, f.now); err != nil {
		t.Fatal(err)
	}
	if err := f.r.ApplySnapshot(ctx, raw, f.now.Add(-301*time.Second)); !errors.Is(err, ErrClock) {
		t.Fatal("exact replay bypassed rollback observation", err)
	}
	f.r = f.restart()
	if err := f.r.ApplySnapshot(ctx, raw, f.now.Add(time.Second)); !errors.Is(err, ErrClock) {
		t.Fatal("replay cleared durable clock latch", err)
	}
	state, err := f.r.ReadState(ctx)
	if err != nil || !state.ClockBlocked {
		t.Fatal("clock latch not persisted", err)
	}
}

func TestExactSignedReplayConcurrentAndEquivocation(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	s := f.snapshot()
	s.Revision = 2
	raw := f.sign(s)
	if err := f.r.ApplySnapshot(ctx, raw, f.now); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	failures := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r := f.restart()
			failures <- r.ApplySnapshot(ctx, raw, f.now.Add(time.Duration(i)*time.Second))
		}(i)
	}
	wg.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	state, err := f.r.ReadState(ctx)
	if err != nil || state.ClockHighWater != f.now.Unix()+11 {
		t.Fatal("concurrent replay lost maximum observed time", err)
	}
	for _, kind := range []string{"equal-different", "older", "future"} {
		t.Run(kind, func(t *testing.T) {
			next := s
			expected := ErrRollback
			switch kind {
			case "equal-different":
				next.SnapshotIssuedAt++
			case "older":
				next.Revision--
			case "future":
				next.SnapshotIssuedAt = f.now.Unix() + 301
				expected = ErrInvalid
			}
			if err := f.r.ApplySnapshot(ctx, f.sign(next), f.now); !errors.Is(err, expected) {
				t.Fatal("nonexact replay accepted", kind, err)
			}
		})
	}
}
