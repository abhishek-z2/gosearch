package search

import (
	"bytes"
	"fmt"
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
	// Updated: pass Options struct with CaseInsensitive: true
	opts := Options{
		CaseInsensitive: true,
	}

	matches, err := SearchFile(&buf, file, "hello", opts)
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

	// Default Options (CaseInsensitive: false)
	opts := Options{}

	matches, err := SearchFile(io.Discard, file, "hello", opts)
	if err != nil {
		t.Fatal(err)
	}

	if matches != 0 {
		t.Errorf("SearchFile() found %d matches, want 0", matches)
	}
}

func TestSearchFileInvertMatch(t *testing.T) {
	file := filepath.Join(t.TempDir(), "test.txt")

	content := `apple
	banana
	cherry`

	err := os.WriteFile(file, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	opts := Options{
		InvertMatch: true,
	}

	matches, err := SearchFile(&buf, file, "banana", opts)
	if err != nil {
		t.Fatal(err)
	}

	// Out of 3 lines, 2 do not contain "banana"
	if matches != 2 {
		t.Fatalf("SearchFile() found %d matches, want 2", matches)
	}

	out := buf.String()
	if !strings.Contains(out, "apple") || !strings.Contains(out, "cherry") {
		t.Errorf("unexpected inverted match output: %q", out)
	}
	if strings.Contains(out, "banana") {
		t.Errorf("inverted output should not contain matched term 'banana'")
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

	opts := Options{}

	matches, err := SearchFile(io.Discard, file, "banana", opts)
	if err != nil {
		t.Fatal(err)
	}

	if matches != 0 {
		t.Errorf("SearchFile() found %d matches, want 0", matches)
	}
}

func TestSearchFileNotFound(t *testing.T) {
	opts := Options{}
	_, err := SearchFile(io.Discard, "nonexistentfile.txt", "Hello", opts)

	if err == nil {
		t.Errorf("expected an error, got nil")
	}
}

func TestIsBinary(t *testing.T) {
	tempDir := t.TempDir()

	textPath := filepath.Join(tempDir, "sample.txt")
	err := os.WriteFile(textPath, []byte("hello world\nthis is text"), 0644)
	if err != nil {
		t.Fatalf("failed to create text file: %v", err)
	}

	binaryPath := filepath.Join(tempDir, "sample.bin")
	err = os.WriteFile(binaryPath, []byte{'E', 'L', 'F', 0, 1, 2, 3}, 0644)
	if err != nil {
		t.Fatalf("failed to create text file: %v", err)
	}

	isBin, err := isBinary(textPath)
	if err != nil {
		t.Errorf("unexpected error for the text file: %v", err)
	}
	if isBin {
		t.Errorf("expected text file to NOT be a binary")
	}

	isBin, err = isBinary(binaryPath)
	if err != nil {
		t.Errorf("unexpected error for the text file: %v", err)
	}
	if !isBin {
		t.Errorf("expected binary file to BE binary")
	}
}

func TestSearchDirectory_PermissionDenied(t *testing.T) {
	tempDir := t.TempDir()

	restrictedDir := filepath.Join(tempDir, "restricted")
	err := os.Mkdir(restrictedDir, 0755)
	if err != nil {
		t.Fatalf("failed to create temp folder: %v", err)
	}

	err = os.Chmod(restrictedDir, 0000)
	if err != nil {
		t.Fatalf("failed to chmod folder: %v", err)
	}

	defer os.Chmod(restrictedDir, 0755)

	opts := Options{
		CaseInsensitive: true,
		Extensions:      []string{"txt"},
	}

	err = SearchDirectory(os.Stdout, restrictedDir, "test", opts)
	if err != nil {
		t.Errorf("expected no fatal error on unreadable directory. got: %v", err)
	}
}

func TestSearchFileCountOnly(t *testing.T) {
	file := filepath.Join(t.TempDir(), "test_count.txt")
	content := `apple
	banana
	apple pie
	cherry`

	err := os.WriteFile(file, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	opts := Options{
		CountOnly: true,
	}

	matches, err := SearchFile(&buf, file, "apple", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if matches != 2 {
		t.Errorf("SearchFile() returned %d matches, want 2", matches)
	}

	expectedOutput := fmt.Sprintf("%s: 2 matches\n", file)
	if buf.String() != expectedOutput {
		t.Errorf("expected %q, got %q", expectedOutput, buf.String())
	}
}

func TestSearchDirectoryConcurrent(t *testing.T) {
	tempDir := t.TempDir()

	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subfolder: %v", err)
	}

	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	ignoredFile := filepath.Join(tempDir, "file3.log")

	if err := os.WriteFile(file1, []byte("match target\nno match"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("another target match"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ignoredFile, []byte("target should be ignored"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	opts := Options{
		Recursive:  true,
		Extensions: []string{"txt"},
	}

	err := SearchDirectoryConcurrent(&buf, tempDir, "target", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "file1.txt") || !strings.Contains(out, "file2.txt") {
		t.Errorf("expected matches from txt files, got %q", out)
	}
	if strings.Contains(out, "file3.log") {
		t.Errorf("file to be ignored, got %q", out)
	}
	if !strings.Contains(out, "2 matches found") {
		t.Errorf("expected match summary '2 matches found',got %q", out)
	}
}
func TestSearchDirectoryConcurrent_CountOnly(t *testing.T) {
	tempDir := t.TempDir()

	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")

	if err := os.WriteFile(file1, []byte("apple pie"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("apple juice\napple cider"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	opts := Options{
		Recursive: true,
		CountOnly: true,
	}

	err := SearchDirectoryConcurrent(&buf, tempDir, "apple", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	// CountOnly should suppress individual line outputs and only report the total summary
	if strings.Contains(out, "apple pie") || strings.Contains(out, "apple juice") {
		t.Errorf("expected no line outputs when CountOnly is true, got: %q", out)
	}
	if !strings.Contains(out, "3 matches found") {
		t.Errorf("expected '3 matches found', got: %q", out)
	}
}

func createBenchmarkDir(b *testing.B, numFiles int) string {
	b.Helper()
	tempDir := b.TempDir()

	for i := 0; i < numFiles; i++ {
		filePath := filepath.Join(tempDir, fmt.Sprintf("file_%d.txt", i))
		content := "golang search test line\nanother line with query word inside\nsome random binary data "
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			b.Fatal(err)
		}
	}
	return tempDir

}

func BenchmarkSearchDirectory(b *testing.B) {
	dir := createBenchmarkDir(b, 500)
	opts := Options{Recursive: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SearchDirectory(io.Discard, dir, "query", opts)
	}
}

func BenchmarkSearchDirectoryConcurrent(b *testing.B) {
	dir := createBenchmarkDir(b, 500)
	opts := Options{Recursive: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SearchDirectoryConcurrent(io.Discard, dir, "query", opts)
	}
}
