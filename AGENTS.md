# AGENTS.md — raz.sh Codebase Guide

This file is intended for agentic coding tools (Claude Code, Copilot, Cursor, etc.) operating in
this repository. It documents build commands, testing, and code style conventions.

---

## Project Overview

`raz.sh` is a personal blog/website server written in Go. It serves Markdown blog posts (with YAML
frontmatter) via a Chi HTTP router, renders them with embedded Go HTML templates, and is deployed as
a single self-contained binary. There is no Node.js, TypeScript, or build pipeline beyond standard
Go tooling.

- **Language**: Go 1.24
- **Router**: `github.com/go-chi/chi/v5`
- **CLI**: `github.com/urfave/cli/v2`
- **Templates**: Embedded via `//go:embed` into the binary at compile time
- **Content**: Blog posts live in `data/blog/*.md` (YAML frontmatter + Markdown body)

---

## Build & Run Commands

### Local Development

```bash
# Run locally in development mode (port 8080)
go run razsh.go serve --http-interface=:8080

# Run with hot-reload using Air (requires Air installed locally)
air

# Build a local binary
go build razsh.go

# Cross-compile for Linux/amd64 (production target)
GOOS=linux GOARCH=amd64 go build razsh.go

# Build and deploy to production server
./helper.build-deploy.sh

# Sync blog post Markdown files to server
./helper.publish.sh

# Run the production binary with TLS
./razsh serve --data-dir=./data --hostname=raz.sh --tls
```

### Docker Development

```bash
# Start development server with live reload (uses Air)
docker-compose up app-dev

# Start production server
docker-compose up app-prod

# Build and run in detached mode
docker-compose up -d app-dev

# View logs
docker-compose logs -f app-dev

# Stop all containers
docker-compose down

# Build fresh images
docker-compose build --no-cache
```

---

## Testing

```bash
# Run all tests
go test ./...

# Run a single test by name
go test ./... -run TestFunctionName

# Run tests in a specific package
go test ./internal/web/...

# Run with verbose output
go test -v ./...
```

Test files follow Go convention: place `foo_test.go` adjacent to `foo.go`, use `package foo` (or
`package foo_test` for black-box tests).

---

## Linting & Formatting

No linter config files are present; use standard Go tooling:

```bash
# Format all Go source files (mandatory before committing)
gofmt -w .

# Or use goimports (also organizes imports)
goimports -w .

# Vet for common mistakes
go vet ./...
```

---

## Project Structure

```
raz.sh/
├── razsh.go                  # Main entry point — CLI app definition and serve command
├── go.mod / go.sum           # Go module and dependency checksums
├── .air.toml                 # Air live reload configuration
├── Dockerfile                # Production Docker image (Alpine)
├── Dockerfile.dev            # Development Docker image with Air
├── docker-compose.yml        # Docker Compose with dev/prod services
├── helper.build-deploy.sh    # Cross-compile + SCP + systemd restart
├── helper.publish.sh         # rsync data/blog/ to production server
├── internal/
│   ├── clix/
│   │   └── clix.go           # Generic CLI flag → struct parser using reflection
│   └── web/
│       ├── web.go            # HTTP server, routes, handler closures, Markdown rendering
│       ├── web_test.go       # Test suite for web package
│       ├── model.go          # Data types: Page, FileHeader, BlogEntry
│       └── tmpl/
│           ├── tmpl.go       # Embedded FS template loading + FuncMap
│           ├── _main.html.tmpl
│           └── pages/
│               ├── index.html.tmpl
│               └── entry.html.tmpl
└── data/
    ├── blog/*.md             # Blog posts (YYYY-MM-DD_slug-title.md convention)
    ├── blog/media/           # Images for blog posts
    └── assets/              # CSS, JS, images served from disk at runtime
```

---

## Code Style Guidelines

### Imports

Group imports in three blocks separated by blank lines: stdlib → third-party → internal. Use
`goimports` to manage this automatically.

```go
import (
    "context"
    "fmt"
    "log"

    "github.com/go-chi/chi/v5"
    "github.com/urfave/cli/v2"

    "github.com/razsh/internal/web"
)
```

