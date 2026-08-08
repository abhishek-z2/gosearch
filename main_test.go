package main

import (
	"os"
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
			got := hasExtension(tt.path, tt.extensions)

			if got != tt.want {
				t.Errorf("hasExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchFile(t *testing.T) {
	file := t.TempDir() + "/test.txt"

	content := `Hello world
This is a test
HELLO again
Nothing here`

	err := os.WriteFile(file, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	matches, err := searchFile(file, "hello", true)
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 2 {
		t.Fatalf("searchFile() found %d matches, want 2", len(matches))
	}

	if matches[0].File != file {
		t.Errorf("matches[0].File = %q, want %q", matches[0].File, file)
	}

	if matches[0].LineNumber != 1 {
		t.Errorf("matches[0].LineNumber = %d, want 1", matches[0].LineNumber)
	}

	if matches[0].Line != "Hello world" {
		t.Errorf("matches[0].Line = %q, want %q", matches[0].Line, "Hello world")
	}

	if matches[1].LineNumber != 3 {
		t.Errorf("matches[1].LineNumber = %d, want 3", matches[1].LineNumber)
	}

	if matches[1].Line != "HELLO again" {
		t.Errorf("matches[1].Line = %q, want %q", matches[1].Line, "HELLO again")
	}
}

func TestSearchFileIgnoreCase(t *testing.T) {
	file := t.TempDir() + "/test.txt"

	content := `Hello world
	This is a test
	HELLO again
	Nothing here`

	err := os.WriteFile(file, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	matches, err := searchFile(file, "hello", false)
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 0 {
		t.Errorf("searchFile() found %d matches, want 0", len(matches))
	}
}

func TestSearchFileNoMatches(t *testing.T) {
	file := t.TempDir() + "/test.txt"

	content := `Hello world
	This is a test
	HELLO again
	Nothing here`

	err := os.WriteFile(file, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	matches, err := searchFile(file, "banana", false)
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 0 {
		t.Errorf("searchFile() found %d matches, want 0", len(matches))
	}
}

func TestSearchFileNotFound(t *testing.T) {
	_, err := searchFile("nonexistentfile.txt", "Hello", false)

	if err == nil {
		t.Errorf("expected an error, got nil")
	}
}
