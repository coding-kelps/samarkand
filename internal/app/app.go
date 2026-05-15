package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	"github.com/coding-kelps/samarkand/internal/clock"
	"github.com/coding-kelps/samarkand/internal/config"
	"github.com/coding-kelps/samarkand/internal/logger"
	"github.com/coding-kelps/samarkand/internal/metadata"
	"github.com/coding-kelps/samarkand/internal/server"
)

type App struct {
	otelProvider *sdklog.LoggerProvider
	logger       *slog.Logger
	clock        *clock.MarketClock
	server       *server.Server
}

func NewApp(ctx context.Context, cfg *config.Config) (App, error) {
	var otelProvider *sdklog.LoggerProvider
	var loggerHandlers []slog.Handler

	level, err := logger.ParseLogLevel(cfg.Log.Level)
	if err != nil {
		return App{}, err
	}
	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})
	loggerHandlers = append(loggerHandlers, consoleHandler)

	if cfg.Log.OpenTelemetry.Enabled {
		otelProvider, err := logger.NewOTelLoggerProvider(ctx)
		if err != nil {
			return App{}, err
		}
		otelHandler := otelslog.NewHandler(metadata.GetName(), otelslog.WithLoggerProvider(otelProvider))
		loggerHandlers = append(loggerHandlers, otelHandler)
	}

	logger := slog.New(logger.NewMultiHandler(loggerHandlers...))
	server := server.NewServer(&server.ServerConfig{
		Addr:   cfg.Server.Addr,
		Logger: logger,
	})
	clock, err := clock.NewMarketClock(&clock.MarketClockConfig{
		Redis: clock.RedisConfig{
			Addr:     cfg.Redis.Addr,
			Username: cfg.Redis.ResolvedUsername(),
			Password: cfg.Redis.ResolvedPassword(),
			DB:       cfg.Redis.DB,
		},
		Logger: logger,
	})
	if err != nil {
		return App{}, err
	}

	return App{
		otelProvider: otelProvider,
		logger:       logger,
		server:       server,
		clock:        clock,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-ctx.Done()
		a.logger.Info("SIGTERM signal catched gracefully shutdown")
		stop()
		a.Shutdown()
	}()

	if err := a.clock.Start(ctx); err != nil {
		return err
	}

	go func() {
		tickInterval := 10 * time.Second
		ticker := time.NewTicker(tickInterval)

		for {
			select {
			case <-ticker.C:
				a.logger.Info("clock time", "time", a.clock.Now().Format("2006-01-02 15:04:05"))
			case <-ctx.Done():
				return
			}
		}
	}()

	if err := a.server.Start(ctx); err != nil {
		return err
	}

	return nil
}

func (a *App) Shutdown() error {
	a.server.Stop()

	if err := a.clock.Close(); err != nil {
		return err
	}

	return nil
}
