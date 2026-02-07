package main

import (
	"hrms/internal/infrastructure/config"
	"hrms/pkg/logger"
	"log/slog"
	"os"
)

func main() {
	configPath := os.Args[1]
	cfg := config.ParseConfig(configPath)
	log := logger.SetupLogger(cfg.Env)

	log.Info("Loading config...",
		slog.String("env", cfg.Env),
		slog.Any("path", cfg.Http.Port))
	log.Warn("Zhansaya loshara")
	log.Error("Adel tozhe")
}
