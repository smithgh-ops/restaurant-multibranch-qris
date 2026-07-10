package main

import (
	"log/slog"
	"os"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/config"
	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/server"
)

func main() {
	// Structured JSON logging in production, text in development
	cfg := config.Load()

	if cfg.AppEnv == "production" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	}

	slog.Info("starting", "app", cfg.AppName, "env", cfg.AppEnv)

	if err := server.Run(cfg); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}
