package logger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	otlp "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"

	"github.com/coding-kelps/samarkand/internal/config"
	"github.com/coding-kelps/samarkand/internal/metadata"
)

func ParseLogLevel(s string) (slog.Level, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(s)); err != nil {
		return level, fmt.Errorf("invalid log level %q: %w", s, err)
	}
	return level, nil
}

func SetupOTelLogger(cfg *config.Config, ctx context.Context) (shutdown func(context.Context) error, err error) {
	exporter, err := otlp.New(ctx)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(metadata.GetName()),
			semconv.ServiceVersionKey.String(metadata.GetVersion()),
		),
	)
	if err != nil {
		return nil, err
	}

	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(
			sdklog.NewBatchProcessor(exporter,
				sdklog.WithExportInterval(5*time.Second),
				sdklog.WithExportTimeout(10*time.Second),
			),
		),
	)

	global.SetLoggerProvider(provider)

	return provider.Shutdown, nil
}

type multiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) slog.Handler {
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, rec slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, rec.Level) {
			if err := h.Handle(ctx, rec.Clone()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: next}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: next}
}