### Naming Conventions

- Use short, expressive variable names in function scope: `r`, `w`, `s`, `t`, `cfg`, `dir`, `err`
- Exported types/functions use `PascalCase`; unexported use `camelCase`
- Package-level constants/vars that act as named constants use `SCREAMING_SNAKE_CASE`
  (e.g., `PAGE_INDEX`, `PAGE_ENTRY`) — consistent with existing code
- Struct fields that map to external formats use struct tags: `yaml:"field_name"`, `cli:"flag-name"`
- Exclude fields from serialization explicitly: `yaml:"-"`
- File names are lowercase with no separators (e.g., `web.go`, `model.go`, `clix.go`)
- Blog post files: `YYYY-MM-DD_slug-title.md`

### Types & Structs

- Define shared data types in a dedicated `model.go` file within the package
- Prefer struct embedding for composition over deep nesting:
  ```go
  type BlogEntry struct {
      FileHeader
      Content any
  }
  ```
- Use `any` (not `interface{}`) for truly dynamic content — Go 1.18+ convention
- Use `template.HTML` to explicitly mark pre-sanitized HTML strings safe for rendering
- Config structs are defined per-package and composed via embedding at the top level:
  ```go
  type Config struct {
      web.Config
      DataDir string `cli:"data-dir"`
  }
  ```

### Functions & Methods

- Constructor pattern: `func New(cfg Config) *Server`
- Methods on pointer receivers: `func (s *Server) Start() error`
- Handler functions are closures that capture config/templates at startup (not per-request):
  ```go
  func renderIndex(dir string, t *template.Template) http.HandlerFunc {
      return func(w http.ResponseWriter, r *http.Request) {
          // handler body
      }
  }
  ```
- Avoid global mutable state; pass dependencies explicitly via function parameters or struct fields

### Error Handling

- Always check errors immediately; never ignore them:
  ```go
  result, err := doSomething()
  if err != nil {
      return fmt.Errorf("doSomething: %w", err)
  }
  ```
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Fatal startup errors (e.g., template load failure) use `panic` — the server cannot function
  without them
- HTTP handler errors: respond with `http.Error` + log with `log.Printf`:
  ```go
  log.Printf("render index: %v", err)
  http.Error(w, "Internal Server Error", http.StatusInternalServerError)
  return
  ```
- Top-level fatal CLI errors: `log.Fatal(err)` or `log.Fatalf("context: %v", err)`
- No custom error types — use plain `errors.New` / `fmt.Errorf` with `%w` wrapping

### Templates

- Templates are embedded into the binary at compile time via `//go:embed`:
  ```go
  //go:embed pages _main.html.tmpl
  var tmplFS embed.FS
  ```
- Parse templates once at startup, not per-request
- Register custom functions via `template.FuncMap` before parsing
- Page name constants are package-level `var`s using `filepath.Join` for portability

### Concurrency & Lifecycle

- Graceful shutdown: use `context.WithTimeout` and listen for OS signals
  (`syscall.SIGTERM`, `os.Interrupt`) on a buffered channel
- Avoid goroutines for request handling (Chi handles this); use goroutines only for background tasks
  like the server's `ListenAndServe`

### Shell Scripts

- All helper scripts use `set -e` (fail-fast on any error)
- Use descriptive `echo` statements for progress logging
- Prefer simple, direct `scp`/`rsync`/`ssh` invocations over abstraction layers

---

## Blog Post Format

```markdown
---
title: "Post Title"
date: 2024-01-15
tags: [go, web]
description: "Short description for SEO/preview"
---

Post body in standard Markdown...
```

File naming: `YYYY-MM-DD_slug-title.md` in `data/blog/`.

---

## Key Dependencies

| Package | Purpose |
|---|---|
| `github.com/go-chi/chi/v5` | HTTP router + middleware |
| `github.com/urfave/cli/v2` | CLI flag parsing |
| `github.com/yuin/goldmark` | Markdown → HTML rendering |
| `gopkg.in/yaml.v3` | YAML frontmatter parsing |
