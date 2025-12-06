package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/s1lentmol/q-flow-backend/services/api-gateway/config"
	"github.com/s1lentmol/q-flow-backend/services/api-gateway/internal/app"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := setupLogger(cfg.Env)

	application, err := app.New(ctx, log, cfg)
	if err != nil {
		log.Error("failed to init application", slog.Any("err", err))
		os.Exit(1)
	}
	defer application.Close()

	go func() {
		addr := app.ListenAddr(cfg.HTTP.Port)
		if err := application.Server.Run(addr); err != nil {
			log.Error("failed to run http server", slog.Any("err", err))
			stop()
		}
	}()

	<-ctx.Done()

	if err := application.Server.Shutdown(context.Background()); err != nil {
		log.Error("failed to shutdown http server", slog.Any("err", err))
	}
	log.Info("api-gateway stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
