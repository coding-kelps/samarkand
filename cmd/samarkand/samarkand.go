package main

import (
	"context"
	"log"
	"os"

	"github.com/coding-kelps/samarkand/internal/app"
	"github.com/coding-kelps/samarkand/internal/config"
)

func main() {
	cmd := app.NewCommand(os.Stdout, config.Load)
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
