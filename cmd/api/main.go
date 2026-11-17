// Comando api: serve l'API HTTP di Linkboard.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alessiogarifo92/linkboard/internal/config"
	"github.com/alessiogarifo92/linkboard/internal/httpapi"
	"github.com/alessiogarifo92/linkboard/internal/links"
	"github.com/alessiogarifo92/linkboard/internal/postgres"
	"github.com/alessiogarifo92/linkboard/internal/redisq"
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
	if err := migrateWithRetry(ctx, store, log); err != nil {
		log.Error("migrazione fallita", "err", err)
		os.Exit(1)
	}
	queue := redisq.New(config.Get("REDIS_ADDR", "localhost:6379"))

	srv := &httpapi.Server{
		Svc:     &links.Service{Store: store, Queue: queue},
		Checks:  map[string]httpapi.Checker{"postgres": store, "redis": queue},
		Version: config.Get("APP_VERSION", "dev"),
		Log:     log,
	}
	httpSrv := &http.Server{Addr: ":" + config.Get("PORT", "8080"), Handler: srv.Routes(), ReadHeaderTimeout: 5 * time.Second}

	go func() {
		log.Info("api in ascolto", "addr", httpSrv.Addr, "version", srv.Version)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server fermato", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done() // SIGTERM: Kubernetes sta terminando il pod
	log.Info("arresto in corso: completo le richieste aperte")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}

// migrateWithRetry aspetta Postgres: all'avvio del cluster il DB può arrivare dopo l'api.
func migrateWithRetry(ctx context.Context, s *postgres.Store, log *slog.Logger) error {
	var err error
	for i := 0; i < 30; i++ {
		if err = s.Migrate(ctx); err == nil {
			return nil
		}
		log.Warn("Postgres non pronto, riprovo", "tentativo", i+1, "err", err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return err
}
