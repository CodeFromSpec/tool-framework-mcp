// code-from-spec: ROOT/golang/tests/os/file_writer@Ix2BDjC8CBJpfwd1HXIhMcHEbbw

package filewriter_test

import (
	"errors"
	"os"
	"path/filepath"
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

	cfsPath := &pathutils.PathCfs{Value: "newfile.txt"}
	content := "hello world"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tempDir, "newfile.txt"))
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != content {
		t.Errorf("content = %q, want %q", string(data), content)
	}
}

func TestFileWrite_OverwritesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("target.txt", []byte("old"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "target.txt"}
	err := filewriter.FileWrite(cfsPath, "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tempDir, "target.txt"))
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
	content := "deep content"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tempDir, "a", "b", "c", "file.txt"))
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != content {
		t.Errorf("content = %q, want %q", string(data), content)
	}

	// Verify intermediate directories exist.
	for _, dir := range []string{"a", filepath.Join("a", "b"), filepath.Join("a", "b", "c")} {
		info, err := os.Stat(filepath.Join(tempDir, dir))
		if err != nil {
			t.Errorf("expected dir %q to exist: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %q to be a directory", dir)
		}
	}
}

func TestFileWrite_PreservesUTF8Content(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "utf8.txt"}
	content := "café 日本語 🎉"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tempDir, "utf8.txt"))
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

	cfsPath := &pathutils.PathCfs{Value: "crlf.txt"}
	content := "alpha\r\nbeta\r\n"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tempDir, "crlf.txt"))
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != content {
		t.Errorf("content bytes = %q, want %q", string(data), content)
	}
}

func TestFileWrite_WritesEmptyContent(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "empty.txt"}

	err := filewriter.FileWrite(cfsPath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(filepath.Join(tempDir, "empty.txt"))
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

	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

func TestFileWrite_CannotCreateDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create "a/b" as a regular file so that it cannot be used as a directory.
	if err := os.MkdirAll("a", 0755); err != nil {
		t.Fatalf("setup mkdir: %v", err)
	}
	if err := os.WriteFile("a/b", []byte("I am a file"), 0644); err != nil {
		t.Fatalf("setup write: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "a/b/file.txt"}

	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filewriter.ErrCannotCreateDirectory) {
		t.Errorf("expected ErrCannotCreateDirectory, got: %v", err)
	}
}

func TestFileWrite_CannotWriteFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create a directory at the target path so writing a file there fails.
	if err := os.MkdirAll("targetdir", 0755); err != nil {
		t.Fatalf("setup mkdir: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "targetdir"}

	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filewriter.ErrCannotWriteFile) {
		t.Errorf("expected ErrCannotWriteFile, got: %v", err)
	}
}
