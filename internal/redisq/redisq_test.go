package redisq

import (
	"context"
	"os"
	"testing"
	"time"
)

// Test di integrazione: girano solo con LINKBOARD_TEST_REDIS_ADDR. Usano una chiave propria.
func TestPushPop(t *testing.T) {
	addr := os.Getenv("LINKBOARD_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("LINKBOARD_TEST_REDIS_ADDR non impostata")
	}
	ctx := context.Background()
	q := New(addr)
	q.Key = "linkboard:test:clicks"
	t.Cleanup(func() { q.Client.Del(ctx, q.Key); q.Client.Close() })

	if err := q.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	q.Push(ctx, "a")
	q.Push(ctx, "b")
	for _, want := range []string{"a", "b"} { // FIFO: LPUSH + BRPOP
		got, err := q.Pop(ctx, time.Second)
		if err != nil || got != want {
			t.Fatalf("Pop = %q, %v; atteso %q", got, err, want)
		}
	}
	if got, err := q.Pop(ctx, time.Second); got != "" || err != nil {
		t.Fatalf("coda vuota: %q, %v", got, err)
	}
}
