package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kodeshack/stuff/config"
	"github.com/kodeshack/stuff/server"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("error starting stuff", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start, err := setup()
	if err != nil {
		return err
	}

	return start(ctx)
}

func setup() (func(context.Context) error, error) {
	config, err := config.NewConfigFromEnv()
	if err != nil {
		return nil, err
	}

	srv, err := server.NewServer(config.Addr)
	if err != nil {
		return nil, err
	}

	return srv.Start, nil
}
