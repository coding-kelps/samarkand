package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/coding-kelps/samarkand/internal/version"
)

func main() {
	cmd := &cli.Command{
		Flags: []cli.Flag{
		},
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "start samarkand market server.",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					panic("not implemented")
				},
			},
			{
				Name:  "version",
				Usage: "Get the version samarkand.",
				Action: func(context.Context, *cli.Command) error {
					fmt.Println(version.GetVersionWithBuildInfo())
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
