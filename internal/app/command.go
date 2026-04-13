package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/contrib/bridges/otelslog"

	"github.com/coding-kelps/samarkand/internal/config"
	"github.com/coding-kelps/samarkand/internal/logger"
	"github.com/coding-kelps/samarkand/internal/metadata"
	"github.com/coding-kelps/samarkand/internal/server"
)

func NewCommand(w io.Writer, loader func(string) (config.Config, error)) *cli.Command {
	return &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to configuration file",
				Sources: cli.EnvVars("SAMARKAND__CONFIG"),
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "start",
				Usage:  "start samarkand market server",
				Action: startAction(w, loader),
			},
			{
				Name:  "validate",
				Usage: "validate the loaded configuration",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "insecure-debug",
						Usage: "Display loaded configuration (for debugging purposes only!)",
					},
				},
				Action: validateAction(w, loader),
			},
			{
				Name:  "version",
				Usage: "Get the version of samarkand",
				Action: func(_ context.Context, _ *cli.Command) error {
					fmt.Fprintln(w, metadata.GetVersionWithBuildInfo())
					return nil
				},
			},
		},
	}
}

func startAction(w io.Writer, loader func(string) (config.Config, error)) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		cfg, err := loader(cmd.Root().String("config"))
		if err != nil {
			return err
		}

		shutdown, err := logger.SetupOTelLogger(&cfg, ctx)
		if err != nil {
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

		log := slog.New(logger.NewMultiHandler(otelHandler, consoleHandler))
		slog.SetDefault(log)

		s := server.New(&cfg, log)
		return s.Start()
	}
}

func validateAction(w io.Writer, loader func(string) (config.Config, error)) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		cfg, err := loader(cmd.Root().String("config"))
		if err != nil {
			return err
		}

		fmt.Fprintln(w, "configuration is valid")

		if cmd.Bool("insecure-debug") {

			cfgJSON, err := json.Marshal(cfg)
			if err != nil {
				return err
			}

			fmt.Println(string(cfgJSON))
		}

		return nil
	}
}
