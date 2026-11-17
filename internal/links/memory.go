package links

import (
	"context"
	"sort"
	"sync"
	"time"
)

// MemoryStore è uno Store in memoria per i test (non usarlo in produzione: ogni replica avrebbe dati suoi).
type MemoryStore struct {
	mu    sync.Mutex
	links map[string]Link
	Now   func() time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{links: map[string]Link{}, Now: time.Now}
}

func (m *MemoryStore) Create(_ context.Context, code, rawURL string) (Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.links[code]; ok {
		return Link{}, ErrCodeTaken
	}
	l := Link{Code: code, URL: rawURL, CreatedAt: m.Now()}
	m.links[code] = l
	return l, nil
}

func (m *MemoryStore) Get(_ context.Context, code string) (Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.links[code]
	if !ok {
		return Link{}, ErrNotFound
	}
	return l, nil
}

func (m *MemoryStore) AddClicks(_ context.Context, code string, n int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.links[code]
	if !ok {
		return ErrNotFound
	}
	l.Clicks += n
	m.links[code] = l
	return nil
}

func (m *MemoryStore) Recent(_ context.Context, limit int) ([]Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Link, 0, len(m.links))
	for _, l := range m.links {
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// MemoryQueue registra i codici accodati; Err, se impostato, viene restituito da Push.
type MemoryQueue struct {
	mu    sync.Mutex
	Codes []string
	Err   error
}

func (q *MemoryQueue) Push(_ context.Context, code string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.Err != nil {
		return q.Err
	}
	q.Codes = append(q.Codes, code)
	return nil
}
