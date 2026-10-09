package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/usecase"
	inhttp "github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/infrastructure/adapters/inbound/http"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/infrastructure/adapters/outbound/messaging"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/infrastructure/adapters/outbound/persistence"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/infrastructure/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// ---------- outbound ----------
	pool, err := persistence.NewPool(ctx, cfg.DSN())
	if err != nil {
		slog.Error("connect postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	consultationRepo := persistence.NewConsultationRepo(pool)
	attentionRepo := persistence.NewAttentionRepo(pool)
	postSummaryRepo := persistence.NewPostSummaryRepo(pool)
	eventPublisher := messaging.NewNoopPublisher()

	// ---------- application ----------
	recordAttention := usecase.NewRecordAttention(consultationRepo, attentionRepo)
	getAttention := usecase.NewGetAttention(attentionRepo)
	getPostSummary := usecase.NewGetPostSummary(postSummaryRepo)
	generatePostSummary := usecase.NewGeneratePostSummary(consultationRepo, attentionRepo, postSummaryRepo)

	// ---------- inbound ----------
	handlers := inhttp.NewHandlers(recordAttention, getAttention, getPostSummary, generatePostSummary)
	router := inhttp.NewRouter(handlers, cfg.JWTPublicKey)

	// ---------- http server ----------
	srv := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           router.Handler(),
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		ReadTimeout:       cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}

	// Silence unused-var warnings for the publisher until it is wired to a
	// real use case that emits events (follow-up PR once ADR-011 is accepted).
	_ = eventPublisher

	go func() {
		slog.Info("starting server", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTPShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
	slog.Info("stopped")
}
