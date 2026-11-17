package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alessiogarifo92/linkboard/internal/links"
)

type fakePopper struct {
	codes []string
	err   error
}

func (f *fakePopper) Pop(context.Context, time.Duration) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if len(f.codes) == 0 {
		return "", nil
	}
	c := f.codes[0]
	f.codes = f.codes[1:]
	return c, nil
}

func TestProcessOneCountsClicks(t *testing.T) {
	ctx := context.Background()
	store := links.NewMemoryStore()
	store.Create(ctx, "abc", "https://example.com")
	w := &Worker{Queue: &fakePopper{codes: []string{"abc", "abc", "gone"}}, Store: store}

	for i := 0; i < 3; i++ {
		if ok, err := w.ProcessOne(ctx); !ok || err != nil {
			t.Fatalf("giro %d: ok=%v err=%v", i, ok, err)
		}
	}
	if ok, err := w.ProcessOne(ctx); ok || err != nil {
		t.Fatalf("coda vuota: ok=%v err=%v", ok, err)
	}
	l, _ := store.Get(ctx, "abc")
	if l.Clicks != 2 {
		t.Fatalf("clicks = %d, atteso 2 (il click su 'gone' va scartato)", l.Clicks)
	}
}

func TestProcessOneQueueError(t *testing.T) {
	w := &Worker{Queue: &fakePopper{err: errors.New("redis giù")}, Store: links.NewMemoryStore()}
	if _, err := w.ProcessOne(context.Background()); err == nil {
		t.Fatal("atteso errore dalla coda")
	}
}

func TestRunStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Worker{Queue: &fakePopper{err: errors.New("redis giù")}, Store: links.NewMemoryStore(), Backoff: time.Hour}
	done := make(chan struct{})
	go func() { w.Run(ctx); close(done) }()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run non si ferma alla cancellazione del contesto")
	}
}
