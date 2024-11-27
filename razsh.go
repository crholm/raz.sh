package main

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"raz.sh/internal/clix"
	"raz.sh/internal/web"
	"syscall"
	"time"
)

func main() {
	app := &cli.App{
		Name:  "razsh",
		Usage: "a blog service",
		Commands: []*cli.Command{
			{
				Name: "serve",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "tls"},
					&cli.StringSliceFlag{Name: "hostname"},
					&cli.StringFlag{Name: "data-dir", Value: "./data"},

					&cli.StringFlag{Name: "http-interface", Value: ":80"},
					&cli.StringFlag{Name: "https-interface", Value: ":443"},

					&cli.StringFlag{Name: "smtp-interface", Value: ":25"},
					&cli.StringFlag{Name: "smtp-publish-to-address", Usage: "the email address that will receive email an publish it, eg. publish@blog.example.com"},
					&cli.StringFlag{Name: "smtp-publish-from-address", Usage: "the email address that will send an email to publish it, eg. my-addess@gmail.comgit"},

					&cli.StringFlag{Name: "ga"},
				},
				Action: serve,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func serve(c *cli.Context) error {
	dataDir := c.String("data-dir")
	dirs := []string{"tmpl", "blog", "acme", "assets"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(dataDir, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	cfg := clix.Parse[Config](c)
	s := web.New(cfg.Web)
	err := s.Start()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGKILL,
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
