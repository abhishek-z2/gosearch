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
	file := createTestFile(t, `Hello world
This is a test
HELLO again
Nothing here`)

	opts := Options{
		CaseInsensitive: true,
	}

	results, matches, err := SearchFile(file, "hello", opts)
	if err != nil {
		t.Fatal(err)
	}

	if matches != 2 {
		t.Fatalf("SearchFile() found %d matches, want 2", matches)
	}

	if len(results) != 2 {
		t.Fatalf("SearchFile() returned %d results, want 2", len(results))
	}

	if results[0].File != file {
		t.Errorf("first result file = %q, want %q", results[0].File, file)
	}

	if results[0].LineNumber != 1 {
		t.Errorf("first result line = %d, want 1", results[0].LineNumber)
	}

	if results[0].Line != "Hello world" {
		t.Errorf("first result text = %q, want %q", results[0].Line, "Hello world")
	}

	if !results[0].isMatch {
		t.Errorf("first result should be a match")
	}

	if results[1].LineNumber != 3 {
		t.Errorf("second result line = %d, want 3", results[1].LineNumber)
	}

	if !results[1].isMatch {
		t.Errorf("second result should be a match")
	}
}

func TestSearchFileCaseSensitive(t *testing.T) {
	file := createTestFile(t, `Hello world
This is a test
HELLO again
Nothing here`)

	results, matches, err := SearchFile(file, "hello", Options{})
	if err != nil {
		t.Fatal(err)
	}

	if matches != 0 {
		t.Errorf("SearchFile() found %d matches, want 0", matches)
	}

	if len(results) != 0 {
		t.Errorf("SearchFile() returned %d results, want 0", len(results))
	}
}

func TestSearchFileInvertMatch(t *testing.T) {
	file := createTestFile(t, `apple
banana
cherry`)

	opts := Options{
		InvertMatch: true,
	}

	results, matches, err := SearchFile(file, "banana", opts)
	if err != nil {
		t.Fatal(err)
	}

	if matches != 2 {
		t.Fatalf("SearchFile() found %d matches, want 2", matches)
	}

	if len(results) != 2 {
		t.Fatalf("SearchFile() returned %d results, want 2", len(results))
	}

	if results[0].LineNumber != 1 {
		t.Errorf("first result line = %d, want 1", results[0].LineNumber)
	}

	if results[1].LineNumber != 3 {
		t.Errorf("second result line = %d, want 3", results[1].LineNumber)
	}

	for _, result := range results {
		if !result.isMatch {
			t.Errorf("inverted matching result should be marked as match")
		}
	}

	for _, result := range results {
		if strings.Contains(result.Line, "banana") {
			t.Errorf("inverted results should not contain banana")
		}
	}
}

func TestSearchFileNoMatches(t *testing.T) {
	file := createTestFile(t, `Hello world
This is a test
HELLO again
Nothing here`)

	results, matches, err := SearchFile(file, "banana", Options{})
	if err != nil {
		t.Fatal(err)
	}

	if matches != 0 {
		t.Errorf("SearchFile() found %d matches, want 0", matches)
	}

	if len(results) != 0 {
		t.Errorf("SearchFile() returned %d results, want 0", len(results))
	}
}

func TestSearchFileNotFound(t *testing.T) {
	_, _, err := SearchFile(
		"nonexistentfile.txt",
		"Hello",
		Options{},
	)

	if err == nil {
		t.Errorf("expected an error, got nil")
	}
}

