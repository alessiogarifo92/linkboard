package postgres

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/alessiogarifo92/linkboard/internal/links"
)

// Test di integrazione: girano solo con LINKBOARD_TEST_DATABASE_URL (in CI c'è un Postgres di servizio).
// Usano codici con prefisso "test-" e li cancellano alla fine: non toccano altri dati.
func testStore(t *testing.T) *Store {
	dsn := os.Getenv("LINKBOARD_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("LINKBOARD_TEST_DATABASE_URL non impostata")
	}
	ctx := context.Background()
	s, err := Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		s.Pool.Exec(ctx, `DELETE FROM links WHERE code LIKE 'test-%'`)
		s.Pool.Close()
	})
	return s
}

func TestStoreRoundTrip(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	l, err := s.Create(ctx, "test-a", "https://example.com")
	if err != nil || l.CreatedAt.IsZero() {
		t.Fatalf("Create: %+v %v", l, err)
	}
	if _, err := s.Create(ctx, "test-a", "https://other.example"); !errors.Is(err, links.ErrCodeTaken) {
		t.Fatalf("atteso ErrCodeTaken, ottenuto %v", err)
	}
	if err := s.AddClicks(ctx, "test-a", 2); err != nil {
		t.Fatal(err)
	}
	if err := s.AddClicks(ctx, "test-missing", 1); !errors.Is(err, links.ErrNotFound) {
		t.Fatalf("atteso ErrNotFound, ottenuto %v", err)
	}
	got, err := s.Get(ctx, "test-a")
	if err != nil || got.Clicks != 2 || got.URL != "https://example.com" {
		t.Fatalf("Get: %+v %v", got, err)
	}
	if _, err := s.Get(ctx, "test-missing"); !errors.Is(err, links.ErrNotFound) {
		t.Fatalf("atteso ErrNotFound, ottenuto %v", err)
	}
	recent, err := s.Recent(ctx, 50)
	if err != nil || len(recent) == 0 {
		t.Fatalf("Recent: %v %v", recent, err)
	}
	if err := s.Ping(ctx); err != nil {
		t.Fatal(err)
	}
}
