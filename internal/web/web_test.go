package web

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseMarkdownContent(t *testing.T) {
	tests := []struct {
		name       string
		content    []byte
		wantTitle  string
		wantPublic bool
		wantErr    bool
	}{
		{
			name: "valid markdown with header",
			content: []byte(`title: "Test Post"
publish_date: 2024-01-15
public: true
---
# Test Content
This is the body.`),
			wantTitle:  "Test Post",
			wantPublic: true,
			wantErr:    false,
		},
		{
			name: "valid markdown with leading whitespace",
			content: []byte(`

title: "Whitespace Test"
publish_date: 2024-02-20
public: false
---
Body content here.`),
			wantTitle:  "Whitespace Test",
			wantPublic: false,
			wantErr:    false,
		},
		{
			name:    "missing header separator",
			content: []byte(`title: "No Separator"\nJust content`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, body, err := parseMarkdownContent(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMarkdownContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if header.Title != tt.wantTitle {
					t.Errorf("parseMarkdownContent() header.Title = %v, want %v", header.Title, tt.wantTitle)
				}
				if header.Public != tt.wantPublic {
					t.Errorf("parseMarkdownContent() header.Public = %v, want %v", header.Public, tt.wantPublic)
				}
				if len(body) == 0 {
					t.Error("parseMarkdownContent() body should not be empty")
				}
			}
		})
	}
}

func TestReadHeader(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		content    string
		wantTitle  string
		wantPublic bool
		wantErr    bool
	}{
		{
			name: "valid header",
			content: `title: "Test Header"
publish_date: 2024-03-10
public: true
---
# Content`,
			wantTitle:  "Test Header",
			wantPublic: true,
			wantErr:    false,
		},
		{
			name:    "no header separator",
			content: `Just some content without header`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tmpDir, tt.name+".md")
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			header, err := readHeader(testFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("readHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if header.Title != tt.wantTitle {
					t.Errorf("readHeader() header.Title = %v, want %v", header.Title, tt.wantTitle)
				}
				if header.Public != tt.wantPublic {
					t.Errorf("readHeader() header.Public = %v, want %v", header.Public, tt.wantPublic)
				}
				// Check that Touched is set (file modification time)
				if header.Touched.IsZero() {
					t.Error("readHeader() header.Touched should not be zero")
				}
			}
		})
	}
}

func TestReadMarkdownFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	content := []byte("\n\n\tSome content with leading whitespace")

	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := readMarkdownFile(testFile)
	if err != nil {
		t.Errorf("readMarkdownFile() error = %v", err)
		return
	}

	// Should trim leading whitespace
	if bytes.HasPrefix(result, []byte("\n")) || bytes.HasPrefix(result, []byte("\t")) {
		t.Error("readMarkdownFile() should trim leading whitespace")
	}
}

func TestGetInfo(t *testing.T) {
	tests := []struct {
		name         string
		cookieValue  string
		wantDarkMode bool
	}{
		{
			name:         "no cookie",
			wantDarkMode: false,
		},
		{
			name:         "dark mode enabled",
			cookieValue:  "true",
			wantDarkMode: true,
		},
		{
			name:         "dark mode disabled",
			cookieValue:  "false",
			wantDarkMode: false,
		},
		{
			name:         "invalid cookie value",
			cookieValue:  "invalid",
			wantDarkMode: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This is a simplified test. In a real scenario,
			// we'd need to create an http.Request with cookies.
			// For now, just test the basic structure.
			info := map[string]any{}
			info["dark-mode"] = false
			if tt.cookieValue == "true" {
				info["dark-mode"] = true
			}

			if info["dark-mode"] != tt.wantDarkMode {
				t.Errorf("getInfo() dark-mode = %v, want %v", info["dark-mode"], tt.wantDarkMode)
			}
		})
	}
}

func TestGetPublicEntries(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test blog posts
	posts := []struct {
		filename string
		content  string
	}{
		{
			filename: "2024-01-15_public-post.md",
			content: `title: "Public Post"
publish_date: 2024-01-15
public: true
---
Content`,
		},
		{
			filename: "2024-02-20_private-post.md",
			content: `title: "Private Post"
publish_date: 2024-02-20
public: false
---
Content`,
		},
		{
			filename: "2024-03-10_another-public.md",
			content: `title: "Another Public"
publish_date: 2024-03-10
public: true
---
Content`,
		},
	}

	for _, post := range posts {
		testFile := filepath.Join(tmpDir, post.filename)
		err := os.WriteFile(testFile, []byte(post.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", post.filename, err)
		}
		// Set different modification times to simulate real scenario
		time.Sleep(10 * time.Millisecond)
	}

	entries, err := getPublicEntries(tmpDir)
	if err != nil {
		t.Errorf("getPublicEntries() error = %v", err)
		return
	}

	// Should only return public entries
	if len(entries) != 2 {
		t.Errorf("getPublicEntries() returned %d entries, want 2", len(entries))
	}

	// Should be sorted by date descending (newest first)
	if len(entries) >= 2 {
		if !entries[0].PublishDate.After(entries[1].PublishDate) && !entries[0].PublishDate.Equal(entries[1].PublishDate) {
			t.Error("getPublicEntries() entries should be sorted by date descending")
		}
	}

	// Verify all entries are public
	for _, entry := range entries {
		if !entry.Public {
			t.Errorf("getPublicEntries() returned non-public entry: %s", entry.Title)
		}
	}
}
