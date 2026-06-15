// code-from-spec: SPEC/golang/tests/os/file_writer@2CDIybEHOX8FNT6lpzrLIiYI1pE
package filewriter_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/filewriter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
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

	path := &pathutils.PathCfs{Value: "newfile.txt"}
	err := filewriter.FileWrite(path, "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("newfile.txt")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("content = %q, want %q", string(got), "hello world")
	}
}

func TestFileWrite_OverwritesExistingFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("target.txt", []byte("old"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	path := &pathutils.PathCfs{Value: "target.txt"}
	err := filewriter.FileWrite(path, "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("target.txt")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(got) != "new" {
		t.Errorf("content = %q, want %q", string(got), "new")
	}
}

func TestFileWrite_CreatesIntermediateDirectories(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	path := &pathutils.PathCfs{Value: "a/b/c/file.txt"}
	err := filewriter.FileWrite(path, "content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("a/b/c/file.txt")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(got) != "content" {
		t.Errorf("content = %q, want %q", string(got), "content")
	}
}

func TestFileWrite_PreservesUTF8Content(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "café 日本語 🎉"
	path := &pathutils.PathCfs{Value: "utf8file.txt"}
	err := filewriter.FileWrite(path, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("utf8file.txt")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(got) != content {
		t.Errorf("content = %q, want %q", string(got), content)
	}
}

func TestFileWrite_PreservesLineEndings(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "alpha\r\nbeta\r\n"
	path := &pathutils.PathCfs{Value: "crlf.txt"}
	err := filewriter.FileWrite(path, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("crlf.txt")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(got) != content {
		t.Errorf("content = %q, want %q", string(got), content)
	}
}

func TestFileWrite_WritesEmptyContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	path := &pathutils.PathCfs{Value: "empty.txt"}
	err := filewriter.FileWrite(path, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat("empty.txt")
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("file size = %d, want 0", info.Size())
	}
}

func TestFileWrite_PropagatesValidationError(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	path := &pathutils.PathCfs{Value: "../../outside"}
	err := filewriter.FileWrite(path, "data")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("error = %v, want ErrDirectoryTraversal", err)
	}
}

func TestFileWrite_CannotCreateDirectory(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("a", []byte("block"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	path := &pathutils.PathCfs{Value: "a/b/file.txt"}
	err := filewriter.FileWrite(path, "data")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filewriter.ErrCannotCreateDirectory) {
		t.Errorf("error = %v, want ErrCannotCreateDirectory", err)
	}
}

func TestFileWrite_CannotWriteFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.Mkdir("file.txt", 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	path := &pathutils.PathCfs{Value: "file.txt"}
	err := filewriter.FileWrite(path, "data")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filewriter.ErrCannotWriteFile) {
		t.Errorf("error = %v, want ErrCannotWriteFile", err)
	}
}
