// code-from-spec: ROOT/golang/tests/os/file_writer@e_Iu0juGJlhxpu689g8rh1v2Nsw
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
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "output/file.txt"}
	if err := os.MkdirAll("output", 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := filewriter.FileWrite(cfsPath, "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile("output/file.txt")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("content = %q, want %q", string(data), "hello world")
	}
}

func TestFileWrite_OverwritesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("output", 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile("output/file.txt", []byte("old"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "output/file.txt"}
	err := filewriter.FileWrite(cfsPath, "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile("output/file.txt")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != "new" {
		t.Errorf("content = %q, want %q", string(data), "new")
	}
}

func TestFileWrite_CreatesIntermediateDirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "a/b/c/file.txt"}
	err := filewriter.FileWrite(cfsPath, "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile("a/b/c/file.txt")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("content = %q, want %q", string(data), "hello")
	}
}

func TestFileWrite_PreservesUTF8Content(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "café 日本語 🎉"
	cfsPath := &pathutils.PathCfs{Value: "output/utf8.txt"}
	if err := os.MkdirAll("output", 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile("output/utf8.txt")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != content {
		t.Errorf("content = %q, want %q", string(data), content)
	}
}

func TestFileWrite_PreservesLineEndings(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "alpha\r\nbeta\r\n"
	cfsPath := &pathutils.PathCfs{Value: "output/crlf.txt"}
	if err := os.MkdirAll("output", 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile("output/crlf.txt")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != content {
		t.Errorf("content = %q, want %q", string(data), content)
	}
}

func TestFileWrite_WritesEmptyContent(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "output/empty.txt"}
	if err := os.MkdirAll("output", 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := filewriter.FileWrite(cfsPath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat("output/empty.txt")
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("file size = %d, want 0", info.Size())
	}
}

func TestFileWrite_PropagatesValidationErrors(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "../../outside"}

	err := filewriter.FileWrite(cfsPath, "x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("error = %v, want ErrDirectoryTraversal", err)
	}
}

func TestFileWrite_CannotCreateDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("a", []byte("block"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "a/b/file.txt"}
	err := filewriter.FileWrite(cfsPath, "x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filewriter.ErrCannotCreateDirectory) {
		t.Errorf("error = %v, want ErrCannotCreateDirectory", err)
	}
}

func TestFileWrite_CannotWriteFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("output/target", 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "output/target"}
	err := filewriter.FileWrite(cfsPath, "x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filewriter.ErrCannotWriteFile) {
		t.Errorf("error = %v, want ErrCannotWriteFile", err)
	}
}
