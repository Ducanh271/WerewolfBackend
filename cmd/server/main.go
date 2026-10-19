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
	"masoi/internal/transport/ws"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	rooms := room.NewManager(room.ManagerConfig{Logger: logger})
	defer rooms.Close()

	cfg := ws.ConfigFromEnv(os.Getenv)
	cfg.Rooms = rooms
	cfg.Logger = logger
	handler, err := ws.New(cfg)
	if err != nil {
		logger.Error("server not started", "err", err)
		os.Exit(1)
	}
	srv := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 10 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
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
