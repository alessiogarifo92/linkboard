// Package httpapi espone il Service di links via HTTP.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alessiogarifo92/linkboard/internal/links"
)

// Checker è una dipendenza da verificare in /readyz (Postgres, Redis).
type Checker interface {
	Ping(ctx context.Context) error
}

// Server raccoglie le dipendenze dei handler.
type Server struct {
	Svc     *links.Service
	Checks  map[string]Checker
	Version string // da env APP_VERSION: utile per vedere i rolling update
	Log     *slog.Logger
}

// Routes restituisce il router dell'api.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/links", s.createLink)
	mux.HandleFunc("GET /api/links", s.recentLinks)
	mux.HandleFunc("GET /api/links/{code}", s.getLink)
	mux.HandleFunc("GET /r/{code}", s.redirect)
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)
	mux.HandleFunc("GET /version", s.version)
	return mux
}

func (s *Server) createLink(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "JSON non valido: atteso {\"url\": \"...\"}")
		return
	}
	l, err := s.Svc.Shorten(r.Context(), body.URL)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, l)
}

func (s *Server) recentLinks(w http.ResponseWriter, r *http.Request) {
	ls, err := s.Svc.Store.Recent(r.Context(), 20)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ls)
}

func (s *Server) getLink(w http.ResponseWriter, r *http.Request) {
	l, err := s.Svc.Store.Get(r.Context(), r.PathValue("code"))
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, l)
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	target, err := s.Svc.Resolve(r.Context(), r.PathValue("code"))
	if err != nil {
		s.fail(w, err)
		return
	}
	http.Redirect(w, r, target, http.StatusFound)
}

// healthz (liveness): il processo risponde. Non controlla le dipendenze, altrimenti
// un Postgres lento farebbe riavviare tutti i pod dell'api.
func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// readyz (readiness): il pod può ricevere traffico solo se le dipendenze rispondono.
func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	result := map[string]string{}
	status := http.StatusOK
	for name, c := range s.Checks {
		if err := c.Ping(ctx); err != nil {
			result[name] = err.Error()
			status = http.StatusServiceUnavailable
		} else {
			result[name] = "ok"
		}
	}
	writeJSON(w, status, result)
}

// version mostra versione e nome del pod: chiamalo in loop per vedere il load balancing tra repliche.
func (s *Server) version(w http.ResponseWriter, _ *http.Request) {
	host, _ := os.Hostname()
	writeJSON(w, http.StatusOK, map[string]string{"version": s.Version, "pod": host})
}

func (s *Server) fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, links.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, links.ErrInvalidURL):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		if s.Log != nil {
			s.Log.Error("errore interno", "err", err)
		}
		writeError(w, http.StatusInternalServerError, "errore interno")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
