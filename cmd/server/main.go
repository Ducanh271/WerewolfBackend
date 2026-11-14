package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "masoi/internal/roles" // registers every role
	"masoi/internal/room"
	"masoi/internal/store"
	"masoi/internal/store/mysql"
	"masoi/internal/transport/ws"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	mcfg := room.ManagerConfig{Logger: logger}
	// MASOI_MYSQL_DSN keeps every room's events and every finished game in
	// MySQL (docs/database.md); without it everything lives in memory.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	db, err := mysql.OpenEnv(ctx, os.Getenv)
	cancel()
	if err != nil {
		logger.Error("server not started", "err", err)
		os.Exit(1)
	}
	if db != nil {
		defer func() { _ = db.Close() }()
		mcfg.Store, mcfg.Archive = db, db
		logger.Info("storing games in MySQL")
	}
	rooms := room.NewManager(mcfg)
	defer rooms.Close()

	cfg := ws.ConfigFromEnv(os.Getenv)
	cfg.Rooms = rooms
	cfg.Logger = logger
	if db != nil { // never a nil *mysql.Store in the interface
		cfg.History = db
	}
	handler, err := ws.New(cfg)
	if err != nil {
		logger.Error("server not started", "err", err)
		os.Exit(1)
	}
	srv := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 10 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if db != nil {
		go purgeDaily(ctx, db, logger)
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()

	logger.Info("listening", "addr", addr, "allowAnyOrigin", cfg.AllowAnyOrigin)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

// purgeDaily deletes the rooms older than store.Retention, with their events
// and games (docs/database.md "Retention"): at start, then once a day.
func purgeDaily(ctx context.Context, db *mysql.Store, logger *slog.Logger) {
	for {
		n, err := db.Purge(time.Now().Add(-store.Retention))
		if err != nil {
			logger.Error("old games not purged", "err", err)
		} else if n > 0 {
			logger.Info("old games purged", "rooms", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(24 * time.Hour):
		}
	}
}
