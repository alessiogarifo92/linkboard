// Package worker consuma la coda dei click e aggiorna i contatori.
package worker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/alessiogarifo92/linkboard/internal/links"
)

// Popper estrae un click dalla coda; "" significa coda vuota allo scadere del timeout.
type Popper interface {
	Pop(ctx context.Context, timeout time.Duration) (string, error)
}

type Worker struct {
	Queue   Popper
	Store   links.Store
	Log     *slog.Logger
	Timeout time.Duration // attesa massima di BRPOP
	Backoff time.Duration // pausa dopo un errore della coda
}

// ProcessOne gestisce al massimo un click. Restituisce true se ne ha elaborato uno.
// Un click per un link cancellato viene scartato: riprovarlo non lo farebbe mai riuscire.
func (w *Worker) ProcessOne(ctx context.Context) (bool, error) {
	code, err := w.Queue.Pop(ctx, w.Timeout)
	if err != nil || code == "" {
		return false, err
	}
	if err := w.Store.AddClicks(ctx, code, 1); err != nil && !errors.Is(err, links.ErrNotFound) {
		return false, err
	}
	return true, nil
}

// Run elabora click finché ctx non viene cancellato (SIGTERM da Kubernetes).
func (w *Worker) Run(ctx context.Context) {
	for ctx.Err() == nil {
		if _, err := w.ProcessOne(ctx); err != nil && ctx.Err() == nil {
			if w.Log != nil {
				w.Log.Error("elaborazione click fallita", "err", err)
			}
			select {
			case <-ctx.Done():
			case <-time.After(w.Backoff):
			}
		}
	}
}
