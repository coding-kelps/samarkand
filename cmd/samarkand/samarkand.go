package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/contrib/bridges/otelslog"

	"github.com/coding-kelps/samarkand/internal/config"
	"github.com/coding-kelps/samarkand/internal/logger"
	"github.com/coding-kelps/samarkand/internal/server"
	"github.com/coding-kelps/samarkand/internal/metadata"
)

func main() {
	cmd := newCommand(os.Stdout, config.Load)
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func newCommand(w io.Writer, loader func(string) (config.Config, error)) *cli.Command {
	return &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to configuration file",
				Sources: cli.EnvVars("SAMARKAND_CONFIG"),
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "start samarkand market server.",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg, err := loader(cmd.Root().String("config"))
					if err != nil {
						slog.Error("failed to load configuration", "error", err)
						return err
					}

					shutdown, err := logger.SetupOTelLogger(&cfg, ctx)
					if err != nil {
						slog.Error("failed to set up OTel logger", "error", err)
						return err
					}
					defer shutdown(ctx)

					level, err := logger.ParseLogLevel(cfg.Log.Level)
					if err != nil {
						return err
					}

					otelHandler := otelslog.NewHandler(metadata.GetName())
					consoleHandler := slog.NewTextHandler(w, &slog.HandlerOptions{
						Level:     level,
						AddSource: true,
					})

					logger := slog.New(logger.NewMultiHandler(otelHandler, consoleHandler))
					slog.SetDefault(logger)

					s := server.New(&cfg, logger)

					return s.Start()
				},
			},
			{
				Name:  "version",
				Usage: "Get the version samarkand.",
				Action: func(context.Context, *cli.Command) error {
					fmt.Fprintln(w, metadata.GetVersionWithBuildInfo())
					return nil
				},
			},
		},
	}
}
