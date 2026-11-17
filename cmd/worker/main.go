// Comando worker: consuma i click da Redis e aggiorna i contatori in Postgres.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alessiogarifo92/linkboard/internal/config"
	"github.com/alessiogarifo92/linkboard/internal/postgres"
	"github.com/alessiogarifo92/linkboard/internal/redisq"
	"github.com/alessiogarifo92/linkboard/internal/worker"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	store, err := postgres.Connect(ctx, config.Get("DATABASE_URL", "postgres://linkboard:linkboard@localhost:5432/linkboard"))
	if err != nil {
		log.Error("connessione a Postgres fallita", "err", err)
		os.Exit(1)
	}
	queue := redisq.New(config.Get("REDIS_ADDR", "localhost:6379"))

	// Endpoint per la liveness probe (settimana 5).
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })
		srv := &http.Server{Addr: ":" + config.Get("HEALTH_PORT", "8081"), Handler: mux, ReadHeaderTimeout: 5 * time.Second}
		_ = srv.ListenAndServe()
	}()

	w := &worker.Worker{Queue: queue, Store: store, Log: log, Timeout: 5 * time.Second, Backoff: 2 * time.Second}
	log.Info("worker avviato")
	w.Run(ctx)
	log.Info("worker fermato")
}
