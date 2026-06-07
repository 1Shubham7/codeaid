package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- readFile ---

func TestReadFile_Success(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "hello.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	got := readFile(path)
	if got != "hello world" {
		t.Errorf("readFile = %q, want %q", got, "hello world")
	}
}

func TestReadFile_NotFound(t *testing.T) {
	got := readFile("/nonexistent/path/missing.txt")
	if !strings.HasPrefix(got, "error") {
		t.Errorf("expected error message, got %q", got)
	}
}

func TestReadFile_ExceedsLimit(t *testing.T) {
	SetMaxFileSizeKB(1)
	defer SetMaxFileSizeKB(100)

	tmp := t.TempDir()
	path := filepath.Join(tmp, "big.bin")
	if err := os.WriteFile(path, make([]byte, 2*1024), 0644); err != nil {
		t.Fatal(err)
	}

	got := readFile(path)
	if !strings.HasPrefix(got, "Blocked:") {
		t.Errorf("expected Blocked message for oversized file, got %q", got)
	}
}

func TestReadFile_AtLimit(t *testing.T) {
	SetMaxFileSizeKB(1)
	defer SetMaxFileSizeKB(100)

	tmp := t.TempDir()
	path := filepath.Join(tmp, "small.txt")
	if err := os.WriteFile(path, []byte("small"), 0644); err != nil {
		t.Fatal(err)
	}

	got := readFile(path)
	if got != "small" {
		t.Errorf("readFile = %q, want %q", got, "small")
	}
}

// --- writeFile ---

func TestWriteFile_Success(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "out.txt")

	got := writeFile(path, "test content")
	if !strings.Contains(got, "successfully wrote") {
		t.Errorf("expected success message, got %q", got)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not found after write: %v", err)
	}
	if string(data) != "test content" {
		t.Errorf("file content = %q, want %q", string(data), "test content")
	}
}

func TestWriteFile_CreatesParentDirs(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "a", "b", "c", "file.txt")

	got := writeFile(path, "nested")
	if !strings.Contains(got, "successfully wrote") {
		t.Errorf("expected success, got %q", got)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if string(data) != "nested" {
		t.Errorf("content = %q, want %q", string(data), "nested")
	}
}

func TestWriteFile_EmptyContent(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "empty.txt")

	got := writeFile(path, "")
	if !strings.Contains(got, "successfully wrote") {
		t.Errorf("expected success for empty file, got %q", got)
	}
}

// --- listDirectory ---

func TestListDirectory_WithContent(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmp, "mydir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "myfile.txt"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	got := listDirectory(tmp)
	if !strings.Contains(got, "mydir/") {
		t.Errorf("expected 'mydir/' in result, got:\n%s", got)
	}
	if !strings.Contains(got, "myfile.txt") {
		t.Errorf("expected 'myfile.txt' in result, got:\n%s", got)
	}
}

func TestListDirectory_Empty(t *testing.T) {
	tmp := t.TempDir()
	got := listDirectory(tmp)
	if !strings.Contains(got, "(empty)") {
		t.Errorf("expected '(empty)', got %q", got)
	}
}

func TestListDirectory_DefaultsToCurrentDir(t *testing.T) {
	got := listDirectory("")
	if strings.HasPrefix(got, "error") {
		t.Errorf("expected success for default '.', got %q", got)
	}
	if !strings.Contains(got, "Directory: .") {
		t.Errorf("expected 'Directory: .' in result, got %q", got)
	}
}

func TestListDirectory_NotFound(t *testing.T) {
	got := listDirectory("/nonexistent/path/xyz")
	if !strings.HasPrefix(got, "error") {
		t.Errorf("expected error, got %q", got)
	}
}
