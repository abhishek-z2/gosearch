package main

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
	matches, err := searchFile(&buf, file, "hello", true)
	if err != nil {
		t.Fatal(err)
	}

	if matches != 2 {
		t.Fatalf("searchFile() found %d matches, want 2", matches)
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

	matches, err := searchFile(io.Discard, file, "hello", false)
	if err != nil {
		t.Fatal(err)
	}

	if matches != 0 {
		t.Errorf("searchFile() found %d matches, want 0", matches)
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

	matches, err := searchFile(io.Discard, file, "banana", false)
	if err != nil {
		t.Fatal(err)
	}

	if matches != 0 {
		t.Errorf("searchFile() found %d matches, want 0", matches)
	}
}

func TestSearchFileNotFound(t *testing.T) {
	_, err := searchFile(io.Discard, "nonexistentfile.txt", "Hello", false)

	if err == nil {
		t.Errorf("expected an error, got nil")
	}
}

func TestRun(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "sample.txt")
	err := os.WriteFile(testFile, []byte("hello world\nfoo bar"), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	tests := []struct {
		name      string
		args      []string
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "missing positional arguments",
			args:      []string{},
			wantErr:   true,
			errSubstr: "usage: gosearch",
		},
		{
			name:      "only one positional argument",
			args:      []string{"hello"},
			wantErr:   true,
			errSubstr: "usage: gosearch",
		},
		{
			name:      "non-existent file path",
			args:      []string{"hello", "nonexistent.txt"},
			wantErr:   true,
			errSubstr: "couldn't find path",
		},
		{
			name:      "directory path without recursive flag",
			args:      []string{"hello", tempDir},
			wantErr:   true,
			errSubstr: "path is a directory",
		},
		{
			name:    "valid file search",
			args:    []string{"hello", testFile},
			wantErr: false,
		},
		{
			name:    "valid directory recursive search",
			args:    []string{"-r", "hello", tempDir},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("run(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errSubstr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("run(%v) error = %q, want error containing %q", tt.args, err, tt.errSubstr)
				}
			}
		})
	}
}
