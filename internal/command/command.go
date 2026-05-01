package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"github.com/coding-kelps/samarkand/internal/app"
	"github.com/coding-kelps/samarkand/internal/config"
	"github.com/coding-kelps/samarkand/internal/metadata"
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
				Action: start(w, loader),
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
				Action: validate(w, loader),
			},
			{
				Name:   "version",
				Usage:  "Get the version of samarkand",
				Action: version(w, loader),
			},
		},
	}
}

func start(w io.Writer, loader func(string) (config.Config, error)) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		cfg, err := loader(cmd.Root().String("config"))
		if err != nil {
			return err
		}

		a, err := app.NewApp(ctx, &cfg)
		if err != nil {
			return err
		}

		err = a.Run()
		if err != nil {
			return err
		}

		return nil
	}
}

func validate(w io.Writer, loader func(string) (config.Config, error)) cli.ActionFunc {
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

func version(w io.Writer, _ func(string) (config.Config, error)) cli.ActionFunc {
	return func(_ context.Context, _ *cli.Command) error {
		fmt.Fprintln(w, metadata.GetVersionWithBuildInfo())
		return nil
	}
}
