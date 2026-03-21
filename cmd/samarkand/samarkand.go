package main

import (
	"context"
	"fmt"
	"log"
	"io"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/coding-kelps/samarkand/internal/config"
	"github.com/coding-kelps/samarkand/internal/version"
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
						return fmt.Errorf("loading config: %w", err)
					}

					fmt.Fprintf(w, "Starting with config: %+v\n", cfg)
					return nil
				},
			},
			{
				Name:  "version",
				Usage: "Get the version samarkand.",
				Action: func(context.Context, *cli.Command) error {
					fmt.Fprintln(w, version.GetVersionWithBuildInfo())
					return nil
				},
			},
		},
	}
}