func TestSearchFileCountOnly(t *testing.T) {
	file := createTestFile(t, `apple
banana
apple pie
cherry`)

	opts := Options{
		CountOnly: true,
	}

	results, matches, err := SearchFile(file, "apple", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if matches != 2 {
		t.Errorf("SearchFile() returned %d matches, want 2", matches)
	}

	if len(results) != 0 {
		t.Errorf("CountOnly should return no results, got %d", len(results))
	}
}

func TestSearchFileBeforeContext(t *testing.T) {
	file := createTestFile(t, `line 1
line 2
line 3
line 4
target
line 6
line 7`)

	opts := Options{
		BeforeContext: 2,
	}

	results, matches, err := SearchFile(file, "target", opts)
	if err != nil {
		t.Fatal(err)
	}

	if matches != 1 {
		t.Fatalf("matches = %d, want 1", matches)
	}

	if len(results) != 3 {
		t.Fatalf("results = %d, want 3", len(results))
	}

	wantLines := []int{3, 4, 5}

	for i, want := range wantLines {
		if results[i].LineNumber != want {
			t.Errorf(
				"result[%d].LineNumber = %d, want %d",
				i,
				results[i].LineNumber,
				want,
			)
		}
	}

	if results[0].isMatch || results[1].isMatch {
		t.Errorf("before-context lines should not be marked as matches")
	}

	if !results[2].isMatch {
		t.Errorf("target line should be marked as a match")
	}
}

func TestSearchFileAfterContext(t *testing.T) {
	file := createTestFile(t, `line 1
target
line 3
line 4
line 5
line 6
line 7`)

	opts := Options{
		AfterContext: 3,
	}

	results, matches, err := SearchFile(file, "target", opts)
	if err != nil {
		t.Fatal(err)
	}

	if matches != 1 {
		t.Fatalf("matches = %d, want 1", matches)
	}

	if len(results) != 4 {
		t.Fatalf("results = %d, want 4", len(results))
	}

	wantLines := []int{2, 3, 4, 5}

	for i, want := range wantLines {
		if results[i].LineNumber != want {
			t.Errorf(
				"result[%d].LineNumber = %d, want %d",
				i,
				results[i].LineNumber,
				want,
			)
		}
	}

	if !results[0].isMatch {
		t.Errorf("target line should be marked as a match")
	}

	for i := 1; i < len(results); i++ {
		if results[i].isMatch {
			t.Errorf("after-context line %d should not be marked as a match", results[i].LineNumber)
		}
	}
}

func TestSearchFileBeforeAndAfterContext(t *testing.T) {
	file := createTestFile(t, `line 1
line 2
line 3
target
line 5
line 6
line 7
line 8`)

	opts := Options{
		BeforeContext: 2,
		AfterContext:  2,
	}

	results, matches, err := SearchFile(file, "target", opts)
	if err != nil {
		t.Fatal(err)
	}

	if matches != 1 {
		t.Fatalf("matches = %d, want 1", matches)
	}

	if len(results) != 5 {
		t.Fatalf("results = %d, want 5", len(results))
	}

	wantLines := []int{2, 3, 4, 5, 6}

	for i, want := range wantLines {
		if results[i].LineNumber != want {
			t.Errorf(
				"result[%d].LineNumber = %d, want %d",
				i,
				results[i].LineNumber,
				want,
			)
		}
	}

	if results[0].isMatch || results[1].isMatch {
		t.Errorf("before-context lines should not be marked as matches")
	}

	if !results[2].isMatch {
		t.Errorf("target should be marked as a match")
	}

	if results[3].isMatch || results[4].isMatch {
		t.Errorf("after-context lines should not be marked as matches")
	}
}

func TestSearchFileOverlappingContext(t *testing.T) {
	file := createTestFile(t, `line 1
line 2
target
line 4
target
line 6
line 7
line 8`)

	opts := Options{
		BeforeContext: 1,
		AfterContext:  1,
	}

	results, matches, err := SearchFile(file, "target", opts)
	if err != nil {
		t.Fatal(err)
	}

	if matches != 2 {
		t.Fatalf("matches = %d, want 2", matches)
	}

	// The two context windows overlap:
	//
	// first match:  2-4
	// second match: 4-6
	//
	// Therefore line 4 should only appear once.
	wantLines := []int{2, 3, 4, 5, 6}

	if len(results) != len(wantLines) {
		t.Fatalf(
			"results = %d, want %d",
			len(results),
			len(wantLines),
		)
	}

	for i, want := range wantLines {
		if results[i].LineNumber != want {
			t.Errorf(
				"result[%d].LineNumber = %d, want %d",
				i,
				results[i].LineNumber,
				want,
			)
		}
	}
}

func TestSearchFileMultipleContextBlocks(t *testing.T) {
	file := createTestFile(t, `target
line 2
line 3
line 4
line 5
target
line 7`)

	opts := Options{
		AfterContext: 1,
	}

	results, matches, err := SearchFile(file, "target", opts)
	if err != nil {
		t.Fatal(err)
	}

	if matches != 2 {
		t.Fatalf("matches = %d, want 2", matches)
	}

	wantLines := []int{1, 2, 6, 7}

	if len(results) != len(wantLines) {
		t.Fatalf(
			"results = %d, want %d",
			len(results),
			len(wantLines),
		)
	}

	for i, want := range wantLines {
		if results[i].LineNumber != want {
			t.Errorf(
				"result[%d].LineNumber = %d, want %d",
				i,
				results[i].LineNumber,
				want,
			)
		}
	}
}

func TestPrintResult(t *testing.T) {
	results := []SearchResult{
		{
			File:       "test.txt",
			LineNumber: 1,
			Line:       "hello",
			isMatch:    true,
		},
		{
			File:       "test.txt",
			LineNumber: 2,
			Line:       "world",
			isMatch:    false,
		},
	}

	var buf bytes.Buffer

	PrintResult(&buf, results, "hello", Options{})

	expected := `test.txt:1: hello
test.txt:2: world
`

	if buf.String() != expected {
		t.Errorf(
			"PrintResult() = %q, want %q",
			buf.String(),
			expected,
		)
	}
}

func TestPrintResultColor(t *testing.T) {
	results := []SearchResult{
		{
			File:       "test.txt",
			LineNumber: 1,
			Line:       "Hello world",
			isMatch:    true,
		},
	}

	var buf bytes.Buffer

	opts := Options{
		Color: true,
	}

	PrintResult(&buf, results, "Hello", opts)

	expected := "test.txt:1: \033[1;31mHello\033[0m world\n"

	if buf.String() != expected {
		t.Errorf(
			"PrintResult() = %q, want %q",
			buf.String(),
			expected,
		)
	}
}

func TestPrintResultCaseInsensitiveColor(t *testing.T) {
	results := []SearchResult{
		{
			File:       "test.txt",
			LineNumber: 1,
			Line:       "Hello HELLO hello",
			isMatch:    true,
		},
	}

	var buf bytes.Buffer

	opts := Options{
		Color:           true,
		CaseInsensitive: true,
	}

	PrintResult(&buf, results, "hello", opts)

	expected := "test.txt:1: \033[1;31mHello\033[0m \033[1;31mHELLO\033[0m \033[1;31mhello\033[0m\n"

	if buf.String() != expected {
		t.Errorf(
			"PrintResult() = %q, want %q",
			buf.String(),
			expected,
		)
	}
}

func TestPrintResultContextSeparator(t *testing.T) {
	results := []SearchResult{
		{
			File:       "test.txt",
			LineNumber: 1,
			Line:       "target",
			isMatch:    true,
		},
		{
			File:       "test.txt",
			LineNumber: 2,
			Line:       "context",
			isMatch:    false,
		},
		{
			File:       "test.txt",
			LineNumber: 10,
			Line:       "target",
			isMatch:    true,
		},
	}

	var buf bytes.Buffer

	opts := Options{
		AfterContext: 1,
	}

	PrintResult(&buf, results, "target", opts)

	expected := `test.txt:1: target
test.txt:2: context
--
test.txt:10: target
`

	if buf.String() != expected {
		t.Errorf(
			"PrintResult() = %q, want %q",
			buf.String(),
			expected,
		)
	}
}

func TestIsBinary(t *testing.T) {
	tempDir := t.TempDir()

	textPath := filepath.Join(tempDir, "sample.txt")
	if err := os.WriteFile(
		textPath,
		[]byte("hello world\nthis is text"),
		0644,
	); err != nil {
		t.Fatalf("failed to create text file: %v", err)
	}

	binaryPath := filepath.Join(tempDir, "sample.bin")
	if err := os.WriteFile(
		binaryPath,
		[]byte{'E', 'L', 'F', 0, 1, 2, 3},
		0644,
	); err != nil {
		t.Fatalf("failed to create binary file: %v", err)
	}

	isBin, err := isBinary(textPath)
	if err != nil {
		t.Errorf("unexpected error for text file: %v", err)
	}

	if isBin {
		t.Errorf("expected text file to NOT be binary")
	}

	isBin, err = isBinary(binaryPath)
	if err != nil {
		t.Errorf("unexpected error for binary file: %v", err)
	}

	if !isBin {
		t.Errorf("expected binary file to BE binary")
	}
}

func TestSearchDirectory_PermissionDenied(t *testing.T) {
	tempDir := t.TempDir()

	restrictedDir := filepath.Join(tempDir, "restricted")

	if err := os.Mkdir(restrictedDir, 0755); err != nil {
		t.Fatalf("failed to create temp folder: %v", err)
	}

	if err := os.Chmod(restrictedDir, 0000); err != nil {
		t.Fatalf("failed to chmod folder: %v", err)
	}

	defer os.Chmod(restrictedDir, 0755)

	opts := Options{
		CaseInsensitive: true,
		Extensions:      []string{"txt"},
	}

	err := SearchDirectory(io.Discard, restrictedDir, "test", opts)
	if err != nil {
		t.Errorf(
			"expected no fatal error on unreadable directory, got: %v",
			err,
		)
	}
}

func TestSearchDirectoryConcurrent(t *testing.T) {
	tempDir := t.TempDir()

	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	ignoredFile := filepath.Join(tempDir, "file3.log")

	if err := os.WriteFile(
		file1,
		[]byte("match target\nno match"),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		file2,
		[]byte("another target match"),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		ignoredFile,
		[]byte("target should be ignored"),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	opts := Options{
		Recursive:  true,
		Extensions: []string{"txt"},
	}

	// The current concurrent implementation only counts matches.
	err := SearchDirectoryConcurrent(&buf, tempDir, "target", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "2 total matches found") {
		t.Errorf("expected total match summary, got %q", out)
	}

	if strings.Contains(out, "file3.log") {
		t.Errorf("log file should be ignored, got %q", out)
	}
}

func TestSearchDirectoryConcurrentCountOnly(t *testing.T) {
	tempDir := t.TempDir()

	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")

	if err := os.WriteFile(
		file1,
		[]byte("apple pie"),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		file2,
		[]byte("apple juice\napple cider"),
		0644,
	); err != nil {
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

	if strings.Contains(out, "apple pie") ||
		strings.Contains(out, "apple juice") {
		t.Errorf(
			"expected no individual line output when CountOnly is true, got: %q",
			out,
		)
	}

	if !strings.Contains(out, "3 total matches found") {
		t.Errorf(
			"expected '3 total matches found', got: %q",
			out,
		)
	}
}

func createTestFile(t *testing.T, content string) string {
	t.Helper()

	file := filepath.Join(t.TempDir(), "test.txt")

	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	return file
}

func createBenchmarkDir(b *testing.B, numFiles int) string {
	b.Helper()

	tempDir := b.TempDir()

	for i := 0; i < numFiles; i++ {
		filePath := filepath.Join(
			tempDir,
			fmt.Sprintf("file_%d.txt", i),
		)

		content := "golang search test line\n" +
			"another line with query word inside\n" +
			"some random binary data "

		if err := os.WriteFile(
			filePath,
			[]byte(content),
			0644,
		); err != nil {
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
		_ = SearchDirectory(
			io.Discard,
			dir,
			"query",
			opts,
		)
	}
}

func BenchmarkSearchDirectoryConcurrent(b *testing.B) {
	dir := createBenchmarkDir(b, 500)
	opts := Options{Recursive: true}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = SearchDirectoryConcurrent(
			io.Discard,
			dir,
			"query",
			opts,
		)
	}
}
