package web

import (
	"html/template"
	"time"
)

type Page struct {
	Title   string
	Content any
}

type FileHeader struct {
	Title       string    `yaml:"title"`
	Slug        string    `yaml:"slug"`
	PublishDate time.Time `yaml:"publish_date"`
	Public      bool      `yaml:"public"`
}
type BlogEntry struct {
	FileHeader
	Body template.HTML
}
