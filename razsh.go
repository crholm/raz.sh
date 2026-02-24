package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/modfin/clix"
	"github.com/urfave/cli/v3"

	"raz.sh/internal/web"
)

func main() {
	app := &cli.Command{
		Name:  "razsh",
		Usage: "a blog service",
		Commands: []*cli.Command{
			{
				Name: "serve",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "tls"},
					&cli.StringFlag{Name: "data-dir", Value: "./data"},
					&cli.StringSliceFlag{Name: "hostname"},
					&cli.StringFlag{Name: "http-interface", Value: ":8080"},
					&cli.StringFlag{Name: "https-interface", Value: ":8443"},
					&cli.StringFlag{Name: "ga"},
				},
				Action: serve,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func serve(ctx context.Context, cmd *cli.Command) error {
	dataDir := cmd.String("data-dir")
	dirs := []string{"tmpl", "blog", "acme", "assets"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(dataDir, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	cfg := clix.ParseCommand[Config](cmd)
	s := web.New(cfg.Web)
	err := s.Start()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	<-sigc

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = s.Stop(ctx)
	if err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}
	return nil
}

type Config struct {
	Web web.Config
}
