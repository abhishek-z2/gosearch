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

	found, err := searchFile(file, "hello", true)
	if err != nil {
		t.Fatal(err)
	}

	if found != 2 {
		t.Errorf("searchFile() found %d matches, want 2", found)
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

	found, err := searchFile(file, "hello", false)
	if err != nil {
		t.Fatal(err)
	}

	if found != 0 {
		t.Errorf("searchFile() found %d matches, want 0", found)
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

	found, err := searchFile(file, "banana", false)
	if err != nil {
		t.Fatal(err)
	}

	if found != 0 {
		t.Errorf("searchFile() found %d matches, want 0", found)
	}
}

func TestSearchFileNotFound(t *testing.T) {
	_, err := searchFile("nonexistentfile.txt", "Hello", false)

	if err == nil {
		t.Errorf("expected an error, got nil")
	}
}
