package web

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"golang.org/x/crypto/acme/autocert"
	"gopkg.in/yaml.v3"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type FileHeader struct {
	Title       string    `yaml:"title"`
	Slug        string    `yaml:"slug"`
	PublishDate time.Time `yaml:"publish_date"`
	Public      bool      `yaml:"public"`
}

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
	r.Get("/", renderIndex(s.cfg.DataDir, gaScript(s.cfg.GaTag)))
	r.Get("/blog/{entry}", renderEntry(s.cfg.DataDir, gaScript(s.cfg.GaTag)))

	r.Get("/blog/media/{file}", serveFiles(filepath.Join(s.cfg.DataDir, "blog", "media")))
	r.Get("/assets/{file}", serveFiles(filepath.Join(s.cfg.DataDir, "assets")))
	r.Get("/assets/img/{file}", serveFiles(filepath.Join(s.cfg.DataDir, "assets/img")))

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
	tmpl, err := template.New("entry.html.tmpl").Funcs(template.FuncMap{
		"toDate": func(t any) string {
			tt, ok := t.(time.Time)
			if !ok {
				return ""
			}
			//return tt.Format("2006-01-02")
			return tt.Format("Mon, 02 Jan 2006")
		},
	}).ParseFiles(filepath.Join(dir, "tmpl", "entry.html.tmpl"))

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

		p := bluemonday.UGCPolicy().AllowElements("nav").AddTargetBlankToFullyQualifiedLinks(true)
		htmlBody = p.SanitizeBytes(htmlBody)

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
</script>

<script>
    !function(t,e){var o,n,p,r;e.__SV||(window.posthog=e,e._i=[],e.init=function(i,s,a){function g(t,e){var o=e.split(".");2==o.length&&(t=t[o[0]],e=o[1]),t[e]=function(){t.push([e].concat(Array.prototype.slice.call(arguments,0)))}}(p=t.createElement("script")).type="text/javascript",p.crossOrigin="anonymous",p.async=!0,p.src=s.api_host.replace(".i.posthog.com","-assets.i.posthog.com")+"/static/array.js",(r=t.getElementsByTagName("script")[0]).parentNode.insertBefore(p,r);var u=e;for(void 0!==a?u=e[a]=[]:a="posthog",u.people=u.people||[],u.toString=function(t){var e="posthog";return"posthog"!==a&&(e+="."+a),t||(e+=" (stub)"),e},u.people.toString=function(){return u.toString(1)+".people (stub)"},o="init bs ws ge fs capture De Ai $s register register_once register_for_session unregister unregister_for_session Is getFeatureFlag getFeatureFlagPayload isFeatureEnabled reloadFeatureFlags updateEarlyAccessFeatureEnrollment getEarlyAccessFeatures on onFeatureFlags onSurveysLoaded onSessionId getSurveys getActiveMatchingSurveys renderSurvey canRenderSurvey canRenderSurveyAsync identify setPersonProperties group resetGroups setPersonPropertiesForFlags resetPersonPropertiesForFlags setGroupPropertiesForFlags resetGroupPropertiesForFlags reset get_distinct_id getGroups get_session_id get_session_replay_url alias set_config startSessionRecording stopSessionRecording sessionRecordingStarted captureException loadToolbar get_property getSessionProperty xs Ss createPersonProfile Es gs opt_in_capturing opt_out_capturing has_opted_in_capturing has_opted_out_capturing clear_opt_in_out_capturing ys debug ks getPageViewId captureTraceFeedback captureTraceMetric".split(" "),n=0;n<o.length;n++)g(u,o[n]);e._i.push([i,s,a])},e.__SV=1)}(document,window.posthog||[]);
    posthog.init('phc_yhgjSRCX6HQobUB51GDN4uaOWjH10T1yvrF9a9le8Yb', {
        api_host: 'https://eu.i.posthog.com',
        person_profiles: 'always',
    })
</script>


`, tag, tag)
}
