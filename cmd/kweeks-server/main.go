// Command kweeks-server runs the kweeks backend: quiz authoring, live rooms,
// server-authoritative scoring, podium declaration, and winner-driven prize
// redemption.
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

	"github.com/osugodbless/kweeks/internal/adapters/bmoni"
	"github.com/osugodbless/kweeks/internal/adapters/clock"
	"github.com/osugodbless/kweeks/internal/adapters/httpapi"
	"github.com/osugodbless/kweeks/internal/adapters/mailer"
	"github.com/osugodbless/kweeks/internal/adapters/scheduler"
	"github.com/osugodbless/kweeks/internal/adapters/store/postgres"
	"github.com/osugodbless/kweeks/internal/adapters/ws"
	"github.com/osugodbless/kweeks/internal/app"
	"github.com/osugodbless/kweeks/internal/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("kweeks-server exited with error", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// --- Storage ---
	st, err := postgres.OpenMigrated(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer st.Close()

	realClock := clock.NewReal()

	// Realtime: player sockets stream room events via the websocket hub, which
	// is also the event bus the Game publishes through.
	hub := ws.NewHub(logger)

	// --- Application services (manual constructor injection) ---
	game := app.NewGame(st, realClock, hub)
	join := app.NewJoin(st, realClock)

	money := bmoni.New(cfg.BmoniBaseURL, cfg.BmoniAPIKey, cfg.BmoniOwnerKey, cfg.BmoniInstructorUserID, cfg.BmoniWalletID)
	mail := mailer.New(cfg.SmtpHost, cfg.SmtpPort, cfg.SmtpUser, cfg.SmtpPass, cfg.FromAddr, logger)
	red := app.NewRedemption(st, realClock, money, mail)

	// --- HTTP + realtime transport ---
	api := httpapi.New(game, join, red).WithWS(hub)
	mux := http.NewServeMux()
	api.Routes(mux)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// --- AUTO-pacing scheduler ---
	sched := scheduler.New(game, st, logger, 500*time.Millisecond)
	go func() {
		if err := sched.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("scheduler stopped", "err", err)
		}
	}()

	// --- Serve ---
	errCh := make(chan error, 1)
	go func() {
		logger.Info("kweeks-server listening", "addr", cfg.HTTPAddr, "env", cfg.Env)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTO)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
