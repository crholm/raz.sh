package tmpl

import (
	"embed"
	"fmt"
	"html/template"
	"path/filepath"
	"time"
)

//go:embed _main.html.tmpl pages/*.html.tmpl
var fs embed.FS

var PAGE_INDEX = filepath.Join("pages", "index.html.tmpl")
var PAGE_ENTRY = filepath.Join("pages", "entry.html.tmpl")

var funcs = template.FuncMap{
	"format_time": func(t any, format string) string {
		tt, ok := t.(time.Time)
		if !ok {
			return ""
		}
		//return tt.Format("2006-01-02")
		//return tt.Format("Mon, 02 Jan 2006")
		return tt.Format(format)
	},
}

func BaseTemplate() (*template.Template, error) {
	return template.New("").Funcs(funcs).ParseFS(fs, "_main.html.tmpl")
}

func AddPage(t *template.Template, page string) (*template.Template, error) {
	fmt.Println("Adding page", page)
	f, err := fs.ReadFile(page)
	if err != nil {
		return nil, fmt.Errorf("failed to read page %s: %w", page, err)
	}
	return t.New(page).Parse(string(f))
}
