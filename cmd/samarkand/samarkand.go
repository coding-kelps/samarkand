package main

import (
	"context"
	"log"
	"os"

	"github.com/coding-kelps/samarkand/internal/command"
	"github.com/coding-kelps/samarkand/internal/config"
)

func main() {
	cmd := command.NewCommand(os.Stdout, config.Load)
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
