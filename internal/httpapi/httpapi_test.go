package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alessiogarifo92/linkboard/internal/links"
)

type fakeCheck struct{ err error }

func (f fakeCheck) Ping(context.Context) error { return f.err }

func newServer(checks map[string]Checker) (*Server, *links.MemoryQueue) {
	q := &links.MemoryQueue{}
	return &Server{Svc: &links.Service{Store: links.NewMemoryStore(), Queue: q}, Checks: checks, Version: "test"}, q
}

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCreateRedirectAndGet(t *testing.T) {
	s, q := newServer(nil)
	h := s.Routes()

	rec := do(t, h, "POST", "/api/links", `{"url":"https://example.com"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST status %d: %s", rec.Code, rec.Body)
	}
	var l links.Link
	json.Unmarshal(rec.Body.Bytes(), &l)

	rec = do(t, h, "GET", "/r/"+l.Code, "")
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "https://example.com" {
		t.Fatalf("redirect: %d %q", rec.Code, rec.Header().Get("Location"))
	}
	if len(q.Codes) != 1 {
		t.Fatalf("click non accodato")
	}

	rec = do(t, h, "GET", "/api/links/"+l.Code, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET link status %d", rec.Code)
	}
	rec = do(t, h, "GET", "/api/links", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), l.Code) {
		t.Fatalf("lista recenti: %d %s", rec.Code, rec.Body)
	}
}

func TestErrors(t *testing.T) {
	s, _ := newServer(nil)
	h := s.Routes()
	cases := []struct {
		method, path, body string
		want               int
	}{
		{"POST", "/api/links", `not json`, http.StatusBadRequest},
		{"POST", "/api/links", `{"url":"ftp://x"}`, http.StatusBadRequest},
		{"POST", "/api/links", `{"url":"https://example.com/` + strings.Repeat("a", 5000) + `"}`, http.StatusBadRequest},
		{"GET", "/r/missing", "", http.StatusNotFound},
		{"GET", "/api/links/missing", "", http.StatusNotFound},
		{"DELETE", "/api/links", "", http.StatusMethodNotAllowed},
	}
	for _, c := range cases {
		if rec := do(t, h, c.method, c.path, c.body); rec.Code != c.want {
			t.Errorf("%s %s: status %d, atteso %d", c.method, c.path, rec.Code, c.want)
		}
	}
}

func TestInternalErrorHidesDetails(t *testing.T) {
	s, _ := newServer(nil)
	rec := httptest.NewRecorder()
	s.fail(rec, errors.New("password=segreta"))
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "segreta") {
		t.Fatalf("500 deve nascondere i dettagli: %d %s", rec.Code, rec.Body)
	}
}

func TestProbes(t *testing.T) {
	s, _ := newServer(map[string]Checker{"postgres": fakeCheck{}, "redis": fakeCheck{err: errors.New("down")}})
	h := s.Routes()
	if rec := do(t, h, "GET", "/healthz", ""); rec.Code != http.StatusOK {
		t.Fatalf("healthz deve essere 200 anche con dipendenze giù, ottenuto %d", rec.Code)
	}
	rec := do(t, h, "GET", "/readyz", "")
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "down") {
		t.Fatalf("readyz: %d %s", rec.Code, rec.Body)
	}
	s.Checks["redis"] = fakeCheck{}
	if rec := do(t, h, "GET", "/readyz", ""); rec.Code != http.StatusOK {
		t.Fatalf("readyz con dipendenze ok: %d", rec.Code)
	}
	if rec := do(t, h, "GET", "/version", ""); !strings.Contains(rec.Body.String(), `"version":"test"`) {
		t.Fatalf("version: %s", rec.Body)
	}
}
