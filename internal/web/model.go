package web

import (
	"html/template"
	"time"
)

type Page struct {
	Info    map[string]any
	Title   string
	Content any
}

type FileHeader struct {
	Title       string    `yaml:"title"`
	Slug        string    `yaml:"slug"`
	Image       string    `yaml:"image"`
	Description string    `yaml:"description"`
	Touched     time.Time `yaml:"-"`
	PublishDate time.Time `yaml:"publish_date"`
	Public      bool      `yaml:"public"`
}
type BlogEntry struct {
	FileHeader
	Body template.HTML
}
