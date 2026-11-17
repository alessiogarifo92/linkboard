package links

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNewCode(t *testing.T) {
	code, err := NewCode(7)
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 7 {
		t.Fatalf("lunghezza = %d, atteso 7", len(code))
	}
	for _, c := range code {
		if !strings.ContainsRune(alphabet, c) {
			t.Fatalf("carattere %q fuori alfabeto", c)
		}
	}
	if _, err := NewCode(0); err == nil {
		t.Fatal("atteso errore per lunghezza 0")
	}
}

func TestValidateURL(t *testing.T) {
	ok := []string{"https://example.com", "http://example.com/a?b=c", "  https://example.com/x  "}
	for _, u := range ok {
		if _, err := ValidateURL(u); err != nil {
			t.Errorf("%q: errore inatteso %v", u, err)
		}
	}
	bad := []string{"", "example.com", "ftp://example.com", "javascript:alert(1)", "https://", "https://" + strings.Repeat("a", 2050)}
	for _, u := range bad {
		if _, err := ValidateURL(u); !errors.Is(err, ErrInvalidURL) {
			t.Errorf("%q: atteso ErrInvalidURL, ottenuto %v", u, err)
		}
	}
}

// takenStore simula collisioni: le prime n Create rispondono ErrCodeTaken.
type takenStore struct {
	*MemoryStore
	n int
}

func (s *takenStore) Create(ctx context.Context, code, u string) (Link, error) {
	if s.n > 0 {
		s.n--
		return Link{}, ErrCodeTaken
	}
	return s.MemoryStore.Create(ctx, code, u)
}

func TestShortenRetriesOnCollision(t *testing.T) {
	svc := &Service{Store: &takenStore{MemoryStore: NewMemoryStore(), n: 2}, Queue: &MemoryQueue{}, Attempts: 3}
	l, err := svc.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if l.URL != "https://example.com" || l.Code == "" {
		t.Fatalf("link inatteso: %+v", l)
	}
}

func TestShortenGivesUpAfterAttempts(t *testing.T) {
	svc := &Service{Store: &takenStore{MemoryStore: NewMemoryStore(), n: 10}, Queue: &MemoryQueue{}, Attempts: 3}
	if _, err := svc.Shorten(context.Background(), "https://example.com"); err == nil {
		t.Fatal("atteso errore dopo i tentativi esauriti")
	}
}

func TestShortenRejectsInvalidURL(t *testing.T) {
	svc := &Service{Store: NewMemoryStore(), Queue: &MemoryQueue{}}
	if _, err := svc.Shorten(context.Background(), "not a url"); !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("atteso ErrInvalidURL, ottenuto %v", err)
	}
}

func TestResolve(t *testing.T) {
	ctx := context.Background()
	q := &MemoryQueue{}
	svc := &Service{Store: NewMemoryStore(), Queue: q}
	l, _ := svc.Shorten(ctx, "https://example.com")

	got, err := svc.Resolve(ctx, l.Code)
	if err != nil || got != "https://example.com" {
		t.Fatalf("Resolve = %q, %v", got, err)
	}
	if len(q.Codes) != 1 || q.Codes[0] != l.Code {
		t.Fatalf("click non accodato: %v", q.Codes)
	}
	if _, err := svc.Resolve(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("atteso ErrNotFound, ottenuto %v", err)
	}
}

func TestResolveSurvivesQueueFailure(t *testing.T) {
	ctx := context.Background()
	svc := &Service{Store: NewMemoryStore(), Queue: &MemoryQueue{Err: errors.New("redis giù")}}
	l, _ := svc.Shorten(ctx, "https://example.com")
	if got, err := svc.Resolve(ctx, l.Code); err != nil || got != "https://example.com" {
		t.Fatalf("il redirect deve funzionare anche senza coda: %q, %v", got, err)
	}
}

func TestMemoryStoreAddClicksAndRecent(t *testing.T) {
	ctx := context.Background()
	m := NewMemoryStore()
	m.Create(ctx, "a", "https://a.example")
	m.Create(ctx, "b", "https://b.example")
	if err := m.AddClicks(ctx, "a", 3); err != nil {
		t.Fatal(err)
	}
	if err := m.AddClicks(ctx, "zzz", 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("atteso ErrNotFound, ottenuto %v", err)
	}
	l, _ := m.Get(ctx, "a")
	if l.Clicks != 3 {
		t.Fatalf("clicks = %d, atteso 3", l.Clicks)
	}
	if got, _ := m.Recent(ctx, 1); len(got) != 1 {
		t.Fatalf("Recent(1) = %d elementi", len(got))
	}
}
