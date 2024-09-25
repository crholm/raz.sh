package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/modfin/henry/slicez"
	"github.com/russross/blackfriday/v2"
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
	// Define the CLI application
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
					&cli.BoolFlag{
						Name: "verbose",
					},
					&cli.StringFlag{
						Name:  "data-dir",
						Value: "./data",
					},
					&cli.StringFlag{
						Name: "ga",
					},
				},
				Action: serve,
			},
		},
	}

	// Run the CLI application
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// Helper function to check for errors
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Struct to hold the file header information
type FileHeader struct {
	Title       string    `yaml:"title"`
	Slug        string    `yaml:"slug"`
	PublishDate time.Time `yaml:"publish_date"`
	Public      bool      `yaml:"public"`
}

// Function to serve the blog
func serve(c *cli.Context) error {
	dataDir := c.String("data-dir")

	// Create the data directory

	// Create the tmpl directory for the go templates, expected to contain index.html.tmpl and entry.html.tmpl
	check(os.MkdirAll(filepath.Join(dataDir, "tmpl"), 0755))

	// Create the blog directory for the markdown files
	check(os.MkdirAll(filepath.Join(dataDir, "blog"), 0755))

	// Create the acme directory, for Let's Encrypt
	check(os.MkdirAll(filepath.Join(dataDir, "acme"), 0755))

	// Create the assets directory, for static files such as CSS and Js
	check(os.MkdirAll(filepath.Join(dataDir, "assets"), 0755))

	// Create a new router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", renderIndex(dataDir, ga(c.String("ga"))))
	r.Get("/blog/{entry}", renderEntry(dataDir, ga(c.String("ga"))))
	r.Get("/blog/media/{file}", assets(filepath.Join(dataDir, "blog", "media")))
	r.Get("/assets/{file}", assets(filepath.Join(dataDir, "assets")))

	// If TLS is not enabled, start a regular HTTP server
	if !c.Bool("tls") {
		server := &http.Server{
			Addr:    c.String("http-interface"),
			Handler: r,
		}

		fmt.Println("Starting http server")
		if c.Bool("verbose") {
			server.ErrorLog = log.New(os.Stderr, "http: ", log.LstdFlags)
		}
		return server.ListenAndServe()
	}

	// If TLS is enabled, start an HTTPS server and a cert manager for Let's Encrypt
	if c.Bool("tls") {
		fmt.Println("Setting tls @", c.String("https-interface"), "for", c.String("hostname"))

		if c.String("hostname") == "" {
			log.Fatal("hostname is required")
		}

		// Create a new cert manager using Let's Encrypt and provide the hostname
		// storing the certificates in the acme directory
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(c.StringSlice("hostname")...),
			Cache:      autocert.DirCache(filepath.Join(c.String("data-dir"), "acme")),
		}

		// Start an HTTP server to redirect HTTP requests to HTTPS and used to serve the Let's Encrypt challenge
		go func() {
			fmt.Println("Starting http server for redirect")
			http.ListenAndServe(":80", certManager.HTTPHandler(nil))
		}()

		server := &http.Server{
			Addr:    c.String("https-interface"),
			Handler: r,
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
				MinVersion:     tls.VersionTLS12,
			},
		}

		fmt.Println("Starting https server")
		if c.Bool("verbose") {
			server.ErrorLog = log.New(os.Stderr, "[https] ", log.LstdFlags)
		}
		return server.ListenAndServeTLS("", "")
	}
	return nil
}

// Function to serve static assets
func assets(dir string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		item := chi.URLParam(request, "file")

		file, err := filepath.Abs(filepath.Join(dir, item))
		if err != nil {
			log.Println(err)
			http.NotFound(writer, request)
			return
		}
		base, err := filepath.Abs(dir)
		if err != nil {
			log.Println(err)
			http.NotFound(writer, request)
			return
		}
		// Check if the file is within the base directory
		// inorder to prevent directory traversal, hopefully...
		if !strings.HasPrefix(file, base) {
			http.Error(writer, "Forbidden", http.StatusForbidden)
			return
		}

		http.ServeFile(writer, request, filepath.Join(dir, item))
	}
}

