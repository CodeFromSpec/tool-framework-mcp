// code-from-spec: ROOT/golang/tests/os/file_writer@-ZX_hN72AszWO6O-GXSlhOK3WqE

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

	cfsPath := &pathutils.PathCfs{Value: "newfile.txt"}
	content := "hello world"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("newfile.txt")
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(got) != content {
		t.Errorf("file content = %q, want %q", string(got), content)
	}
}

func TestFileWrite_OverwritesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("existing.txt", []byte("old"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "existing.txt"}

	err := filewriter.FileWrite(cfsPath, "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile("existing.txt")
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(got) != "new" {
		t.Errorf("file content = %q, want %q", string(got), "new")
	}
}

func TestFileWrite_CreatesIntermediateDirectories(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfsPath := &pathutils.PathCfs{Value: "a/b/c/file.txt"}
	content := "nested content"

	err := filewriter.FileWrite(cfsPath, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, dir := range []string{"a", "a/b", "a/b/c"} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("expected directory %q to exist: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %q to be a directory", dir)
		}
	}

	got, err := os.ReadFile("a/b/c/file.txt")
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(got) != content {
		t.Errorf("file content = %q, want %q", string(got), content)
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

	got, err := os.ReadFile("utf8.txt")
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(got) != content {
		t.Errorf("file bytes = %q, want %q", string(got), content)
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

	got, err := os.ReadFile("crlf.txt")
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(got) != content {
		t.Errorf("file bytes = %q, want %q", string(got), content)
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

	info, err := os.Stat("empty.txt")
	if err != nil {
		t.Fatalf("expected file to exist: %v", err)
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
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) && !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected directory traversal error, got: %v", err)
	}

	// No file should have been created outside the root
	if _, statErr := os.Stat("../../outside"); statErr == nil {
		t.Error("file should not have been created")
	}
}

func TestFileWrite_CannotCreateDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create a regular file named "a" so that "a/b/file.txt" cannot be created
	if err := os.WriteFile("a", []byte("not a dir"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "a/b/file.txt"}

	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, filewriter.ErrCannotCreateDirectory) {
		t.Errorf("expected ErrCannotCreateDirectory, got: %v", err)
	}
}

func TestFileWrite_CannotWriteFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create a directory at the target path
	if err := os.MkdirAll("target", 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: "target"}

	err := filewriter.FileWrite(cfsPath, "content")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, filewriter.ErrCannotWriteFile) {
		t.Errorf("expected ErrCannotWriteFile, got: %v", err)
	}
}
