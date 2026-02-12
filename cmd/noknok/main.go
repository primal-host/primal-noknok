package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/primal-host/noknok/internal/atproto"
	"github.com/primal-host/noknok/internal/config"
	"github.com/primal-host/noknok/internal/database"
	"github.com/primal-host/noknok/internal/server"
	"github.com/primal-host/noknok/internal/session"
)

func main() {
	slog.Info("noknok starting", "version", config.Version)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	db, err := database.Open(ctx, cfg.DSN())
	cancel()
	if err != nil {
		slog.Error("database open failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("database connected")

	// Seed owner user.
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	if err := db.SeedOwner(ctx, cfg.OwnerDID); err != nil {
		cancel()
		slog.Error("failed to seed owner", "error", err)
		os.Exit(1)
	}
	cancel()
	slog.Info("owner seeded", "did", cfg.OwnerDID)

	// OAuth client.
	store := atproto.NewPgStore(db.Pool)
	oauthClient, err := atproto.NewOAuthClient(cfg.PublicURL, cfg.OAuthPrivateKey, store)
	if err != nil {
		slog.Error("OAuth client init failed", "error", err)
		os.Exit(1)
	}
	slog.Info("OAuth client initialized")

	// Session manager.
	ttl, err := time.ParseDuration(cfg.SessionTTL)
	if err != nil {
		slog.Error("invalid SESSION_TTL", "error", err)
		os.Exit(1)
	}
	secure := strings.HasPrefix(cfg.PublicURL, "https://")
	sess := session.NewManager(db.Pool, ttl, cfg.CookieDomain, secure)
	sess.StartCleanup()

	srv := server.New(db, sess, cfg, oauthClient)

	go func() {
		if err := srv.Start(); err != nil {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("shutting down", "signal", sig.String())

	sess.StopCleanup()

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("stopped")
}
