package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
