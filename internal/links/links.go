// Package links contiene la logica di dominio di Linkboard, indipendente da HTTP, Postgres e Redis.
package links

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"
)

// Errori di dominio, tradotti in status HTTP da httpapi.
var (
	ErrNotFound   = errors.New("link non trovato")
	ErrInvalidURL = errors.New("URL non valido: serve http:// o https:// e un host")
	ErrCodeTaken  = errors.New("codice già in uso")
)

// Link è un URL accorciato con il suo conteggio di click.
type Link struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	Clicks    int64     `json:"clicks"`
	CreatedAt time.Time `json:"created_at"`
}

// Store persiste i link (in produzione: Postgres).
type Store interface {
	Create(ctx context.Context, code, rawURL string) (Link, error) // ErrCodeTaken se il codice esiste
	Get(ctx context.Context, code string) (Link, error)            // ErrNotFound se assente
	AddClicks(ctx context.Context, code string, n int64) error
	Recent(ctx context.Context, limit int) ([]Link, error)
}

// Queue trasporta gli eventi di click dall'api al worker (in produzione: una lista Redis).
type Queue interface {
	Push(ctx context.Context, code string) error
}

const alphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789" // senza 0/O/1/l/I ambigui

// NewCode genera un codice casuale di n caratteri da alphabet.
func NewCode(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("lunghezza codice non valida: %d", n)
	}
	var b strings.Builder
	max := big.NewInt(int64(len(alphabet)))
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b.WriteByte(alphabet[idx.Int64()])
	}
	return b.String(), nil
}

// ValidateURL accetta solo URL assoluti http/https con host, entro 2048 caratteri.
func ValidateURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 2048 {
		return "", ErrInvalidURL
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", ErrInvalidURL
	}
	return u.String(), nil
}

// Service unisce Store e Queue nei casi d'uso dell'api.
type Service struct {
	Store    Store
	Queue    Queue
	CodeLen  int
	Attempts int // tentativi in caso di collisione del codice
}

// Shorten valida l'URL e crea un link con un codice nuovo, riprovando sulle collisioni.
func (s *Service) Shorten(ctx context.Context, rawURL string) (Link, error) {
	clean, err := ValidateURL(rawURL)
	if err != nil {
		return Link{}, err
	}
	attempts := s.Attempts
	if attempts <= 0 {
		attempts = 5
	}
	codeLen := s.CodeLen
	if codeLen <= 0 {
		codeLen = 7
	}
	for i := 0; i < attempts; i++ {
		code, err := NewCode(codeLen)
		if err != nil {
			return Link{}, err
		}
		l, err := s.Store.Create(ctx, code, clean)
		if errors.Is(err, ErrCodeTaken) {
			continue
		}
		return l, err
	}
	return Link{}, fmt.Errorf("nessun codice libero dopo %d tentativi", attempts)
}

// Resolve restituisce l'URL di destinazione e accoda un click.
// Un errore della coda non blocca il redirect: il click va perso, l'utente no.
func (s *Service) Resolve(ctx context.Context, code string) (string, error) {
	l, err := s.Store.Get(ctx, code)
	if err != nil {
		return "", err
	}
	_ = s.Queue.Push(ctx, code)
	return l.URL, nil
}
