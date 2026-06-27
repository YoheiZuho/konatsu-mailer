package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yoheizuho/konatsu-mailer/internal/analysis"
	"github.com/yoheizuho/konatsu-mailer/internal/api"
	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/imapsync"
	"github.com/yoheizuho/konatsu-mailer/internal/push"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
	"github.com/yoheizuho/konatsu-mailer/internal/ws"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	db, err := store.New(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	if err := store.Migrate(cfg.DatabaseURL, "./migrations"); err != nil {
		logger.Error("failed to run migrations", slog.Any("error", err))
		os.Exit(1)
	}

	// Root context cancelled on SIGINT/SIGTERM; ties background workers to the
	// process lifetime.
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Realtime hub.
	hub := ws.NewHub()
	go hub.Run(rootCtx)

	// LLM analysis pipeline + Web Push.
	pusher := push.NewPusher(cfg.VapidPublicKey, cfg.VapidPrivateKey, cfg.VapidSubject)
	pipeline := analysis.New(db, cfg, hub, pusher)
	pipeline.Start(rootCtx)

	// IMAP sync workers (enqueue new mail for analysis).
	syncMgr := imapsync.NewManager(db, cfg, hub, pipeline)
	go func() {
		if err := syncMgr.Start(rootCtx); err != nil {
			logger.Error("sync manager stopped", slog.Any("error", err))
		}
	}()

	// Reusable IMAP connections for on-demand operations (body fetch, flags, move).
	pool := imapsync.NewPool()
	defer pool.Close()

	gin.SetMode(gin.ReleaseMode)
	r := api.NewRouter(cfg, db, hub, pipeline, pool)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
		// No Read/Write timeouts: the WebSocket (/api/ws) and SSE (/api/ai/draft)
		// endpoints are long-lived. ReadHeaderTimeout still guards the handshake.
		ReadHeaderTimeout: 15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("server starting", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server listen error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	<-rootCtx.Done()
	logger.Info("server shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", slog.Any("error", err))
	}
	logger.Info("server exited")
}
