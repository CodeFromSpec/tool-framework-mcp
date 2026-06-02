// code-from-spec: ROOT/golang/tests/os/file_writer@ogLK2IKZUTH3kmMBKYcwqHNvHIA
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

func TestFileWriter_WritesContentToNewFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	err := filewriter.FileWrite(&pathutils.PathCfs{Value: "file.txt"}, "hello world")
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello world" {
		t.Errorf("got %q, want %q", string(data), "hello world")
	}
}

func TestFileWriter_OverwritesExistingFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := filewriter.FileWrite(&pathutils.PathCfs{Value: "file.txt"}, "new"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Errorf("got %q, want %q", string(data), "new")
	}
}

func TestFileWriter_CreatesIntermediateDirectories(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := filewriter.FileWrite(&pathutils.PathCfs{Value: "a/b/c/file.txt"}, "content"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile("a/b/c/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "content" {
		t.Errorf("got %q, want %q", string(data), "content")
	}
}

func TestFileWriter_PreservesUTF8Content(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "café 日本語 🎉"
	if err := filewriter.FileWrite(&pathutils.PathCfs{Value: "file.txt"}, content); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Errorf("got %q, want %q", string(data), content)
	}
}

func TestFileWriter_PreservesLineEndings(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "alpha\r\nbeta\r\n"
	if err := filewriter.FileWrite(&pathutils.PathCfs{Value: "file.txt"}, content); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Errorf("got %q, want %q", string(data), content)
	}
}

func TestFileWriter_WritesEmptyContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := filewriter.FileWrite(&pathutils.PathCfs{Value: "file.txt"}, ""); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 0 {
		t.Errorf("expected zero bytes, got %d", info.Size())
	}
}

func TestFileWriter_PropagatesValidationErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	err := filewriter.FileWrite(&pathutils.PathCfs{Value: "../../outside"}, "content")
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}

	_, statErr := os.Stat("../../outside")
	if !os.IsNotExist(statErr) {
		t.Errorf("expected no file to be created, but stat returned: %v", statErr)
	}
}

func TestFileWriter_CannotCreateDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.MkdirAll("a", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("a/b", []byte("blocker"), 0644); err != nil {
		t.Fatal(err)
	}

	err := filewriter.FileWrite(&pathutils.PathCfs{Value: "a/b/c/file.txt"}, "content")
	if !errors.Is(err, filewriter.ErrCannotCreateDirectory) {
		t.Errorf("expected ErrCannotCreateDirectory, got %v", err)
	}
}

func TestFileWriter_CannotWriteFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.MkdirAll("a/b/target", 0755); err != nil {
		t.Fatal(err)
	}

	err := filewriter.FileWrite(&pathutils.PathCfs{Value: "a/b/target"}, "content")
	if !errors.Is(err, filewriter.ErrCannotWriteFile) {
		t.Errorf("expected ErrCannotWriteFile, got %v", err)
	}
}
