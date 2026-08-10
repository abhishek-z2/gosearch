package search

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHasExtension(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		extensions []string
		want       bool
	}{
		{
			name:       "go file matches",
			path:       "main.go",
			extensions: []string{"go"},
			want:       true,
		},
		{
			name:       "js file does not match go",
			path:       "app.js",
			extensions: []string{"go"},
			want:       false,
		},
		{
			name:       "multiple extensions",
			path:       "app.js",
			extensions: []string{"go", "js", "ts"},
			want:       true,
		},
		{
			name:       "no extensions means everything matches",
			path:       "main.go",
			extensions: []string{},
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Unexported helper inside package search: stays lowercase!
			got := hasExtension(tt.path, tt.extensions)

			if got != tt.want {
				t.Errorf("hasExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "test.txt")

	content := `Hello world
This is a test
HELLO again
Nothing here`

	err := os.WriteFile(file, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	// Call exported SearchFile
	matches, err := SearchFile(&buf, file, "hello", true)
	if err != nil {
		t.Fatal(err)
	}

	if matches != 2 {
		t.Fatalf("SearchFile() found %d matches, want 2", matches)
	}

	out := buf.String()
	if !strings.Contains(out, "Hello world") || !strings.Contains(out, "HELLO again") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestSearchFileCaseSensitive(t *testing.T) {
	file := filepath.Join(t.TempDir(), "test.txt")

	content := `Hello world
This is a test
HELLO again
Nothing here`

	err := os.WriteFile(file, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	matches, err := SearchFile(io.Discard, file, "hello", false)
	if err != nil {
		t.Fatal(err)
	}

	if matches != 0 {
		t.Errorf("SearchFile() found %d matches, want 0", matches)
	}
}

func TestSearchFileNoMatches(t *testing.T) {
	file := filepath.Join(t.TempDir(), "test.txt")

	content := `Hello world
This is a test
HELLO again
Nothing here`

	err := os.WriteFile(file, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	matches, err := SearchFile(io.Discard, file, "banana", false)
	if err != nil {
		t.Fatal(err)
	}

	if matches != 0 {
		t.Errorf("SearchFile() found %d matches, want 0", matches)
	}
}

func TestSearchFileNotFound(t *testing.T) {
	_, err := SearchFile(io.Discard, "nonexistentfile.txt", "Hello", false)

	if err == nil {
		t.Errorf("expected an error, got nil")
	}
}
