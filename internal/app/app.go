package app

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	"github.com/coding-kelps/samarkand/internal/config"
	"github.com/coding-kelps/samarkand/internal/logger"
	"github.com/coding-kelps/samarkand/internal/metadata"
	"github.com/coding-kelps/samarkand/internal/server"
)

type App struct {
	otelProvider *sdklog.LoggerProvider
	logger       *slog.Logger
	server       *server.Server
}

func NewApp(ctx context.Context, cfg *config.Config) (App, error) {
	level, err := logger.ParseLogLevel(cfg.Log.Level)
	if err != nil {
		return App{}, err
	}

	otelProvider, err := logger.NewOTelLoggerProvider(ctx)
	if err != nil {
		return App{}, err
	}
	otelHandler := otelslog.NewHandler(metadata.GetName(), otelslog.WithLoggerProvider(otelProvider))
	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})

	logger := slog.New(logger.NewMultiHandler(otelHandler, consoleHandler))
	server := server.NewServer(&server.ServerConfig{
		Addr:   cfg.Server.Addr,
		Logger: logger,
	})

	return App{
		logger: logger,
		server: server,
	}, nil
}

func (a *App) Run() error {

	err := a.server.Start()
	if err != nil {
		return err
	}
	return nil
}
