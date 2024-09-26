package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/russross/blackfriday/v2"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/acme/autocert"
	"gopkg.in/yaml.v3"
)

type FileHeader struct {
	Title       string    `yaml:"title"`
	Slug        string    `yaml:"slug"`
	PublishDate time.Time `yaml:"publish_date"`
	Public      bool      `yaml:"public"`
}

func main() {
	app := &cli.App{
		Name:  "razsh",
		Usage: "a blog service",
		Commands: []*cli.Command{
			{
				Name: "serve",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "http-interface", Value: ":80"},
					&cli.StringFlag{Name: "https-interface", Value: ":443"},
					&cli.BoolFlag{Name: "tls"},
					&cli.StringSliceFlag{Name: "hostname"},
					&cli.BoolFlag{Name: "verbose"},
					&cli.StringFlag{Name: "data-dir", Value: "./data"},
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

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", renderIndex(dataDir, gaScript(c.String("ga"))))
	r.Get("/blog/{entry}", renderEntry(dataDir, gaScript(c.String("ga"))))
	r.Get("/blog/media/{file}", serveFiles(filepath.Join(dataDir, "blog", "media")))
	r.Get("/assets/{file}", serveFiles(filepath.Join(dataDir, "assets")))

	if !c.Bool("tls") {
		return startHTTPServer(c.String("http-interface"), r, c.Bool("verbose"))
	}

	return startHTTPSServer(c, r, dataDir)
}

func startHTTPServer(addr string, handler http.Handler, verbose bool) error {
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	fmt.Println("Starting HTTP server")
	if verbose {
		server.ErrorLog = log.New(os.Stderr, "http: ", log.LstdFlags)
	}
	return server.ListenAndServe()
}

func startHTTPSServer(c *cli.Context, handler http.Handler, dataDir string) error {
	hostnames := c.StringSlice("hostname")
	if len(hostnames) == 0 {
		return fmt.Errorf("hostname is required for TLS")
	}

	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(hostnames...),
		Cache:      autocert.DirCache(filepath.Join(dataDir, "acme")),
	}

	go func() {
		fmt.Println("Starting HTTP server for redirect")
		if err := http.ListenAndServe(":80", certManager.HTTPHandler(nil)); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	server := &http.Server{
		Addr:    c.String("https-interface"),
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		},
	}

	fmt.Println("Starting HTTPS server")
	if c.Bool("verbose") {
		server.ErrorLog = log.New(os.Stderr, "[https] ", log.LstdFlags)
	}
	return server.ListenAndServeTLS("", "")
}

func serveFiles(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file := chi.URLParam(r, "file")
		if err := serveFile(w, r, dir, file); err != nil {
			log.Printf("Error serving file: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func serveFile(w http.ResponseWriter, r *http.Request, dir, file string) error {
	path, err := filepath.Abs(filepath.Join(dir, file))
	if err != nil {
		return err
	}

	base, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(path, base) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil
	}

	http.ServeFile(w, r, path)
	return nil
}

func renderEntry(dir, ga string) http.HandlerFunc {
	tmpl, err := template.ParseFiles(filepath.Join(dir, "tmpl", "entry.html.tmpl"))
	if err != nil {
		log.Fatalf("Failed to parse entry template: %v", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		item := chi.URLParam(r, "entry")
		entry, err := readMarkdownFile(filepath.Join(dir, "blog", item+".md"))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		header, body, err := parseMarkdownContent(entry)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !header.Public {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		htmlBody := blackfriday.Run(body,
			blackfriday.WithExtensions(blackfriday.CommonExtensions|blackfriday.AutoHeadingIDs),
			blackfriday.WithRenderer(blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{Flags: blackfriday.TOC})),
		)

		if err := tmpl.Execute(w, struct {
			FileHeader
			Body template.HTML
			GA   template.HTML
		}{
			FileHeader: header,
			Body:       template.HTML(htmlBody),
			GA:         template.HTML(ga),
		}); err != nil {
			log.Printf("Error rendering entry template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func renderIndex(dir, ga string) http.HandlerFunc {
	tmpl := template.Must(template.New("index.html.tmpl").Funcs(template.FuncMap{
		"toDate": func(t any) string {
			tt, ok := t.(time.Time)
			if !ok {
				return ""
			}
			return tt.Format("2006-01-02")
		},
	}).ParseFiles(filepath.Join(dir, "tmpl", "index.html.tmpl")))

	return func(w http.ResponseWriter, r *http.Request) {
		entries, err := getPublicEntries(filepath.Join(dir, "blog"))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, struct {
			Items []FileHeader
			GA    template.HTML
		}{Items: entries, GA: template.HTML(ga)}); err != nil {
			log.Printf("Error rendering index template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func getPublicEntries(dir string) ([]FileHeader, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil, err
	}

	var headers []FileHeader
	for _, file := range files {
		h, err := readHeader(file)
		if err != nil {
			log.Printf("Error reading header for %s: %v", file, err)
			continue
		}
		if !h.Public {
			continue
		}
		h.Slug = strings.TrimSuffix(filepath.Base(file), ".md")
		headers = append(headers, h)
	}

	// Sort headers by PublishDate in descending order
	sort.Slice(headers, func(i, j int) bool {
		return headers[i].PublishDate.After(headers[j].PublishDate)
	})

	// Take the latest 20 entries
	if len(headers) > 20 {
		headers = headers[:20]
	}

	return headers, nil
}

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

	var header FileHeader
	if err := yaml.Unmarshal(headerBytes, &header); err != nil {
		return FileHeader{}, err
	}

	return header, nil
}

func readMarkdownFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return bytes.TrimLeft(content, "\t\n\r- "), nil
}

func parseMarkdownContent(content []byte) (FileHeader, []byte, error) {
	headerBytes, bodyBytes, found := bytes.Cut(content, []byte("---\n"))
	if !found {
		return FileHeader{}, nil, fmt.Errorf("could not find header")
	}

	var header FileHeader
	if err := yaml.Unmarshal(headerBytes, &header); err != nil {
		return FileHeader{}, nil, fmt.Errorf("could not parse header: %w", err)
	}

	return header, bodyBytes, nil
}

func gaScript(tag string) string {
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