// Function to render a blog entry
func renderEntry(dir string, ga string) http.HandlerFunc {

	// Parse the entry.html.tmpl file to be used as a template
	tmpl, err := template.ParseFiles(filepath.Join(dir, "tmpl", "entry.html.tmpl"))
	if err != nil {
		log.Fatal(fmt.Errorf("file %s must exist, %w", filepath.Join(dir, "tmpl", "entry.html.tmpl"), err))
	}

	dir = filepath.Join(dir, "blog")
	return func(writer http.ResponseWriter, request *http.Request) {
		item := chi.URLParam(request, "entry")

		basePath, err := filepath.Abs(dir)
		if err != nil {
			log.Println(err)
			http.NotFound(writer, request)
			return
		}
		requestedPath, err := filepath.Abs(filepath.Join(basePath, item+".md"))
		if err != nil {
			log.Println(err)
			http.NotFound(writer, request)
			return
		}
		// Check if the file is within the base directory
		// inorder to prevent directory traversal, hopefully...
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

		//render toc and body

		htmlBody := blackfriday.Run(bodyBytes,
			blackfriday.WithExtensions(blackfriday.CommonExtensions|blackfriday.AutoHeadingIDs),
			blackfriday.WithRenderer(blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{Flags: blackfriday.TOC})),
		)

		err = tmpl.Execute(writer, struct {
			FileHeader
			Body string
			GA   string
		}{
			FileHeader: header,
			Body:       string(htmlBody),
			GA:         ga,
		},
		)

		if err != nil {
			log.Println("could not render template ", err)
			http.Error(writer, "could not render template", http.StatusInternalServerError)
			return
		}
	}
}

// Function to render the index page
func renderIndex(dir string, ga string) http.HandlerFunc {
	tmpl := template.New("index.html.tmpl").Funcs(template.FuncMap{
		"toDate": func(t any) string {
			tt, ok := t.(time.Time)
			if !ok {
				return ""
			}
			return tt.Format("2006-01-02")
		},
	})

	tmpl, err := tmpl.ParseFiles(filepath.Join(dir, "tmpl", "index.html.tmpl"))
	if err != nil {
		log.Fatal(fmt.Errorf("file %s must exist, %w", filepath.Join(dir, "tmpl", "index.html.tmpl"), err))
	}

	dir = filepath.Join(dir, "blog")

	return func(writer http.ResponseWriter, request *http.Request) {
		var mdFiles []string

		entries, err := os.ReadDir(dir)
		if err != nil {
			log.Println(err)
			http.Error(writer, "could not read dir", http.StatusInternalServerError)
			return
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				mdFiles = append(mdFiles, filepath.Join(dir, entry.Name()))
			}
		}
		var headers []FileHeader
		for _, file := range slicez.Take(slicez.Reverse(slicez.Sort(mdFiles)), 20) {
			fmt.Println(file)
			h, err := readHeader(file)
			if err != nil {
				log.Println("could not read header for ", file, " ", err)
				continue
			}
			if !h.Public {
				continue
			}
			h.Slug = strings.TrimSuffix(filepath.Base(file), ".md")
			headers = append(headers, h)
		}

		err = tmpl.Execute(writer, struct {
			Items []FileHeader
			GA    string
		}{Items: headers, GA: ga})

		if err != nil {
			log.Println("could not render template ", err)
		}
	}
}

// Function to read the yaml header of a markdown file
func readHeader(file string) (FileHeader, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return FileHeader{}, err
	}
	data = bytes.TrimLeft(data, "\t\n\r- ")
	headerBytes, _, found := bytes.Cut(data, []byte("---\n"))
	if !found {
		return FileHeader{}, fmt.Errorf("could not find header")
	}

	header := FileHeader{}
	err = yaml.Unmarshal(headerBytes, &header)
	if err != nil {
		return FileHeader{}, err
	}

	return header, nil
}

func ga(tag string) string {
	if tag == "" {
		return ""
	}
	return fmt.Sprintf(`
<script async src="https://www.googletagmanager.com/gtag/js?id=%s"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());

  gtag('config', '%s');
</script>`, tag, tag)
}
