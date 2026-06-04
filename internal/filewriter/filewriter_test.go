// code-from-spec: ROOT/golang/tests/os/file_writer@hUxmpzZ6hfg5oqPWcTIqeND0ceg
package filewriter_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func testChdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("testChdir cleanup: %v", err)
		}
	})
}

func TestFileWrite_WritesContentToNewFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	cfsPath := &pathutils.PathCfs{Value: "newfile.txt"}
	err := filewriter.FileWrite(cfsPath, "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("newfile.txt")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("got %q, want %q", string(got), "hello world")
	}
}

func TestFileWrite_OverwritesExistingFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("existing.txt", []byte("old"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "existing.txt"}
	err := filewriter.FileWrite(cfsPath, "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("existing.txt")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(got) != "new" {
		t.Errorf("got %q, want %q", string(got), "new")
	}
}

func TestFileWrite_CreatesIntermediateDirectories(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	cfsPath := &pathutils.PathCfs{Value: "a/b/c/file.txt"}
	err := filewriter.FileWrite(cfsPath, "content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("a/b/c/file.txt")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(got) != "content" {
		t.Errorf("got %q, want %q", string(got), "content")
	}
}

func TestFileWrite_PreservesUTF8Content(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "café 日本語 🎉"
	cfsPath := &pathutils.PathCfs{Value: "utf8.txt"}
	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("utf8.txt")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(got) != content {
		t.Errorf("got %q, want %q", string(got), content)
	}
}

func TestFileWrite_PreservesLineEndings(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "alpha\r\nbeta\r\n"
	cfsPath := &pathutils.PathCfs{Value: "crlf.txt"}
	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("crlf.txt")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(got) != content {
		t.Errorf("got %q, want %q", string(got), content)
	}
}

func TestFileWrite_WritesEmptyContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	cfsPath := &pathutils.PathCfs{Value: "empty.txt"}
	err := filewriter.FileWrite(cfsPath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("empty.txt")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty file, got %d bytes", len(got))
	}
}

func TestFileWrite_PropagatesValidationErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	cfsPath := &pathutils.PathCfs{Value: "../../outside"}
	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestFileWrite_CannotCreateDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("a", []byte("file"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "a/b/file.txt"}
	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filewriter.ErrCannotCreateDirectory) {
		t.Errorf("expected ErrCannotCreateDirectory, got %v", err)
	}
}

func TestFileWrite_CannotWriteFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.Mkdir("targetdir", 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "targetdir"}
	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filewriter.ErrCannotWriteFile) {
		t.Errorf("expected ErrCannotWriteFile, got %v", err)
	}
}
