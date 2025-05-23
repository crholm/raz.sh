package web

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/feeds"
	"github.com/microcosm-cc/bluemonday"
	"github.com/modfin/henry/slicez"
	"github.com/russross/blackfriday/v2"
	"golang.org/x/crypto/acme/autocert"
	"gopkg.in/yaml.v3"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"raz.sh/internal/web/tmpl"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	cfg     Config
	servers []*http.Server
}

type Config struct {
	UseTLS         bool     `cli:"tls"`
	DataDir        string   `cli:"data-dir"`
	Hostname       []string `cli:"hostname"`
	HttpInterface  string   `cli:"http-interface"`
	HttpsInterface string   `cli:"https-interface"`
	GaTag          string   `cli:"ga"`
}

func New(cfg Config) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Stop(ctx context.Context) error {
	for _, server := range s.servers {
		if err := server.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}
func (s *Server) Start() error {

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", renderIndex(s.cfg.DataDir))
	r.Get("/blog.atom", renderFeed(s.cfg.DataDir, "atom"))
	r.Get("/blog.rss", renderFeed(s.cfg.DataDir, "rss"))
	r.Get("/blog.json", renderFeed(s.cfg.DataDir, "json"))
	r.Get("/blog/{entry}", renderEntry(s.cfg.DataDir))

	r.Get("/blog/media/*", serveDir(filepath.Join(s.cfg.DataDir, "blog", "media")))
	r.Get("/assets/*", serveDir(filepath.Join(s.cfg.DataDir, "assets")))

	if !s.cfg.UseTLS {
		return s.startHTTPServer(r)
	}

	return s.startHTTPSServer(r)

}

func (s *Server) startHTTPServer(handler http.Handler) error {
	server := &http.Server{
		Addr:    s.cfg.HttpInterface,
		Handler: handler,
	}
	s.servers = append(s.servers, server)

	fmt.Println("Starting HTTP server")
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()
	return nil
}

func (s *Server) startHTTPSServer(handler http.Handler) error {
	if len(s.cfg.Hostname) == 0 {
		return fmt.Errorf("hostname is required for TLS")
	}

	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(s.cfg.Hostname...),
		Cache:      autocert.DirCache(filepath.Join(s.cfg.DataDir, "acme")),
	}

	go func() {
		fmt.Println("Starting HTTP server for redirect")

		redirect := &http.Server{
			Addr:    ":80",
			Handler: certManager.HTTPHandler(nil),
		}
		s.servers = append(s.servers, redirect)
		err := redirect.ListenAndServe()
		if err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	server := &http.Server{
		Addr:    s.cfg.HttpsInterface,
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		},
	}
	s.servers = append(s.servers, server)

	fmt.Println("Starting HTTPS server")
	go func() {
		err := server.ListenAndServeTLS("", "")
		if err != nil {
			log.Printf("HTTPS server error: %v", err)
		}
	}()

	return nil
}

func serveDir(dir string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		file := chi.URLParam(r, "*")

		f, err := os.OpenInRoot(dir, file)
		defer f.Close()

		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.ServeContent(w, r, file, time.Now(), f)
	}
}

func getInfo(r *http.Request) map[string]any {
	info := map[string]any{}
	info["dark-mode"] = false
	if c, err := r.Cookie("dark-mode"); err == nil {
		info["dark-mode"], _ = strconv.ParseBool(c.Value)
	}
	return info
}

func renderFeed(dir string, _type string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		entries, err := getPublicEntries(filepath.Join(dir, "blog"))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		feed := feeds.Feed{
			Title:       "Raz Blog",
			Link:        &feeds.Link{Href: "https://raz.sh"},
			Description: "A fintech CTOs takes on tech and development",
			Author:      &feeds.Author{Name: "Rasmus Holm"},
			Created:     time.Now(),
			Updated:     time.Now(),
			Image:       &feeds.Image{Link: "https://raz.sh/assets/img/blog-banner-5-small-2.jpg"},
		}

		feed.Items = slicez.Map(entries, func(entry FileHeader) *feeds.Item {
			return &feeds.Item{
				Id:          "https://raz.sh/blog/" + entry.Slug,
				Title:       entry.Title,
				Link:        &feeds.Link{Href: "https://raz.sh/blog/" + entry.Slug},
				Description: entry.Description,
				Created:     entry.PublishDate,
				Updated:     entry.Touched,
			}
		})

		feed.Updated = slicez.Fold(entries, func(a time.Time, b FileHeader) time.Time {
			if b.Touched.After(a) {
				return b.Touched
			}
			return a
		}, time.Time{})

		switch _type {
		case "json":
			w.Header().Set("Content-Type", "application/json")
			feed.WriteJSON(w)
		case "atom":
			w.Header().Set("Content-Type", "application/atom+xml")
			feed.WriteAtom(w)
		default:
			w.Header().Set("Content-Type", "application/rss+xml")
			feed.WriteRss(w)
		}
	}
}

func renderEntry(dir string) http.HandlerFunc {

	t, err := tmpl.BaseTemplate()
	if err != nil {
		log.Printf("Error loading base template: %v", err)
		panic(fmt.Errorf("error loading base template: %v", err))
	}
	t, err = tmpl.AddPage(t, tmpl.PAGE_ENTRY)
	if err != nil {
		log.Printf("Error loading entry template: %v", err)
		panic(fmt.Errorf("error loading entry template: %v", err))
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

		p := bluemonday.UGCPolicy().AllowElements("nav").AddTargetBlankToFullyQualifiedLinks(true)
		htmlBody = p.SanitizeBytes(htmlBody)

		if err := t.Lookup(tmpl.PAGE_ENTRY).Execute(w, Page{
			Info:  getInfo(r),
			Title: header.Title + "- Raz Blog",
			Content: BlogEntry{
				FileHeader: header,
				Body:       template.HTML(htmlBody),
			},
		}); err != nil {
			log.Printf("Error rendering entry template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func renderIndex(dir string) http.HandlerFunc {

	t, err := tmpl.BaseTemplate()
	if err != nil {
		log.Printf("Error loading base template: %v", err)
		panic(fmt.Errorf("error loading base template: %v", err))
	}
	t, err = tmpl.AddPage(t, tmpl.PAGE_INDEX)
	if err != nil {
		log.Printf("Error loading index template: %v", err)
		panic(fmt.Errorf("error loading index template: %v", err))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		entries, err := getPublicEntries(filepath.Join(dir, "blog"))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = t.ExecuteTemplate(w, tmpl.PAGE_INDEX, Page{
			Info:    getInfo(r),
			Title:   "Raz Blog",
			Content: entries,
		})
		if err != nil {
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

	stat, err := os.Stat(file)
	if err != nil {
		return FileHeader{}, fmt.Errorf("failed to get file info: %w", err)
	}
	header.Touched = stat.ModTime()
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
