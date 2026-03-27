package offline

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type stubRefresher struct {
	mu        sync.Mutex
	calls     int
	startedAt []time.Time
	err       error
}

func (s *stubRefresher) RefreshAll(context.Context) error {
	s.mu.Lock()
	s.calls++
	call := s.calls
	s.startedAt = append(s.startedAt, time.Now())
	s.mu.Unlock()

	if call == 1 {
		time.Sleep(30 * time.Millisecond)
		return nil
	}
	return s.err
}

func TestRunLoopWaitsAfterEachRound(t *testing.T) {
	refresher := &stubRefresher{err: errors.New("stop")}
	app := NewApp(refresher, 10*time.Millisecond)
	errCh := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go app.runLoop(ctx, errCh)

	select {
	case err := <-errCh:
		if !errors.Is(err, refresher.err) {
			t.Fatalf("expected %v, got %v", refresher.err, err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("runLoop did not finish in time")
	}

	refresher.mu.Lock()
	defer refresher.mu.Unlock()
	if refresher.calls != 2 {
		t.Fatalf("expected 2 refresh calls, got %d", refresher.calls)
	}
	if len(refresher.startedAt) != 2 {
		t.Fatalf("expected 2 timestamps, got %d", len(refresher.startedAt))
	}
	if gap := refresher.startedAt[1].Sub(refresher.startedAt[0]); gap < 40*time.Millisecond {
		t.Fatalf("expected second round to start after first finishes and interval passes, got %v", gap)
	}
}
