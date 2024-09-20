package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/acme/autocert"
	"gopkg.in/yaml.v3"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

func main() {

	app := cli.App{
		Name:  "razsh",
		Usage: "a blog service",
		Commands: []*cli.Command{
			{
				Name: "serve",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "http-interface",
						Value: ":80",
					},
					&cli.StringFlag{
						Name:  "https-interface",
						Value: ":443",
					},
					&cli.BoolFlag{
						Name: "tls",
					},
					&cli.StringSliceFlag{
						Name: "hostname",
					},
					&cli.StringFlag{
						Name:  "data-dir",
						Value: "./data",
					},
				},
				Action: serve,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func serve(c *cli.Context) error {

	dataDir := c.String("data-dir")
	check(os.MkdirAll(filepath.Join(dataDir, "tmpl"), 0755))
	check(os.MkdirAll(filepath.Join(dataDir, "blog"), 0755))
	check(os.MkdirAll(filepath.Join(dataDir, "acme"), 0755))
	check(os.MkdirAll(filepath.Join(dataDir, "assets"), 0755))

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", renderIndex(dataDir))
	r.Get("/blog/{entry}", renderEntry(dataDir))

	server := &http.Server{
		Addr:    c.String("http-interface"),
		Handler: r,
	}
	if c.Bool("tls") {

		if c.String("hostname") == "" {
			log.Fatal("hostname is required")
		}

		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(c.StringSlice("hostname")...),
			Cache:      autocert.DirCache(filepath.Join(c.String("data-dir"), "acme")),
		}

		server.Addr = c.String("https-interface")
		server.TLSConfig = &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12, // improves cert reputation score at https://www.ssllabs.com/ssltest/
		}
	}

	server.Handler = r
	return server.ListenAndServe()
}

func renderEntry(dir string) http.HandlerFunc {

	tmpl, err := template.ParseFiles(filepath.Join(dir, "tmpl", "entry.html.tmpl"))
	if err != nil {
		log.Fatal(fmt.Errorf("file %s must exist, %w", filepath.Join(dir, "tmpl", "entry.html.tmpl"), err))
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		item := chi.URLParam(request, "entry")

		basePath, err := filepath.Abs(filepath.Join(dir, "blog"))
		if err != nil {
			log.Println(err)
			http.NotFound(writer, request)
			return
		}
		//Potential file traversal bugg....
		requestedPath, err := filepath.Abs(filepath.Join(basePath, item+".md"))
		if err != nil {
			log.Println(err)
			http.NotFound(writer, request)
			return
		}
		if !strings.HasPrefix(requestedPath, basePath) {
			http.Error(writer, "Forbidden", http.StatusForbidden)
			return
		}
		entry, err := os.ReadFile(requestedPath)
		if err != nil {
			log.Println(err)
			http.Error(writer, "could not read file", http.StatusInternalServerError)
			return
		}

		entry = bytes.TrimLeft(entry, "\t\n\r- ")

		headerBytes, bodyBytes, found := bytes.Cut(entry, []byte("---\n"))
		if !found {
			log.Println("could not find header ", requestedPath)
			http.Error(writer, "could not find header", http.StatusInternalServerError)
			return
		}

		header := FileHeader{}
		err = yaml.Unmarshal(headerBytes, &header)
		if err != nil {
			log.Println("could not parse header of ", requestedPath, " ", err)
			http.Error(writer, "could not parse header", http.StatusInternalServerError)
			return
		}
		if !header.Public {
			http.Error(writer, "Forbidden", http.StatusForbidden)
			return
		}
		err = tmpl.Execute(writer, struct {
			FileHeader
			Body string
		}{
			FileHeader: header,
			Body:       string(bodyBytes),
		},
		)

		if err != nil {
			log.Println("could not render template ", err)
			http.Error(writer, "could not render template", http.StatusInternalServerError)
			return
		}

	}
}

func renderIndex(dir string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

	}

}

type FileHeader struct {
	Title       string    `yaml:"title"`
	PublishDate time.Time `yaml:"publish_date"`
	Public      bool      `yaml:"public"`
}
