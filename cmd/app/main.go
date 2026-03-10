package main

import (
	"context"
	"errors"
	"hrms/internal/infrastructure/config"
	"hrms/pkg/log"
	"os"
)

func main() {
	configPath := os.Args[1]
	cfg := config.ParseConfig(configPath)
	logger := log.NewLog(cfg.LogLevel)
	logger.Info("Some Info Message")
	logger.With(
		log.String("Structured log KEY", "Structured log VALUE"),
	).Info("With some info message")

	ctx := context.Background()
	logger.WithContext(ctx).Info("Logging with context",
		log.String("key", "value"),
	)

	logger.Warn("Some warning message")

	err := errors.New("some Error Message")
	logger.Error("501", log.Error(err))
}
